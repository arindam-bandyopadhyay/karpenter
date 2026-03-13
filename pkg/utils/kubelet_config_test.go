/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import (
	"testing"

	ociv1beta1 "github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetKubeletMaxPodsWithoutVnic(t *testing.T) {
	// default
	require.Equal(t, 110, GetKubeletMaxPods(nil, nil, 32))

	val := int32(50)
	cfg := &ociv1beta1.KubeletConfiguration{MaxPods: &val}
	require.Equal(t, 50, GetKubeletMaxPods(cfg, nil, 32))
	require.Equal(t, 40, GetKubeletMaxPodsWithDefault(cfg, 40)) // min of 50 and 40
}

func TestGetKubeletMaxPodsWithVnics(t *testing.T) {
	testCases := []struct {
		maxPods  *int32
		ipCounts []*int
		expect   int
	}{
		{
			maxPods:  nil,
			ipCounts: nil,
			expect:   110,
		},
		{
			maxPods:  lo.ToPtr(int32(0)),
			ipCounts: nil,
			expect:   110,
		},
		{
			maxPods:  nil,
			ipCounts: []*int{lo.ToPtr(16)},
			expect:   16,
		},
		{
			maxPods:  lo.ToPtr(int32(50)),
			ipCounts: []*int{lo.ToPtr(16), lo.ToPtr(10)},
			expect:   26,
		},
		{
			maxPods:  lo.ToPtr(int32(20)),
			ipCounts: []*int{lo.ToPtr(16), lo.ToPtr(10)},
			expect:   20,
		},
	}

	testSubnetId := "ocid1.subnet.oc1..aaa"
	for _, tc := range testCases {
		secondaryVnicConfigs := make([]*ociv1beta1.SecondaryVnicConfig, 0)
		for _, ipCount := range tc.ipCounts {
			secondaryVnicConfigs = append(secondaryVnicConfigs, &ociv1beta1.SecondaryVnicConfig{
				SimpleVnicConfig: ociv1beta1.SimpleVnicConfig{
					SubnetAndNsgConfig: &ociv1beta1.SubnetAndNsgConfig{
						SubnetConfig: &ociv1beta1.SubnetConfig{SubnetId: &testSubnetId},
					},
				},
				IpCount: ipCount,
			})
		}

		result := GetKubeletMaxPods(&ociv1beta1.KubeletConfiguration{
			MaxPods: tc.maxPods,
		}, secondaryVnicConfigs, 32)

		assert.Equal(t, tc.expect, result)
	}
}
