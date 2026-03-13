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

	"github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/cache"
	"github.com/oracle/karpenter-provider-oci/pkg/oci"
	"github.com/oracle/karpenter-provider-oci/pkg/utils"
	ocicpg "github.com/oracle/oci-go-sdk/v65/clusterplacementgroups"
)

var (
	InvalidClusterPlacementGroupConfigError = errors.New("define either ocid or clusterPlacementGroup selector")
)

type Provider interface {
	ResolveClusterPlacementGroups(ctx context.Context,
		clusterPlacementGroupConfigs []*v1beta1.ClusterPlacementGroupConfig) ([]ResolveResult, error)
}

type DefaultProvider struct {
	clusterPlacementGroupClient oci.ClusterPlacementGroupClient
	clusterCompartmentId        string
	cpgResCache                 *cache.GetOrLoadCache[*ocicpg.ClusterPlacementGroup]
	cpgResSelectorCache         *cache.GetOrLoadCache[[]*ocicpg.ClusterPlacementGroupSummary]
}

func NewProvider(ctx context.Context, clusterPlacementGroupClient oci.ClusterPlacementGroupClient,
	clusterCompartmentId string) *DefaultProvider {
	return &DefaultProvider{
		clusterPlacementGroupClient: clusterPlacementGroupClient,
		clusterCompartmentId:        clusterCompartmentId,
		cpgResCache:                 cache.NewDefaultGetOrLoadCache[*ocicpg.ClusterPlacementGroup](),
		cpgResSelectorCache:         cache.NewDefaultGetOrLoadCache[[]*ocicpg.ClusterPlacementGroupSummary](),
	}
}

func (p *DefaultProvider) ResolveClusterPlacementGroups(ctx context.Context,
	clusterPlacementGroupConfigs []*v1beta1.ClusterPlacementGroupConfig) ([]ResolveResult, error) {
	if len(clusterPlacementGroupConfigs) == 0 {
		return nil, errors.New("no clusterPlacementGroup specified")
	}

	cpgRes := make([]ResolveResult, 0)
	for _, cfg := range clusterPlacementGroupConfigs {
		if cfg.ClusterPlacementGroupId != nil && cfg.ClusterPlacementGroupFilter != nil {
			return nil, InvalidClusterPlacementGroupConfigError
		}

		if cfg.ClusterPlacementGroupId != nil {
			c, err := p.getClusterPlacementGroup(ctx, *cfg.ClusterPlacementGroupId)
			if err != nil {
				return cpgRes, err
			}

			cpgRes = append(cpgRes, fromClusterPlacementGroup(c))
		} else {
			cs, err := p.filterClusterPlacementGroups(ctx, cfg.ClusterPlacementGroupFilter)
			if err != nil {
				return cpgRes, err
			}

			cpgRes = append(cpgRes, utils.MapNoIndex(cs, fromClusterPlacementGroupSummary)...)
		}
	}

	return cpgRes, nil
}

//nolint:dupl
func (p *DefaultProvider) filterClusterPlacementGroups(ctx context.Context,
	selector *v1beta1.OciResourceSelectorTerm) ([]*ocicpg.ClusterPlacementGroupSummary, error) {
	key, err := utils.HashFor(selector)
	if err != nil {
		return nil, err
	}

	compartmentId := p.clusterCompartmentId
	if selector.CompartmentId != nil {
		compartmentId = *selector.CompartmentId
	}

	var displayName *string
	if selector.DisplayName != nil {
		displayName = selector.DisplayName
	}

	listReq := ocicpg.ListClusterPlacementGroupsRequest{
		CompartmentId: &compartmentId,
		Name:          displayName,
	}

	tagFilterFunc := utils.ToTagFilterFunc(selector,
		func(s *ocicpg.ClusterPlacementGroupSummary) map[string]string {
			return s.FreeformTags
		},
		func(s *ocicpg.ClusterPlacementGroupSummary) map[string]map[string]interface{} {
			return s.DefinedTags
		},
	)

	return p.cpgResSelectorCache.GetOrLoad(ctx, key,
		func(ctx context.Context, s string) ([]*ocicpg.ClusterPlacementGroupSummary, error) {
			return p.listAndFilterClusterPlacementGroups(ctx, listReq, tagFilterFunc)
		})
}

//nolint:dupl
func (p *DefaultProvider) getClusterPlacementGroup(ctx context.Context,
	ocid string) (*ocicpg.ClusterPlacementGroup, error) {
	return p.cpgResCache.GetOrLoad(ctx, ocid,
		func(ctx context.Context, s string) (*ocicpg.ClusterPlacementGroup, error) {
			resp, err := p.clusterPlacementGroupClient.GetClusterPlacementGroup(ctx,
				ocicpg.GetClusterPlacementGroupRequest{
					ClusterPlacementGroupId: &ocid,
				})

			if err != nil {
				return nil, err
			}

			return &resp.ClusterPlacementGroup, nil
		})
}

//nolint:dupl
func (p *DefaultProvider) listAndFilterClusterPlacementGroups(ctx context.Context,
	request ocicpg.ListClusterPlacementGroupsRequest,
	extraFilterFunc func(i *ocicpg.ClusterPlacementGroupSummary) bool) ([]*ocicpg.ClusterPlacementGroupSummary,
	error) {
	var cpgSummaries []*ocicpg.ClusterPlacementGroupSummary
	for {
		resp, err := p.clusterPlacementGroupClient.ListClusterPlacementGroups(ctx, request)

		if err != nil {
			return nil, err
		}

		for _, item := range resp.Items {
			if item.LifecycleState != ocicpg.ClusterPlacementGroupLifecycleStateDeleted {
				continue
			}

			if extraFilterFunc(&item) {
				cpgSummaries = append(cpgSummaries, &item)
			}
		}

		request.Page = resp.OpcNextPage
		if request.Page == nil {
			break
		}
	}

	return cpgSummaries, nil
}
