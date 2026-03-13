/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrettySlice(t *testing.T) {
	in := []int{1, 2, 3, 4}
	got := PrettySlice(in, 3)
	require.Equal(t, "1, 2, 3 and 1 other(s)", got)

	require.Equal(t, "", PrettySlice([]int{}, 3))
	require.Equal(t, "1", PrettySlice([]int{1}, 3))
}

func TestPrettyString(t *testing.T) {
	short := PrettyString("hello", 10)
	require.Equal(t, "hello", short)

	long := PrettyString("0123456789abcdef", 10)
	require.Equal(t, "0123456...", long)
}
