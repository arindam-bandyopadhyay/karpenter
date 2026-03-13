/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestForEachAndMapNoIndex(t *testing.T) {
	src := []int{1, 2, 3}

	sum := 0
	ForEachNoIndex(src, func(i int) { sum += i })
	require.Equal(t, 6, sum)

	strs := MapNoIndex(src, func(i int) string { return fmt.Sprintf("n%d", i) })
	require.Equal(t, []string{"n1", "n2", "n3"}, strs)
}
