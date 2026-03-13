/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package placement

type Proposal struct {
	Ad                      string
	Fd                      *string
	CapacityReservationId   *string
	ClusterPlacementGroupId *string
	ComputeClusterId        *string
	tempIdentifier          string
}
