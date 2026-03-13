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
	"time"

	"github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/capacityreservation"
	"github.com/oracle/karpenter-provider-oci/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const CapacityReservationInDifferentCompartment = "capacity reservation(s) is in" +
	" a different compartment from node compartment"

type CapacityReservationReconciler struct {
	capacityReservationProvider capacityreservation.Provider
}

func (c *CapacityReservationReconciler) Reconcile(ctx context.Context,
	nodeClass *v1beta1.OCINodeClass) (reconcile.Result, error) {
	if len(nodeClass.Spec.CapacityReservationConfigs) > 0 {
		results, err := c.capacityReservationProvider.
			ResolveCapacityReservations(ctx, nodeClass.Spec.CapacityReservationConfigs)
		nodeClass.Status.CapacityReservations = toCapacityReservations(results)

		if err != nil {
			log.FromContext(ctx).Error(err, "failed to resolve capacity reservations")
			nodeClass.StatusConditions().SetFalse(v1beta1.ConditionTypeCapacityReservation,
				v1beta1.ConditionCapacityReservationNotReadyReason, utils.PrettyString(err.Error(), 1024))

			// TODO - classify error and decide whether to retry, we don't return error so as not
			// to disturb other reconciliation.
			return reconcile.Result{
				RequeueAfter: 5 * time.Minute,
			}, nil
		}
		log.FromContext(ctx).Info("capacity reservation resolved")
		nodeClass.StatusConditions().SetTrue(v1beta1.ConditionTypeCapacityReservation)
		return reconcile.Result{}, nil
	} else {
		nodeClass.Status.CapacityReservations = nil
		err := nodeClass.StatusConditions().Clear(v1beta1.ConditionTypeCapacityReservation)

		if err != nil {
			log.FromContext(ctx).Error(err, "failed to clear capacity reservation condition")

			return reconcile.Result{
				RequeueAfter: 5 * time.Minute,
			}, nil
		}
	}

	return reconcile.Result{}, nil
}

func toCapacityReservations(results []capacityreservation.ResolveResult) []v1beta1.CapacityReservation {
	return utils.MapNoIndex(results, func(capRes capacityreservation.ResolveResult) v1beta1.CapacityReservation {
		return v1beta1.CapacityReservation{
			CapacityReservationId: capRes.Ocid,
			DisplayName:           capRes.Name,
			AvailabilityDomain:    capRes.Ad,
		}
	})
}
