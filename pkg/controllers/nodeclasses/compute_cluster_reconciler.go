/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package nodeclasses

import (
	"context"
	"time"

	"github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/computecluster"
	"github.com/oracle/karpenter-provider-oci/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const ComputeClusterInDifferentCompartment = "compute cluster is in a different compartment from node compartment"

type ComputeClusterReconciler struct {
	computeClusterProvider computecluster.Provider
}

func (c *ComputeClusterReconciler) Reconcile(ctx context.Context,
	nodeClass *v1beta1.OCINodeClass) (reconcile.Result, error) {
	if nodeClass.Spec.ComputeClusterConfig != nil {
		result, err := c.computeClusterProvider.ResolveComputeCluster(ctx,
			nodeClass.Spec.ComputeClusterConfig)

		if err != nil {
			log.FromContext(ctx).Error(err, "failed to resolve compute cluster")
			nodeClass.Status.ComputeCluster = nil
			nodeClass.StatusConditions().SetFalse(v1beta1.ConditionTypeComputeCluster,
				v1beta1.ConditionComputeClusterNotReadyReason, utils.PrettyString(err.Error(), 1024))

			// TODO - classify error and decide whether to retry, we don't return error so as not
			// to disturb other reconciliation.
			return reconcile.Result{
				RequeueAfter: 5 * time.Minute,
			}, nil
		}

		log.FromContext(ctx).Info("compute cluster resolved")
		nodeClass.Status.ComputeCluster = toComputeCluster(result)
		nodeClass.StatusConditions().SetTrue(v1beta1.ConditionTypeComputeCluster)
	}

	return reconcile.Result{}, nil
}

func toComputeCluster(res *computecluster.ResolveResult) *v1beta1.ComputeCluster {
	return &v1beta1.ComputeCluster{
		ComputeClusterId: res.Ocid,
		DisplayName:      res.Name,
	}
}
