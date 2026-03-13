/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package nodeclasses

import (
	"github.com/oracle/karpenter-provider-oci/pkg/fakes"
	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NodeClass Event", func() {
	It("should create event properly", func() {
		eventTestNodeClass := fakes.CreateBasicOciNodeClass()
		event := PendingNodeClaimTerminationEvent(&eventTestNodeClass, []string{"n1", "n2"})

		Expect(event.InvolvedObject).To(Equal(&eventTestNodeClass))
		Expect(event.Type).To(Equal(corev1.EventTypeNormal))
		Expect(event.Reason).To(Equal("WaitingOnNodeClaimTermination"))
		Expect(event.Message).To(Equal("Waiting on NodeClaim termination for n1, n2"))
	})
})
