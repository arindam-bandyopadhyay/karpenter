/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package instance

import ocicore "github.com/oracle/oci-go-sdk/v65/core"

type InstanceInfo struct {
	*ocicore.Instance
	etag string
}
