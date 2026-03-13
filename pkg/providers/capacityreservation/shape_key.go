/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package capacityreservation

import (
	"fmt"

	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	"github.com/samber/lo"
)

func KeyOfShapeAvailability(s ShapeAvailability) string {
	return KeyOfShape(s.Shape, s.Ocpu, s.MemoryInGbs)
}

func KeyOfInstance(i *ocicore.Instance) string {
	var ocpu, gbs *float32
	if i.ShapeConfig != nil {
		ocpu = i.ShapeConfig.Ocpus
		gbs = i.ShapeConfig.MemoryInGBs
	}

	return KeyOfShape(*i.Shape, ocpu, gbs)
}

func KeyOfShape(shape string, ocpu *float32, gbs *float32) string {
	if ocpu != nil && gbs != nil {
		return fmt.Sprintf("%s.%f.%f", shape, lo.FromPtr(ocpu), lo.FromPtr(gbs))
	}

	return fmt.Sprintf("%s..", shape)
}
