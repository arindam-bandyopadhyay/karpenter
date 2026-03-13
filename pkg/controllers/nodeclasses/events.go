/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package nodeclasses

import (
	"fmt"

	ociv1beta1 "github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/utils"

	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/karpenter/pkg/events"
)

func PendingNodeClaimTerminationEvent(nodeClass *ociv1beta1.OCINodeClass, names []string) events.Event {
	return events.Event{
		InvolvedObject: nodeClass,
		Type:           corev1.EventTypeNormal,
		Reason:         "WaitingOnNodeClaimTermination",
		Message:        fmt.Sprintf("Waiting on NodeClaim termination for %s", utils.PrettySlice(names, 5)),
		DedupeValues:   []string{string(nodeClass.UID)},
	}
}
