/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package capacityreservation

import (
	"strings"
	"time"

	"github.com/samber/lo"
)

type ResolveResult struct {
	Ocid                string
	Name                string
	Ad                  string
	CompartmentId       string
	ShapeAvailabilities []ShapeAvailability
	time                time.Time
}

// AvailabilityForShape calculates shape availability for a specific shape in a capacity reservation.
// returns a map of availability for each FD, when Fd is not specified empty FD shapeKey is used as
// theoretically a capacity reservation can mix ad and fd reservations, we don't allow this case as it would
// cause confusion when estimate capacity reservation usage.
func (rr ResolveResult) AvailabilityForShape(shape string,
	ocpu *float32, gbs *float32, shapeConfigSupported bool) map[string]ShapeAvailability {
	out := make(map[string]ShapeAvailability)

	// merge multiple capRes for the same shape, should only happen in theory that a capRes has multiple records f
	// or the same, :) you never know.
	shapeKey := KeyOfShape(shape, ocpu, gbs)
	for _, sa := range rr.ShapeAvailabilities {
		// capacity reservation allows ocpu/memory settings for non-flexible shape, so need check both entry.
		// in the output non-flexible shape ocpu/memory is removed.
		if KeyOfShapeAvailability(sa) == shapeKey || (!shapeConfigSupported && sa.Shape == shape) {
			shapeFd := lo.FromPtr(sa.FaultDomain)

			v, ok := out[shapeFd]
			if ok {
				// merge into existing fd, be noticed in case not FD specified, use empty shapeKey for ad.
				v.Total += sa.Total
				v.Used += sa.Used
			} else {
				v = ShapeAvailability{
					Shape: shape,
					Total: sa.Total,
					Used:  sa.Used,
				}

				if shapeConfigSupported {
					v.Ocpu = ocpu
					v.MemoryInGbs = gbs
				}
			}

			out[shapeFd] = v
		}
	}

	return out
}

func (rr ResolveResult) adjustCapacityAgainstUsage(usage *Usage) ResolveResult {
	output := ResolveResult{
		Ocid:                rr.Ocid,
		Ad:                  rr.Ad,
		Name:                rr.Name,
		CompartmentId:       rr.CompartmentId,
		ShapeAvailabilities: make([]ShapeAvailability, 0),
	}

	output.ShapeAvailabilities = lo.Map(rr.ShapeAvailabilities, func(item ShapeAvailability, _ int) ShapeAvailability {
		uk := usageKey{
			shape: item.Shape,
			key:   KeyOfShapeAvailability(item),
		}

		// this works well for flexible shape w/ shape config (mandatory) and non-flexible shape w/o shape config.
		if fdUsage, ok := usage.shapeFdUsageMap[uk]; ok {
			return adjustShapeAvailability(rr.Ad, item, fdUsage)
		}

		// capacity reservation allows specifying shape config for non-flexible shape, which does not take effect.
		// in such case, only shape is relevant during consumption, usage is tracked against shape only
		possibleUk := usageKey{
			shape: item.Shape,
			key:   KeyOfShape(item.Shape, nil, nil),
		}

		if fdUsage, ok := usage.shapeFdUsageMap[possibleUk]; ok {
			return adjustShapeAvailability(rr.Ad, item, fdUsage)
		}

		return item
	})

	return output
}

func adjustShapeAvailability(ad string, input ShapeAvailability, fdUsage map[string]int64) ShapeAvailability {
	sa := ShapeAvailability{
		Ad:          ad,
		FaultDomain: input.FaultDomain,
		Shape:       input.Shape,
		Ocpu:        input.Ocpu,
		MemoryInGbs: input.MemoryInGbs,
		Total:       input.Total,
	}

	// we don't allow mix of fd capRes vs. ad capRes. if capRes is declared without fd, we combine all fd usage
	// and adjust against that ad.
	var instanceCount int64
	if input.FaultDomain == nil {
		for _, v := range fdUsage {
			instanceCount += v
		}
	} else if u, fdOk := fdUsage[lo.FromPtr(input.FaultDomain)]; fdOk {
		instanceCount = u
	}

	sa.Used = min(input.Total, input.Used+instanceCount)
	return sa
}

type ResolveResultSlice []ResolveResult

func (rrs ResolveResultSlice) AvailabilityForShape(shape string, ocpu *float32, memoryInGbs *float32,
	shapeConfigSupported bool) map[CapacityReserveIdAndAd]map[string]ShapeAvailability {
	out := make(map[CapacityReserveIdAndAd]map[string]ShapeAvailability)

	for _, rr := range rrs {
		innerMap := rr.AvailabilityForShape(shape, ocpu, memoryInGbs, shapeConfigSupported)

		if len(innerMap) > 0 {
			idAndAd := CapacityReserveIdAndAd{
				Ocid: rr.Ocid,
				Ad:   rr.Ad,
			}
			out[idAndAd] = innerMap
		}
	}

	return out
}

type CapacityReserveIdAndAd struct {
	Ocid string
	Ad   string
}

func toCapacityResolveResult(c *CapResWithLoadTime) ResolveResult {
	shapeAvailabilities := make([]ShapeAvailability, 0)

	for _, instanceReserveConfig := range c.InstanceReservationConfigs {
		var ocpus, memoryInGbs *float32
		if instanceReserveConfig.InstanceShapeConfig != nil {
			ocpus = instanceReserveConfig.InstanceShapeConfig.Ocpus
			memoryInGbs = instanceReserveConfig.InstanceShapeConfig.MemoryInGBs
		}

		sa := ShapeAvailability{
			Ad:          *c.AvailabilityDomain,
			FaultDomain: instanceReserveConfig.FaultDomain,
			Total:       *instanceReserveConfig.ReservedCount,
			Used:        *instanceReserveConfig.UsedCount,
			Shape:       *instanceReserveConfig.InstanceShape,
			Ocpu:        ocpus,
			MemoryInGbs: memoryInGbs,
		}
		shapeAvailabilities = append(shapeAvailabilities, sa)
	}

	return ResolveResult{
		Ocid:                *c.Id,
		Name:                *c.DisplayName,
		Ad:                  *c.AvailabilityDomain,
		CompartmentId:       *c.CompartmentId,
		ShapeAvailabilities: shapeAvailabilities,
		time:                c.time,
	}
}

type ShapeAvailability struct {
	Ad          string
	FaultDomain *string
	Total       int64
	Used        int64
	Shape       string
	Ocpu        *float32
	MemoryInGbs *float32
}

func OcidToLabelValue(ocid string) string {
	elements := strings.Split(ocid, ".")
	lastEle := elements[len(elements)-1]

	if len(lastEle) <= 63 {
		return lastEle
	}

	// currently ocid last part is 60 characters long so this should be unreachable,
	// keep the last 63 bytes in case it is too long in the future.
	return lastEle[len(lastEle)-63:]
}
