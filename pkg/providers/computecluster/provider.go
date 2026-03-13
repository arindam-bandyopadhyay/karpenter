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
	"fmt"

	ociv1beta1 "github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/cache"
	"github.com/oracle/karpenter-provider-oci/pkg/oci"
	"github.com/oracle/karpenter-provider-oci/pkg/utils"
	ocicore "github.com/oracle/oci-go-sdk/v65/core"
)

var (
	InvalidComputeClusterConfigError = errors.New("define either ocid or compute cluster selector")
)

type Provider interface {
	ResolveComputeCluster(ctx context.Context,
		cfg *ociv1beta1.ComputeClusterConfig) (*ResolveResult, error)
}

type DefaultProvider struct {
	computeClient               oci.ComputeClient
	clusterCompartmentId        string
	computeClusterCache         *cache.GetOrLoadCache[*ocicore.ComputeCluster]
	computeClusterSelectorCache *cache.GetOrLoadCache[[]*ocicore.ComputeClusterSummary]
}

func NewProvider(ctx context.Context, computeClient oci.ComputeClient,
	clusterCompartmentId string) *DefaultProvider {
	return &DefaultProvider{
		computeClient:               computeClient,
		clusterCompartmentId:        clusterCompartmentId,
		computeClusterCache:         cache.NewDefaultGetOrLoadCache[*ocicore.ComputeCluster](),
		computeClusterSelectorCache: cache.NewDefaultGetOrLoadCache[[]*ocicore.ComputeClusterSummary](),
	}
}

func (p *DefaultProvider) ResolveComputeCluster(ctx context.Context,
	cfg *ociv1beta1.ComputeClusterConfig) (*ResolveResult, error) {
	if cfg == nil {
		return nil, errors.New("no computeClusterConfig specified")
	}

	if cfg.ComputeClusterId != nil && cfg.ComputeClusterFilter != nil {
		return nil, InvalidComputeClusterConfigError
	}

	if cfg.ComputeClusterId != nil {
		c, err := p.getComputeCluster(ctx, *cfg.ComputeClusterId)
		if err != nil {
			return nil, err
		}

		return fromComputeCluster(c), nil
	} else {
		cs, err := p.filterComputeClusters(ctx, cfg.ComputeClusterFilter)
		if err != nil {
			return nil, err
		}

		if len(cs) != 1 {
			return nil, fmt.Errorf("unique computeCluster is required, actual: %d", len(cs))
		}

		return fromComputeClusterSummary(cs[0]), nil
	}
}

//nolint:dupl
func (p *DefaultProvider) filterComputeClusters(ctx context.Context,
	selector *ociv1beta1.OciResourceSelectorTerm) ([]*ocicore.ComputeClusterSummary, error) {
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

	listReq := ocicore.ListComputeClustersRequest{
		CompartmentId: &compartmentId,
		DisplayName:   displayName,
	}

	tagFilterFunc := utils.ToTagFilterFunc(selector,
		func(s *ocicore.ComputeClusterSummary) map[string]string {
			return s.FreeformTags
		},
		func(s *ocicore.ComputeClusterSummary) map[string]map[string]interface{} {
			return s.DefinedTags
		},
	)

	return p.computeClusterSelectorCache.GetOrLoad(ctx, key,
		func(ctx context.Context, s string) ([]*ocicore.ComputeClusterSummary, error) {
			return p.listAndFilterComputeClusters(ctx, listReq, tagFilterFunc)
		})
}

//nolint:dupl
func (p *DefaultProvider) getComputeCluster(ctx context.Context, ocid string) (*ocicore.ComputeCluster, error) {
	return p.computeClusterCache.GetOrLoad(ctx, ocid,
		func(ctx context.Context, s string) (*ocicore.ComputeCluster, error) {
			resp, err := p.computeClient.GetComputeCluster(ctx,
				ocicore.GetComputeClusterRequest{
					ComputeClusterId: &ocid,
				})

			if err != nil {
				return nil, err
			}

			return &resp.ComputeCluster, nil
		})
}

//nolint:dupl
func (p *DefaultProvider) listAndFilterComputeClusters(ctx context.Context,
	request ocicore.ListComputeClustersRequest,
	extraFilterFunc func(i *ocicore.ComputeClusterSummary) bool) ([]*ocicore.ComputeClusterSummary,
	error) {
	var computeClusterSummaries []*ocicore.ComputeClusterSummary
	for {
		resp, err := p.computeClient.ListComputeClusters(ctx, request)

		if err != nil {
			return nil, err
		}

		for _, item := range resp.Items {
			if item.LifecycleState != ocicore.ComputeClusterLifecycleStateActive {
				continue
			}

			// Compute API has a bug that filtering by displayName doesn't work
			if request.DisplayName != nil {
				if *item.DisplayName != *request.DisplayName {
					continue
				}
			}

			if extraFilterFunc == nil || extraFilterFunc(&item) {
				computeClusterSummaries = append(computeClusterSummaries, &item)
			}
		}

		request.Page = resp.OpcNextPage
		if request.Page == nil {
			break
		}
	}

	return computeClusterSummaries, nil
}
