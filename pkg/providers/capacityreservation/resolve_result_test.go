/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package capacityreservation

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestResolveResult_AvailabilityForShape(t *testing.T) {
	tests := []struct {
		name          string
		resolveResult ResolveResult
		shape         string
		ocpu          *float32
		memoryInGbs   *float32
		expected      map[string]ShapeAvailability
	}{
		{
			name: "single shape fd match",
			resolveResult: ResolveResult{
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: "VM.Standard1.1", Total: 10, Used: 5},
				},
			},
			shape: "VM.Standard1.1",
			expected: map[string]ShapeAvailability{
				"fd1": {FaultDomain: nil, Shape: "VM.Standard1.1", Total: 10, Used: 5},
			},
		},
		{
			name: "merge same shape different fd",
			resolveResult: ResolveResult{
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: "VM.Standard1.1", Total: 10, Used: 5},
					{FaultDomain: lo.ToPtr("fd2"), Shape: "VM.Standard1.1", Total: 20, Used: 10},
				},
			},
			shape: "VM.Standard1.1",
			expected: map[string]ShapeAvailability{
				"fd1": {FaultDomain: nil, Shape: "VM.Standard1.1", Total: 10, Used: 5},
				"fd2": {FaultDomain: nil, Shape: "VM.Standard1.1", Total: 20, Used: 10},
			},
		},
		{
			name: "no match",
			resolveResult: ResolveResult{
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: "VM.Standard1.1", Total: 10, Used: 5},
				},
			},
			shape:    "VM.Standard2.1",
			expected: map[string]ShapeAvailability{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.resolveResult.AvailabilityForShape(tt.shape, tt.ocpu, tt.memoryInGbs, false)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveResult_adjustCapacityAgainstUsage(t *testing.T) {
	ocpu := float32(1)
	memoryInGbs := float32(16)
	nonFlexibleShape := "VM.Standard1.1"
	flexibleShape := "VM.Standard.E5.Flex"

	nonFlexibleShapeUsageKey := usageKey{
		shape: nonFlexibleShape,
		key:   KeyOfShape(nonFlexibleShape, nil, nil),
	}
	flexibleShapeUsageKey := usageKey{
		shape: flexibleShape,
		key:   KeyOfShape(flexibleShape, &ocpu, &memoryInGbs),
	}

	tests := []struct {
		name          string
		resolveResult ResolveResult
		usage         *Usage
		expected      ResolveResult
	}{
		{
			name: "fd level adjustment",
			resolveResult: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: nonFlexibleShape, Total: 10, Used: 5},
				},
			},
			usage: &Usage{
				shapeFdUsageMap: map[usageKey]map[string]int64{
					nonFlexibleShapeUsageKey: {"fd1": 2}, // KeyOfShape
				},
			},
			expected: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: nonFlexibleShape, Total: 10, Used: 7}, // 5 + 2
				},
			},
		},
		{
			name: "fd level adjustment for non-flexible shape w/ ocpu + memory",
			resolveResult: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: nonFlexibleShape,
						Ocpu: &ocpu, MemoryInGbs: &memoryInGbs, Total: 10, Used: 5},
				},
			},
			usage: &Usage{
				shapeFdUsageMap: map[usageKey]map[string]int64{
					nonFlexibleShapeUsageKey: {"fd1": 2}, // KeyOfShape
				},
			},
			expected: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: nonFlexibleShape,
						Ocpu: &ocpu, MemoryInGbs: &memoryInGbs, Total: 10, Used: 7}, // 5 + 2
				},
			},
		},
		{
			name: "fd level adjustment for flexible shape",
			resolveResult: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: flexibleShape,
						Ocpu: &ocpu, MemoryInGbs: &memoryInGbs, Total: 10, Used: 5},
				},
			},
			usage: &Usage{
				shapeFdUsageMap: map[usageKey]map[string]int64{
					flexibleShapeUsageKey: {"fd1": 2}, // KeyOfShape
				},
			},
			expected: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: flexibleShape,
						Ocpu: &ocpu, MemoryInGbs: &memoryInGbs, Total: 10, Used: 7}, // 5 + 2
				},
			},
		},
		{
			name: "ad level adjustment (nil fd)",
			resolveResult: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: nil, Shape: nonFlexibleShape, Total: 10, Used: 5},
				},
			},
			usage: &Usage{
				shapeFdUsageMap: map[usageKey]map[string]int64{
					nonFlexibleShapeUsageKey: {"fd1": 1, "fd2": 2}, // sum all fds
				},
			},
			expected: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: nil, Shape: nonFlexibleShape, Total: 10, Used: 8}, // 5 + 3
				},
			},
		},
		{
			name: "ad level adjustment (nil fd) for non-flexible w/ ocpu + memory",
			resolveResult: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: nil, Shape: nonFlexibleShape, Ocpu: &ocpu, MemoryInGbs: &memoryInGbs, Total: 10, Used: 5},
				},
			},
			usage: &Usage{
				shapeFdUsageMap: map[usageKey]map[string]int64{
					nonFlexibleShapeUsageKey: {"fd1": 1, "fd2": 2}, // sum all fds
				},
			},
			expected: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: nil, Shape: nonFlexibleShape, Ocpu: &ocpu, MemoryInGbs: &memoryInGbs, Total: 10, Used: 8}, // 5 + 3
				},
			},
		},
		{
			name: "ad level adjustment (nil fd) for flexible shape",
			resolveResult: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: nil, Shape: flexibleShape, Ocpu: &ocpu, MemoryInGbs: &memoryInGbs, Total: 10, Used: 5},
				},
			},
			usage: &Usage{
				shapeFdUsageMap: map[usageKey]map[string]int64{
					flexibleShapeUsageKey: {"fd1": 1, "fd2": 2}, // sum all fds
				},
			},
			expected: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: nil, Shape: flexibleShape, Ocpu: &ocpu, MemoryInGbs: &memoryInGbs, Total: 10, Used: 8}, // 5 + 3
				},
			},
		},
		{
			name: "usage exceeds total, cap at total",
			resolveResult: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: nonFlexibleShape, Total: 10, Used: 5},
				},
			},
			usage: &Usage{
				shapeFdUsageMap: map[usageKey]map[string]int64{
					nonFlexibleShapeUsageKey: {"fd1": 10}, // 5 + 10 = 15 > 10
				},
			},
			expected: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: nonFlexibleShape, Total: 10, Used: 10}, // capped
				},
			},
		},
		{
			name: "usage exceeds total, cap at total for non-flexible shape w/ ocpu + memory",
			resolveResult: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: nonFlexibleShape,
						Ocpu: &ocpu, MemoryInGbs: &memoryInGbs, Total: 10, Used: 5},
				},
			},
			usage: &Usage{
				shapeFdUsageMap: map[usageKey]map[string]int64{
					nonFlexibleShapeUsageKey: {"fd1": 10}, // 5 + 10 = 15 > 10
				},
			},
			expected: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: nonFlexibleShape,
						Ocpu: &ocpu, MemoryInGbs: &memoryInGbs, Total: 10, Used: 10}, // capped
				},
			},
		},
		{
			name: "usage exceeds total, cap at total for flexible shape",
			resolveResult: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: flexibleShape,
						Ocpu: &ocpu, MemoryInGbs: &memoryInGbs, Total: 10, Used: 5},
				},
			},
			usage: &Usage{
				shapeFdUsageMap: map[usageKey]map[string]int64{
					flexibleShapeUsageKey: {"fd1": 10}, // 5 + 10 = 15 > 10
				},
			},
			expected: ResolveResult{
				Ocid: "test-ocid",
				ShapeAvailabilities: []ShapeAvailability{
					{FaultDomain: lo.ToPtr("fd1"), Shape: flexibleShape,
						Ocpu: &ocpu, MemoryInGbs: &memoryInGbs, Total: 10, Used: 10}, // capped
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.resolveResult.adjustCapacityAgainstUsage(tt.usage)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveResultSlice_AvailabilityForShape(t *testing.T) {
	slice := ResolveResultSlice{
		{
			Ocid: "ocid1",
			Ad:   "ad1",
			ShapeAvailabilities: []ShapeAvailability{
				{FaultDomain: lo.ToPtr("fd1"), Shape: "VM.Standard1.1", Total: 10, Used: 5},
				{Shape: "VM.Standard.E5.Flex", Ocpu: lo.ToPtr(float32(1)), MemoryInGbs: lo.ToPtr(float32(12)), Total: 6, Used: 3},
			},
		},
		{
			Ocid: "ocid2",
			Ad:   "ad2",
			ShapeAvailabilities: []ShapeAvailability{
				{FaultDomain: lo.ToPtr("fd1"), Shape: "VM.Standard1.1", Total: 20, Used: 10},
				{Shape: "VM.Standard2.1", Ocpu: lo.ToPtr(float32(1)), MemoryInGbs: lo.ToPtr(float32(12)), Total: 8, Used: 4},
			},
		},
	}

	result := slice.AvailabilityForShape("VM.Standard1.1", nil, nil, false)
	key1 := CapacityReserveIdAndAd{
		Ocid: "ocid1",
		Ad:   "ad1",
	}

	key2 := CapacityReserveIdAndAd{
		Ocid: "ocid2",
		Ad:   "ad2",
	}
	expected := map[CapacityReserveIdAndAd]map[string]ShapeAvailability{
		key1: {
			"fd1": {FaultDomain: nil, Shape: "VM.Standard1.1", Total: 10, Used: 5},
		},
		key2: {
			"fd1": {FaultDomain: nil, Shape: "VM.Standard1.1", Total: 20, Used: 10},
		},
	}
	assert.Equal(t, expected, result)

	result = slice.AvailabilityForShape("VM.Standard.E5.Flex", lo.ToPtr(float32(1)), lo.ToPtr(float32(12)), true)
	expected = map[CapacityReserveIdAndAd]map[string]ShapeAvailability{
		key1: {
			"": {FaultDomain: nil, Shape: "VM.Standard.E5.Flex",
				Ocpu: lo.ToPtr(float32(1)), MemoryInGbs: lo.ToPtr(float32(12)), Total: 6, Used: 3},
		},
	}
	assert.Equal(t, expected, result)

	// wrong memory should return empty
	result = slice.AvailabilityForShape("VM.Standard.E5.Flex", lo.ToPtr(float32(1)), lo.ToPtr(float32(32)), true)
	expected = map[CapacityReserveIdAndAd]map[string]ShapeAvailability{}
	assert.Equal(t, expected, result)

	result = slice.AvailabilityForShape("VM.Standard2.1", nil, nil, false)
	expected = map[CapacityReserveIdAndAd]map[string]ShapeAvailability{
		key2: {
			"": {FaultDomain: nil, Shape: "VM.Standard2.1", Total: 8, Used: 4},
		},
	}
	assert.Equal(t, expected, result)
}

func TestOcidToLabelValue(t *testing.T) {
	testData := []string{
		makeCapacityReservationId(60),
		makeCapacityReservationId(64),
	}

	for _, d := range testData {
		start := len(d) - 63
		lastDotIndex := strings.LastIndex(d, ".")
		if lastDotIndex+1 >= start {
			start = lastDotIndex + 1
		}

		assert.Equal(t, d[start:], OcidToLabelValue(d))
	}
}

func makeCapacityReservationId(lastPartLen int) string {
	var b bytes.Buffer
	b.WriteString("ocid1.capacityreservation.ocx.xyz")
	for ; lastPartLen > 0; lastPartLen-- {
		b.WriteByte(byte('A' + rand.Intn(26)))
	}
	return b.String()
}
