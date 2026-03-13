/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package capacityreservation

import (
	"sync"
	"time"

	ocicore "github.com/oracle/oci-go-sdk/v65/core"
)

type Usage struct {
	sync.Mutex
	lastCommit      time.Time
	shapeFdUsageMap map[usageKey]map[string]int64
}

func newUsage() *Usage {
	return &Usage{
		lastCommit:      time.Now(),
		shapeFdUsageMap: make(map[usageKey]map[string]int64),
	}
}

func (u *Usage) IncreaseUsage(i *ocicore.Instance) {
	u.adjust(*i.Shape, KeyOfInstance(i), *i.FaultDomain, 1)
}

func (u *Usage) DecreaseUsage(i *ocicore.Instance) {
	u.adjust(*i.Shape, KeyOfInstance(i), *i.FaultDomain, -1)
}

func (u *Usage) adjust(shape string, key string, fd string, delta int64) {
	u.Lock()
	defer u.Unlock()

	mapKey := usageKey{
		shape: shape,
		key:   key,
	}

	fdUsage, ok := u.shapeFdUsageMap[mapKey]
	if ok {
		fdDelta, fdOk := fdUsage[fd]
		if fdOk {
			fdUsage[fd] = fdDelta + delta
		} else {
			fdUsage[fd] = delta
		}
	} else {
		fdUsage = make(map[string]int64)
		fdUsage[fd] = delta
	}

	u.shapeFdUsageMap[mapKey] = fdUsage
	u.lastCommit = time.Now()
}

type usageKey struct {
	shape string
	key   string
}
