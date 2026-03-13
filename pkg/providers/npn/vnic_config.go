/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package npn

import (
	ociv1beta1 "github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	npnv1beta1 "github.com/oracle/karpenter-provider-oci/pkg/npn/apis/v1beta1"
	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	"github.com/samber/lo"
)

func MapValueStringToMapValueInterface(input map[string]map[string]string) map[string]map[string]interface{} {
	return lo.MapEntries(input,
		func(k string, v map[string]string) (string, map[string]interface{}) {
			return k, lo.MapEntries(v, func(ki string, vi string) (string, interface{}) {
				return ki, vi
			})
		})
}

func ToNpnIpvAddressCidrPair(
	vnicCidrPair []*ociv1beta1.Ipv6AddressIpv6SubnetCidrPairDetails) []npnv1beta1.Ipv6AddressCidrPair {
	var ipv6IpCidrPairs []npnv1beta1.Ipv6AddressCidrPair
	if len(vnicCidrPair) > 0 {
		for _, value := range vnicCidrPair {
			if value != nil {
				ipv6IpCidrPairs = append(ipv6IpCidrPairs, npnv1beta1.Ipv6AddressCidrPair{
					Ipv6SubnetCidr: value.SubnetCidr,
				})
			}
		}
	}
	return ipv6IpCidrPairs
}

func ToOciCoreIpvAddressCidrPair(
	vnicCidrPair []*ociv1beta1.Ipv6AddressIpv6SubnetCidrPairDetails) []ocicore.Ipv6AddressIpv6SubnetCidrPairDetails {
	var ipv6IpCidrPairs []ocicore.Ipv6AddressIpv6SubnetCidrPairDetails
	if len(vnicCidrPair) > 0 {
		for _, value := range vnicCidrPair {
			if value != nil {
				ipv6IpCidrPairs = append(ipv6IpCidrPairs, ocicore.Ipv6AddressIpv6SubnetCidrPairDetails{
					Ipv6SubnetCidr: &value.SubnetCidr,
				})
			}
		}
	}
	return ipv6IpCidrPairs
}

func NsgIdsToNetworkSecurityGroupObjects(networkSecurityGroupIds []string) []*ocicore.NetworkSecurityGroup {
	if networkSecurityGroupIds == nil {
		return nil
	}

	nsgDetails := make([]*ocicore.NetworkSecurityGroup, 0)
	for _, value := range networkSecurityGroupIds {
		pValue := &value
		nsgDetails = append(nsgDetails, &ocicore.NetworkSecurityGroup{
			Id: pValue,
		})
	}

	return nsgDetails
}

func StringArrayToIpv6AddressIpv6SubnetCidrPairs(
	ipv6IpCidrPairs *[]string) []*ociv1beta1.Ipv6AddressIpv6SubnetCidrPairDetails {
	var ipv6IpCidrPairObjects []*ociv1beta1.Ipv6AddressIpv6SubnetCidrPairDetails
	if ipv6IpCidrPairs != nil {
		for _, v := range *ipv6IpCidrPairs {
			ipv6IpCidrPairObjects = append(ipv6IpCidrPairObjects, &ociv1beta1.Ipv6AddressIpv6SubnetCidrPairDetails{
				SubnetCidr: v,
			})
		}
	}
	return ipv6IpCidrPairObjects
}
