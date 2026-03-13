/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

func RefreshAtInterval(ctx context.Context,
	initRefresh bool,
	startAsync <-chan struct{},
	interval time.Duration,
	refreshFunc func(context.Context) error) func() {

	startTime := time.Now()
	return func() {
		if initRefresh {
			invokeFuncAndLogError(ctx, refreshFunc)
		}

		select {
		case <-ctx.Done():
			return
		case <-startAsync:
		}

		if time.Since(startTime) > interval {
			invokeFuncAndLogError(ctx, refreshFunc)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				invokeFuncAndLogError(ctx, refreshFunc)
			}
		}
	}
}

func invokeFuncAndLogError(ctx context.Context, f func(context.Context) error) {
	if err := f(ctx); err != nil {
		log.FromContext(ctx).Error(err, "failed to invoke func")
	}
}
