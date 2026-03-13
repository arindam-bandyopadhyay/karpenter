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
	"github.com/oracle/karpenter-provider-oci/pkg/providers/network"
	"github.com/oracle/karpenter-provider-oci/pkg/utils"
	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	"github.com/samber/lo"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type NetworkReconciler struct {
	networkProvider network.Provider
}

func (n *NetworkReconciler) Reconcile(ctx context.Context, nodeClass *v1beta1.OCINodeClass) (reconcile.Result, error) {
	// resolve image, if there is an image ocid, use it directly
	result, err := n.networkProvider.ResolveNetworkConfig(ctx, nodeClass.Spec.NetworkConfig)

	return updateNetworkReconcileResult(ctx, nodeClass, result, err)
}

func updateNetworkReconcileResult(ctx context.Context, class *v1beta1.OCINodeClass,
	result *network.NetworkResolveResult, err error) (reconcile.Result, error) {
	if err == nil {
		log.FromContext(ctx).Info("network resolved")
		class.Status.Network = toNodeClassNetwork(result)

		class.StatusConditions().SetTrue(v1beta1.ConditionTypeNetworkReady)
		return reconcile.Result{}, nil
	} else {
		log.FromContext(ctx).Error(err, "failed to resolve network")
		class.Status.Network = toNodeClassNetwork(result)
		class.StatusConditions().SetFalse(v1beta1.ConditionTypeNetworkReady,
			v1beta1.ConditionNetworkNotReadyReason, utils.PrettyString(err.Error(), 1024))

		// TODO - classify error and decide whether to retry, we don't return error so as not
		// to disturb other reconciliation.
		return reconcile.Result{
			RequeueAfter: 5 * time.Minute,
		}, nil
	}
}

func toNodeClassNetwork(input *network.NetworkResolveResult) *v1beta1.Network {
	n := &v1beta1.Network{}

	if input.PrimaryVnicSubnet != nil {
		n.PrimaryVnic = toVnic(*input.PrimaryVnicSubnet)
	}

	n.SecondaryVnics = lo.Map(input.OtherVnicSubnets, func(item *network.SubnetAndNsgs, _ int) *v1beta1.Vnic {
		return toVnic(*item)
	})

	return n
}

func toVnic(input network.SubnetAndNsgs) *v1beta1.Vnic {
	var subnet v1beta1.Subnet
	if input.Subnet != nil {
		subnet.SubnetId = *input.Subnet.Id
		subnet.DisplayName = *input.Subnet.DisplayName
	}

	return &v1beta1.Vnic{
		Subnet: subnet,
		NetworkSecurityGroups: lo.Map(input.NetworkSecurityGroups,
			func(item *ocicore.NetworkSecurityGroup, _ int) v1beta1.NetworkSecurityGroup {
				var id, name string
				if item != nil {
					id = *item.Id
					name = *item.DisplayName
				}
				return v1beta1.NetworkSecurityGroup{
					NetworkSecurityGroupId: id,
					DisplayName:            name,
				}
			}),
	}
}
