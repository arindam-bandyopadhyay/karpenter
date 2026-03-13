/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package npn

import (
	"testing"

	ociv1beta1 "github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	npnv1beta1 "github.com/oracle/karpenter-provider-oci/pkg/npn/apis/v1beta1"
	"github.com/stretchr/testify/require"
)

func TestMapValueStringToInterface(t *testing.T) {
	in := map[string]map[string]string{"ns": {"k": "v"}}
	out := MapValueStringToMapValueInterface(in)
	require.Equal(t, "v", out["ns"]["k"])
}

func TestIpv6PairConversions(t *testing.T) {
	cidr := "2001:db8::/64"
	in := []*ociv1beta1.Ipv6AddressIpv6SubnetCidrPairDetails{{SubnetCidr: cidr}, nil}

	npn := ToNpnIpvAddressCidrPair(in)
	require.Equal(t, []npnv1beta1.Ipv6AddressCidrPair{{Ipv6SubnetCidr: cidr}}, npn)

	oci := ToOciCoreIpvAddressCidrPair(in)
	require.Len(t, oci, 1)
	require.Equal(t, cidr, *oci[0].Ipv6SubnetCidr)
}

func TestNsgHelpers(t *testing.T) {
	ids := []string{"nsg1", "nsg2"}
	out := NsgIdsToNetworkSecurityGroupObjects(ids)
	require.Len(t, out, 2)
	require.Equal(t, "nsg1", *out[0].Id)

	// nil input
	out = NsgIdsToNetworkSecurityGroupObjects(nil)
	require.Nil(t, out)
}

func TestStringArrayToIpv6Pairs(t *testing.T) {
	arr := []string{"c1", "c2"}
	out := StringArrayToIpv6AddressIpv6SubnetCidrPairs(&arr)
	require.Len(t, out, 2)
	require.Equal(t, "c1", out[0].SubnetCidr)
}
