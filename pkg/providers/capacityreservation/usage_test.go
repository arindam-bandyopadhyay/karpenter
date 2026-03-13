/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package capacityreservation

import (
	"sync"
	"testing"
	"time"

	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestUsage_Concurrency(t *testing.T) {
	usage := newUsage()
	instance := &ocicore.Instance{
		Shape:       lo.ToPtr("VM.Standard1.1"),
		FaultDomain: lo.ToPtr("fd1"),
		ShapeConfig: &ocicore.InstanceShapeConfig{
			Ocpus:       float32Ptr(2.0),
			MemoryInGBs: float32Ptr(15.0),
		},
	}

	const numGoroutines = 100
	const operationsPerGoroutine = 10
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				usage.IncreaseUsage(instance)
				usage.DecreaseUsage(instance)
			}
		}()
	}

	wg.Wait()

	// Should be back to zero since we increased and decreased equally
	uk := usageKey{
		shape: *instance.Shape,
		key:   KeyOfInstance(instance),
	}
	fdUsage := usage.shapeFdUsageMap[uk]["fd1"]
	assert.Equal(t, int64(0), fdUsage)

	// Last commit should be updated
	assert.True(t, usage.lastCommit.After(start))
}
