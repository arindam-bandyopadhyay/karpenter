/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

//nolint:dupl
package nodeclasses

import (
	"context"
	"errors"
	"time"

	"github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/clusterplacementgroup"
	"github.com/oracle/karpenter-provider-oci/pkg/utils"
	"github.com/samber/lo"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const ClusterPlacementGroupInTheSameAd = "multiple cluster placement group(s) is in an ad"

type ClusterPlacementGroupReconciler struct {
	clusterPlacementGroupProvider clusterplacementgroup.Provider
}

func (c *ClusterPlacementGroupReconciler) Reconcile(ctx context.Context,
	nodeClass *v1beta1.OCINodeClass) (reconcile.Result, error) {
	if len(nodeClass.Spec.ClusterPlacementGroupConfigs) > 0 {
		results, err := c.clusterPlacementGroupProvider.
			ResolveClusterPlacementGroups(ctx, nodeClass.Spec.ClusterPlacementGroupConfigs)

		nodeClass.Status.ClusterPlacementGroups = toClusterPlacementGroups(results)
		if err != nil {
			log.FromContext(ctx).Error(err, "failed to resolve cluster placement group")
		} else {
			multipleCpgInSameAd := clusterPlacementGroupsInTheSameAd(results)
			if len(multipleCpgInSameAd) > 0 {
				log.FromContext(ctx).Info("found multiple cpgs in the same ad", "cpgs",
					multipleCpgInSameAd)
				err = errors.New(ClusterPlacementGroupInTheSameAd)
			}
		}

		if err != nil {
			nodeClass.StatusConditions().SetFalse(v1beta1.ConditionTypeClusterPlacementGroup,
				v1beta1.ConditionClusterPlacementGroupNotReadyReason, utils.PrettyString(err.Error(), 1024))

			// TODO - classify error and decide whether to retry, we don't return error so as not
			// to disturb other reconciliation.
			return reconcile.Result{
				RequeueAfter: 5 * time.Minute,
			}, nil
		}
		log.FromContext(ctx).Info("cluster placement group resolved")
		nodeClass.StatusConditions().SetTrue(v1beta1.ConditionTypeClusterPlacementGroup)
		return reconcile.Result{}, nil
	} else {
		nodeClass.Status.ClusterPlacementGroups = nil
		err := nodeClass.StatusConditions().Clear(v1beta1.ConditionTypeClusterPlacementGroup)

		if err != nil {
			log.FromContext(ctx).Error(err, "failed to clear cluster placement group condition")

			return reconcile.Result{
				RequeueAfter: 5 * time.Minute,
			}, nil
		}
	}

	return reconcile.Result{}, nil
}

func toClusterPlacementGroups(results []clusterplacementgroup.ResolveResult) []v1beta1.ClusterPlacementGroup {
	return utils.MapNoIndex(results, func(cpgRes clusterplacementgroup.ResolveResult) v1beta1.ClusterPlacementGroup {
		return v1beta1.ClusterPlacementGroup{
			ClusterPlacementGroupId: cpgRes.Ocid,
			DisplayName:             cpgRes.Name,
			AvailabilityDomain:      cpgRes.Ad,
		}
	})
}

func clusterPlacementGroupsInTheSameAd(results []clusterplacementgroup.ResolveResult) map[string][]string {
	groupByAds := lo.GroupBy(results, func(item clusterplacementgroup.ResolveResult) string {
		return item.Ad
	})

	out := make(map[string][]string)
	for k, v := range groupByAds {
		if len(v) > 1 {
			out[k] = lo.Map(v, func(item clusterplacementgroup.ResolveResult, _ int) string {
				return item.Ocid
			})
		}
	}

	return out
}
