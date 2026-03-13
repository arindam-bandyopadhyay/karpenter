/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package instancetype

import (
	"github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	ocicore "github.com/oracle/oci-go-sdk/v65/core"
)

func ToLaunchInstanceCpuBaseline(
	utilization v1beta1.BaselineOcpuUtilization) ocicore.LaunchInstanceShapeConfigDetailsBaselineOcpuUtilizationEnum {
	cpuBaseline := ocicore.LaunchInstanceShapeConfigDetailsBaselineOcpuUtilization1
	switch utilization {
	case v1beta1.BASELINE_1_2:
		cpuBaseline = ocicore.LaunchInstanceShapeConfigDetailsBaselineOcpuUtilization2
	case v1beta1.BASELINE_1_8:
		cpuBaseline = ocicore.LaunchInstanceShapeConfigDetailsBaselineOcpuUtilization8
	}

	return cpuBaseline
}

func FromInstanceCpuBaseline(
	utilization ocicore.InstanceShapeConfigBaselineOcpuUtilizationEnum) v1beta1.BaselineOcpuUtilization {
	switch utilization {
	case ocicore.InstanceShapeConfigBaselineOcpuUtilization8:
		return v1beta1.BASELINE_1_8
	case ocicore.InstanceShapeConfigBaselineOcpuUtilization2:
		return v1beta1.BASELINE_1_2
	default:
		return v1beta1.BASELINE_1_1
	}
}
