/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import (
	ociv1beta1 "github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
)

const (
	DefaultMaxFlannelClusterPod = 110
)

func GetKubeletMaxPods(kubeletCfg *ociv1beta1.KubeletConfiguration,
	secondaryVnicConfigs []*ociv1beta1.SecondaryVnicConfig,
	defaultSecondaryVnicIpCount int) int {
	return GetKubeletMaxPodsWithDefault(kubeletCfg,
		GetDefaultMaxPods(secondaryVnicConfigs, defaultSecondaryVnicIpCount))
}

func GetDefaultMaxPods(secondaryVnicConfigs []*ociv1beta1.SecondaryVnicConfig,
	defaultSecondaryVnicIpCount int) int {
	maxPods := DefaultMaxFlannelClusterPod // default for flannel cluster
	if len(secondaryVnicConfigs) > 0 {
		allIps := 0
		for _, secondaryVnicConfig := range secondaryVnicConfigs {
			if secondaryVnicConfig.IpCount != nil {
				allIps += *secondaryVnicConfig.IpCount
			} else {
				allIps += defaultSecondaryVnicIpCount
			}
		}
		maxPods = allIps
	}
	return maxPods
}

func GetKubeletMaxPodsWithDefault(kubeletCfg *ociv1beta1.KubeletConfiguration, maxPodsInput int) int {
	maxPods := maxPodsInput
	if kubeletCfg != nil {
		if kubeletCfg.MaxPods != nil && *kubeletCfg.MaxPods != 0 {
			maxPods = min(maxPods, int(*kubeletCfg.MaxPods))
		}
	}

	return maxPods
}
