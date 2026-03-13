/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package fakes

import "sync"

// Counter provides thread-safe increment/decrement operations for test metrics
type Counter struct {
	n  int
	mu sync.Mutex
}

func (c *Counter) Inc() {
	c.mu.Lock()
	c.n++
	c.mu.Unlock()
}

func (c *Counter) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

// Pagination helpers for simulating OCI API pagination
func pageIndex(p *string) int {
	if p == nil {
		return 0
	}
	switch *p {
	case "p2":
		return 1
	case "p3":
		return 2
	default:
		return 0
	}
}

func nextPageToken(i int) string {
	switch i {
	case 1:
		return "p2"
	case 2:
		return "p3"
	default:
		return ""
	}
}
