/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import "github.com/samber/lo"

func ForEachNoIndex[T any](collection []T, iteratee func(item T)) {
	lo.ForEach(collection, func(item T, _ int) {
		iteratee(item)
	})
}

func MapNoIndex[T any, R any](collection []T, iteratee func(item T) R) []R {
	return lo.Map(collection, func(item T, _ int) R {
		return iteratee(item)
	})
}
