/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package cloudprovider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ocioraclecouldcomv1beta1 "github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/operator/options"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/image"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/instance"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/instancetype"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/kms"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/network"
	"github.com/oracle/karpenter-provider-oci/pkg/providers/placement"
	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	v1 "k8s.io/api/core/v1"
	corev1 "sigs.k8s.io/karpenter/pkg/apis/v1"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	ctx       context.Context
	cancel    context.CancelFunc
	testEnv   *envtest.Environment
	cfg       *rest.Config
	k8sClient client.Client
)

func TestCloudProvider(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "CloudProvider Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())
	ctx = options.ToContext(ctx, &options.Options{})

	var err error
	err = ocioraclecouldcomv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "apis", "crds")},
		ErrorIfCRDPathMissing: true,
	}

	// Retrieve the first found binary directory to allow running tests from IDEs
	if getFirstFoundEnvTestBinaryDir() != "" {
		testEnv.BinaryAssetsDirectory = getFirstFoundEnvTestBinaryDir()
	}

	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

// getFirstFoundEnvTestBinaryDir locates the first binary in the specified path.
// ENVTEST-based tests depend on specific binaries, usually located in paths set by
// controller-runtime. When running tests directly (e.g., via an IDE) without using
// Makefile targets, the 'BinaryAssetsDirectory' must be explicitly configured.
//
// This function streamlines the process by finding the required binaries, similar to
// setting the 'KUBEBUILDER_ASSETS' environment variable. To ensure the binaries are
// properly set up, run 'make setup-envtest' beforehand.
func getFirstFoundEnvTestBinaryDir() string {
	basePath := filepath.Join("..", "..", "bin", "k8s")
	entries, err := os.ReadDir(basePath)
	if err != nil {
		logf.Log.Error(err, "Failed to read directory", "path", basePath)
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return filepath.Join(basePath, entry.Name())
		}
	}
	return ""
}

// Mock for InstanceProvider
type FakeInstanceProvider struct {
	TestInstance *instance.InstanceInfo
}

func NewFakeInstanceProvider(inputInstance *instance.InstanceInfo) *FakeInstanceProvider {
	return &FakeInstanceProvider{
		TestInstance: inputInstance,
	}
}

func (ip *FakeInstanceProvider) LaunchInstance(ctx context.Context,
	nodeClaim *corev1.NodeClaim,
	nodeClass *ocioraclecouldcomv1beta1.OCINodeClass,
	instanceType *instancetype.OciInstanceType,
	imageResolveResult *image.ImageResolveResult,
	networkResolveResult *network.NetworkResolveResult,
	kmsKeyResolveResult *kms.KmsKeyResolveResult,
	placementProposal *placement.Proposal) (*instance.InstanceInfo, error) {
	return ip.TestInstance, nil
}

func (ip *FakeInstanceProvider) DeleteInstance(ctx context.Context, instanceOcid string) error {
	return nil
}

func (ip *FakeInstanceProvider) GetInstance(ctx context.Context,
	instanceOcid string) (*instance.InstanceInfo, error) {
	return &instance.InstanceInfo{}, nil
}

func (ip *FakeInstanceProvider) GetInstanceCached(ctx context.Context,
	instanceOcid string) (*instance.InstanceInfo, error) {
	return ip.GetInstance(ctx, instanceOcid)
}

func (ip *FakeInstanceProvider) GetInstanceCompartment(nodeClass *ocioraclecouldcomv1beta1.OCINodeClass) string {
	return ""
}

func (ip *FakeInstanceProvider) ListInstances(ctx context.Context,
	compartmentId string) ([]*ocicore.Instance, error) {
	return []*ocicore.Instance{}, nil
}

func (ip *FakeInstanceProvider) ListInstanceBootVolumeAttachments(ctx context.Context,
	compartmentOcid string, instanceOcid string, ad string) ([]*ocicore.BootVolumeAttachment, error) {
	return []*ocicore.BootVolumeAttachment{}, nil
}

func (ip *FakeInstanceProvider) ListInstanceBootVolumeAttachmentsCached(ctx context.Context,
	compartmentOcid, instanceOcid, ad string) ([]*ocicore.BootVolumeAttachment, error) {
	return ip.ListInstanceBootVolumeAttachments(ctx, compartmentOcid, instanceOcid, ad)
}

func (ip *FakeInstanceProvider) ListInstanceVnicAttachments(ctx context.Context,
	compartmentOcid string, instanceOcid string) ([]*ocicore.VnicAttachment, error) {
	return []*ocicore.VnicAttachment{}, nil
}

func (ip *FakeInstanceProvider) ListInstanceVnicAttachmentsCached(ctx context.Context,
	compartmentOcid, instanceOcid string) ([]*ocicore.VnicAttachment, error) {
	return ip.ListInstanceVnicAttachments(ctx, compartmentOcid, instanceOcid)
}

// Mock for InstanceTypeProvider
type FakeInstanceTypeProvider struct {
	InstancesTypes []*instancetype.OciInstanceType
}

func NewFakeInstanceTypeProvider(testInstanceTypes []*instancetype.OciInstanceType) *FakeInstanceTypeProvider {
	return &FakeInstanceTypeProvider{
		InstancesTypes: testInstanceTypes,
	}
}

func (itp *FakeInstanceTypeProvider) ListInstanceTypes(ctx context.Context,
	nodeClass *ocioraclecouldcomv1beta1.OCINodeClass, taints []v1.Taint) ([]*instancetype.OciInstanceType, error) {
	return itp.InstancesTypes, nil
}
