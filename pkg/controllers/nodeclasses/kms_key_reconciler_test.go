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
	"github.com/oracle/karpenter-provider-oci/pkg/providers/kms"
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "sigs.k8s.io/karpenter/pkg/test/expectations"
)

var kmsTestNodeClass v1beta1.OCINodeClass
var kmsController *Controller
var testKeyId string

var _ = Describe("KMS Key Reconciler", func() {
	BeforeEach(func() {
		kmsTestNodeClass = fakes.CreateBasicOciNodeClass()

		testKeyId = "ocid1.key.oc1.iad.testvalut.kms123"
		fakeKmsClient := fakes.NewFakeKmsClient()

		kmsProvider := lo.Must(kms.NewProvider(ctx, "testCompartmentId", fakes.NewDummyConfigurationProvider()))
		kmsProvider.SetKmsClient("https://testvalut-management.kms.us-ashburn-1.oraclecloud.com", fakeKmsClient)

		kmsController = &Controller{
			Client:   k8sClient,
			recorder: &fakes.FakeEventRecorder{},
			reconcilers: []nodeClassReconciler{
				&InitStatus{},
				&KmsKeyReconciler{
					kmsKeyProvider: kmsProvider,
				},
			},
		}
	})

	It("should set KmsKey condition to true if kms key found", func() {
		kmsTestNodeClass.Spec.VolumeConfig.BootVolumeConfig.KmsKeyConfig = &v1beta1.KmsKeyConfig{
			KmsKeyId: &testKeyId,
		}

		nodeClassPtr := &kmsTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, kmsController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		_, found := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeKmsKeyReady && item.Status == metav1.ConditionTrue
		})
		Expect(found).To(BeTrue())

		_, imageFound := lo.Find(nodeClassPtr.Status.Volume.KmsKeys, func(item *v1beta1.KmsKey) bool {
			return item.KmsKeyId == testKeyId
		})
		Expect(imageFound).To(BeTrue())
	})

	It("should set KmsKey condition to false if kms key not found", func() {
		kmsTestNodeClass.Spec.VolumeConfig.BootVolumeConfig.KmsKeyConfig = &v1beta1.KmsKeyConfig{
			KmsKeyId: lo.ToPtr("dummyTestKey"),
		}

		nodeClassPtr := &kmsTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, kmsController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.GetConditions()).ToNot(BeEmpty())
		_, found := lo.Find(nodeClassPtr.GetConditions(), func(item status.Condition) bool {
			return item.Type == v1beta1.ConditionTypeKmsKeyReady && item.Status == metav1.ConditionFalse
		})
		Expect(found).To(BeTrue())
	})

	It("should not set KmsKey condition if kms key not set", func() {
		nodeClassPtr := &kmsTestNodeClass
		nodeClassPtr.Status.Volume = &v1beta1.Volume{
			KmsKeys: []*v1beta1.KmsKey{
				{
					KmsKeyId:    "xxx",
					DisplayName: "yyy",
				},
			},
		}
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, kmsController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.StatusConditions().Get(v1beta1.ConditionTypeKmsKeyReady)).To(BeNil())
		Expect(nodeClassPtr.Status.Volume.KmsKeys).To(BeNil())
	})

	It("should unset status.volume.kmsKeys if kms key not set", func() {
		nodeClassPtr := &kmsTestNodeClass
		ExpectApplied(ctx, k8sClient, nodeClassPtr)
		ExpectObjectReconciled(ctx, k8sClient, kmsController, nodeClassPtr)
		nodeClassPtr = ExpectExists(ctx, k8sClient, nodeClassPtr)

		Expect(nodeClassPtr.StatusConditions().Get(v1beta1.ConditionTypeKmsKeyReady)).To(BeNil())
	})
})
