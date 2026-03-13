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
	"github.com/oracle/karpenter-provider-oci/pkg/providers/clusterplacementgroup"
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "sigs.k8s.io/karpenter/pkg/test/expectations"
)

var cgpTestNodeClass v1beta1.OCINodeClass
var cpgController *Controller

var (
	testCgpId1 = "ocid1.cpg.1"
	testCgpId2 = "ocid1.cpg.2"
	testCgpId3 = "ocid1.cpg.3"
)

var _ = Describe("Cluster Placement Group Reconciler", func() {
	BeforeEach(func() {
		cgpTestNodeClass = fakes.CreateBasicOciNodeClass()
		cgpTestNodeClass.Spec.ClusterPlacementGroupConfigs = []*v1beta1.ClusterPlacementGroupConfig{}

		testClusterCompartmentId := "testClusterCompartmentId"
		fakeCpgClient := fakes.NewFakeClusterPlacementGroupClient(testClusterCompartmentId)

		cgpProvider := clusterplacementgroup.NewProvider(ctx, fakeCpgClient, testClusterCompartmentId)

		cpgController = &Controller{
			Client:   k8sClient,
			recorder: &fakes.FakeEventRecorder{},
			reconcilers: []nodeClassReconciler{
				&InitStatus{},
				&ClusterPlacementGroupReconciler{
					clusterPlacementGroupProvider: cgpProvider,
				},
			},
		}
	})

	It("should set ClusterPlacementGroup condition to true if ClusterPlacementGroup found", func() {
		cgpTestNodeClass.Spec.ClusterPlacementGroupConfigs = append(cgpTestNodeClass.Spec.ClusterPlacementGroupConfigs,
			&v1beta1.ClusterPlacementGroupConfig{
				ClusterPlacementGroupId: &testCgpId1,
			})

		nodeClassPtr := &cgpTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, cpgController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		_, found := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeClusterPlacementGroup && item.Status == metav1.ConditionTrue
		})
		Expect(found).To(BeTrue())
	})

	It("should set ClusterPlacementGroup condition to false if ClusterPlacementGroup not found", func() {
		cgpTestNodeClass.Spec.ClusterPlacementGroupConfigs = append(cgpTestNodeClass.Spec.ClusterPlacementGroupConfigs,
			&v1beta1.ClusterPlacementGroupConfig{
				ClusterPlacementGroupId: lo.ToPtr("dummyCgpId"),
			})

		nodeClassPtr := &cgpTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, cpgController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		_, found := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeClusterPlacementGroup && item.Status == metav1.ConditionFalse
		})
		Expect(found).To(BeTrue())
	})

	It("should set ClusterPlacementGroup condition to true if ClusterPlacementGroup compartment not matching", func() {
		cgpTestNodeClass.Spec.ClusterPlacementGroupConfigs = append(cgpTestNodeClass.Spec.ClusterPlacementGroupConfigs,
			&v1beta1.ClusterPlacementGroupConfig{
				ClusterPlacementGroupId: &testCgpId2,
			})

		nodeClassPtr := &cgpTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, cpgController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		_, found := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeClusterPlacementGroup && item.Status == metav1.ConditionTrue
		})
		Expect(found).To(BeTrue())
	})

	It("should set ClusterPlacementGroup condition to false if multi ClusterPlacementGroup in same AD", func() {
		cgpTestNodeClass.Spec.ClusterPlacementGroupConfigs = append(cgpTestNodeClass.Spec.ClusterPlacementGroupConfigs,
			&v1beta1.ClusterPlacementGroupConfig{
				ClusterPlacementGroupId: &testCgpId1,
			},
			&v1beta1.ClusterPlacementGroupConfig{
				ClusterPlacementGroupId: &testCgpId3,
			})

		nodeClassPtr := &cgpTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, cpgController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		condition, found := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeClusterPlacementGroup && item.Status == metav1.ConditionFalse
		})
		Expect(found).To(BeTrue())
		Expect(condition.Reason).To(Equal(v1beta1.ConditionClusterPlacementGroupNotReadyReason))
		Expect(condition.Message).To(Equal(ClusterPlacementGroupInTheSameAd))
	})

	It("should unset status.clusterPlacementGroups if no clusterPlacementGroupConfigs", func() {
		cgpTestNodeClass.Status.ClusterPlacementGroups = []v1beta1.ClusterPlacementGroup{
			{
				ClusterPlacementGroupId: testCgpId1,
				DisplayName:             "xxxx",
				AvailabilityDomain:      "ad-1",
			},
		}

		nodeClassPtr := &cgpTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, cpgController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.StatusConditions().Get(v1beta1.ConditionTypeClusterPlacementGroup)).To(BeNil())
		Expect(nodeClassPtr.Status.ClusterPlacementGroups).To(BeNil())
	})
})
