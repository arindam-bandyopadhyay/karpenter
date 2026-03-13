/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package blockstorage

import (
	"context"

	"github.com/oracle/karpenter-provider-oci/pkg/cache"
	"github.com/oracle/karpenter-provider-oci/pkg/oci"
	ocicore "github.com/oracle/oci-go-sdk/v65/core"
)

type Provider interface {
	GetBootVolume(ctx context.Context, bootVolumeOcid string) (*ocicore.BootVolume, error)
	GetBootVolumeCached(ctx context.Context, bootVolumeOcid string) (*ocicore.BootVolume, error)
}

type DefaultProvider struct {
	blockStorageClient oci.BlockStorageClient
	bootVolumeCache    *cache.GetOrLoadCache[*ocicore.BootVolume]
}

func NewProvider(ctx context.Context, blockStorageClient oci.BlockStorageClient) (*DefaultProvider, error) {
	p := &DefaultProvider{
		blockStorageClient: blockStorageClient,
		bootVolumeCache:    cache.NewDefaultGetOrLoadCache[*ocicore.BootVolume](),
	}

	return p, nil
}

func (p DefaultProvider) GetBootVolume(ctx context.Context, bootVolumeOcid string) (*ocicore.BootVolume, error) {
	p.bootVolumeCache.Evict(ctx, bootVolumeOcid)
	return p.GetBootVolumeCached(ctx, bootVolumeOcid)
}

func (p DefaultProvider) getBootVolumeImpl(ctx context.Context, bootVolumeOcid string) (*ocicore.BootVolume, error) {
	getResp, err := p.blockStorageClient.GetBootVolume(ctx, ocicore.GetBootVolumeRequest{
		BootVolumeId: &bootVolumeOcid,
	})
	if err != nil {
		return nil, err
	}
	return &getResp.BootVolume, nil
}

func (p *DefaultProvider) GetBootVolumeCached(ctx context.Context, bootVolumeOcid string) (*ocicore.BootVolume, error) {
	return p.bootVolumeCache.GetOrLoad(ctx, bootVolumeOcid,
		func(ctx context.Context, key string) (*ocicore.BootVolume, error) {
			return p.getBootVolumeImpl(ctx, key)

		})
}
