/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package cloudprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/oracle/karpenter-provider-oci/pkg/apis/v1beta1"
	"github.com/oracle/karpenter-provider-oci/pkg/fakes"
	ocicore "github.com/oracle/oci-go-sdk/v65/core"
	"github.com/samber/lo"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
)

var _ = Describe("Drift Tests", func() {
	It("should handle second vnic drift properly", func() {
		testCases := []struct {
			VnicAttachments     []*ocicore.VnicAttachment
			NsgIds              []*string
			SecondVnics         []*v1beta1.Vnic
			ExpectedDriftReason cloudprovider.DriftReason
			ExpectedResult      bool
		}{
			{ // Empty vnic attachement and second vnics config
				make([]*ocicore.VnicAttachment, 0),
				make([]*string, 0),
				make([]*v1beta1.Vnic, 0),
				"",
				false,
			},
			{ // Vnic attachements more than second vnics
				[]*ocicore.VnicAttachment{createVnicAttachment("vnic1", "subnet1"),
					createVnicAttachment("vnic2", "subnet2")},
				[]*string{nil, nil},
				[]*v1beta1.Vnic{createOciVnic("subnet2", []string{})},
				"SecondaryVnicsNumberMismatch",
				true,
			},
			{ // Second vnics more than Vnic attachements
				[]*ocicore.VnicAttachment{createVnicAttachment("vnic1", "subnet1")},
				[]*string{nil, nil},
				[]*v1beta1.Vnic{createOciVnic("subnet1", []string{}), createOciVnic("subnet2", []string{})},
				"SecondaryVnicsNumberMismatch",
				true,
			},
			{ // Vnic attachements match second vnics, empty nsgIds
				[]*ocicore.VnicAttachment{createVnicAttachment("vnic1", "subnet1"),
					createVnicAttachment("vnic2", "subnet2")},
				[]*string{nil, nil},
				[]*v1beta1.Vnic{createOciVnic("subnet1", []string{}), createOciVnic("subnet2", []string{})},
				"",
				false,
			},
			{ // Vnic attachements match second vnics, nsgIds matching
				[]*ocicore.VnicAttachment{createVnicAttachment("vnic1", "subnet1"),
					createVnicAttachment("vnic2", "subnet2")},
				[]*string{lo.ToPtr("nsg1,nsg2"), lo.ToPtr("nsg3")},
				[]*v1beta1.Vnic{createOciVnic("subnet1", []string{"nsg1", "nsg2"}),
					createOciVnic("subnet2", []string{"nsg3"})},
				"",
				false,
			},
			{ // Vnic attachements match second vnics, nsgIds not matching
				[]*ocicore.VnicAttachment{createVnicAttachment("vnic1", "subnet1"),
					createVnicAttachment("vnic2", "subnet2")},
				[]*string{lo.ToPtr("nsg1,nsg2"), nil},
				[]*v1beta1.Vnic{createOciVnic("subnet1", []string{"nsg1", "nsg2"}),
					createOciVnic("subnet2", []string{"nsg3"})},
				"SecondaryVnicNsgIdsMismatch",
				true,
			},
			{ // Vnic attachements match second vnics, same subnet same nsgs
				[]*ocicore.VnicAttachment{createVnicAttachment("vnic1", "subnet1"),
					createVnicAttachment("vnic2", "subnet1")},
				[]*string{lo.ToPtr("nsg1,nsg2"), lo.ToPtr("nsg1,nsg2")},
				[]*v1beta1.Vnic{createOciVnic("subnet1", []string{"nsg1", "nsg2"}),
					createOciVnic("subnet1", []string{"nsg1", "nsg2"})},
				"",
				false,
			},
			{ // Vnic attachements using same subne match one second vnic but not all
				[]*ocicore.VnicAttachment{createVnicAttachment("vnic1", "subnet1"),
					createVnicAttachment("vnic2", "subnet1")},
				[]*string{lo.ToPtr("nsg1,nsg2"), lo.ToPtr("nsg1,nsg2")},
				[]*v1beta1.Vnic{createOciVnic("subnet1", []string{"nsg1", "nsg2"}),
					createOciVnic("subnet2", []string{"nsg3"})},
				"SecondaryVnicNsgIdsMismatch",
				true,
			},
		}
		ctx := context.Background()
		for _, tc := range testCases {
			data, err := json.Marshal(tc)
			if err == nil {
				lo.Must(fmt.Fprintf(GinkgoWriter, "Testing testcase: %v", string(data)))
			} else {
				lo.Must(fmt.Fprintf(GinkgoWriter, "Testing testcase: nil"))
			}
			var nsgsMap = make(map[string][]string)
			var subNetIdMap = make(map[string]string)
			for index, vnicAttachment := range tc.VnicAttachments {
				nsgIds := make([]string, 0)
				if tc.NsgIds[index] != nil {
					nsgIds = strings.Split(*tc.NsgIds[index], ",")
				}
				nsgsMap[*vnicAttachment.VnicId] = nsgIds
				subNetIdMap[*vnicAttachment.VnicId] = *vnicAttachment.SubnetId
			}
			driftReason, err, isDrifted := isNetworkSecondVnicDrifted(ctx, tc.VnicAttachments, tc.SecondVnics,
				func(ctx context.Context, vnicOcid string) (*ocicore.Vnic, error) {
					nsgIds, ok1 := nsgsMap[vnicOcid]
					subNetId, ok2 := subNetIdMap[vnicOcid]
					if ok1 && ok2 {
						return &ocicore.Vnic{
							Id:        &vnicOcid,
							NsgIds:    nsgIds,
							SubnetId:  &subNetId,
							IsPrimary: lo.ToPtr(false),
						}, nil
					}
					return nil, nil
				})
			Expect(err).ToNot(HaveOccurred())
			Expect(isDrifted).To(Equal(tc.ExpectedResult))
			Expect(driftReason).To(Equal(tc.ExpectedDriftReason))
		}
	})

	It("should handle LaunchOptions mismatch properly", func() {
		testCases := []struct {
			desired  *ocicore.LaunchOptions
			actual   *ocicore.LaunchOptions
			expected bool
		}{
			{nil, &defaultLaunchOption, false}, // Nil desired against default actual
			{
				nil,
				&ocicore.LaunchOptions{
					BootVolumeType:       ocicore.LaunchOptionsBootVolumeTypeIscsi,
					Firmware:             ocicore.LaunchOptionsFirmwareBios,
					NetworkType:          ocicore.LaunchOptionsNetworkTypeParavirtualized,
					RemoteDataVolumeType: ocicore.LaunchOptionsRemoteDataVolumeTypeParavirtualized,
				},
				false,
			}, // Nil desired against customised actual
			{
				&defaultLaunchOption,
				&defaultLaunchOption,
				false,
			}, // Desired match actual
			{
				&ocicore.LaunchOptions{
					BootVolumeType:                  ocicore.LaunchOptionsBootVolumeTypeScsi,
					Firmware:                        ocicore.LaunchOptionsFirmwareUefi64,
					NetworkType:                     ocicore.LaunchOptionsNetworkTypeParavirtualized,
					RemoteDataVolumeType:            ocicore.LaunchOptionsRemoteDataVolumeTypeParavirtualized,
					IsConsistentVolumeNamingEnabled: lo.ToPtr(true),
				},
				&defaultLaunchOption,
				true,
			}, // Mismatch BootVolumeType
			{
				&ocicore.LaunchOptions{
					BootVolumeType:                  ocicore.LaunchOptionsBootVolumeTypeParavirtualized,
					Firmware:                        ocicore.LaunchOptionsFirmwareBios,
					NetworkType:                     ocicore.LaunchOptionsNetworkTypeParavirtualized,
					RemoteDataVolumeType:            ocicore.LaunchOptionsRemoteDataVolumeTypeParavirtualized,
					IsConsistentVolumeNamingEnabled: lo.ToPtr(true),
				},
				&defaultLaunchOption,
				true,
			}, // Mismatch Firmware
			{
				&ocicore.LaunchOptions{
					BootVolumeType:                  ocicore.LaunchOptionsBootVolumeTypeParavirtualized,
					Firmware:                        ocicore.LaunchOptionsFirmwareUefi64,
					NetworkType:                     ocicore.LaunchOptionsNetworkTypeE1000,
					RemoteDataVolumeType:            ocicore.LaunchOptionsRemoteDataVolumeTypeParavirtualized,
					IsConsistentVolumeNamingEnabled: lo.ToPtr(true),
				},
				&defaultLaunchOption,
				true,
			}, // Mismatch NetworkType
			{
				&ocicore.LaunchOptions{
					BootVolumeType:                  ocicore.LaunchOptionsBootVolumeTypeParavirtualized,
					Firmware:                        ocicore.LaunchOptionsFirmwareUefi64,
					NetworkType:                     ocicore.LaunchOptionsNetworkTypeParavirtualized,
					RemoteDataVolumeType:            ocicore.LaunchOptionsRemoteDataVolumeTypeIscsi,
					IsConsistentVolumeNamingEnabled: lo.ToPtr(true),
				},
				&defaultLaunchOption,
				true,
			}, // Mismatch RemoteDataVolumeType
			{
				&ocicore.LaunchOptions{
					BootVolumeType:                  ocicore.LaunchOptionsBootVolumeTypeParavirtualized,
					Firmware:                        ocicore.LaunchOptionsFirmwareUefi64,
					NetworkType:                     ocicore.LaunchOptionsNetworkTypeParavirtualized,
					RemoteDataVolumeType:            ocicore.LaunchOptionsRemoteDataVolumeTypeIscsi,
					IsConsistentVolumeNamingEnabled: lo.ToPtr(false),
				},
				&defaultLaunchOption,
				true,
			}, // Mismatch IsConsistentVolumeNamingEnabled
			{
				&ocicore.LaunchOptions{
					BootVolumeType: ocicore.LaunchOptionsBootVolumeTypeParavirtualized,
				},
				&defaultLaunchOption,
				false,
			}, // Matching only set values
			{
				&ocicore.LaunchOptions{
					NetworkType: ocicore.LaunchOptionsNetworkTypeVfio,
				},
				&defaultLaunchOption,
				true,
			}, // Mismatching on partially set values
			{
				&ocicore.LaunchOptions{
					IsConsistentVolumeNamingEnabled: lo.ToPtr(false),
				},
				&defaultLaunchOption,
				true,
			}, // Mismatch on IsConsistentVolumeNamingEnabled set
			{
				&ocicore.LaunchOptions{},
				&defaultLaunchOption,
				false,
			}, // Matching on empty launch options
		}
		for _, tc := range testCases {
			result := isLaunchOptionMismatch(tc.desired, tc.actual)
			Expect(result).To(Equal(tc.expected))
		}
	})

	It("Attaching vnic shouldn't cause panic", func() {
		vincAttachments := []*ocicore.VnicAttachment{createVnicAttachment("vnic1", "subnet1"),
			createVnicAttachment("vnic2", "subnet2")}
		vincAttachments[0].LifecycleState = ocicore.VnicAttachmentLifecycleStateAttaching
		vincAttachments[1].LifecycleState = ocicore.VnicAttachmentLifecycleStateAttached

		nc := fakes.CreateBasicOciNodeClass()
		nc.Status = v1beta1.OCINodeClassStatus{
			Network: &v1beta1.Network{
				PrimaryVnic: &v1beta1.Vnic{
					Subnet: v1beta1.Subnet{
						SubnetId: "subnet2",
					},
				},
			},
		}
		desireState := &InstanceDesiredState{
			NodeClass: &nc,
		}

		result, err := IsInstanceNetworkDrifted(context.TODO(), desireState, vincAttachments,
			func(ctx context.Context, vnicOcid string) (*ocicore.Vnic, error) {
				return &ocicore.Vnic{
					Id:        lo.ToPtr("vnic2"),
					SubnetId:  lo.ToPtr("subnet2"),
					IsPrimary: lo.ToPtr(true),
				}, nil
			}, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(cloudprovider.DriftReason("")))
	})

	It("Secondary vnics check can be check optionally", func() {
		vincAttachments := []*ocicore.VnicAttachment{createVnicAttachment("vnic1", "subnet1"),
			createVnicAttachment("vnic2", "subnet2")}
		vincAttachments[0].LifecycleState = ocicore.VnicAttachmentLifecycleStateAttached
		vincAttachments[1].LifecycleState = ocicore.VnicAttachmentLifecycleStateAttached

		nc := fakes.CreateBasicOciNodeClass()
		nc.Status = v1beta1.OCINodeClassStatus{
			Network: &v1beta1.Network{
				PrimaryVnic: &v1beta1.Vnic{
					Subnet: v1beta1.Subnet{
						SubnetId: "subnet2",
					},
				},
				SecondaryVnics: []*v1beta1.Vnic{
					{
						Subnet: v1beta1.Subnet{
							SubnetId: "subnet1",
						},
					},
					{
						Subnet: v1beta1.Subnet{
							SubnetId: "subnet3",
						},
					},
				},
			},
		}

		desireState := &InstanceDesiredState{
			NodeClass: &nc,
		}

		getVnicFunc := func(ctx context.Context, vnicOcid string) (*ocicore.Vnic, error) {
			vnicMap := map[string]*ocicore.Vnic{
				"vnic2": {
					Id:        lo.ToPtr("vnic2"),
					SubnetId:  lo.ToPtr("subnet2"),
					IsPrimary: lo.ToPtr(true),
				},
				"vnic1": {
					Id:        lo.ToPtr("vnic1"),
					SubnetId:  lo.ToPtr("subnet1"),
					IsPrimary: lo.ToPtr(false),
				},
				"vnic3": {
					Id:        lo.ToPtr("vnic3"),
					SubnetId:  lo.ToPtr("subnet3"),
					IsPrimary: lo.ToPtr(false),
				},
			}

			return vnicMap[vnicOcid], nil
		}

		result, err := IsInstanceNetworkDrifted(context.TODO(), desireState, vincAttachments, getVnicFunc, true)

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(cloudprovider.DriftReason("")))

		result, err = IsInstanceNetworkDrifted(context.TODO(), desireState, vincAttachments, getVnicFunc, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(cloudprovider.DriftReason("SecondaryVnicsNumberMismatch")))

		vnicAtt3 := createVnicAttachment("vnic3", "subnet3")
		vnicAtt3.LifecycleState = ocicore.VnicAttachmentLifecycleStateAttached
		vincAttachments = append(vincAttachments, vnicAtt3)

		result, err = IsInstanceNetworkDrifted(context.TODO(), desireState, vincAttachments, getVnicFunc, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(cloudprovider.DriftReason("")))
	})
})

func createVnicAttachment(vnicOcid string, subnetId string) *ocicore.VnicAttachment {
	return &ocicore.VnicAttachment{
		Id:       &vnicOcid,
		VnicId:   &vnicOcid,
		SubnetId: &subnetId,
	}
}
func createOciVnic(subnetId string, nsgIds []string) *v1beta1.Vnic {
	return &v1beta1.Vnic{
		Subnet: v1beta1.Subnet{
			SubnetId: subnetId,
		},
		NetworkSecurityGroups: lo.Map(nsgIds, func(item string, _ int) v1beta1.NetworkSecurityGroup {
			return v1beta1.NetworkSecurityGroup{
				NetworkSecurityGroupId: item,
			}
		}),
	}
}

var defaultLaunchOption = ocicore.LaunchOptions{
	BootVolumeType:                  ocicore.LaunchOptionsBootVolumeTypeParavirtualized,
	Firmware:                        ocicore.LaunchOptionsFirmwareUefi64,
	NetworkType:                     ocicore.LaunchOptionsNetworkTypeParavirtualized,
	RemoteDataVolumeType:            ocicore.LaunchOptionsRemoteDataVolumeTypeParavirtualized,
	IsConsistentVolumeNamingEnabled: lo.ToPtr(true),
}
