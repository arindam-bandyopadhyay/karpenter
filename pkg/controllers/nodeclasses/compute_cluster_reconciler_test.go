/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package nodeclasses

import (
	"github.com/awslabs/operatorpkg/status"
	"github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/fakes"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/computecluster"
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "sigs.k8s.io/karpenter/pkg/test/expectations"
)

var comClusterTestNodeClass v1beta1.OCINodeClass
var comClusterController *Controller

var (
	testComputeClusterId1 = "ocid1.computecluster.1"
	testComputeClusterId2 = "ocid1.computecluster.2"
)

var _ = Describe("Compute Cluster Reconciler", func() {
	BeforeEach(func() {
		comClusterTestNodeClass = fakes.CreateBasicOciNodeClass()
		comClusterTestNodeClass.Spec.ComputeClusterConfig = &v1beta1.ComputeClusterConfig{}

		testClusterCompartmentId := "testClusterCompartmentId"
		fakeComputeClient := fakes.NewFakeComputeClient(testClusterCompartmentId)

		computeClusterProvider := computecluster.NewProvider(ctx, fakeComputeClient, testClusterCompartmentId)

		comClusterController = &Controller{
			Client:   k8sClient,
			recorder: &fakes.FakeEventRecorder{},
			reconcilers: []nodeClassReconciler{
				&InitStatus{},
				&ComputeClusterReconciler{
					computeClusterProvider: computeClusterProvider,
				},
			},
		}
	})

	It("should set ComputeCluster condition to true if ComputeCluster found", func() {
		comClusterTestNodeClass.Spec.ComputeClusterConfig.ComputeClusterId = &testComputeClusterId1

		nodeClassPtr := &comClusterTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, comClusterController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		_, found := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeComputeCluster && item.Status == metav1.ConditionTrue
		})
		Expect(found).To(BeTrue())
	})

	It("should set ComputeCluster condition to false if ComputeCluster not found", func() {
		comClusterTestNodeClass.Spec.ComputeClusterConfig.ComputeClusterId = lo.ToPtr("dummyComputeClusterId")
		comClusterTestNodeClass.Name = "test-oci-nodeclass-cs-cluster-false"

		nodeClassPtr := &comClusterTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, comClusterController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		_, found := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeComputeCluster && item.Status == metav1.ConditionFalse
		})
		Expect(found).To(BeTrue())
	})

	It("should set ComputeCluster condition to true if ComputeCluster compartment not matching", func() {
		comClusterTestNodeClass.Spec.ComputeClusterConfig.ComputeClusterId = &testComputeClusterId2
		comClusterTestNodeClass.Name = "test-oci-nodeclass-cs-cluster-compartment"

		nodeClassPtr := &comClusterTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, comClusterController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		_, found := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeComputeCluster && item.Status == metav1.ConditionTrue
		})
		Expect(found).To(BeTrue())
	})
})
