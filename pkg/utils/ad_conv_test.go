/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import "testing"

func TestAdConversions(t *testing.T) {
	ad := "phx:AD-1"

	zone := AdToZoneLabelValue(ad)
	if zone != "AD-1" {
		t.Fatalf("AdToZoneLabelValue(%q) = %q, want %q", ad, zone, "AD-1")
	}

	recomposed := ZoneLabelValueToAd(zone, "phx")
	if recomposed != ad {
		t.Fatalf("ZoneLabelValueToAd(%q, phx) = %q, want %q", zone, recomposed, ad)
	}

	prefix := ExtractLogicalAdPrefix(ad)
	if prefix != "phx" {
		t.Fatalf("ExtractLogicalAdPrefix(%q) = %q, want %q", ad, prefix, "phx")
	}
}
