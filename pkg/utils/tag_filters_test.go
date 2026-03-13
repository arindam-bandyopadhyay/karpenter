/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import (
	"testing"

	"github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/stretchr/testify/require"
)

type tfResource struct {
	ft map[string]string
	dt map[string]map[string]interface{}
}

func TestTagFilterFunc(t *testing.T) {
	selector := &v1beta1.OciResourceSelectorTerm{
		FreeformTags: map[string]string{"env": "prod"},
		DefinedTags:  map[string]map[string]string{"ns": {"team": "core"}},
	}
	filter := ToTagFilterFunc(selector,
		func(r tfResource) map[string]string { return r.ft },
		func(r tfResource) map[string]map[string]interface{} {
			result := make(map[string]map[string]interface{})
			for k, v := range r.dt {
				result[k] = make(map[string]interface{})
				for k2, v2 := range v {
					result[k][k2] = v2
				}
			}
			return result
		})

	ok := tfResource{
		ft: map[string]string{"env": "prod"},
		dt: map[string]map[string]interface{}{"ns": {"team": "core"}},
	}
	badFF := tfResource{
		ft: map[string]string{"env": "dev"},
		dt: map[string]map[string]interface{}{"ns": {"team": "core"}},
	}
	badDef := tfResource{
		ft: map[string]string{"env": "prod"},
		dt: map[string]map[string]interface{}{"ns": {"team": "other"}},
	}

	require.True(t, filter(ok))
	require.False(t, filter(badFF))
	require.False(t, filter(badDef))
}
