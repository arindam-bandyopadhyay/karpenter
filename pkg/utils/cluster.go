/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package utils

import (
	"strings"

	"github.com/coreos/go-semver/semver"
	"k8s.io/client-go/kubernetes"
)

func GetClusterVersion(kubernetesInterface kubernetes.Interface) (*semver.Version, error) {
	k8sVersion, err := kubernetesInterface.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}
	return semver.New(strings.Trim(strings.ToLower(k8sVersion.String()), "v")), nil
}
