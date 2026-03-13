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

type nodeClassReconciler interface {
	Reconcile(context.Context, *v1beta1.OCINodeClass) (reconcile.Result, error)
}
