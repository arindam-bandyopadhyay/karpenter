# Install KPO Helm Chart

These commands assume you are installing from a checkout of this repository (the chart path is `./chart`).
If you are installing from a packaged chart archive, replace `./chart` with the path to the `.tgz` you are using.

KPO’s Helm chart supports a wide range of configuration. To view all available values:

```shell
helm show values ./chart
```

Only `settings.clusterCompartmentId`, `settings.vcnCompartmentId`, and `settings.apiserverEndpoint` must be provided by the user.

The chart already provides a default OKE image compartment for `settings.preBakedImageCompartmentId`. You only need to set it if you want to override that default. Please refer to [Override the OKE image compartment used by `imageFilter`](advanced-use-cases.md#override-the-oke-image-compartment-used-by-imagefilter) for additional details.

For all the chart values, see [Helm Chart Reference](../reference/helm-chart.md).

### Install

1. Decide the namespace where KPO will run.

2. Create a `values.yaml` containing the required values.

   Minimal example:

   ```yaml
   settings:
     clusterCompartmentId: "<your-cluster-compartment-ocid>"
     vcnCompartmentId: "<your-vcn-compartment-ocid>"
     apiserverEndpoint: "<api-server-endpoint-ip>"
     ociVcnIpNative: false

   # Optional: override image location/tag (example for OCIR)
   image:
     registry: "<registry>"
     repositoryName: "<your-ocir-namespace>/<your-repo>"
     tag: "<image-tag>"
   ```

3. Install:

   ```shell
   helm install karpenter ./chart \
     --values <path-to-values.yaml> \
     --namespace <karpenter-namespace> \
     --create-namespace
   ```

4. Verify:

   ```shell
   kubectl -n <karpenter-namespace> rollout status deploy/karpenter --timeout=120s
   kubectl -n <karpenter-namespace> get pods
   ```

## Upgrade KPO Helm Chart

```shell
helm upgrade karpenter ./chart \
  --namespace <karpenter-namespace> \
  --values <path-to-values.yaml>
```

## Uninstall KPO Helm Chart

```shell
helm uninstall karpenter --namespace <karpenter-namespace>
```
