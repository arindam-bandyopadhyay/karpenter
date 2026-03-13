/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import "strings"

func AdToZoneLabelValue(ad string) string {
	return strings.Split(ad, ":")[1]
}

func ZoneLabelValueToAd(zone string, prefix string) string {
	return prefix + ":" + zone
}

func ExtractLogicalAdPrefix(ad string) string {
	return strings.Split(ad, ":")[0]
}
