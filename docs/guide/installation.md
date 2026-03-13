# Installation

This page covers prerequisites and the recommended installation flow for KPO.

## Prerequisites

- **OKE Cluster (Kubernetes version >= v1.31)**: use an existing cluster or create a new one by following the [instructions](https://docs.oracle.com/en-us/iaas/Content/ContEng/Tasks/create-cluster.htm)
- **OKE Managed Node Pool or Self-Managed Nodes**: provision Kubernetes capacity to run KPO.
- **OciIpNativeCNI**: if using OciIpNativeCNI cluster, ensure add-on version >= **3.0.0**. It is highly recommended to use add-on version **3.2.0** or later. If an earlier add-on version (prior to **3.2.0**) must be used, it is required that the secondary VNIC `ipCount` is explicitly configured to be no greater than 16.
- **Helm client**: [Download](https://helm.sh/docs/intro/quickstart#install-helm) the Helm client binaries


## Next steps

1. [Configure IAM Policies for KPO to manage OCI resources](configure-iam-policies.md)
2. [Install KPO Helm Chart](installation-using-helm.md)
