/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:openapi-gen=true

// +kubebuilder:rbac:groups=oci.oraclecloud.com,resources=nativepodnetworks,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=npn;npns
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.reason`
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NativePodNetwork contains information about the various vnics and ips that are attached to this node
type NativePodNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              NativePodNetworkSpec   `json:"spec"`
	Status            NativePodNetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NativePodNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NativePodNetwork `json:"items"`
}

type NativePodNetworkSpec struct {
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	// ID holds the instance ocid that this CR belongs to
	ID string `json:"id"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=31
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=110
	// +Optional
	// MaxPodCount is the maximum number of pods that can be scheduled onto this node
	MaxPodCount int64 `json:"maxPodCount"`

	// +kubebuilder:validation:MinItems=0
	// +kubebuilder:validation:MaxItems=5
	// +kubebuilder:validation:Optional
	// +Optional
	// NetworkSecurityGroupIDs are the firewall rules associated with this node
	NetworkSecurityGroupIDs []string `json:"networkSecurityGroupIds"`

	// +kubebuilder:validation:MinItems=0
	// +kubebuilder:validation:MaxItems=5
	// +kubebuilder:validation:Optional
	// PodSubnetIDs are the subnets that pods are allowed to be allocated on
	PodSubnetIDs []string `json:"podSubnetIds"`

	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=2
	// +kubebuilder:validation:Optional
	// +Optional
	// IpFamilies are the IP Protocols supported in single/dual stack mode
	IpFamilies []string `json:"ipFamilies"`

	// +kubebuilder:validation:MinItems=0
	// +kubebuilder:validation:MaxItems=64
	// +kubebuilder:validation:Optional
	// +Optional
	// SecondaryVnics are the vnics which are created or attached to the node
	SecondaryVnics []SecondaryVnic `json:"secondaryVnics"`
}

type SecondaryVnic struct {
	CreateVnicDetails CreateVnicDetails `json:"createVnicDetails"`
	DisplayName       string            `json:"displayName"`
	NicIndex          *int              `json:"nicIndex,omitempty"`
}

type CreateVnicDetails struct {
	AssignIpv6Ip                         *bool                             `json:"assignIpv6Ip,omitempty"`
	AssignPublicIp                       *bool                             `json:"assignPublicIp,omitempty"`
	DefinedTags                          map[string]map[string]interface{} `json:"definedTags,omitempty"`
	DisplayName                          *string                           `json:"displayName,omitempty"`
	IpCount                              *int                              `json:"ipCount,omitempty"`
	ApplicationResources                 []string                          `json:"applicationResources,omitempty"`
	FreeformTags                         map[string]string                 `json:"freeformTags,omitempty"`
	SecurityAttributes                   map[string]map[string]interface{} `json:"securityAttributes,omitempty"`
	Ipv6AddressIpv6SubnetCidrPairDetails []Ipv6AddressCidrPair             `json:"ipv6AddressIpv6SubnetCidrPairDetails,omitempty"`
	NsgIds                               []string                          `json:"nsgIds,omitempty"`
	SkipSourceDestCheck                  *bool                             `json:"skipSourceDestCheck,omitempty"`
	SubnetId                             *string                           `json:"subnetId,omitempty"`
}

type Ipv6AddressCidrPair struct {
	Ipv6Address    string `json:"ipv6Address"`
	Ipv6SubnetCidr string `json:"ipv6SubnetCidr"`
}

type NativePodNetworkStatus struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=BACKOFF;IN_PROGRESS;SUCCESS
	// State is a short code of what is happening to the NPN custom resource
	State string `json:"state,omitempty"`

	// +kubebuilder:validation:Optional
	// +Optional
	// Reason is a general statement about any actions or errors that are happening on the NPN custom resource
	Reason string `json:"reason,omitempty"`

	// +kubebuilder:validation:Optional
	// +Optional
	// VNICs contains general information about the vnics/ips that are attached
	VNICs []VNIC `json:"vnics,omitempty"`
}

type IPVersion struct {
	// The IPv4 address/cidr in case of dual stack
	V4 string `json:"v4,omitempty"`

	// The IPv6 address/cidr in case of dual stack
	V6 string `json:"v6,omitempty"`
}

type NICConfiguration struct {
	// +kubebuilder:default=31
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=31
	// +Optional
	// IpCount is the number of ips to be attached to this vnic
	IpCount int64 `json:"ipCount"`

	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=2
	// +kubebuilder:validation:Optional
	// +Optional
	// IpFamilies are the IP Protocols supported in single/dual stack mode
	// Must be a subnet of ipFamilies in NativePodNetworkSpec
	IpFamilies []string `json:"ipFamilies"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=255
	// Subnet id is the subnet that this vnic is on
	SubnetId string `json:"subnet"`

	// +kubebuilder:validation:MinItems=0
	// +kubebuilder:validation:MaxItems=5
	// +kubebuilder:validation:Optional
	// +Optional
	// NetworkSecurityGroupIDs are the firewall rules associated with this vnic
	NetworkSecurityGroupIDs []string `json:"nsgs"`

	// +kubebuilder:validation:MinItems=0
	// +kubebuilder:validation:MaxItems=1
	// +kubebuilder:validation:Optional
	// ApplicationResources are used to determine which pods are to be scheduled on this vnic
	ApplicationResources []string `json:"applicationResources,omitempty"`
}

type VNIC struct {
	// +kubebuilder:validation:MinLength=1
	ID string `json:"vnicId,omitempty"`

	// +kubebuilder:validation:Optional
	// +Optional
	// NicConfiguration contains settings specified for this vnic
	NicConfiguration NICConfiguration `json:"nicConfiguration,omitempty"`

	// +kubebuilder:validation:MinLength=1
	// This field is being deprecated in favor of using subnetCidrs field
	SubnetCIDR string `json:"subnetCidr,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinItems=1
	// +Optional
	SubnetCIDRs []IPVersion `json:"subnetCidrs,omitempty"`

	// +kubebuilder:validation:MinLength=1
	MacAddress string `json:"macAddress,omitempty"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	// This field is being deprecated in favor of using routerIPs field
	RouterIP string `json:"routerIp,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinItems=1
	// +Optional
	RouterIPs []IPVersion `json:"routerIps,omitempty"`

	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=32
	// This field is being deprecated in favor of using podAddresses field
	Addresses []string `json:"addresses,omitempty"`

	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=32
	// +kubebuilder:validation:Optional
	// +Optional
	PodAddresses []IPVersion `json:"podAddresses,omitempty"`

	// +kubebuilder:validation:Optional
	// +Optional
	// Used for bare metal Compute shapes to identify the physical NIC hosting the VNIC
	NICIndex int `json:"nicIndex"`

	// +kubebuilder:validation:Optional
	// +Optional
	// Used for bare metal Compute shapes to isolate VNIC traffic on physical interfaces
	VLANTag int `json:"vlanTag"`

	// +kubebuilder:validation:Optional
	// +Optional
	// Used to enrich with primary VNIC address for the host namespace interface. This field is being deprecated in favor of using hostAddresses field
	HostAddress string `json:"hostAddress"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinItems=1
	// +Optional
	HostAddresses []IPVersion `json:"hostAddresses,omitempty"`
}
