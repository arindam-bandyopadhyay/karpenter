/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import "github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"

func ToTagFilterFunc[R interface{}](s *v1beta1.OciResourceSelectorTerm,
	freeFromTagsGetter func(R) map[string]string,
	definedTagsGetter func(R) map[string]map[string]interface{}) func(R) bool {
	return func(r R) bool {
		if s.FreeformTags != nil {
			rf := freeFromTagsGetter(r)

			for k, v := range s.FreeformTags {
				rv, ok := rf[k]
				if !ok || rv != v {
					return false
				}
			}
		}

		if s.DefinedTags != nil {
			rd := definedTagsGetter(r)

			for sn, sm := range s.DefinedTags {
				rm, ok := rd[sn]
				if !ok {
					return false
				}

				for k, v := range sm {
					if rv, iok := rm[k]; !iok || rv != v {
						return false
					}
				}
			}
		}

		return true
	}
}
