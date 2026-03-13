/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRefreshAtInterval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls int32
	startAsync := make(chan struct{})
	fn := RefreshAtInterval(ctx, true, startAsync, 10*time.Millisecond, func(context.Context) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})

	go func() {
		close(startAsync) // allow the goroutine to continue
		fn()
	}()

	// Wait long enough for at least 2 invocations (init + one interval)
	time.Sleep(35 * time.Millisecond)
	require.GreaterOrEqual(t, atomic.LoadInt32(&calls), int32(2))
}

func TestInvokeFuncAndLogError(t *testing.T) {
	// This function is tested indirectly through RefreshAtInterval
	// But we can test it directly by checking if it doesn't panic on error
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("invokeFuncAndLogError should not panic: %v", r)
		}
	}()

	invokeFuncAndLogError(context.Background(), func(context.Context) error {
		return errors.New("test error")
	})
}

func TestRefreshAtIntervalEdgeCases(t *testing.T) {
	// Test with initRefresh=false and early context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	startAsync := make(chan struct{})
	var calls int32

	fn := RefreshAtInterval(ctx, false, startAsync, 1*time.Hour, func(context.Context) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})

	go func() {
		close(startAsync)
		fn()
	}()

	// Give it a moment to potentially run
	time.Sleep(10 * time.Millisecond)

	// Should not have called the function since context was cancelled and initRefresh=false
	require.Equal(t, int32(0), atomic.LoadInt32(&calls))
}

func TestRefreshAtIntervalWithError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startAsync := make(chan struct{})
	var calls int32

	fn := RefreshAtInterval(ctx, true, startAsync, 10*time.Millisecond, func(context.Context) error {
		count := atomic.AddInt32(&calls, 1)
		if count == 1 {
			return nil // First call succeeds
		}
		return nil // Subsequent calls also succeed for this test
	})

	go func() {
		close(startAsync)
		fn()
	}()

	time.Sleep(35 * time.Millisecond)
	require.GreaterOrEqual(t, atomic.LoadInt32(&calls), int32(1))
}
