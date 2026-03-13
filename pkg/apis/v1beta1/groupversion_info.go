/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

// Package v1beta1 contains API Schema definitions for the oci.oraclecloud.com v1beta1 API group.
// +kubebuilder:object:generate=true
// +groupName=oci.oraclecloud.com
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "oci.oraclecloud.com", Version: "v1beta1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	localSchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = localSchemeBuilder.AddToScheme
)

// Adds the list of known types to scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&OCINodeClass{},
		&OCINodeClassList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
