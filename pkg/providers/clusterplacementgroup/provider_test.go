/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package clusterplacementgroup

import (
	"context"
	"errors"
	"testing"

	"github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/fakes"
	ocicpg "github.com/oracle/oci-go-sdk/v65/clusterplacementgroups"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestNewProvider(t *testing.T) {
	ctx := context.Background()
	fakeClient := &fakes.FakeClusterPlacementGroup{}
	compartmentId := "test-compartment"

	provider := NewProvider(ctx, fakeClient, compartmentId)

	assert.NotNil(t, provider)
	assert.Equal(t, fakeClient, provider.clusterPlacementGroupClient)
	assert.Equal(t, compartmentId, provider.clusterCompartmentId)
	assert.NotNil(t, provider.cpgResCache)
	assert.NotNil(t, provider.cpgResSelectorCache)
}

func TestResolveClusterPlacementGroups(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		configs           []*v1beta1.ClusterPlacementGroupConfig
		setupFake         func(*fakes.FakeClusterPlacementGroup)
		expectedResults   []ResolveResult
		expectError       bool
		expectedGetCount  int
		expectedListCount int
	}{
		{
			name:        "empty configs",
			configs:     []*v1beta1.ClusterPlacementGroupConfig{},
			expectError: true,
		},
		{
			name: "single config with OCID",
			configs: []*v1beta1.ClusterPlacementGroupConfig{
				{
					ClusterPlacementGroupId: lo.ToPtr("ocid1.cpg.123"),
				},
			},
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				f.GetResp = ocicpg.GetClusterPlacementGroupResponse{
					ClusterPlacementGroup: ocicpg.ClusterPlacementGroup{
						Id:                 lo.ToPtr("ocid1.cpg.123"),
						Name:               lo.ToPtr("test-cpg"),
						AvailabilityDomain: lo.ToPtr("AD-1"),
						CompartmentId:      lo.ToPtr("comp-123"),
					},
				}
			},
			expectedResults: []ResolveResult{
				{
					Ocid:          "ocid1.cpg.123",
					Name:          "test-cpg",
					Ad:            "AD-1",
					CompartmentId: "comp-123",
				},
			},
			expectedGetCount: 1,
		},
		{
			name: "single config with filter",
			configs: []*v1beta1.ClusterPlacementGroupConfig{
				{
					ClusterPlacementGroupFilter: &v1beta1.OciResourceSelectorTerm{
						DisplayName: lo.ToPtr("test-cpg"),
					},
				},
			},
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				f.ListResp = ocicpg.ListClusterPlacementGroupsResponse{
					ClusterPlacementGroupCollection: ocicpg.ClusterPlacementGroupCollection{
						Items: []ocicpg.ClusterPlacementGroupSummary{
							{
								Id:                 lo.ToPtr("ocid1.cpg.456"),
								Name:               lo.ToPtr("test-cpg"),
								AvailabilityDomain: lo.ToPtr("AD-2"),
								CompartmentId:      lo.ToPtr("comp-456"),
								LifecycleState:     ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
							},
						},
					},
				}
			},
			expectedResults: []ResolveResult{
				{
					Ocid:          "ocid1.cpg.456",
					Name:          "test-cpg",
					Ad:            "AD-2",
					CompartmentId: "comp-456",
				},
			},
			expectedListCount: 1,
		},
		{
			name: "multiple configs",
			configs: []*v1beta1.ClusterPlacementGroupConfig{
				{
					ClusterPlacementGroupId: lo.ToPtr("ocid1.cpg.123"),
				},
				{
					ClusterPlacementGroupFilter: &v1beta1.OciResourceSelectorTerm{
						DisplayName: lo.ToPtr("test-cpg-2"),
					},
				},
			},
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				f.GetResp = ocicpg.GetClusterPlacementGroupResponse{
					ClusterPlacementGroup: ocicpg.ClusterPlacementGroup{
						Id:                 lo.ToPtr("ocid1.cpg.123"),
						Name:               lo.ToPtr("test-cpg-1"),
						AvailabilityDomain: lo.ToPtr("AD-1"),
						CompartmentId:      lo.ToPtr("comp-123"),
					},
				}
				f.ListResp = ocicpg.ListClusterPlacementGroupsResponse{
					ClusterPlacementGroupCollection: ocicpg.ClusterPlacementGroupCollection{
						Items: []ocicpg.ClusterPlacementGroupSummary{
							{
								Id:                 lo.ToPtr("ocid1.cpg.456"),
								Name:               lo.ToPtr("test-cpg-2"),
								AvailabilityDomain: lo.ToPtr("AD-2"),
								CompartmentId:      lo.ToPtr("comp-456"),
								LifecycleState:     ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
							},
						},
					},
				}
			},
			expectedResults: []ResolveResult{
				{
					Ocid:          "ocid1.cpg.123",
					Name:          "test-cpg-1",
					Ad:            "AD-1",
					CompartmentId: "comp-123",
				},
				{
					Ocid:          "ocid1.cpg.456",
					Name:          "test-cpg-2",
					Ad:            "AD-2",
					CompartmentId: "comp-456",
				},
			},
			expectedGetCount:  1,
			expectedListCount: 1,
		},
		{
			name: "config with both OCID and filter",
			configs: []*v1beta1.ClusterPlacementGroupConfig{
				{
					ClusterPlacementGroupId:     lo.ToPtr("ocid1.cpg.123"),
					ClusterPlacementGroupFilter: &v1beta1.OciResourceSelectorTerm{},
				},
			},
			expectError: true,
		},
		{
			name: "get cluster placement group error",
			configs: []*v1beta1.ClusterPlacementGroupConfig{
				{
					ClusterPlacementGroupId: lo.ToPtr("ocid1.cpg.123"),
				},
			},
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				f.GetErr = errors.New("get error")
			},
			expectError:      true,
			expectedGetCount: 1,
		},
		{
			name: "filter cluster placement groups error",
			configs: []*v1beta1.ClusterPlacementGroupConfig{
				{
					ClusterPlacementGroupFilter: &v1beta1.OciResourceSelectorTerm{
						DisplayName: lo.ToPtr("test"),
					},
				},
			},
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				f.ListErr = errors.New("list error")
			},
			expectError:       true,
			expectedListCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := &fakes.FakeClusterPlacementGroup{}
			if tt.setupFake != nil {
				tt.setupFake(fakeClient)
			}

			provider := NewProvider(ctx, fakeClient, "test-compartment")

			results, err := provider.ResolveClusterPlacementGroups(ctx, tt.configs)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResults, results)
			}

			assert.Equal(t, tt.expectedGetCount, fakeClient.GetCount.Get())
			assert.Equal(t, tt.expectedListCount, fakeClient.ListCount.Get())
		})
	}
}

func TestFilterClusterPlacementGroups(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		selector          *v1beta1.OciResourceSelectorTerm
		setupFake         func(*fakes.FakeClusterPlacementGroup)
		expectedResults   []*ocicpg.ClusterPlacementGroupSummary
		expectError       bool
		expectedListCount int
	}{
		{
			name: "filter by display name",
			selector: &v1beta1.OciResourceSelectorTerm{
				DisplayName: lo.ToPtr("test-cpg"),
			},
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				f.ListResp = ocicpg.ListClusterPlacementGroupsResponse{
					ClusterPlacementGroupCollection: ocicpg.ClusterPlacementGroupCollection{
						Items: []ocicpg.ClusterPlacementGroupSummary{
							{
								Id:             lo.ToPtr("ocid1.cpg.123"),
								Name:           lo.ToPtr("test-cpg"),
								LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
							},
							{
								Id:             lo.ToPtr("ocid1.cpg.456"),
								Name:           lo.ToPtr("other-cpg"),
								LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
							},
						},
					},
				}
			},
			expectedResults: []*ocicpg.ClusterPlacementGroupSummary{
				{
					Id:             lo.ToPtr("ocid1.cpg.123"),
					Name:           lo.ToPtr("test-cpg"),
					LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
				},
			},
			expectedListCount: 1,
		},
		{
			name: "filter by freeform tags",
			selector: &v1beta1.OciResourceSelectorTerm{
				FreeformTags: map[string]string{
					"environment": "test",
				},
			},
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				f.ListResp = ocicpg.ListClusterPlacementGroupsResponse{
					ClusterPlacementGroupCollection: ocicpg.ClusterPlacementGroupCollection{
						Items: []ocicpg.ClusterPlacementGroupSummary{
							{
								Id:             lo.ToPtr("ocid1.cpg.123"),
								Name:           lo.ToPtr("test-cpg"),
								FreeformTags:   map[string]string{"environment": "test"},
								LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
							},
						},
					},
				}
			},
			expectedResults: []*ocicpg.ClusterPlacementGroupSummary{
				{
					Id:             lo.ToPtr("ocid1.cpg.123"),
					Name:           lo.ToPtr("test-cpg"),
					FreeformTags:   map[string]string{"environment": "test"},
					LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
				},
			},
			expectedListCount: 1,
		},
		{
			name: "custom compartment",
			selector: &v1beta1.OciResourceSelectorTerm{
				CompartmentId: lo.ToPtr("custom-compartment"),
				DisplayName:   lo.ToPtr("test-cpg"),
			},
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				f.ListResp = ocicpg.ListClusterPlacementGroupsResponse{
					ClusterPlacementGroupCollection: ocicpg.ClusterPlacementGroupCollection{
						Items: []ocicpg.ClusterPlacementGroupSummary{
							{
								Id:             lo.ToPtr("ocid1.cpg.123"),
								Name:           lo.ToPtr("test-cpg"),
								LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
							},
						},
					},
				}
			},
			expectedResults: []*ocicpg.ClusterPlacementGroupSummary{
				{
					Id:             lo.ToPtr("ocid1.cpg.123"),
					Name:           lo.ToPtr("test-cpg"),
					LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
				},
			},
			expectedListCount: 1,
		},
		{
			name: "list error",
			selector: &v1beta1.OciResourceSelectorTerm{
				DisplayName: lo.ToPtr("test"),
			},
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				f.ListErr = errors.New("list error")
			},
			expectError:       true,
			expectedListCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := &fakes.FakeClusterPlacementGroup{}
			if tt.setupFake != nil {
				tt.setupFake(fakeClient)
			}

			provider := NewProvider(ctx, fakeClient, "default-compartment")

			results, err := provider.filterClusterPlacementGroups(ctx, tt.selector)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResults, results)
			}

			assert.Equal(t, tt.expectedListCount, fakeClient.ListCount.Get())
		})
	}
}

func TestGetClusterPlacementGroup(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		ocid             string
		setupFake        func(*fakes.FakeClusterPlacementGroup)
		expectedResult   *ocicpg.ClusterPlacementGroup
		expectError      bool
		expectedGetCount int
	}{
		{
			name: "successful get",
			ocid: "ocid1.cpg.123",
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				f.GetResp = ocicpg.GetClusterPlacementGroupResponse{
					ClusterPlacementGroup: ocicpg.ClusterPlacementGroup{
						Id:   lo.ToPtr("ocid1.cpg.123"),
						Name: lo.ToPtr("test-cpg"),
					},
				}
			},
			expectedResult: &ocicpg.ClusterPlacementGroup{
				Id:   lo.ToPtr("ocid1.cpg.123"),
				Name: lo.ToPtr("test-cpg"),
			},
			expectedGetCount: 1,
		},
		{
			name: "get error",
			ocid: "ocid1.cpg.123",
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				f.GetErr = errors.New("get error")
			},
			expectError:      true,
			expectedGetCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := &fakes.FakeClusterPlacementGroup{}
			if tt.setupFake != nil {
				tt.setupFake(fakeClient)
			}

			provider := NewProvider(ctx, fakeClient, "test-compartment")

			result, err := provider.getClusterPlacementGroup(ctx, tt.ocid)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			assert.Equal(t, tt.expectedGetCount, fakeClient.GetCount.Get())
		})
	}
}

func TestListAndFilterClusterPlacementGroups(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		request           ocicpg.ListClusterPlacementGroupsRequest
		filterFunc        func(*ocicpg.ClusterPlacementGroupSummary) bool
		setupFake         func(*fakes.FakeClusterPlacementGroup)
		expectedResults   []*ocicpg.ClusterPlacementGroupSummary
		expectedListCount int
	}{
		{
			name: "filter active items",
			request: ocicpg.ListClusterPlacementGroupsRequest{
				CompartmentId: lo.ToPtr("test-compartment"),
			},
			filterFunc: func(c *ocicpg.ClusterPlacementGroupSummary) bool {
				return true // accept all
			},
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				f.ListResp = ocicpg.ListClusterPlacementGroupsResponse{
					ClusterPlacementGroupCollection: ocicpg.ClusterPlacementGroupCollection{
						Items: []ocicpg.ClusterPlacementGroupSummary{
							{
								Id:             lo.ToPtr("ocid1.cpg.123"),
								Name:           lo.ToPtr("cpg-1"),
								LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
							},
							{
								Id:             lo.ToPtr("ocid1.cpg.456"),
								Name:           lo.ToPtr("cpg-2"),
								LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateActive, // should be filtered out
							},
						},
					},
				}
			},
			expectedResults: []*ocicpg.ClusterPlacementGroupSummary{
				{
					Id:             lo.ToPtr("ocid1.cpg.123"),
					Name:           lo.ToPtr("cpg-1"),
					LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
				},
			},
			expectedListCount: 1,
		},
		{
			name: "with pagination",
			request: ocicpg.ListClusterPlacementGroupsRequest{
				CompartmentId: lo.ToPtr("test-compartment"),
			},
			filterFunc: func(c *ocicpg.ClusterPlacementGroupSummary) bool {
				return true
			},
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				// First call returns page with next page token
				f.OnList = func(ctx context.Context,
					req ocicpg.ListClusterPlacementGroupsRequest) (ocicpg.ListClusterPlacementGroupsResponse, error) {
					if req.Page == nil {
						// First page
						return ocicpg.ListClusterPlacementGroupsResponse{
							ClusterPlacementGroupCollection: ocicpg.ClusterPlacementGroupCollection{
								Items: []ocicpg.ClusterPlacementGroupSummary{
									{
										Id:             lo.ToPtr("ocid1.cpg.123"),
										Name:           lo.ToPtr("cpg-1"),
										LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
									},
								},
							},
							OpcNextPage: lo.ToPtr("next-page-token"),
						}, nil
					} else if *req.Page == "next-page-token" {
						// Second page
						return ocicpg.ListClusterPlacementGroupsResponse{
							ClusterPlacementGroupCollection: ocicpg.ClusterPlacementGroupCollection{
								Items: []ocicpg.ClusterPlacementGroupSummary{
									{
										Id:             lo.ToPtr("ocid1.cpg.456"),
										Name:           lo.ToPtr("cpg-2"),
										LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
									},
								},
							},
						}, nil
					}
					return ocicpg.ListClusterPlacementGroupsResponse{}, nil
				}
			},
			expectedResults: []*ocicpg.ClusterPlacementGroupSummary{
				{
					Id:             lo.ToPtr("ocid1.cpg.123"),
					Name:           lo.ToPtr("cpg-1"),
					LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
				},
				{
					Id:             lo.ToPtr("ocid1.cpg.456"),
					Name:           lo.ToPtr("cpg-2"),
					LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
				},
			},
			expectedListCount: 2,
		},
		{
			name: "with custom filter",
			request: ocicpg.ListClusterPlacementGroupsRequest{
				CompartmentId: lo.ToPtr("test-compartment"),
			},
			filterFunc: func(c *ocicpg.ClusterPlacementGroupSummary) bool {
				return *c.Name == "cpg-1" // only accept cpg-1
			},
			setupFake: func(f *fakes.FakeClusterPlacementGroup) {
				f.ListResp = ocicpg.ListClusterPlacementGroupsResponse{
					ClusterPlacementGroupCollection: ocicpg.ClusterPlacementGroupCollection{
						Items: []ocicpg.ClusterPlacementGroupSummary{
							{
								Id:             lo.ToPtr("ocid1.cpg.123"),
								Name:           lo.ToPtr("cpg-1"),
								LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
							},
							{
								Id:             lo.ToPtr("ocid1.cpg.456"),
								Name:           lo.ToPtr("cpg-2"),
								LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
							},
						},
					},
				}
			},
			expectedResults: []*ocicpg.ClusterPlacementGroupSummary{
				{
					Id:             lo.ToPtr("ocid1.cpg.123"),
					Name:           lo.ToPtr("cpg-1"),
					LifecycleState: ocicpg.ClusterPlacementGroupLifecycleStateDeleted,
				},
			},
			expectedListCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := &fakes.FakeClusterPlacementGroup{}
			if tt.setupFake != nil {
				tt.setupFake(fakeClient)
			}

			provider := NewProvider(ctx, fakeClient, "test-compartment")

			results, err := provider.listAndFilterClusterPlacementGroups(ctx, tt.request, tt.filterFunc)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResults, results)
			assert.Equal(t, tt.expectedListCount, fakeClient.ListCount.Get())
		})
	}
}

func TestResolveResultFunctions(t *testing.T) {
	t.Run("fromClusterPlacementGroup", func(t *testing.T) {
		cpg := &ocicpg.ClusterPlacementGroup{
			Id:                 lo.ToPtr("ocid1.cpg.123"),
			Name:               lo.ToPtr("test-cpg"),
			AvailabilityDomain: lo.ToPtr("AD-1"),
			CompartmentId:      lo.ToPtr("comp-123"),
		}

		result := fromClusterPlacementGroup(cpg)

		expected := ResolveResult{
			Ocid:          "ocid1.cpg.123",
			Name:          "test-cpg",
			Ad:            "AD-1",
			CompartmentId: "comp-123",
		}

		assert.Equal(t, expected, result)
	})

	t.Run("fromClusterPlacementGroupSummary", func(t *testing.T) {
		cpgSummary := &ocicpg.ClusterPlacementGroupSummary{
			Id:                 lo.ToPtr("ocid1.cpg.456"),
			Name:               lo.ToPtr("test-cpg-summary"),
			AvailabilityDomain: lo.ToPtr("AD-2"),
			CompartmentId:      lo.ToPtr("comp-456"),
		}

		result := fromClusterPlacementGroupSummary(cpgSummary)

		expected := ResolveResult{
			Ocid:          "ocid1.cpg.456",
			Name:          "test-cpg-summary",
			Ad:            "AD-2",
			CompartmentId: "comp-456",
		}

		assert.Equal(t, expected, result)
	})
}
