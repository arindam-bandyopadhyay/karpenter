/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package fakes

import (
	"context"

	ocicore "github.com/oracle/oci-go-sdk/v65/core"
)

// FakeBlockstorage implements oci.BlockstorageClient for tests
type FakeBlockstorage struct {
	GetBootVolumeResponse ocicore.GetBootVolumeResponse
	GetBootVolumeError    error
	GetCount              Counter
	OnGetBootVolume       func(context.Context, ocicore.GetBootVolumeRequest) (ocicore.GetBootVolumeResponse, error)
}

func (f *FakeBlockstorage) GetBootVolume(ctx context.Context, req ocicore.GetBootVolumeRequest) (
	ocicore.GetBootVolumeResponse, error) {
	f.GetCount.Inc()
	if f.OnGetBootVolume != nil {
		return f.OnGetBootVolume(ctx, req)
	}
	if f.GetBootVolumeError != nil {
		return ocicore.GetBootVolumeResponse{}, f.GetBootVolumeError
	}
	return f.GetBootVolumeResponse, nil
}
