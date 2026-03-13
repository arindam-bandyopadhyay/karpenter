/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package fakes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	ocicpg "github.com/oracle/oci-go-sdk/v65/clusterplacementgroups"
	"github.com/oracle/oci-go-sdk/v65/common"
	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	ocikms "github.com/oracle/oci-go-sdk/v65/keymanagement"
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/karpenter/pkg/events"
)

type FakeEventRecorder struct {
}

func (r *FakeEventRecorder) Publish(evts ...events.Event) {
	// Do nothing for now
}

func CreateBasicOciNodeClass() v1beta1.OCINodeClass {
	return v1beta1.OCINodeClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OCINodeClass",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("test-oci-nodeclass-%d", time.Now().UnixNano()),
		},
		Spec: v1beta1.OCINodeClassSpec{
			VolumeConfig: &v1beta1.VolumeConfig{
				BootVolumeConfig: &v1beta1.BootVolumeConfig{
					ImageConfig: &v1beta1.ImageConfig{
						ImageType: v1beta1.OKEImage,
						ImageId:   lo.ToPtr("testImageId"),
					},
				},
			},
			NetworkConfig: &v1beta1.NetworkConfig{
				PrimaryVnicConfig: &v1beta1.SimpleVnicConfig{
					SubnetAndNsgConfig: &v1beta1.SubnetAndNsgConfig{
						SubnetConfig: &v1beta1.SubnetConfig{
							SubnetId: lo.ToPtr("testSubnetId"),
						},
					},
					VnicDisplayName: lo.ToPtr("testPrimaryVnic"),
				},
			},
		},
	}
}

func CreateOciNodeClassWithMinimumReconcilableSetting(clusterCompartmentId string) v1beta1.OCINodeClass {
	ociTestNodeClass := CreateBasicOciNodeClass()
	ociTestNodeClass.Spec.NodeCompartmentId = &clusterCompartmentId
	ociTestNodeClass.Spec.VolumeConfig.BootVolumeConfig.ImageConfig.ImageId = lo.ToPtr("ocid1.image.123")
	ociTestNodeClass.Spec.NetworkConfig.PrimaryVnicConfig.
		SubnetAndNsgConfig.SubnetConfig.SubnetId = lo.ToPtr("subnet-ipv4")
	return ociTestNodeClass
}

func NewFakeKmsClient() *FakeKms {
	testKeyId := "ocid1.key.oc1.iad.testvalut.kms123"
	fakeKmsClient := &FakeKms{}
	fakeKmsClient.OnGet = func(ctx2 context.Context, request ocikms.GetKeyRequest) (ocikms.GetKeyResponse, error) {
		if request.KeyId != nil && *request.KeyId == testKeyId {
			return ocikms.GetKeyResponse{
				Key: ocikms.Key{
					Id:          &testKeyId,
					DisplayName: lo.ToPtr("testKey"),
				},
			}, nil
		}
		return ocikms.GetKeyResponse{}, errors.New("key not found")
	}
	return fakeKmsClient
}

func NewFakeCapacityReservationClient(testClusterCompartmentId string) *FakeCompute {
	testCrId1 := "ocid1.capacityreservation.1"
	testCrId2 := "ocid1.capacityreservation.2"

	mockCRMap := make(map[string]ocicore.GetComputeCapacityReservationResponse)
	mockCRMap[testCrId1] = ocicore.GetComputeCapacityReservationResponse{
		ComputeCapacityReservation: ocicore.ComputeCapacityReservation{
			Id:                 &testCrId1,
			DisplayName:        lo.ToPtr("TestCapacityReservation1"),
			AvailabilityDomain: lo.ToPtr("ad1"),
			CompartmentId:      &testClusterCompartmentId,
		},
	}
	mockCRMap[testCrId2] = ocicore.GetComputeCapacityReservationResponse{
		ComputeCapacityReservation: ocicore.ComputeCapacityReservation{
			Id:                 &testCrId2,
			DisplayName:        lo.ToPtr("TestCapacityReservation2"),
			AvailabilityDomain: lo.ToPtr("ad1"),
			CompartmentId:      lo.ToPtr("anotherCompartment"),
		},
	}

	fakeComputeClient := &FakeCompute{}
	fakeComputeClient.OnGetComputeCapacityReservation = func(ctx2 context.Context,
		request ocicore.GetComputeCapacityReservationRequest) (ocicore.GetComputeCapacityReservationResponse, error) {
		if request.CapacityReservationId != nil {
			resp, found := mockCRMap[*request.CapacityReservationId]
			if found {
				return resp, nil
			}
		}
		return ocicore.GetComputeCapacityReservationResponse{}, errors.New("capacity reservation not found")
	}
	return fakeComputeClient
}

func NewFakeClusterPlacementGroupClient(testClusterCompartmentId string) *FakeClusterPlacementGroup {
	testCgpId1 := "ocid1.cpg.1"
	testCgpId2 := "ocid1.cpg.2"
	testCgpId3 := "ocid1.cpg.3"

	mockCpgMap := make(map[string]ocicpg.GetClusterPlacementGroupResponse)
	mockCpgMap[testCgpId1] = ocicpg.GetClusterPlacementGroupResponse{
		ClusterPlacementGroup: ocicpg.ClusterPlacementGroup{
			Id:                 &testCgpId1,
			Name:               lo.ToPtr("TestCPG1"),
			AvailabilityDomain: lo.ToPtr("ad1"),
			CompartmentId:      &testClusterCompartmentId,
		},
	}
	mockCpgMap[testCgpId2] = ocicpg.GetClusterPlacementGroupResponse{
		ClusterPlacementGroup: ocicpg.ClusterPlacementGroup{
			Id:                 &testCgpId2,
			Name:               lo.ToPtr("TestCPG2"),
			AvailabilityDomain: lo.ToPtr("ad1"),
			CompartmentId:      lo.ToPtr("anotherCompartment"),
		},
	}
	mockCpgMap[testCgpId3] = ocicpg.GetClusterPlacementGroupResponse{
		ClusterPlacementGroup: ocicpg.ClusterPlacementGroup{
			Id:                 &testCgpId3,
			Name:               lo.ToPtr("TestCPG3"),
			AvailabilityDomain: lo.ToPtr("ad1"),
			CompartmentId:      &testClusterCompartmentId,
		},
	}

	fakeCpgClient := &FakeClusterPlacementGroup{}
	fakeCpgClient.OnGet = func(ctx2 context.Context,
		request ocicpg.GetClusterPlacementGroupRequest) (ocicpg.GetClusterPlacementGroupResponse, error) {
		if request.ClusterPlacementGroupId != nil {
			resp, found := mockCpgMap[*request.ClusterPlacementGroupId]
			if found {
				return resp, nil
			}
		}
		return ocicpg.GetClusterPlacementGroupResponse{}, errors.New("cluster placement group not found")
	}

	return fakeCpgClient
}

func NewFakeComputeClient(testClusterCompartmentId string) *FakeCompute {
	testComputeClusterId1 := "ocid1.computecluster.1"
	testComputeClusterId2 := "ocid1.computecluster.2"

	mockComputeClusterMap := make(map[string]ocicore.GetComputeClusterResponse)
	mockComputeClusterMap[testComputeClusterId1] = ocicore.GetComputeClusterResponse{
		ComputeCluster: ocicore.ComputeCluster{
			Id:                 &testComputeClusterId1,
			DisplayName:        lo.ToPtr("TestComputerCluster1"),
			AvailabilityDomain: lo.ToPtr("ad1"),
			CompartmentId:      &testClusterCompartmentId,
		},
	}
	mockComputeClusterMap[testComputeClusterId2] = ocicore.GetComputeClusterResponse{
		ComputeCluster: ocicore.ComputeCluster{
			Id:                 &testComputeClusterId2,
			DisplayName:        lo.ToPtr("TestComputerCluster2"),
			AvailabilityDomain: lo.ToPtr("ad1"),
			CompartmentId:      lo.ToPtr("anotherCompartment"),
		},
	}

	fakeComputeClient := &FakeCompute{}
	fakeComputeClient.OnGetComputeCluster = func(ctx2 context.Context,
		request ocicore.GetComputeClusterRequest) (ocicore.GetComputeClusterResponse, error) {
		if request.ComputeClusterId != nil {
			resp, found := mockComputeClusterMap[*request.ComputeClusterId]
			if found {
				return resp, nil
			}
		}
		return ocicore.GetComputeClusterResponse{}, errors.New("compute cluster not found")
	}

	testImageId := "ocid1.image.123"
	fakeComputeClient.OnGetImage = func(ctx2 context.Context,
		request ocicore.GetImageRequest) (ocicore.GetImageResponse, error) {
		if request.ImageId != nil && *request.ImageId == testImageId {
			return ocicore.GetImageResponse{
				Image: ocicore.Image{
					Id:                     &testImageId,
					DisplayName:            lo.ToPtr("test-image"),
					OperatingSystem:        lo.ToPtr("Oracle Linux"),
					OperatingSystemVersion: lo.ToPtr("8"),
					CompartmentId:          lo.ToPtr("custom-comp"),
					TimeCreated:            &common.SDKTime{Time: time.Now()},
				},
			}, nil
		}
		return ocicore.GetImageResponse{}, errors.New("image not found")
	}

	fakeComputeClient.listShapesResp = ocicore.ListShapesResponse{
		RawResponse: &http.Response{
			StatusCode: 200,
		},
	}

	fakeComputeClient.ListImageShapeCompatibilityEntriesResp = ocicore.ListImageShapeCompatibilityEntriesResponse{
		RawResponse: &http.Response{
			StatusCode: 200,
		},
		Items: []ocicore.ImageShapeCompatibilitySummary{
			{
				ImageId: &testImageId,
				Shape:   lo.ToPtr("VM.Standard.E4.Flex"),
			},
		},
	}

	return fakeComputeClient
}

func NewFakeVirtualNetworkClient() *FakeVirtualNetwork {
	return &FakeVirtualNetwork{
		UseNetworkTestData: true,
	}
}

func NewFakeIdentityClient() *FakeIdentity {
	return NewFakeIdentity()
}

func NewDummyChannel() chan struct{} {
	startCh := make(chan struct{})
	close(startCh) // Close immediately to prevent refresh goroutine
	return startCh
}

func NewDummyConfigurationProvider() common.ConfigurationProvider {
	return common.NewRawConfigurationProvider(
		"ocid1.tenancy.oc1..dummy", // tenancy
		"ocid1.user.oc1..dummy",    // user
		"us-phoenix-1",             // region (string form, not region code)
		"d3:ad:be:ef",              // fingerprint
		"-----BEGIN PRIVATE KEY-----\nMIICfake\n-----END PRIVATE KEY-----", // dummy key; not used in tests
		nil,
	)
}
