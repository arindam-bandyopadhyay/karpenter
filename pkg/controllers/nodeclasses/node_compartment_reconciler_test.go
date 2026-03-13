/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package nodeclasses

import (
	"time"

	"github.com/awslabs/operatorpkg/status"
	"github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/fakes"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/identity"
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "sigs.k8s.io/karpenter/pkg/test/expectations"
)

var compartmentTestNodeClass v1beta1.OCINodeClass
var compartmentController *Controller
var testCompartmentId string

var _ = Describe("Node Compartment Reconciler", func() {
	BeforeEach(func() {
		compartmentTestNodeClass = fakes.CreateBasicOciNodeClass()
		testCompartmentId = "ocid1.tenancy.oc1..tenancy123"

		identityProvider := lo.Must(identity.NewProvider(ctx, testCompartmentId, fakes.NewFakeIdentityClient()))

		compartmentController = &Controller{
			Client:   k8sClient,
			recorder: &fakes.FakeEventRecorder{},
			reconcilers: []nodeClassReconciler{
				&InitStatus{},
				&NodeCompartmentReconciler{
					identityProvider:     identityProvider,
					clusterCompartmentId: testCompartmentId,
				},
			},
		}
	})

	It("should set NodeCompartment condition to true if compartment found", func() {
		compartmentTestNodeClass.Spec.NodeCompartmentId = lo.ToPtr("ocid1.compartment.oc1..cluster123")

		nodeClassPtr := &compartmentTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, compartmentController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		_, found := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeNodeCompartment && item.Status == metav1.ConditionTrue
		})
		Expect(found).To(BeTrue())
	})

	It("should set NodeCompartment condition to false if compartment not found", func() {
		compartmentTestNodeClass.Spec.NodeCompartmentId = lo.ToPtr("dummyCompartment")

		nodeClassPtr := &compartmentTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		result := ExpectObjectReconciled(ctx, k8sClient, compartmentController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(result.RequeueAfter).To(Equal(5 * time.Minute))

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		_, foundTrue := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeNodeCompartment && item.Status == metav1.ConditionTrue
		})
		Expect(foundTrue).To(BeFalse())

		_, foundFalse := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeNodeCompartment && item.Status == metav1.ConditionFalse
		})
		Expect(foundFalse).To(BeTrue())
	})
})
