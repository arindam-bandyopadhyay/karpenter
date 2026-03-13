/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package capacityreservation

import (
	"testing"

	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestKeyOfShapeAvailability(t *testing.T) {
	sa := ShapeAvailability{
		Shape:       "VM.Standard1.1",
		Ocpu:        nil,
		MemoryInGbs: nil,
	}
	result := KeyOfShapeAvailability(sa)
	expected := "VM.Standard1.1.."
	assert.Equal(t, expected, result)
}

func TestKeyOfInstance(t *testing.T) {
	instance := &ocicore.Instance{
		Shape: lo.ToPtr("VM.Standard1.1"),
		ShapeConfig: &ocicore.InstanceShapeConfig{
			Ocpus:       float32Ptr(2.0),
			MemoryInGBs: float32Ptr(15.0),
		},
		FaultDomain: lo.ToPtr("fd1"),
	}
	result := KeyOfInstance(instance)
	expected := "VM.Standard1.1.2.000000.15.000000"
	assert.Equal(t, expected, result)
}

func TestKeyOfShape(t *testing.T) {
	tests := []struct {
		shape       string
		ocpu        *float32
		memoryInGbs *float32
		expected    string
	}{
		{"VM.Standard1.1", nil, nil, "VM.Standard1.1.."},
		{"VM.Standard1.1", float32Ptr(2.0), float32Ptr(15.0), "VM.Standard1.1.2.000000.15.000000"},
	}
	for _, tt := range tests {
		result := KeyOfShape(tt.shape, tt.ocpu, tt.memoryInGbs)
		assert.Equal(t, tt.expected, result)
	}
}

func float32Ptr(f float32) *float32 { return &f }
