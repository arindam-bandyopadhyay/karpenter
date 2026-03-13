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
	"github.com/oracle/karpenter-provider-oci/pkg/providers/capacityreservation"
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "sigs.k8s.io/karpenter/pkg/test/expectations"
)

var crTestNodeClass v1beta1.OCINodeClass
var crController *Controller

const testClusterCompartmentId = "testClusterCompartmentId"

var (
	testCrId1 = "ocid1.capacityreservation.1"
	testCrId2 = "ocid1.capacityreservation.2"
)

var _ = Describe("Capacity Reservation Reconciler", func() {
	BeforeEach(func() {
		crTestNodeClass = fakes.CreateBasicOciNodeClass()
		crTestNodeClass.Spec.CapacityReservationConfigs = []*v1beta1.CapacityReservationConfig{}

		fakeComputeClient := fakes.NewFakeCapacityReservationClient(testClusterCompartmentId)

		crProvider := capacityreservation.NewProvider(ctx, fakeComputeClient, testClusterCompartmentId)

		crController = &Controller{
			Client:   k8sClient,
			recorder: &fakes.FakeEventRecorder{},
			reconcilers: []nodeClassReconciler{
				&InitStatus{},
				&CapacityReservationReconciler{
					capacityReservationProvider: crProvider,
				},
			},
		}
	})

	It("should set CapacityReservation condition to true if CapacityReservation found", func() {
		crTestNodeClass.Spec.CapacityReservationConfigs = append(crTestNodeClass.Spec.CapacityReservationConfigs,
			&v1beta1.CapacityReservationConfig{
				CapacityReservationId: &testCrId1,
			})

		nodeClassPtr := &crTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, crController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		_, found := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeCapacityReservation && item.Status == metav1.ConditionTrue
		})
		Expect(found).To(BeTrue())
	})

	It("should set CapacityReservation condition to false if CapacityReservation not found", func() {
		crTestNodeClass.Spec.CapacityReservationConfigs = append(crTestNodeClass.Spec.CapacityReservationConfigs,
			&v1beta1.CapacityReservationConfig{
				CapacityReservationId: lo.ToPtr("dummyCrId"),
			})

		nodeClassPtr := &crTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, crController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		_, found := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeCapacityReservation && item.Status == metav1.ConditionFalse
		})
		Expect(found).To(BeTrue())
	})

	It("should set CapacityReservation condition to true if CapacityReservation compartment not matching", func() {
		crTestNodeClass.Spec.CapacityReservationConfigs = append(crTestNodeClass.Spec.CapacityReservationConfigs,
			&v1beta1.CapacityReservationConfig{
				CapacityReservationId: &testCrId2,
			})

		nodeClassPtr := &crTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, crController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		_, found := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeCapacityReservation && item.Status == metav1.ConditionTrue
		})
		Expect(found).To(BeTrue())
	})

	It("unset status.capacityReservations if no capacity reservation config", func() {
		crTestNodeClass.Status.CapacityReservations = []v1beta1.CapacityReservation{
			{
				CapacityReservationId: "xxxx",
				DisplayName:           "yyyy",
				AvailabilityDomain:    "ad-1",
			},
		}

		nodeClassPtr := &crTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, crController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.StatusConditions().Get(v1beta1.ConditionTypeCapacityReservation)).To(BeNil())
		Expect(nodeClassPtr.Status.CapacityReservations).To(BeNil())
	})
})
