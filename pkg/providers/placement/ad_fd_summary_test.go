/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package placement

import (
	"testing"

	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestNewAdFdSummary(t *testing.T) {
	summary := newAdFdSummary()

	assert.NotNil(t, summary)
	assert.NotNil(t, summary.instanceMap)
	assert.Empty(t, summary.instanceMap)
}

func TestAdFdSummary_Forget(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*adFdSummary)
		instanceId  string
		expectedLen int
	}{
		{
			name: "forget existing instance",
			setup: func(s *adFdSummary) {
				s.updateBy("instance1", "AD-1", "FD-1", "VM.Standard2.1", nil, nil, nil)
			},
			instanceId:  "instance1",
			expectedLen: 0,
		},
		{
			name: "forget non-existing instance",
			setup: func(s *adFdSummary) {
				s.updateBy("instance1", "AD-1", "FD-1", "VM.Standard2.1", nil, nil, nil)
			},
			instanceId:  "instance2",
			expectedLen: 1,
		},
		{
			name:        "forget from empty summary",
			setup:       func(s *adFdSummary) {},
			instanceId:  "instance1",
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := newAdFdSummary()
			tt.setup(summary)

			summary.forget(tt.instanceId)

			assert.Len(t, summary.instanceMap, tt.expectedLen)
		})
	}
}

func TestAdFdSummary_Update(t *testing.T) {
	tests := []struct {
		name     string
		instance *ocicore.Instance
		expected *instanceSummary
	}{
		{
			name:     "update with nil instance",
			instance: nil,
			expected: nil,
		},
		{
			name: "update with instance without shape config",
			instance: &ocicore.Instance{
				Id:                 lo.ToPtr("instance1"),
				AvailabilityDomain: lo.ToPtr("AD-1"),
				FaultDomain:        lo.ToPtr("FD-1"),
				Shape:              lo.ToPtr("VM.Standard2.1"),
			},
			expected: &instanceSummary{
				ad:                    "AD-1",
				fd:                    "FD-1",
				shape:                 "VM.Standard2.1",
				capacityReservationId: nil,
				ocpu:                  nil,
				memoryInGbs:           nil,
			},
		},
		{
			name: "update with instance with shape config",
			instance: &ocicore.Instance{
				Id:                 lo.ToPtr("instance1"),
				AvailabilityDomain: lo.ToPtr("AD-1"),
				FaultDomain:        lo.ToPtr("FD-1"),
				Shape:              lo.ToPtr("VM.Standard2.1"),
				ShapeConfig: &ocicore.InstanceShapeConfig{
					Ocpus:       lo.ToPtr(float32(2.0)),
					MemoryInGBs: lo.ToPtr(float32(16.0)),
				},
				CapacityReservationId: lo.ToPtr("reservation1"),
			},
			expected: &instanceSummary{
				ad:                    "AD-1",
				fd:                    "FD-1",
				shape:                 "VM.Standard2.1",
				capacityReservationId: lo.ToPtr("reservation1"),
				ocpu:                  lo.ToPtr(float32(2.0)),
				memoryInGbs:           lo.ToPtr(float32(16.0)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := newAdFdSummary()
			summary.update(tt.instance)

			if tt.expected == nil {
				assert.Empty(t, summary.instanceMap)
			} else {
				actual, exists := summary.instanceMap[*tt.instance.Id]
				assert.True(t, exists)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

func TestAdFdSummary_UpdateBy(t *testing.T) {
	tests := []struct {
		name                  string
		instanceId            string
		ad                    string
		fd                    string
		shape                 string
		capacityReservationId *string
		ocpu                  *float32
		memoryInGbs           *float32
		expected              *instanceSummary
	}{
		{
			name:                  "update new instance",
			instanceId:            "instance1",
			ad:                    "AD-1",
			fd:                    "FD-1",
			shape:                 "VM.Standard2.1",
			capacityReservationId: nil,
			ocpu:                  nil,
			memoryInGbs:           nil,
			expected: &instanceSummary{
				ad:                    "AD-1",
				fd:                    "FD-1",
				shape:                 "VM.Standard2.1",
				capacityReservationId: nil,
				ocpu:                  nil,
				memoryInGbs:           nil,
			},
		},
		{
			name:                  "update existing instance - should update",
			instanceId:            "instance1",
			ad:                    "AD-2",
			fd:                    "FD-2",
			shape:                 "VM.Standard2.2",
			capacityReservationId: lo.ToPtr("reservation1"),
			ocpu:                  lo.ToPtr(float32(4.0)),
			memoryInGbs:           lo.ToPtr(float32(32.0)),
			expected: &instanceSummary{
				ad:                    "AD-2",
				fd:                    "FD-2",
				shape:                 "VM.Standard2.2",
				capacityReservationId: lo.ToPtr("reservation1"),
				ocpu:                  lo.ToPtr(float32(4.0)),
				memoryInGbs:           lo.ToPtr(float32(32.0)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := newAdFdSummary()

			// For existing instance test, add instance first
			if tt.name == "update existing instance" {
				summary.updateBy("instance1", "AD-1", "FD-1", "VM.Standard2.1", nil, nil, nil)
			}

			summary.updateBy(tt.instanceId, tt.ad, tt.fd, tt.shape, tt.capacityReservationId, tt.ocpu, tt.memoryInGbs)

			actual, exists := summary.instanceMap[tt.instanceId]
			assert.True(t, exists)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestAdFdSummary_Propose(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*adFdSummary)
		proposals   []Proposal
		shape       string
		ocpu        *float32
		memoryInGbs *float32
		expected    *Proposal
	}{
		{
			name:      "empty proposals",
			setup:     func(s *adFdSummary) {},
			proposals: []Proposal{},
			expected:  nil,
		},
		{
			name:  "single proposal",
			setup: func(s *adFdSummary) {},
			proposals: []Proposal{
				{Ad: "AD-1", Fd: lo.ToPtr("FD-1"), tempIdentifier: "claim1"},
			},
			shape:       "VM.Standard2.1",
			ocpu:        lo.ToPtr(float32(2.0)),
			memoryInGbs: lo.ToPtr(float32(16.0)),
			expected:    &Proposal{Ad: "AD-1", Fd: lo.ToPtr("FD-1"), tempIdentifier: "claim1"},
		},
		{
			name: "multiple proposals - select lowest count AD/FD",
			setup: func(s *adFdSummary) {
				s.updateBy("inst1", "AD-1", "FD-1", "VM.Standard2.1", nil, nil, nil)
				s.updateBy("inst2", "AD-1", "FD-1", "VM.Standard2.1", nil, nil, nil) // AD-1/FD-1 has 2
				s.updateBy("inst3", "AD-1", "FD-2", "VM.Standard2.1", nil, nil, nil) // AD-1/FD-2 has 1
				s.updateBy("inst4", "AD-2", "FD-1", "VM.Standard2.1", nil, nil, nil) // AD-2/FD-1 has 1
			},
			proposals: []Proposal{
				{Ad: "AD-1", Fd: lo.ToPtr("FD-1"), tempIdentifier: "claim1"}, // count 2
				{Ad: "AD-1", Fd: lo.ToPtr("FD-2"), tempIdentifier: "claim2"}, // count 1 - should be selected
				{Ad: "AD-2", Fd: lo.ToPtr("FD-1"), tempIdentifier: "claim3"}, // count 1
			},
			shape:       "VM.Standard2.1",
			ocpu:        lo.ToPtr(float32(2.0)),
			memoryInGbs: lo.ToPtr(float32(16.0)),
			expected:    &Proposal{Ad: "AD-1", Fd: lo.ToPtr("FD-2"), tempIdentifier: "claim2"},
		},
		{
			name: "multiple proposals with nil FD - select lowest count AD",
			setup: func(s *adFdSummary) {
				s.updateBy("inst1", "AD-1", "FD-1", "VM.Standard2.1", nil, nil, nil)
				s.updateBy("inst2", "AD-2", "FD-1", "VM.Standard2.1", nil, nil, nil)
				s.updateBy("inst3", "AD-2", "FD-2", "VM.Standard2.1", nil, nil, nil) // AD-2 has 2 total
			},
			proposals: []Proposal{
				{Ad: "AD-1", Fd: nil, tempIdentifier: "claim1"}, // AD-1 total count 1
				{Ad: "AD-2", Fd: nil, tempIdentifier: "claim2"}, // AD-2 total count 2
			},
			shape:       "VM.Standard2.1",
			ocpu:        lo.ToPtr(float32(2.0)),
			memoryInGbs: lo.ToPtr(float32(16.0)),
			expected:    &Proposal{Ad: "AD-1", Fd: nil, tempIdentifier: "claim1"},
		},
		{
			name: "multiple proposals with nil FD - select AD with zero count",
			setup: func(s *adFdSummary) {
				s.updateBy("inst1", "AD-1", "FD-1", "VM.Standard2.1", nil, nil, nil)
				s.updateBy("inst2", "AD-2", "FD-1", "VM.Standard2.1", nil, nil, nil)
				s.updateBy("inst3", "AD-2", "FD-2", "VM.Standard2.1", nil, nil, nil) // AD-2 has 2 total
			},
			proposals: []Proposal{
				{Ad: "AD-1", Fd: nil, tempIdentifier: "claim1"}, // AD-1 total count 1
				{Ad: "AD-2", Fd: nil, tempIdentifier: "claim2"}, // AD-2 total count 2
				{Ad: "AD-3", Fd: nil, tempIdentifier: "claim3"}, // AD-3 total count 0
			},
			shape:       "VM.Standard2.1",
			ocpu:        lo.ToPtr(float32(2.0)),
			memoryInGbs: lo.ToPtr(float32(16.0)),
			expected:    &Proposal{Ad: "AD-3", Fd: nil, tempIdentifier: "claim3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := newAdFdSummary()
			tt.setup(summary)

			result := summary.Propose(tt.proposals, tt.shape, tt.ocpu, tt.memoryInGbs)

			assert.Equal(t, tt.expected, result)

			// Verify instance was added to summary if proposal was selected
			if result != nil {
				actual, exists := summary.instanceMap[result.tempIdentifier]
				assert.True(t, exists)
				assert.Equal(t, result.Ad, actual.ad)
				assert.Equal(t, tt.shape, actual.shape)
				assert.Equal(t, tt.ocpu, actual.ocpu)
				assert.Equal(t, tt.memoryInGbs, actual.memoryInGbs)
			}
		})
	}
}

func TestAdFdSummary_UsageForCapacityReservation(t *testing.T) {
	tests := []struct {
		name                  string
		setup                 func(*adFdSummary)
		capacityReservationId string
		shape                 string
		ocpu                  *float32
		memoryInGbs           *float32
		expected              map[string]int64
	}{
		{
			name:                  "empty summary",
			setup:                 func(s *adFdSummary) {},
			capacityReservationId: "reservation1",
			shape:                 "VM.Standard2.1",
			ocpu:                  lo.ToPtr(float32(2.0)),
			memoryInGbs:           lo.ToPtr(float32(16.0)),
			expected:              map[string]int64{},
		},
		{
			name: "single instance matching reservation and shape",
			setup: func(s *adFdSummary) {
				s.updateBy("inst1", "AD-1", "FD-1", "VM.Standard2.1",
					lo.ToPtr("reservation1"), lo.ToPtr(float32(2.0)), lo.ToPtr(float32(16.0)))
			},
			capacityReservationId: "reservation1",
			shape:                 "VM.Standard2.1",
			ocpu:                  lo.ToPtr(float32(2.0)),
			memoryInGbs:           lo.ToPtr(float32(16.0)),
			expected:              map[string]int64{"FD-1": 1},
		},
		{
			name: "multiple instances in same FD",
			setup: func(s *adFdSummary) {
				s.updateBy("inst1", "AD-1", "FD-1", "VM.Standard2.1",
					lo.ToPtr("reservation1"), lo.ToPtr(float32(2.0)), lo.ToPtr(float32(16.0)))
				s.updateBy("inst2", "AD-1", "FD-1", "VM.Standard2.1",
					lo.ToPtr("reservation1"), lo.ToPtr(float32(2.0)), lo.ToPtr(float32(16.0)))
				s.updateBy("inst3", "AD-1", "FD-2", "VM.Standard2.1",
					lo.ToPtr("reservation1"), lo.ToPtr(float32(2.0)), lo.ToPtr(float32(16.0)))
			},
			capacityReservationId: "reservation1",
			shape:                 "VM.Standard2.1",
			ocpu:                  lo.ToPtr(float32(2.0)),
			memoryInGbs:           lo.ToPtr(float32(16.0)),
			expected:              map[string]int64{"FD-1": 2, "FD-2": 1},
		},
		{
			name: "instances with different reservations",
			setup: func(s *adFdSummary) {
				s.updateBy("inst1", "AD-1", "FD-1", "VM.Standard2.1",
					lo.ToPtr("reservation1"), lo.ToPtr(float32(2.0)), lo.ToPtr(float32(16.0)))
				s.updateBy("inst2", "AD-1", "FD-1", "VM.Standard2.1",
					lo.ToPtr("reservation2"), lo.ToPtr(float32(2.0)), lo.ToPtr(float32(16.0)))
			},
			capacityReservationId: "reservation1",
			shape:                 "VM.Standard2.1",
			ocpu:                  lo.ToPtr(float32(2.0)),
			memoryInGbs:           lo.ToPtr(float32(16.0)),
			expected:              map[string]int64{"FD-1": 1},
		},
		{
			name: "instances with different shapes",
			setup: func(s *adFdSummary) {
				s.updateBy("inst1", "AD-1", "FD-1", "VM.Standard2.1",
					lo.ToPtr("reservation1"), lo.ToPtr(float32(2.0)), lo.ToPtr(float32(16.0)))
				s.updateBy("inst2", "AD-1", "FD-1", "VM.Standard2.2",
					lo.ToPtr("reservation1"), lo.ToPtr(float32(4.0)), lo.ToPtr(float32(32.0)))
			},
			capacityReservationId: "reservation1",
			shape:                 "VM.Standard2.1",
			ocpu:                  lo.ToPtr(float32(2.0)),
			memoryInGbs:           lo.ToPtr(float32(16.0)),
			expected:              map[string]int64{"FD-1": 1},
		},
		{
			name: "instance without capacity reservation",
			setup: func(s *adFdSummary) {
				s.updateBy("inst1", "AD-1", "FD-1", "VM.Standard2.1", nil,
					lo.ToPtr(float32(2.0)), lo.ToPtr(float32(16.0)))
			},
			capacityReservationId: "reservation1",
			shape:                 "VM.Standard2.1",
			ocpu:                  lo.ToPtr(float32(2.0)),
			memoryInGbs:           lo.ToPtr(float32(16.0)),
			expected:              map[string]int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := newAdFdSummary()
			tt.setup(summary)

			result := summary.UsageForCapacityReservation(tt.capacityReservationId, tt.shape, tt.ocpu, tt.memoryInGbs)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdFdSummary_InstanceCountForAdFd(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*adFdSummary)
		ad       string
		fd       *string
		expected int64
	}{
		{
			name:     "empty summary",
			setup:    func(s *adFdSummary) {},
			ad:       "AD-1",
			fd:       lo.ToPtr("FD-1"),
			expected: 0,
		},
		{
			name: "count with specific AD and FD",
			setup: func(s *adFdSummary) {
				s.updateBy("inst1", "AD-1", "FD-1", "VM.Standard2.1", nil, nil, nil)
				s.updateBy("inst2", "AD-1", "FD-1", "VM.Standard2.1", nil, nil, nil)
				s.updateBy("inst3", "AD-1", "FD-2", "VM.Standard2.1", nil, nil, nil)
				s.updateBy("inst4", "AD-2", "FD-1", "VM.Standard2.1", nil, nil, nil)
			},
			ad:       "AD-1",
			fd:       lo.ToPtr("FD-1"),
			expected: 2,
		},
		{
			name: "count with specific AD and nil FD (all FDs in AD)",
			setup: func(s *adFdSummary) {
				s.updateBy("inst1", "AD-1", "FD-1", "VM.Standard2.1", nil, nil, nil)
				s.updateBy("inst2", "AD-1", "FD-1", "VM.Standard2.1", nil, nil, nil)
				s.updateBy("inst3", "AD-1", "FD-2", "VM.Standard2.1", nil, nil, nil)
				s.updateBy("inst4", "AD-2", "FD-1", "VM.Standard2.1", nil, nil, nil)
			},
			ad:       "AD-1",
			fd:       nil,
			expected: 3,
		},
		{
			name: "count with different AD",
			setup: func(s *adFdSummary) {
				s.updateBy("inst1", "AD-1", "FD-1", "VM.Standard2.1", nil, nil, nil)
				s.updateBy("inst2", "AD-2", "FD-1", "VM.Standard2.1", nil, nil, nil)
			},
			ad:       "AD-2",
			fd:       lo.ToPtr("FD-1"),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := newAdFdSummary()
			tt.setup(summary)

			result := summary.instanceCountForAdFd(tt.ad, tt.fd)

			assert.Equal(t, tt.expected, result)
		})
	}
}
