/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package nodeclasses

import (
	"context"

	"github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Hash struct {
}

func (h *Hash) Reconcile(ctx context.Context, nodeClass *v1beta1.OCINodeClass) (reconcile.Result, error) {
	// if nodeClass.StatusConditions().Root().IsTrue() {
	// var infs []interface{}
	// infs = append(infs, nodeClass.Spec)
	// infs = append(infs, nodeClass.Status.Volume)
	// infs = append(infs, nodeClass.Status.Network)
	//
	// hash, err := utils.HashForMultiObjects(infs)
	// if err != nil {
	//	return reconcile.Result{
	//		RequeueAfter: 5 * time.Minute,
	//	}, nil
	// }
	//
	// if nodeClass.Annotations == nil {
	//	nodeClass.Annotations = make(map[string]string)
	// }
	//
	// nodeClass.Annotations[v1beta1.NodeClassHash] = hash
	// }

	return reconcile.Result{}, nil
}
