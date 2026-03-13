/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package blockstorage

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/oracle/karpenter-provider-oci/pkg/cache"
	"github.com/oracle/karpenter-provider-oci/pkg/fakes"
	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_GetBootVolume(t *testing.T) {
	ctx := context.TODO()
	volId := "ocid1.bootvolume.oc1..abc123"
	happyBootVol := ocicore.BootVolume{
		Id:        &volId,
		SizeInGBs: lo.ToPtr(int64(50)),
	}
	happyResp := ocicore.GetBootVolumeResponse{
		BootVolume: happyBootVol,
	}

	tests := []struct {
		name       string
		arrange    func() DefaultProvider
		input      string
		want       *ocicore.BootVolume
		wantErr    bool
		wantGetCnt int
	}{
		{
			name: "happy path",
			arrange: func() DefaultProvider {
				fake := &fakes.FakeBlockstorage{
					GetBootVolumeResponse: happyResp,
				}
				return DefaultProvider{
					blockStorageClient: fake,
					bootVolumeCache:    cache.NewDefaultGetOrLoadCache[*ocicore.BootVolume](),
				}
			},
			input:      volId,
			want:       &happyBootVol,
			wantErr:    false,
			wantGetCnt: 1,
		},
		{
			name: "error from client",
			arrange: func() DefaultProvider {
				fake := &fakes.FakeBlockstorage{
					GetBootVolumeError: errors.New("not found"),
				}
				return DefaultProvider{
					blockStorageClient: fake,
					bootVolumeCache:    cache.NewDefaultGetOrLoadCache[*ocicore.BootVolume](),
				}
			},
			input:      volId,
			want:       nil,
			wantErr:    true,
			wantGetCnt: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := tt.arrange()
			fake, _ := provider.blockStorageClient.(*fakes.FakeBlockstorage)
			got, err := provider.GetBootVolume(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.want, got)
			}
			assert.Equal(t, tt.wantGetCnt, fake.GetCount.Get())
		})
	}
}

func TestProvider_GetBootVolumeCached(t *testing.T) {
	ctx := context.TODO()
	volId := "ocid1.bootvolume.oc1..cached"
	bootVol := ocicore.BootVolume{
		Id:        &volId,
		SizeInGBs: lo.ToPtr(int64(10)),
	}
	resp := ocicore.GetBootVolumeResponse{
		BootVolume: bootVol,
	}

	t.Run("caches result, only one client call", func(t *testing.T) {
		fake := &fakes.FakeBlockstorage{
			GetBootVolumeResponse: resp,
		}
		provider := &DefaultProvider{
			blockStorageClient: fake,
			bootVolumeCache:    cache.NewDefaultGetOrLoadCache[*ocicore.BootVolume](),
		}
		// First call
		val1, err := provider.GetBootVolumeCached(ctx, volId)
		require.NoError(t, err)
		require.NotNil(t, val1)
		assert.Equal(t, &bootVol, val1)

		// Second call (should be from cache)
		val2, err := provider.GetBootVolumeCached(ctx, volId)
		require.NoError(t, err)
		require.NotNil(t, val2)
		assert.Equal(t, &bootVol, val2)

		// Only one actual GetBootVolume call should have happened
		assert.Equal(t, 1, fake.GetCount.Get())
	})

	t.Run("error is not cached", func(t *testing.T) {
		fake := &fakes.FakeBlockstorage{
			GetBootVolumeError: errors.New("exploded"),
		}
		provider := &DefaultProvider{
			blockStorageClient: fake,
			bootVolumeCache:    cache.NewDefaultGetOrLoadCache[*ocicore.BootVolume](),
		}
		_, err := provider.GetBootVolumeCached(ctx, volId)
		require.Error(t, err)
		assert.Equal(t, 1, fake.GetCount.Get())

		// On retry, error should occur again (not cached)
		_, err = provider.GetBootVolumeCached(ctx, volId)
		require.Error(t, err)
		assert.Equal(t, 2, fake.GetCount.Get())
	})

	t.Run("cache hit after prefill", func(t *testing.T) {
		fake := &fakes.FakeBlockstorage{
			GetBootVolumeResponse: resp,
		}
		c := cache.NewDefaultGetOrLoadCache[*ocicore.BootVolume]()
		// Prefill cache directly
		_, err := c.GetOrLoad(ctx, volId, func(context.Context, string) (*ocicore.BootVolume, error) {
			return &bootVol, nil
		})
		require.NoError(t, err)

		provider := &DefaultProvider{
			blockStorageClient: fake,
			bootVolumeCache:    c,
		}
		// Call should hit cache without calling fake
		val, err := provider.GetBootVolumeCached(ctx, volId)
		require.NoError(t, err)
		assert.Equal(t, &bootVol, val)
		assert.Equal(t, 0, fake.GetCount.Get(), "should not call client on cache hit")
	})

	t.Run("cache eviction", func(t *testing.T) {
		fake := &fakes.FakeBlockstorage{
			GetBootVolumeResponse: resp,
		}
		c := cache.NewDefaultGetOrLoadCache[*ocicore.BootVolume]()
		provider := &DefaultProvider{
			blockStorageClient: fake,
			bootVolumeCache:    c,
		}
		// First call loads and caches
		_, err := provider.GetBootVolumeCached(ctx, volId)
		require.NoError(t, err)
		assert.Equal(t, 1, fake.GetCount.Get())

		// Evict manually
		c.Evict(ctx, volId)

		// Second call should load again
		_, err = provider.GetBootVolumeCached(ctx, volId)
		require.NoError(t, err)
		assert.Equal(t, 2, fake.GetCount.Get(), "should call client after eviction")
	})

	t.Run("concurrent lookup single load", func(t *testing.T) {
		fake := &fakes.FakeBlockstorage{
			GetBootVolumeResponse: resp,
		}
		provider := &DefaultProvider{
			blockStorageClient: fake,
			bootVolumeCache:    cache.NewDefaultGetOrLoadCache[*ocicore.BootVolume](),
		}

		var wg sync.WaitGroup
		const numGoroutines = 20
		results := make([]*ocicore.BootVolume, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				val, err := provider.GetBootVolumeCached(ctx, volId)
				require.NoError(t, err)
				results[idx] = val
			}(i)
		}
		wg.Wait()

		// All results should be the same object
		first := results[0]
		for _, r := range results[1:] {
			assert.Same(t, first, r, "all should point to same cached instance")
		}
		assert.Equal(t, 1, fake.GetCount.Get(), "loader should execute only once")
	})

	t.Run("error then success caching", func(t *testing.T) {
		callCount := 0
		fake := &fakes.FakeBlockstorage{
			OnGetBootVolume: func(context.Context, ocicore.GetBootVolumeRequest) (ocicore.GetBootVolumeResponse, error) {
				callCount++
				if callCount == 1 {
					return ocicore.GetBootVolumeResponse{}, errors.New("temporary error")
				}
				return resp, nil
			},
		}
		provider := &DefaultProvider{
			blockStorageClient: fake,
			bootVolumeCache:    cache.NewDefaultGetOrLoadCache[*ocicore.BootVolume](),
		}

		// First call fails, not cached
		_, err := provider.GetBootVolumeCached(ctx, volId)
		require.Error(t, err)
		assert.Equal(t, 1, callCount)

		// Second call succeeds, caches
		val, err := provider.GetBootVolumeCached(ctx, volId)
		require.NoError(t, err)
		assert.Equal(t, &bootVol, val)
		assert.Equal(t, 2, callCount)

		// Third call from cache
		val2, err := provider.GetBootVolumeCached(ctx, volId)
		require.NoError(t, err)
		assert.Equal(t, &bootVol, val2)
		assert.Equal(t, 2, callCount, "no additional client call")
	})
}

func TestProvider_NewProvider(t *testing.T) {
	ctx := context.TODO()
	fakeClient := &fakes.FakeBlockstorage{}

	p, err := NewProvider(ctx, fakeClient)
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, fakeClient, p.blockStorageClient)
	assert.NotNil(t, p.bootVolumeCache)
}
