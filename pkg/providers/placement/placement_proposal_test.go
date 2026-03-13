/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package placement

import (
	"context"
	"errors"
	"testing"

	ociv1beta1 "github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/fakes"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/capacityreservation"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/clusterplacementgroup"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/computecluster"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/identity"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/instancetype"
	ocicpg "github.com/oracle/oci-go-sdk/v65/clusterplacementgroups"
	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/scheduling"
)

// helper to create a simple offering
func newOffering(ad string) *cloudprovider.Offering {
	reqs := scheduling.NewRequirements(
		scheduling.NewRequirement(corev1.CapacityTypeLabelKey, v1.NodeSelectorOpIn, corev1.CapacityTypeOnDemand),
		scheduling.NewRequirement(v1.LabelTopologyZone, v1.NodeSelectorOpIn, ad),
		scheduling.NewRequirement(cloudprovider.ReservationIDLabel, v1.NodeSelectorOpDoesNotExist),
	)
	return &cloudprovider.Offering{
		Requirements: reqs,
		Price:        0.1,
		Available:    true,
	}
}

// helper to create instancetype.OciInstanceType
func newInstanceType(shape string, ocpu *float32, memoryInGbs *float32,
	offerings []*cloudprovider.Offering) *instancetype.OciInstanceType {
	it := &instancetype.OciInstanceType{
		Shape:       shape,
		Ocpu:        ocpu,
		MemoryInGbs: memoryInGbs,
	}
	it.InstanceType.Offerings = offerings
	return it
}

// helper to create corev1.NodeClaim
func newNodeClaim(nodePoolName string, requirements ...corev1.NodeSelectorRequirementWithMinValues) *corev1.NodeClaim {
	return &corev1.NodeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-claim",
			Labels: map[string]string{
				corev1.NodePoolLabelKey: nodePoolName,
			},
		},
		Spec: corev1.NodeClaimSpec{
			Requirements: requirements,
		},
	}
}

func TestProvider_PlaceInstance_Success(t *testing.T) {
	ctx := context.Background()
	fakeCompute := &fakes.FakeCompute{}
	fakeIdentity := fakes.NewFakeIdentity()
	capResProvider := capacityreservation.NewProvider(ctx, fakeCompute, "ocid1.compartment.oc1..cluster123")
	identityProvider, _ := identity.NewProvider(ctx, "ocid1.compartment.oc1..cluster123", fakeIdentity)

	provider := &DefaultProvider{
		instancesByNodePool:         make(map[string]*adFdSummary),
		capacityReservationProvider: capResProvider,
		identityProvider:            identityProvider,
	}

	offerings := []*cloudprovider.Offering{
		newOffering("ad-1"),
	}
	instanceType := newInstanceType("VM.Standard2.1", lo.ToPtr(float32(2.0)), lo.ToPtr(float32(16.0)), offerings)
	nodeClass := &ociv1beta1.OCINodeClass{}
	claim := newNodeClaim("nodepool1")

	placeFuncCalled := false
	placeFunc := func(proposal *Proposal) error {
		placeFuncCalled = true
		assert.Equal(t, "PHX:ad-1", proposal.Ad)
		assert.Nil(t, proposal.Fd)
		assert.Nil(t, proposal.CapacityReservationId)
		return nil
	}

	err := provider.PlaceInstance(ctx, claim, nodeClass, instanceType, placeFunc)

	assert.NoError(t, err)
	assert.True(t, placeFuncCalled)
}

func TestProvider_PlaceInstance_RetryThenSuccess(t *testing.T) {
	ctx := context.Background()
	fakeCompute := &fakes.FakeCompute{}
	fakeIdentity := fakes.NewFakeIdentity()
	capResProvider := capacityreservation.NewProvider(ctx, fakeCompute, "ocid1.compartment.oc1..cluster123")
	identityProvider, _ := identity.NewProvider(ctx, "ocid1.compartment.oc1..cluster123", fakeIdentity)

	provider := &DefaultProvider{
		instancesByNodePool:         make(map[string]*adFdSummary),
		capacityReservationProvider: capResProvider,
		identityProvider:            identityProvider,
	}

	offerings := []*cloudprovider.Offering{
		newOffering("ad-1"),
		newOffering("ad-2"),
	}
	instanceType := newInstanceType("VM.Standard2.1", lo.ToPtr(float32(2.0)), lo.ToPtr(float32(16.0)), offerings)
	nodeClass := &ociv1beta1.OCINodeClass{}
	claim := newNodeClaim("nodepool1")

	callCount := 0
	placeFunc := func(proposal *Proposal) error {
		callCount++
		if callCount == 1 {
			return errors.New("first proposal failed")
		}
		return nil
	}

	err := provider.PlaceInstance(ctx, claim, nodeClass, instanceType, placeFunc)

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestProvider_PlaceInstance_NoEligibleOfferings(t *testing.T) {
	ctx := context.Background()
	fakeCompute := &fakes.FakeCompute{}
	fakeIdentity := fakes.NewFakeIdentity()
	capResProvider := capacityreservation.NewProvider(ctx, fakeCompute, "ocid1.compartment.oc1..cluster123")
	identityProvider, _ := identity.NewProvider(ctx, "ocid1.compartment.oc1..cluster123", fakeIdentity)

	provider := &DefaultProvider{
		instancesByNodePool:         make(map[string]*adFdSummary),
		capacityReservationProvider: capResProvider,
		identityProvider:            identityProvider,
	}

	// No offerings
	offerings := []*cloudprovider.Offering{}
	instanceType := newInstanceType("VM.Standard2.1", lo.ToPtr(float32(2.0)), lo.ToPtr(float32(16.0)), offerings)
	nodeClass := &ociv1beta1.OCINodeClass{}
	claim := newNodeClaim("nodepool1")

	placeFunc := func(proposal *Proposal) error {
		t.Fatal("placeFunc should not be called")
		return nil
	}

	err := provider.PlaceInstance(ctx, claim, nodeClass, instanceType, placeFunc)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no compatible offering")
}

func TestProvider_placementDecorateFunc(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		nodeClass     *ociv1beta1.OCINodeClass
		instanceType  *instancetype.OciInstanceType
		inputProposal Proposal
		setupFakes    func(*fakes.FakeCompute, *fakes.FakeClusterPlacementGroup)
		expected      []Proposal
		expectError   bool
	}{
		{
			name:          "no special placement - basic proposal unchanged",
			nodeClass:     &ociv1beta1.OCINodeClass{},
			instanceType:  &instancetype.OciInstanceType{Shape: "VM.Standard2.1"},
			inputProposal: Proposal{Ad: "AD-1", Fd: lo.ToPtr("FD-1")},
			expected:      []Proposal{{Ad: "AD-1", Fd: lo.ToPtr("FD-1")}},
		},

		{
			name: "compute cluster - matching AD",
			nodeClass: &ociv1beta1.OCINodeClass{
				Spec: ociv1beta1.OCINodeClassSpec{
					ComputeClusterConfig: &ociv1beta1.ComputeClusterConfig{
						ComputeClusterId: lo.ToPtr("cluster1"),
					},
				},
			},
			instanceType:  &instancetype.OciInstanceType{Shape: "VM.Standard2.1"},
			inputProposal: Proposal{Ad: "AD-1"},
			setupFakes: func(fc *fakes.FakeCompute, _ *fakes.FakeClusterPlacementGroup) {
				fc.OnGetComputeCluster = func(ctx context.Context, req ocicore.GetComputeClusterRequest) (
					ocicore.GetComputeClusterResponse, error) {
					return ocicore.GetComputeClusterResponse{
						ComputeCluster: ocicore.ComputeCluster{
							Id:                 lo.ToPtr("cluster1"),
							DisplayName:        lo.ToPtr("test-cluster"),
							AvailabilityDomain: lo.ToPtr("AD-1"),
							CompartmentId:      lo.ToPtr("comp-123"),
						},
					}, nil
				}
			},
			expected: []Proposal{{Ad: "AD-1", ComputeClusterId: lo.ToPtr("cluster1")}},
		},
		{
			name: "compute cluster - different AD",
			nodeClass: &ociv1beta1.OCINodeClass{
				Spec: ociv1beta1.OCINodeClassSpec{
					ComputeClusterConfig: &ociv1beta1.ComputeClusterConfig{
						ComputeClusterId: lo.ToPtr("cluster1"),
					},
				},
			},
			instanceType:  &instancetype.OciInstanceType{Shape: "VM.Standard2.1"},
			inputProposal: Proposal{Ad: "AD-2"},
			setupFakes: func(fc *fakes.FakeCompute, _ *fakes.FakeClusterPlacementGroup) {
				fc.OnGetComputeCluster = func(ctx context.Context, req ocicore.GetComputeClusterRequest) (
					ocicore.GetComputeClusterResponse, error) {
					return ocicore.GetComputeClusterResponse{
						ComputeCluster: ocicore.ComputeCluster{
							Id:                 lo.ToPtr("cluster1"),
							DisplayName:        lo.ToPtr("test-cluster"),
							AvailabilityDomain: lo.ToPtr("AD-1"),
							CompartmentId:      lo.ToPtr("comp-123"),
						},
					}, nil
				}
			},
			expected: nil, // filtered out due to AD mismatch
		},
		{
			name: "cluster placement group - matching AD",
			nodeClass: &ociv1beta1.OCINodeClass{
				Spec: ociv1beta1.OCINodeClassSpec{
					ClusterPlacementGroupConfigs: []*ociv1beta1.ClusterPlacementGroupConfig{
						{ClusterPlacementGroupId: lo.ToPtr("cpg1")},
					},
				},
			},
			instanceType:  &instancetype.OciInstanceType{Shape: "VM.Standard2.1"},
			inputProposal: Proposal{Ad: "AD-1"},
			setupFakes: func(_ *fakes.FakeCompute, fcpg *fakes.FakeClusterPlacementGroup) {
				fcpg.GetResp = ocicpg.GetClusterPlacementGroupResponse{
					ClusterPlacementGroup: ocicpg.ClusterPlacementGroup{
						Id:                 lo.ToPtr("cpg1"),
						Name:               lo.ToPtr("test-cpg"),
						AvailabilityDomain: lo.ToPtr("AD-1"),
						CompartmentId:      lo.ToPtr("comp-123"),
					},
				}
			},
			expected: []Proposal{{Ad: "AD-1", ClusterPlacementGroupId: lo.ToPtr("cpg1")}},
		},
		{
			name: "cluster placement group - different AD",
			nodeClass: &ociv1beta1.OCINodeClass{
				Spec: ociv1beta1.OCINodeClassSpec{
					ClusterPlacementGroupConfigs: []*ociv1beta1.ClusterPlacementGroupConfig{
						{ClusterPlacementGroupId: lo.ToPtr("cpg1")},
					},
				},
			},
			instanceType:  &instancetype.OciInstanceType{Shape: "VM.Standard2.1"},
			inputProposal: Proposal{Ad: "AD-2"},
			setupFakes: func(_ *fakes.FakeCompute, fcpg *fakes.FakeClusterPlacementGroup) {
				fcpg.GetResp = ocicpg.GetClusterPlacementGroupResponse{
					ClusterPlacementGroup: ocicpg.ClusterPlacementGroup{
						Id:                 lo.ToPtr("cpg1"),
						Name:               lo.ToPtr("test-cpg"),
						AvailabilityDomain: lo.ToPtr("AD-1"),
						CompartmentId:      lo.ToPtr("comp-123"),
					},
				}
			},
			expected: nil, // filtered out due to AD mismatch
		},
		{
			name: "capacity reservation - available",
			nodeClass: &ociv1beta1.OCINodeClass{
				Spec: ociv1beta1.OCINodeClassSpec{
					CapacityReservationConfigs: []*ociv1beta1.CapacityReservationConfig{
						{CapacityReservationId: lo.ToPtr("res1")},
					},
				},
			},
			instanceType: &instancetype.OciInstanceType{Shape: "VM.Standard2.1",
				Ocpu: lo.ToPtr(float32(2.0)), MemoryInGbs: lo.ToPtr(float32(16.0))},
			inputProposal: Proposal{Ad: "AD-1", CapacityReservationId: lo.ToPtr("res1")},
			setupFakes: func(fc *fakes.FakeCompute, _ *fakes.FakeClusterPlacementGroup) {
				fc.OnGetComputeCapacityReservation = func(ctx context.Context,
					req ocicore.GetComputeCapacityReservationRequest) (
					ocicore.GetComputeCapacityReservationResponse, error) {
					return ocicore.GetComputeCapacityReservationResponse{
						ComputeCapacityReservation: ocicore.ComputeCapacityReservation{
							Id:                 lo.ToPtr("res1"),
							DisplayName:        lo.ToPtr("test-res"),
							CompartmentId:      lo.ToPtr("comp-123"),
							AvailabilityDomain: lo.ToPtr("AD-1"),
							InstanceReservationConfigs: []ocicore.InstanceReservationConfig{
								{
									InstanceShape: lo.ToPtr("VM.Standard2.1"),
									InstanceShapeConfig: &ocicore.InstanceReservationShapeConfigDetails{
										Ocpus:       lo.ToPtr(float32(2.0)),
										MemoryInGBs: lo.ToPtr(float32(16.0)),
									},
									ReservedCount: int64Ptr(10),
									UsedCount:     int64Ptr(5), // available
								},
							},
						},
					}, nil
				}
			},
			expected: []Proposal{{Ad: "AD-1", CapacityReservationId: lo.ToPtr("res1")}},
		},
		{
			name: "capacity reservation - full",
			nodeClass: &ociv1beta1.OCINodeClass{
				Spec: ociv1beta1.OCINodeClassSpec{
					CapacityReservationConfigs: []*ociv1beta1.CapacityReservationConfig{
						{CapacityReservationId: lo.ToPtr("res1")},
					},
				},
			},
			instanceType: &instancetype.OciInstanceType{Shape: "VM.Standard2.1",
				Ocpu: lo.ToPtr(float32(2.0)), MemoryInGbs: lo.ToPtr(float32(16.0))},
			inputProposal: Proposal{Ad: "AD-1", CapacityReservationId: lo.ToPtr("res1")},
			setupFakes: func(fc *fakes.FakeCompute, _ *fakes.FakeClusterPlacementGroup) {
				fc.OnGetComputeCapacityReservation = func(ctx context.Context,
					req ocicore.GetComputeCapacityReservationRequest) (ocicore.GetComputeCapacityReservationResponse, error) {
					return ocicore.GetComputeCapacityReservationResponse{
						ComputeCapacityReservation: ocicore.ComputeCapacityReservation{
							Id:                 lo.ToPtr("res1"),
							DisplayName:        lo.ToPtr("test-res"),
							CompartmentId:      lo.ToPtr("comp-123"),
							AvailabilityDomain: lo.ToPtr("AD-1"),
							InstanceReservationConfigs: []ocicore.InstanceReservationConfig{
								{
									InstanceShape: lo.ToPtr("VM.Standard2.1"),
									InstanceShapeConfig: &ocicore.InstanceReservationShapeConfigDetails{
										Ocpus:       lo.ToPtr(float32(2.0)),
										MemoryInGBs: lo.ToPtr(float32(16.0)),
									},
									ReservedCount: int64Ptr(10),
									UsedCount:     int64Ptr(10), // full
								},
							},
						},
					}, nil
				}
			},
			expected: nil, // filtered out due to full capacity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeCompute := &fakes.FakeCompute{}
			fakeCPG := &fakes.FakeClusterPlacementGroup{}

			if tt.setupFakes != nil {
				tt.setupFakes(fakeCompute, fakeCPG)
			}

			// Create provider with fake dependencies
			provider := &DefaultProvider{
				capacityReservationProvider:   capacityreservation.NewProvider(ctx, fakeCompute, "test-compartment"),
				computeClusterProvider:        computecluster.NewProvider(ctx, fakeCompute, "test-compartment"),
				clusterPlacementGroupProvider: clusterplacementgroup.NewProvider(ctx, fakeCPG, "test-compartment"),
			}

			decorateFunc, err := provider.placementDecorateFunc(ctx, tt.nodeClass, tt.instanceType)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			results := decorateFunc(tt.inputProposal)
			assert.Equal(t, tt.expected, results)
		})
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}
