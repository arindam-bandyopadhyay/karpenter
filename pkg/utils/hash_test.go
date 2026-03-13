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

type sample struct {
	A string
	B int
}

func TestHashFor(t *testing.T) {
	obj := sample{"hello", 42}

	h1, err := HashFor(obj)
	require.NoError(t, err)

	// Deterministic
	h2, err := HashFor(obj)
	require.NoError(t, err)
	require.Equal(t, h1, h2, "hash must be deterministic")

	// Different object -> different hash
	other := sample{"bye", 24}
	h3, err := HashFor(other)
	require.NoError(t, err)
	require.NotEqual(t, h1, h3, "different objects must hash differently")

	// Nil object - json.Marshal(nil) works, returns "null"
	hNil, err := HashFor(nil)
	require.NoError(t, err)
	require.Equal(t, Digest([]byte("null")), hNil)

	// Unmarshalable object (channel)
	ch := make(chan int)
	_, err = HashFor(ch)
	require.Error(t, err)
}

func TestHashForMultiObjects(t *testing.T) {
	arr := []interface{}{sample{"a", 1}, sample{"b", 2}}

	h1, err := HashForMultiObjects(arr)
	require.NoError(t, err)

	// same order, same hash
	h2, err := HashForMultiObjects(arr)
	require.NoError(t, err)
	require.Equal(t, h1, h2)

	// with nil in slice should produce same hash
	arrWithNil := append(arr, nil)
	h3, err := HashForMultiObjects(arrWithNil)
	require.NoError(t, err)
	require.Equal(t, h1, h3)

	// with unmarshalable object
	ch := make(chan int)
	arrWithChan := []interface{}{sample{"a", 1}, ch}
	_, err = HashForMultiObjects(arrWithChan)
	require.Error(t, err)
}

func TestHashForMarshalError(t *testing.T) {
	// Test that HashFor returns an error when json.Marshal fails
	// This happens with types that cannot be marshaled (like functions, channels, etc.)
	ch := make(chan int)
	_, err := HashFor(ch)
	require.Error(t, err)

	// Test with a function
	fn := func() {}
	_, err = HashFor(fn)
	require.Error(t, err)
}
