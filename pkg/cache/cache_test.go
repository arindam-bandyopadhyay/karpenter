/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package cache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestNewGetOrLoadCache(t *testing.T) {
	defaultExp := 5 * time.Minute
	cleanup := 10 * time.Minute
	c := NewGetOrLoadCache[string](defaultExp, cleanup)
	assert.NotNil(t, c)
	assert.NotNil(t, c.cache)
}

func TestNewDefaultGetOrLoadCache(t *testing.T) {
	c := NewDefaultGetOrLoadCache[string]()
	assert.NotNil(t, c)
	assert.NotNil(t, c.cache)
}

func TestGetOrLoad_HappyPath(t *testing.T) {
	c := NewDefaultGetOrLoadCache[string]()
	ctx := context.Background()
	key := "test-key"
	expected := "loaded-value"

	loader := func(ctx context.Context, k string) (string, error) {
		assert.Equal(t, key, k)
		return expected, nil
	}

	val, err := c.GetOrLoad(ctx, key, loader)
	assert.NoError(t, err)
	assert.Equal(t, expected, val)

	// Second call should hit cache
	val2, err := c.GetOrLoad(ctx, key, loader)
	assert.NoError(t, err)
	assert.Equal(t, expected, val2)
}

func TestGetOrLoad_LoaderError(t *testing.T) {
	c := NewDefaultGetOrLoadCache[string]()
	ctx := context.Background()
	key := "error-key"
	expectedErr := errors.New("load failed")

	loader := func(ctx context.Context, k string) (string, error) {
		return "", expectedErr
	}

	val, err := c.GetOrLoad(ctx, key, loader)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, "", val)

	// Second call should try to load again (since error not cached)
	val2, err := c.GetOrLoad(ctx, key, loader)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, "", val2)
}

func TestGetOrLoad_CacheHit(t *testing.T) {
	c := NewDefaultGetOrLoadCache[string]()
	ctx := context.Background()
	key := "hit-key"
	expected := "cached-value"

	// Manually set cache (for test purposes)
	c.cache.Set(key, expected, cache.DefaultExpiration)

	loader := func(ctx context.Context, k string) (string, error) {
		t.Fatal("loader should not be called")
		return "", nil
	}

	val, err := c.GetOrLoad(ctx, key, loader)
	assert.NoError(t, err)
	assert.Equal(t, expected, val)
}

func TestGetOrLoad_Concurrent(t *testing.T) {
	c := NewDefaultGetOrLoadCache[string]()
	ctx := context.Background()
	key := "concurrent-key"
	expected := "concurrent-value"

	callCount := 0
	var mu sync.Mutex

	loader := func(_ context.Context, _ string) (string, error) { //nolint:unparam
		mu.Lock()
		callCount++
		mu.Unlock()
		time.Sleep(10 * time.Millisecond) // simulate work
		return expected, nil
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	results := make([]string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			val, err := c.GetOrLoad(ctx, key, loader)
			assert.NoError(t, err)
			results[idx] = val
		}(i)
	}

	wg.Wait()

	assert.Equal(t, 1, callCount, "loader should be called only once")
	for _, res := range results {
		assert.Equal(t, expected, res)
	}
}

func TestEvict(t *testing.T) {
	c := NewDefaultGetOrLoadCache[string]()
	ctx := context.Background()
	key := "evict-key"
	expected := "to-evict"

	// Set in cache
	c.cache.Set(key, expected, cache.DefaultExpiration)

	// Verify hit
	val, err := c.GetOrLoad(ctx, key, func(ctx context.Context, k string) (string, error) {
		t.Fatal("should hit cache")
		return "", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, expected, val)

	// Evict
	c.Evict(ctx, key)

	// Next call should miss cache
	loaderCalled := false
	val2, err := c.GetOrLoad(ctx, key, func(ctx context.Context, k string) (string, error) {
		loaderCalled = true
		return "new-value", nil
	})
	assert.NoError(t, err)
	assert.True(t, loaderCalled)
	assert.Equal(t, "new-value", val2)
}

func TestMakeCompositeKey(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected string
	}{
		{"empty", []string{}, ""},
		{"single", []string{"a"}, "a"},
		{"multiple", []string{"a", "b", "c"}, "a|b|c"},
		{"with empty", []string{"a", "", "c"}, "a||c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MakeCompositeKey(tt.values...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
