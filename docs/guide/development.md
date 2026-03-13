# Development

## Running Karpenter Controller locally for rapid development

1. Obtain the OKE cluster's CA certificate:
```sh
kubectl config view --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}' | base64 -d > ca.crt
```

2. Export environment variables:
```sh
export CLUSTER_COMPARTMENT_ID=<cluster compartment id>
export APISERVER_ENDPOINT=<OKE Cluster API endpoint IP>
export VCN_COMPARTMENT_ID=<vcn compartment id>
export DISABLE_LEADER_ELECTION=true
export KUBERNETES_CA_CERT_FILE=<OKE cluster CA cert obtained in Step 1>
export OCI_AUTH_METHOD=SESSION
export OCI_REGION=<Region of OKE cluster>
export KUBECONFIG=<OKE cluster KubeConfig location>
```

If you want to run KPO with an OCI profile other than `DEFAULT`, export:
```sh
export OCI_PROFILE_NAME=<OCI profile name>
```

3. Run:
```sh
make run
```

After making code or documentation changes, run:
```sh
make verify
```
This is the recommended local validation step before opening a pull request.

## Running E2E Tests

E2E tests install the Karpenter Provider OCI Helm chart into an existing OKE cluster, create an `OCINodeClass` + `NodePool`, validate scale-up/scale-down, and (for flannel) run drift + consolidation scenarios.

### Prerequisites

- Tools: `go`, `kubectl`, `helm`, `oci` CLI, `jq`
- An existing OKE cluster (flannel) kubeconfig exported as `KUBECONFIG`
- Optional (NPN): a second OKE cluster kubeconfig exported as `KUBECONFIG_NPN`
- A packaged Helm chart archive path exported as `KARPENTER_CHART_TGZ` (must be a path relative to the repo root; e.g., `dist/karpenter-0.1.0.tgz`)

### Configure test data (recommended: generate from templates)

The E2E tests read:
- `test/e2e/testdata/e2e_test_config_flannel.json` (flannel)
- `test/e2e/testdata/e2e_test_config_npn.json` (NPN)
- `test/e2e/testdata/e2e_test_helm_values_flannel.yaml` (flannel chart values)
- `test/e2e/testdata/e2e_test_helm_values_npn.yaml` (NPN chart values)

#### OCI resources required by `generateTestConfig.sh`

`test/e2e/testdata/generateTestConfig.sh` fills in the templates by looking up OCI resources by **compartment name** and **resource display-name**. Before running it, ensure your test environment includes:

- A VCN with worker subnets and NSGs.
- A flannel OKE cluster, and optionally an NPN OKE cluster.
- Capacity reservations, a compute cluster, and KMS keys if you want the related test coverage.

All resource names used by the generator are overridable via environment variables. The most commonly customized values are:

- `DRIFT_COMPARTMENT_NAME`
- `KEYS_COMPARTMENT_NAME`
- `VAULT_NAME`
- `KMS_KEY1_NAME`, `KMS_KEY2_NAME`
- `NODE_SUBNET1_NAME`, `NODE_SUBNET2_NAME`
- `NSG1_NAME`, `NSG2_NAME`
- `CAPACITY_RESERVATION1_NAME`, `CAPACITY_RESERVATION2_NAME`
- `COMPUTE_CLUSTER_NAME`
- `FLANNEL_CLUSTER_NAME`, `NPN_CLUSTER_NAME`
- `UBUNTU_IMAGE_NAME`, `CUSTOM_IMAGE_NAME`
- `SSH_PUB_KEY`
- `IMAGE_REGISTRY`, `IMAGE_REPOSITORY_NAME`

To generate these from the templates using OCI lookups:

```sh
cd test/e2e/testdata

# Optional overrides:
# export OCI_CLI_AUTH=instance_principal|api_key|security_token
# export ENDPOINT="https://containerengine.<region>.oci.oraclecloud.com"   # if needed
# export OCI_AUTH_METHOD_FOR_TEST=PROFILE_SESSION|INSTANCE_PRINCIPAL       # written into JSON config
# export SSH_PUB_KEY="$(cat ~/.ssh/id_rsa.pub)"                            # optional; auto-detected when possible
# export IMAGE_REGISTRY="ghcr.io"
# export IMAGE_REPOSITORY_NAME="oracle/karpenter-provider-oci"

./generateTestConfig.sh <TENANCY_OCID> <COMPARTMENT_NAME> <IMAGE_TAG>
```

Notes:
- `IMAGE_TAG` is substituted into the Helm values templates (`image.tag`). Use `IMAGE_REGISTRY` and `IMAGE_REPOSITORY_NAME` if your image is published somewhere else.
- If the generator doesn’t match your environment, you can edit the generated `.json`/`.yaml` directly; field meanings are in `test/e2e/config.go`.

### Package the Helm chart

```sh
mkdir -p dist
helm package ./chart -d dist
export KARPENTER_CHART_TGZ="$(ls dist/karpenter-*.tgz | tail -n 1)"
```

### Run E2E

```sh
# Flannel only
KUBECONFIG=<path/to/flannel/kubeconfig> KARPENTER_CHART_TGZ="${KARPENTER_CHART_TGZ}" make test-e2e-flannel

# Flannel + NPN (only runs NPN suite when KUBECONFIG_NPN is set)
KUBECONFIG=<path/to/flannel/kubeconfig> \
KUBECONFIG_NPN=<path/to/npn/kubeconfig> \
KARPENTER_CHART_TGZ="${KARPENTER_CHART_TGZ}" \
make test-e2e
```

Helpful flags:
- `SKIP_CLEANUP=true` skips teardown and Helm uninstall (useful for debugging).
- `LARGE_SHAPE_TEST_ENABLED=false` skips large/expensive shape tests (GPU/baremetal/denseIO/compute-cluster).


