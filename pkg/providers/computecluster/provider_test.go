/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package computecluster

import (
	"context"
	"errors"
	"testing"

	ociv1beta1 "github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/fakes"
	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestNewProvider(t *testing.T) {
	fakeClient := &fakes.FakeCompute{}
	compartmentId := "test-compartment"

	provider := NewProvider(context.TODO(), fakeClient, compartmentId)

	assert.NotNil(t, provider)
	assert.Equal(t, fakeClient, provider.computeClient)
	assert.Equal(t, compartmentId, provider.clusterCompartmentId)
	assert.NotNil(t, provider.computeClusterCache)
	assert.NotNil(t, provider.computeClusterSelectorCache)
}

func TestResolveComputeCluster(t *testing.T) {
	ctx := context.Background()
	compartmentId := "test-compartment"

	tests := []struct {
		name        string
		config      *ociv1beta1.ComputeClusterConfig
		setupFake   func(*fakes.FakeCompute)
		expected    *ResolveResult
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "config with both OCID and filter",
			config: &ociv1beta1.ComputeClusterConfig{
				ComputeClusterId:     lo.ToPtr("ocid1.computecluster.123"),
				ComputeClusterFilter: &ociv1beta1.OciResourceSelectorTerm{},
			},
			expectError: true,
		},
		{
			name: "single config with OCID",
			config: &ociv1beta1.ComputeClusterConfig{
				ComputeClusterId: lo.ToPtr("ocid1.computecluster.123"),
			},
			setupFake: func(f *fakes.FakeCompute) {
				f.OnGetComputeCluster = func(ctx context.Context,
					req ocicore.GetComputeClusterRequest) (ocicore.GetComputeClusterResponse, error) {
					return ocicore.GetComputeClusterResponse{
						ComputeCluster: ocicore.ComputeCluster{
							Id:                 lo.ToPtr("ocid1.computecluster.123"),
							DisplayName:        lo.ToPtr("test-compute-cluster"),
							AvailabilityDomain: lo.ToPtr("AD-1"),
							CompartmentId:      lo.ToPtr("comp-123"),
						},
					}, nil
				}
			},
			expected: &ResolveResult{
				Ocid:          "ocid1.computecluster.123",
				Name:          "test-compute-cluster",
				Ad:            "AD-1",
				CompartmentId: "comp-123",
			},
		},
		{
			name: "single config with filter - single match",
			config: &ociv1beta1.ComputeClusterConfig{
				ComputeClusterFilter: &ociv1beta1.OciResourceSelectorTerm{
					DisplayName: lo.ToPtr("test-compute-cluster"),
				},
			},
			setupFake: func(f *fakes.FakeCompute) {
				f.OnListComputeClusters = func(ctx context.Context,
					req ocicore.ListComputeClustersRequest) (ocicore.ListComputeClustersResponse, error) {
					return ocicore.ListComputeClustersResponse{
						ComputeClusterCollection: ocicore.ComputeClusterCollection{
							Items: []ocicore.ComputeClusterSummary{
								{
									Id:                 lo.ToPtr("ocid1.computecluster.456"),
									DisplayName:        lo.ToPtr("test-compute-cluster"),
									AvailabilityDomain: lo.ToPtr("AD-2"),
									CompartmentId:      lo.ToPtr("comp-456"),
									LifecycleState:     ocicore.ComputeClusterLifecycleStateActive,
								},
							},
						},
					}, nil
				}
			},
			expected: &ResolveResult{
				Ocid:          "ocid1.computecluster.456",
				Name:          "test-compute-cluster",
				Ad:            "AD-2",
				CompartmentId: "comp-456",
			},
		},
		{
			name: "filter with multiple matches",
			config: &ociv1beta1.ComputeClusterConfig{
				ComputeClusterFilter: &ociv1beta1.OciResourceSelectorTerm{
					DisplayName: lo.ToPtr("test-compute-cluster"),
				},
			},
			setupFake: func(f *fakes.FakeCompute) {
				f.OnListComputeClusters = func(ctx context.Context,
					req ocicore.ListComputeClustersRequest) (ocicore.ListComputeClustersResponse, error) {
					return ocicore.ListComputeClustersResponse{
						ComputeClusterCollection: ocicore.ComputeClusterCollection{
							Items: []ocicore.ComputeClusterSummary{
								{
									Id:             lo.ToPtr("ocid1.computecluster.456"),
									DisplayName:    lo.ToPtr("test-compute-cluster"),
									LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
								},
								{
									Id:             lo.ToPtr("ocid1.computecluster.789"),
									DisplayName:    lo.ToPtr("test-compute-cluster"),
									LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
								},
							},
						},
					}, nil
				}
			},
			expectError: true,
		},
		{
			name: "filter with no matches",
			config: &ociv1beta1.ComputeClusterConfig{
				ComputeClusterFilter: &ociv1beta1.OciResourceSelectorTerm{
					DisplayName: lo.ToPtr("nonexistent"),
				},
			},
			setupFake: func(f *fakes.FakeCompute) {
				f.OnListComputeClusters = func(ctx context.Context,
					req ocicore.ListComputeClustersRequest) (ocicore.ListComputeClustersResponse, error) {
					return ocicore.ListComputeClustersResponse{
						ComputeClusterCollection: ocicore.ComputeClusterCollection{
							Items: []ocicore.ComputeClusterSummary{},
						},
					}, nil
				}
			},
			expectError: true,
		},
		{
			name: "get compute cluster error",
			config: &ociv1beta1.ComputeClusterConfig{
				ComputeClusterId: lo.ToPtr("ocid1.computecluster.123"),
			},
			setupFake: func(f *fakes.FakeCompute) {
				f.OnGetComputeCluster = func(ctx context.Context,
					req ocicore.GetComputeClusterRequest) (ocicore.GetComputeClusterResponse, error) {
					return ocicore.GetComputeClusterResponse{}, errors.New("get error")
				}
			},
			expectError: true,
		},
		{
			name: "filter compute clusters error",
			config: &ociv1beta1.ComputeClusterConfig{
				ComputeClusterFilter: &ociv1beta1.OciResourceSelectorTerm{
					DisplayName: lo.ToPtr("test"),
				},
			},
			setupFake: func(f *fakes.FakeCompute) {
				f.OnListComputeClusters = func(ctx context.Context,
					req ocicore.ListComputeClustersRequest) (ocicore.ListComputeClustersResponse, error) {
					return ocicore.ListComputeClustersResponse{}, errors.New("list error")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := &fakes.FakeCompute{}
			if tt.setupFake != nil {
				tt.setupFake(fakeClient)
			}

			provider := NewProvider(ctx, fakeClient, compartmentId)

			result, err := provider.ResolveComputeCluster(ctx, tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFilterComputeClusters(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		selector          *ociv1beta1.OciResourceSelectorTerm
		setupFake         func(*fakes.FakeCompute)
		expectedResults   []*ocicore.ComputeClusterSummary
		expectError       bool
		expectedListCount int
	}{
		{
			name: "filter by display name",
			selector: &ociv1beta1.OciResourceSelectorTerm{
				DisplayName: lo.ToPtr("test-compute-cluster"),
			},
			setupFake: func(f *fakes.FakeCompute) {
				f.ListResp = ocicore.ListComputeClustersResponse{
					ComputeClusterCollection: ocicore.ComputeClusterCollection{
						Items: []ocicore.ComputeClusterSummary{
							{
								Id:             lo.ToPtr("ocid1.computecluster.123"),
								DisplayName:    lo.ToPtr("test-compute-cluster"),
								LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
							},
							{
								Id:             lo.ToPtr("ocid1.computecluster.456"),
								DisplayName:    lo.ToPtr("other-cluster"),
								LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
							},
						},
					},
				}
			},
			expectedResults: []*ocicore.ComputeClusterSummary{
				{
					Id:             lo.ToPtr("ocid1.computecluster.123"),
					DisplayName:    lo.ToPtr("test-compute-cluster"),
					LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
				},
			},
			expectedListCount: 1,
		},
		{
			name: "filter by freeform tags",
			selector: &ociv1beta1.OciResourceSelectorTerm{
				FreeformTags: map[string]string{
					"environment": "test",
				},
			},
			setupFake: func(f *fakes.FakeCompute) {
				f.OnListComputeClusters = func(ctx context.Context,
					req ocicore.ListComputeClustersRequest) (ocicore.ListComputeClustersResponse, error) {
					return ocicore.ListComputeClustersResponse{
						ComputeClusterCollection: ocicore.ComputeClusterCollection{
							Items: []ocicore.ComputeClusterSummary{
								{
									Id:             lo.ToPtr("ocid1.computecluster.123"),
									DisplayName:    lo.ToPtr("test-compute-cluster"),
									FreeformTags:   map[string]string{"environment": "test"},
									LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
								},
								{
									Id:             lo.ToPtr("ocid1.computecluster.456"),
									DisplayName:    lo.ToPtr("other-cluster"),
									LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
								},
							},
						},
					}, nil
				}
			},
			expectedResults: []*ocicore.ComputeClusterSummary{
				{
					Id:             lo.ToPtr("ocid1.computecluster.123"),
					DisplayName:    lo.ToPtr("test-compute-cluster"),
					FreeformTags:   map[string]string{"environment": "test"},
					LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
				},
			},
			expectedListCount: 1,
		},
		{
			name: "custom compartment",
			selector: &ociv1beta1.OciResourceSelectorTerm{
				CompartmentId: lo.ToPtr("custom-compartment"),
				DisplayName:   lo.ToPtr("test-compute-cluster"),
			},
			setupFake: func(f *fakes.FakeCompute) {
				f.OnListComputeClusters = func(ctx context.Context,
					req ocicore.ListComputeClustersRequest) (ocicore.ListComputeClustersResponse, error) {
					// Verify custom compartment is used
					assert.Equal(t, "custom-compartment", *req.CompartmentId)

					return ocicore.ListComputeClustersResponse{
						ComputeClusterCollection: ocicore.ComputeClusterCollection{
							Items: []ocicore.ComputeClusterSummary{
								{
									Id:             lo.ToPtr("ocid1.computecluster.123"),
									DisplayName:    lo.ToPtr("test-compute-cluster"),
									LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
								},
							},
						},
					}, nil
				}
			},
			expectedResults: []*ocicore.ComputeClusterSummary{
				{
					Id:             lo.ToPtr("ocid1.computecluster.123"),
					DisplayName:    lo.ToPtr("test-compute-cluster"),
					LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
				},
			},
			expectedListCount: 1,
		},
		{
			name: "list error",
			selector: &ociv1beta1.OciResourceSelectorTerm{
				DisplayName: lo.ToPtr("test"),
			},
			setupFake: func(f *fakes.FakeCompute) {
				f.OnListComputeClusters = func(ctx context.Context,
					req ocicore.ListComputeClustersRequest) (ocicore.ListComputeClustersResponse, error) {
					return ocicore.ListComputeClustersResponse{}, errors.New("list error")
				}
			},
			expectError:       true,
			expectedListCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := &fakes.FakeCompute{}
			if tt.setupFake != nil {
				tt.setupFake(fakeClient)
			}

			provider := NewProvider(ctx, fakeClient, "default-compartment")

			results, err := provider.filterComputeClusters(ctx, tt.selector)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResults, results)
			}

			assert.Equal(t, tt.expectedListCount, fakeClient.ListComputeClustersCount.Get())
		})
	}
}

func TestGetComputeCluster(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		ocid             string
		setupFake        func(*fakes.FakeCompute)
		expectedResult   *ocicore.ComputeCluster
		expectError      bool
		expectedGetCount int
	}{
		{
			name: "successful get",
			ocid: "ocid1.computecluster.123",
			setupFake: func(f *fakes.FakeCompute) {
				f.OnGetComputeCluster = func(ctx context.Context,
					req ocicore.GetComputeClusterRequest) (ocicore.GetComputeClusterResponse, error) {
					assert.Equal(t, "ocid1.computecluster.123", *req.ComputeClusterId)
					return ocicore.GetComputeClusterResponse{
						ComputeCluster: ocicore.ComputeCluster{
							Id:          lo.ToPtr("ocid1.computecluster.123"),
							DisplayName: lo.ToPtr("test-compute-cluster"),
						},
					}, nil
				}
			},
			expectedResult: &ocicore.ComputeCluster{
				Id:          lo.ToPtr("ocid1.computecluster.123"),
				DisplayName: lo.ToPtr("test-compute-cluster"),
			},
			expectedGetCount: 1,
		},
		{
			name: "get error",
			ocid: "ocid1.computecluster.123",
			setupFake: func(f *fakes.FakeCompute) {
				f.OnGetComputeCluster = func(ctx context.Context,
					req ocicore.GetComputeClusterRequest) (ocicore.GetComputeClusterResponse, error) {
					return ocicore.GetComputeClusterResponse{}, errors.New("get error")
				}
			},
			expectError:      true,
			expectedGetCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := &fakes.FakeCompute{}
			if tt.setupFake != nil {
				tt.setupFake(fakeClient)
			}

			provider := NewProvider(ctx, fakeClient, "test-compartment")

			result, err := provider.getComputeCluster(ctx, tt.ocid)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			assert.Equal(t, tt.expectedGetCount, fakeClient.GetComputeClusterCount.Get())
		})
	}
}

func TestListAndFilterComputeClusters(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		request           ocicore.ListComputeClustersRequest
		filterFunc        func(*ocicore.ComputeClusterSummary) bool
		setupFake         func(*fakes.FakeCompute)
		expectedResults   []*ocicore.ComputeClusterSummary
		expectedListCount int
	}{
		{
			name: "filter active items",
			request: ocicore.ListComputeClustersRequest{
				CompartmentId: lo.ToPtr("test-compartment"),
			},
			filterFunc: func(c *ocicore.ComputeClusterSummary) bool {
				return true // accept all
			},
			setupFake: func(f *fakes.FakeCompute) {
				f.OnListComputeClusters = func(ctx context.Context,
					req ocicore.ListComputeClustersRequest) (ocicore.ListComputeClustersResponse, error) {
					return ocicore.ListComputeClustersResponse{
						ComputeClusterCollection: ocicore.ComputeClusterCollection{
							Items: []ocicore.ComputeClusterSummary{
								{
									Id:             lo.ToPtr("ocid1.computecluster.123"),
									DisplayName:    lo.ToPtr("cluster-1"),
									LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
								},
								{
									Id:             lo.ToPtr("ocid1.computecluster.456"),
									DisplayName:    lo.ToPtr("cluster-2"),
									LifecycleState: ocicore.ComputeClusterLifecycleStateDeleted, // should be filtered out
								},
							},
						},
					}, nil
				}
			},
			expectedResults: []*ocicore.ComputeClusterSummary{
				{
					Id:             lo.ToPtr("ocid1.computecluster.123"),
					DisplayName:    lo.ToPtr("cluster-1"),
					LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
				},
			},
			expectedListCount: 1,
		},
		{
			name: "with pagination",
			request: ocicore.ListComputeClustersRequest{
				CompartmentId: lo.ToPtr("test-compartment"),
			},
			filterFunc: func(c *ocicore.ComputeClusterSummary) bool {
				return true
			},
			setupFake: func(f *fakes.FakeCompute) {
				callCount := 0
				f.OnListComputeClusters = func(ctx context.Context,
					req ocicore.ListComputeClustersRequest) (ocicore.ListComputeClustersResponse, error) {
					callCount++
					if req.Page == nil {
						// First page
						return ocicore.ListComputeClustersResponse{
							ComputeClusterCollection: ocicore.ComputeClusterCollection{
								Items: []ocicore.ComputeClusterSummary{
									{
										Id:             lo.ToPtr("ocid1.computecluster.123"),
										DisplayName:    lo.ToPtr("cluster-1"),
										LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
									},
								},
							},
							OpcNextPage: lo.ToPtr("next-page-token"),
						}, nil
					} else if *req.Page == "next-page-token" {
						// Second page
						return ocicore.ListComputeClustersResponse{
							ComputeClusterCollection: ocicore.ComputeClusterCollection{
								Items: []ocicore.ComputeClusterSummary{
									{
										Id:             lo.ToPtr("ocid1.computecluster.456"),
										DisplayName:    lo.ToPtr("cluster-2"),
										LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
									},
								},
							},
						}, nil
					}
					return ocicore.ListComputeClustersResponse{}, nil
				}
			},
			expectedResults: []*ocicore.ComputeClusterSummary{
				{
					Id:             lo.ToPtr("ocid1.computecluster.123"),
					DisplayName:    lo.ToPtr("cluster-1"),
					LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
				},
				{
					Id:             lo.ToPtr("ocid1.computecluster.456"),
					DisplayName:    lo.ToPtr("cluster-2"),
					LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
				},
			},
			expectedListCount: 2,
		},
		{
			name: "with custom filter",
			request: ocicore.ListComputeClustersRequest{
				CompartmentId: lo.ToPtr("test-compartment"),
			},
			filterFunc: func(c *ocicore.ComputeClusterSummary) bool {
				return *c.DisplayName == "cluster-1" // only accept cluster-1
			},
			setupFake: func(f *fakes.FakeCompute) {
				f.OnListComputeClusters = func(ctx context.Context,
					req ocicore.ListComputeClustersRequest) (ocicore.ListComputeClustersResponse, error) {
					return ocicore.ListComputeClustersResponse{
						ComputeClusterCollection: ocicore.ComputeClusterCollection{
							Items: []ocicore.ComputeClusterSummary{
								{
									Id:             lo.ToPtr("ocid1.computecluster.123"),
									DisplayName:    lo.ToPtr("cluster-1"),
									LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
								},
								{
									Id:             lo.ToPtr("ocid1.computecluster.456"),
									DisplayName:    lo.ToPtr("cluster-2"),
									LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
								},
							},
						},
					}, nil
				}
			},
			expectedResults: []*ocicore.ComputeClusterSummary{
				{
					Id:             lo.ToPtr("ocid1.computecluster.123"),
					DisplayName:    lo.ToPtr("cluster-1"),
					LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
				},
			},
			expectedListCount: 1,
		},
		{
			name: "nil filter func",
			request: ocicore.ListComputeClustersRequest{
				CompartmentId: lo.ToPtr("test-compartment"),
			},
			filterFunc: nil,
			setupFake: func(f *fakes.FakeCompute) {
				f.OnListComputeClusters = func(ctx context.Context,
					req ocicore.ListComputeClustersRequest) (ocicore.ListComputeClustersResponse, error) {
					return ocicore.ListComputeClustersResponse{
						ComputeClusterCollection: ocicore.ComputeClusterCollection{
							Items: []ocicore.ComputeClusterSummary{
								{
									Id:             lo.ToPtr("ocid1.computecluster.123"),
									DisplayName:    lo.ToPtr("cluster-1"),
									LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
								},
							},
						},
					}, nil
				}
			},
			expectedResults: []*ocicore.ComputeClusterSummary{
				{
					Id:             lo.ToPtr("ocid1.computecluster.123"),
					DisplayName:    lo.ToPtr("cluster-1"),
					LifecycleState: ocicore.ComputeClusterLifecycleStateActive,
				},
			},
			expectedListCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := &fakes.FakeCompute{}
			if tt.setupFake != nil {
				tt.setupFake(fakeClient)
			}

			provider := NewProvider(ctx, fakeClient, "test-compartment")

			results, err := provider.listAndFilterComputeClusters(ctx, tt.request, tt.filterFunc)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResults, results)
			assert.Equal(t, tt.expectedListCount, fakeClient.ListComputeClustersCount.Get())
		})
	}
}

func TestResolveResultFunctions(t *testing.T) {
	t.Run("fromComputeCluster", func(t *testing.T) {
		cluster := &ocicore.ComputeCluster{
			Id:                 lo.ToPtr("ocid1.computecluster.123"),
			DisplayName:        lo.ToPtr("test-cluster"),
			AvailabilityDomain: lo.ToPtr("AD-1"),
			CompartmentId:      lo.ToPtr("comp-123"),
		}

		result := fromComputeCluster(cluster)

		expected := &ResolveResult{
			Ocid:          "ocid1.computecluster.123",
			Name:          "test-cluster",
			Ad:            "AD-1",
			CompartmentId: "comp-123",
		}

		assert.Equal(t, expected, result)
	})

	t.Run("fromComputeClusterSummary", func(t *testing.T) {
		clusterSummary := &ocicore.ComputeClusterSummary{
			Id:                 lo.ToPtr("ocid1.computecluster.456"),
			DisplayName:        lo.ToPtr("test-cluster-summary"),
			AvailabilityDomain: lo.ToPtr("AD-2"),
			CompartmentId:      lo.ToPtr("comp-456"),
		}

		result := fromComputeClusterSummary(clusterSummary)

		expected := &ResolveResult{
			Ocid:          "ocid1.computecluster.456",
			Name:          "test-cluster-summary",
			Ad:            "AD-2",
			CompartmentId: "comp-456",
		}

		assert.Equal(t, expected, result)
	})
}
