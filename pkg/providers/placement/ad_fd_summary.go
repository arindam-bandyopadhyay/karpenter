/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package placement

import (
	"sort"
	"sync"

	"github.com/oracle/karpenter-provider-oci/pkg/providers/capacityreservation"
	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	"github.com/samber/lo"
)

type adFdSummary struct {
	mutex       sync.Mutex
	instanceMap map[string]*instanceSummary
}

type instanceSummary struct {
	ad                    string
	fd                    string
	shape                 string
	capacityReservationId *string
	ocpu                  *float32
	memoryInGbs           *float32
}

func newAdFdSummary() *adFdSummary {
	return &adFdSummary{
		instanceMap: make(map[string]*instanceSummary),
	}
}

func (a *adFdSummary) forget(instanceId string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	delete(a.instanceMap, instanceId)
}

func (a *adFdSummary) update(i *ocicore.Instance) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if i == nil {
		return
	}

	var ocpu, memoryInGbs *float32
	if i.ShapeConfig != nil {
		ocpu = i.ShapeConfig.Ocpus
		memoryInGbs = i.ShapeConfig.MemoryInGBs
	}

	a.updateBy(*i.Id, *i.AvailabilityDomain, *i.FaultDomain, *i.Shape, i.CapacityReservationId, ocpu, memoryInGbs)
}

func (a *adFdSummary) updateBy(instanceId, ad, fd, shape string, capacityReservationId *string,
	ocpu, memoryInGbs *float32) {
	_, ok := a.instanceMap[instanceId]
	if !ok {
		a.instanceMap[instanceId] = &instanceSummary{
			ad:                    ad,
			fd:                    fd,
			capacityReservationId: capacityReservationId,
			shape:                 shape,
			ocpu:                  ocpu,
			memoryInGbs:           memoryInGbs,
		}
	}
}

func (a *adFdSummary) Propose(proposals []Proposal, shape string, ocpu, memoryInGbs *float32) *Proposal {
	if len(proposals) == 0 {
		return nil
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()

	sort.Slice(proposals, func(i, j int) bool {
		return a.instanceCountForAdFd(proposals[i].Ad, proposals[i].Fd) <
			a.instanceCountForAdFd(proposals[j].Ad, proposals[j].Fd)
	})

	p := proposals[0]
	a.updateBy(p.tempIdentifier, p.Ad, lo.FromPtr(p.Fd), shape, p.CapacityReservationId, ocpu, memoryInGbs)
	return &p
}

// UsageForCapacityReservation return usage from instance cache, it is not in use for now
func (a *adFdSummary) UsageForCapacityReservation(capacityReservationId, shape string,
	ocpu, memoryInGbs *float32) map[string]int64 {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	output := make(map[string]int64)
	shapeKey := capacityreservation.KeyOfShape(shape, ocpu, memoryInGbs)

	for _, v := range a.instanceMap {
		if v.capacityReservationId != nil && *v.capacityReservationId == capacityReservationId {
			keyOfInstance := capacityreservation.KeyOfShape(v.shape, v.ocpu, v.memoryInGbs)

			if shapeKey == keyOfInstance {
				fdEntry, ok := output[v.fd]
				if ok {
					output[v.fd] = fdEntry + 1
				} else {
					output[v.fd] = 1
				}
			}
		}
	}

	return output
}

func (a *adFdSummary) instanceCountForAdFd(ad string, fd *string) int64 {
	var count int64

	for _, v := range a.instanceMap {
		if v.ad == ad {
			if fd != nil && v.fd == *fd {
				count++
			} else if fd == nil {
				count++
			}
		}
	}

	return count
}
