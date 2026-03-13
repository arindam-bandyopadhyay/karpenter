/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRegion(t *testing.T) {
	origFn := getInstanceMetadata
	defer func() { getInstanceMetadata = origFn }()

	// Happy-path: IMDS returns region
	getInstanceMetadata = func(_, _ string) (int, []byte, error) {
		return 200, []byte(`{"region":"us-phoenix-1"}`), nil
	}
	r, err := GetRegion()
	require.NoError(t, err)
	require.Equal(t, "us-phoenix-1", r)

	// IMDS returns invalid JSON → expect error
	getInstanceMetadata = func(_, _ string) (int, []byte, error) {
		return 200, []byte(`invalid json`), nil
	}
	_, err = GetRegion()
	require.Error(t, err)

	// IMDS returns JSON without region field → expect ErrRegionNotFound
	getInstanceMetadata = func(_, _ string) (int, []byte, error) {
		return 200, []byte(`{"other":"value"}`), nil
	}
	_, err = GetRegion()
	require.Equal(t, ErrRegionNotFound, err)

	// IMDS failure → fallback to env var
	getInstanceMetadata = func(_, _ string) (int, []byte, error) { return 500, nil, nil }
	require.NoError(t, os.Setenv("OCI_REGION", "us-ashburn-1"))
	r, err = GetRegion()
	require.NoError(t, err)
	require.Equal(t, "us-ashburn-1", r)
	require.NoError(t, os.Unsetenv("OCI_REGION"))

	// Both IMDS & env fail → expect error
	getInstanceMetadata = func(_, _ string) (int, []byte, error) { return 500, nil, nil }
	_, err = GetRegion()
	require.Error(t, err)
}
