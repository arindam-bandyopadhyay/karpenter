// Karpenter Provider OCI
//
// Copyright (c) 2026 Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

package network

import (
	"testing"
)

func TestIsIPv6SingleStack(t *testing.T) {
	tests := []struct {
		name       string
		ipFamilies []IpFamily
		want       bool
	}{
		{
			name:       "Single stack IPv6",
			ipFamilies: []IpFamily{IPv6},
			want:       true,
		},
		{
			name:       "Single stack IPv4",
			ipFamilies: []IpFamily{IPv4},
			want:       false,
		},
		{
			name:       "Dual stack IPv4+IPv6",
			ipFamilies: []IpFamily{IPv4, IPv6},
			want:       false,
		},
		{
			name:       "Empty ipFamilies",
			ipFamilies: []IpFamily{},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsIPv6SingleStack(tt.ipFamilies)
			if got != tt.want {
				t.Errorf("IsIPv6SingleStack() = %v, want %v, ipFamilies = %v", got, tt.want, tt.ipFamilies)
			}
		})
	}
}

func TestGetDefaultSecondaryVnicIPCount(t *testing.T) {
	tests := []struct {
		name       string
		ipFamilies []IpFamily
		want       int
	}{
		{
			name:       "Single stack IPv6",
			ipFamilies: []IpFamily{IPv6},
			want:       DefaultIPv6SecondaryVnicIPCount,
		},
		{
			name:       "Single stack IPv4",
			ipFamilies: []IpFamily{IPv4},
			want:       DefaultSecondaryVnicIPCount,
		},
		{
			name:       "Dual stack IPv4+IPv6",
			ipFamilies: []IpFamily{IPv4, IPv6},
			want:       DefaultSecondaryVnicIPCount,
		},
		{
			name:       "Empty ipFamilies",
			ipFamilies: []IpFamily{},
			want:       DefaultSecondaryVnicIPCount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDefaultSecondaryVnicIPCount(tt.ipFamilies)
			if got != tt.want {
				t.Errorf("GetDefaultSecondaryVnicIPCount() = %v, want %v, ipFamilies = %v", got, tt.want, tt.ipFamilies)
			}
		})
	}
}
