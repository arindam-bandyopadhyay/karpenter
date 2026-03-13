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
	"github.com/oracle/karpenter-provider-oci/pkg/providers/kms"
	"github.com/oracle/karpenter-provider-oci/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type KmsKeyReconciler struct {
	kmsKeyProvider kms.Provider
}

func (k *KmsKeyReconciler) Reconcile(ctx context.Context, nodeClass *v1beta1.OCINodeClass) (reconcile.Result, error) {
	var kmsKeyReses []*kms.KmsKeyResolveResult
	if nodeClass.Spec.VolumeConfig.BootVolumeConfig.KmsKeyConfig != nil {
		kmsKeyRes, err := k.kmsKeyProvider.ResolveKmsKeyConfig(ctx,
			nodeClass.Spec.VolumeConfig.BootVolumeConfig.KmsKeyConfig)

		if err == nil {
			kmsKeyReses = append(kmsKeyReses, kmsKeyRes)
		}

		return updateKmsKeyReconcileResult(ctx, nodeClass, kmsKeyReses, err)
	} else {
		nodeClass.Status.Volume.KmsKeys = nil
		err := nodeClass.StatusConditions().Clear(v1beta1.ConditionTypeKmsKeyReady)

		if err != nil {
			log.FromContext(ctx).Error(err, "failed to clear kms key condition")

			return reconcile.Result{
				RequeueAfter: 5 * time.Minute,
			}, nil
		}
	}

	return reconcile.Result{}, nil
}

func updateKmsKeyReconcileResult(ctx context.Context, class *v1beta1.OCINodeClass,
	kmsKeyReses []*kms.KmsKeyResolveResult, err error) (reconcile.Result, error) {
	class.Status.Volume.KmsKeys = nil
	if err == nil {
		log.FromContext(ctx).Info("kms key resolved")

		for _, kmsKeyRes := range kmsKeyReses {
			class.Status.Volume.KmsKeys = append(class.Status.Volume.KmsKeys, &v1beta1.KmsKey{
				KmsKeyId:    kmsKeyRes.Ocid,
				DisplayName: kmsKeyRes.DisplayName,
			})
		}

		class.StatusConditions().SetTrue(v1beta1.ConditionTypeKmsKeyReady)
	} else {
		log.FromContext(ctx).Error(err, "failed to resolve kms key")
		class.StatusConditions().SetFalse(v1beta1.ConditionTypeKmsKeyReady,
			v1beta1.ConditionKmsKeyNotReadyReason, utils.PrettyString(err.Error(), 1024))

		// TODO - classify error and decide whether to retry, we don't return error so as not to
		// disturb other reconciliation.
		return reconcile.Result{
			RequeueAfter: 5 * time.Minute,
		}, nil
	}

	return reconcile.Result{}, nil
}
