/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package computecluster

import (
	ocicore "github.com/oracle/oci-go-sdk/v65/core"
)

type ResolveResult struct {
	Ocid          string
	Name          string
	Ad            string
	CompartmentId string
}

func fromComputeCluster(c *ocicore.ComputeCluster) *ResolveResult {
	return &ResolveResult{
		Ocid:          *c.Id,
		Name:          *c.DisplayName,
		Ad:            *c.AvailabilityDomain,
		CompartmentId: *c.CompartmentId,
	}
}

func fromComputeClusterSummary(c *ocicore.ComputeClusterSummary) *ResolveResult {
	return &ResolveResult{
		Ocid:          *c.Id,
		Name:          *c.DisplayName,
		Ad:            *c.AvailabilityDomain,
		CompartmentId: *c.CompartmentId,
	}
}
