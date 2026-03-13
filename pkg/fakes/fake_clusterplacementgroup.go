/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package fakes

import (
	"context"

	ocicpg "github.com/oracle/oci-go-sdk/v65/clusterplacementgroups"
)

// FakeClusterPlacementGroup implements oci.ClusterPlacementGroupClient for tests.
type FakeClusterPlacementGroup struct {
	GetResp  ocicpg.GetClusterPlacementGroupResponse
	GetErr   error
	ListResp ocicpg.ListClusterPlacementGroupsResponse
	ListErr  error

	GetCount  Counter
	ListCount Counter

	// Optional hooks override canned behavior
	OnGet  func(context.Context, ocicpg.GetClusterPlacementGroupRequest) (ocicpg.GetClusterPlacementGroupResponse, error)
	OnList func(context.Context, ocicpg.ListClusterPlacementGroupsRequest) (
		ocicpg.ListClusterPlacementGroupsResponse, error)
}

func (f *FakeClusterPlacementGroup) GetClusterPlacementGroup(ctx context.Context,
	request ocicpg.GetClusterPlacementGroupRequest) (ocicpg.GetClusterPlacementGroupResponse, error) {
	f.GetCount.Inc()
	if f.OnGet != nil {
		return f.OnGet(ctx, request)
	}
	if f.GetErr != nil {
		return ocicpg.GetClusterPlacementGroupResponse{}, f.GetErr
	}
	return f.GetResp, nil
}

func (f *FakeClusterPlacementGroup) ListClusterPlacementGroups(ctx context.Context,
	request ocicpg.ListClusterPlacementGroupsRequest) (ocicpg.ListClusterPlacementGroupsResponse, error) {
	f.ListCount.Inc()
	if f.OnList != nil {
		return f.OnList(ctx, request)
	}
	if f.ListErr != nil {
		return ocicpg.ListClusterPlacementGroupsResponse{}, f.ListErr
	}

	// Filter items based on request parameters
	items := f.ListResp.ClusterPlacementGroupCollection.Items
	if request.Name != nil {
		// Filter by display name
		filteredItems := make([]ocicpg.ClusterPlacementGroupSummary, 0)
		for _, item := range items {
			if item.Name != nil && *item.Name == *request.Name {
				filteredItems = append(filteredItems, item)
			}
		}
		items = filteredItems
	}

	return ocicpg.ListClusterPlacementGroupsResponse{
		ClusterPlacementGroupCollection: ocicpg.ClusterPlacementGroupCollection{
			Items: items,
		},
		OpcNextPage: f.ListResp.OpcNextPage,
	}, nil
}
