// Karpenter Provider OCI
//
// Copyright (c) 2026 Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPowerOfTwo(t *testing.T) {
	testCases := []struct {
		value  int
		expect bool
	}{
		{0, false},
		{1, true},
		{2, true},
		{3, false},
		{4, true},
		{8, true},
		{16, true},
		{32, true},
		{64, true},
		{128, true},
		{256, true},
		{7, false},
		{25, false},
		{57, false},
		{99, false},
		{130, false},
		{210, false},
	}

	for _, tc := range testCases {
		result := IsPowerOfTwo(tc.value)
		require.Equal(t, tc.expect, result,
			fmt.Sprintf("value %d expect IsPowerOfTwo %v, expected value %v", tc.value, result, tc.expect))
	}
}
