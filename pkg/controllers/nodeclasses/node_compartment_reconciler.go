/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package nodeclasses

import (
	"context"
	"errors"
	"time"

	"github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/identity"
	"github.com/oracle/karpenter-provider-oci/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type NodeCompartmentReconciler struct {
	// TODO - support node compartment resolution
	identityProvider     identity.Provider
	clusterCompartmentId string
}

func (n *NodeCompartmentReconciler) Reconcile(ctx context.Context,
	nodeClass *v1beta1.OCINodeClass) (reconcile.Result, error) {

	_, err := n.identityProvider.ResolveCompartment(ctx, n.getNodeCompartment(nodeClass))

	if err != nil {
		log.FromContext(ctx).Error(err, "failed to resolve node compartment")
		nodeClass.StatusConditions().SetFalse(v1beta1.ConditionTypeNodeCompartment,
			v1beta1.ConditionNodeCompartmentNotReadyReason, utils.PrettyString(err.Error(), 1024))

		// TODO - classify error and decide whether to retry, we don't return error so as not
		// to disturb other reconciliation.
		return reconcile.Result{
			RequeueAfter: 5 * time.Minute,
		}, nil
	}

	log.FromContext(ctx).Info("node compartment resolved")
	nodeClass.StatusConditions().SetTrue(v1beta1.ConditionTypeNodeCompartment)

	return reconcile.Result{}, nil
}

func (n *NodeCompartmentReconciler) getNodeCompartment(nodeClass *v1beta1.OCINodeClass) string {
	compartmentId := n.clusterCompartmentId

	if nodeClass.Spec.NodeCompartmentId != nil {
		compartmentId = *nodeClass.Spec.NodeCompartmentId
	}

	return compartmentId
}

// nolint:unused
func (n *NodeCompartmentReconciler) ResolvedCompartmentProviderFunc() func(*v1beta1.OCINodeClass) (string, error) {
	return func(nodeClass *v1beta1.OCINodeClass) (string, error) {
		if !nodeClass.StatusConditions().Get(v1beta1.ConditionTypeNodeCompartment).IsTrue() {
			return "", errors.New("nodeCompartmentResolveFailure")
		}

		return n.getNodeCompartment(nodeClass), nil
	}
}
