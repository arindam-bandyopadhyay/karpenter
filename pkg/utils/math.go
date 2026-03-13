// Karpenter Provider OCI
//
// Copyright (c) 2026 Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

package utils

// IsPowerOfTwo checks if the input number is a power of two.
// Returns false if n is less than or equal to zero.
func IsPowerOfTwo(n int) bool {
	if n <= 0 {
		return false
	}
	return (n & (n - 1)) == 0
}
