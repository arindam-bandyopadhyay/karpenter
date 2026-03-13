/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package clusterplacementgroup

import (
	ocicpg "github.com/oracle/oci-go-sdk/v65/clusterplacementgroups"
)

type ResolveResult struct {
	Ocid          string
	Name          string
	Ad            string
	CompartmentId string
}

func fromClusterPlacementGroup(c *ocicpg.ClusterPlacementGroup) ResolveResult {
	return ResolveResult{
		Ocid:          *c.Id,
		Name:          *c.Name,
		Ad:            *c.AvailabilityDomain,
		CompartmentId: *c.CompartmentId,
	}
}

func fromClusterPlacementGroupSummary(c *ocicpg.ClusterPlacementGroupSummary) ResolveResult {
	return ResolveResult{
		Ocid:          *c.Id,
		Name:          *c.Name,
		Ad:            *c.AvailabilityDomain,
		CompartmentId: *c.CompartmentId,
	}
}
