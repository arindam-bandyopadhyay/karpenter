# Usage

Here are a few common use cases. For solutions addressing additional scenarios, refer to the [Advanced Use Cases](advanced-use-cases.md) or check the [OCINodeClass](ocinodeclass.md) specification.

---

## Use KPO to manage nodes with OCI flexible shapes and an OKE image

The YAML below creates a Karpenter `NodePool` that can launch nodes with one of `VM.Standard.E3.Flex`, `VM.Standard.E4.Flex`, `VM.Standard.E5.Flex` shapes. The `OCINodeClass` provides two `shapeConfigs` (`2 ocpus / 8 GiB` and `4 ocpus / 16 GiB`) and selects an OKE pre-baked image by OCID.

```yaml
---
apiVersion: karpenter.sh/v1
kind: NodePool
metadata:
  name: my-nodepool
spec:
  template:
    spec:
      expireAfter: Never
      nodeClassRef:
        group: oci.oraclecloud.com
        kind: OCINodeClass
        name: my-ocinodeclass
      requirements:
        - key: karpenter.sh/capacity-type
          operator: In
          values:
            - on-demand
        - key: oci.oraclecloud.com/instance-shape # extend this list as needed
          operator: In
          values:
            - VM.Standard.E3.Flex
            - VM.Standard.E4.Flex
            - VM.Standard.E5.Flex
      terminationGracePeriod: 120m
  disruption:
    budgets:
      - nodes: 5%
    consolidateAfter: 60m
    consolidationPolicy: WhenEmpty
  limits:
    cpu: 64
    memory: 256Gi
---
apiVersion: oci.oraclecloud.com/v1beta1
kind: OCINodeClass
metadata:
  name: my-ocinodeclass
spec:
  shapeConfigs:
    - ocpus: 2
      memoryInGbs: 8
    - ocpus: 4
      memoryInGbs: 16
  volumeConfig:
    bootVolumeConfig:
      imageConfig:
        imageType: OKEImage
        imageId: <oke-image-ocid>
  networkConfig:
    primaryVnicConfig:
      subnetConfig:
        subnetId: <subnet-ocid>
```        

## Ensure worker nodes using an OKE image are always updated to the latest image

The sample OCINodeClass below specifies an image filter to select OKE images. The resolved image depends on the cluster's Kubernetes version and the available OKE images. When the cluster control plane is upgraded or new OKE images are released, the desired worker node image will also change—nodes launched with an outdated image will be considered as "Drifted". To minimize unexpected disruption during these events, it is recommended to configure an appropriate disruption budget in the Karpenter node pool, specifying reasons, disruption percentage, and schedule.

```yaml
---
apiVersion: karpenter.sh/v1
kind: NodePool
metadata:
  name: my-nodepool
spec:
  template:
    spec:
      expireAfter: Never
      nodeClassRef:
        group: oci.oraclecloud.com
        kind: OCINodeClass
        name: my-ocinodeclass
      requirements:
        - key: karpenter.sh/capacity-type
          operator: In
          values:
            - on-demand
        - key: oci.oraclecloud.com/instance-shape # extend this list as needed
          operator: In
          values:
            - VM.Standard.E3.Flex
            - VM.Standard.E4.Flex
            - VM.Standard.E5.Flex
      terminationGracePeriod: 120m
  disruption:
    budgets:
      - nodes: 5%
        reasons:
          - Drifted
        schedule: "@daily" # customize for your needs (see https://karpenter.sh/docs/concepts/disruption/)
        duration: 10m
    consolidateAfter: 60m
    consolidationPolicy: WhenEmpty
  limits:
    cpu: 64
    memory: 256Gi
---
apiVersion: oci.oraclecloud.com/v1beta1
kind: OCINodeClass
metadata:
  name: my-ocinodeclass
spec:
  shapeConfigs:
    - ocpus: 2
      memoryInGbs: 8
    - ocpus: 4
      memoryInGbs: 16
  volumeConfig:
    bootVolumeConfig:
      imageConfig:
        imageType: OKEImage
        imageFilter:
          osFilter: "Oracle Linux"
          osVersionFilter: "8"  # see OCINodeClass docs for imageFilter behavior
  networkConfig:
    primaryVnicConfig:
      subnetConfig:
        subnetId: <subnet-ocid>
```
## Launch worker nodes for an OciIpNativeCNI cluster

The sample `OCINodeClass` below includes a secondary VNIC configuration. In clusters using the OciIpNativeCNI add-on, worker nodes provisioned by Karpenter will attach a secondary VNIC. All pods will receive a VCN-routable IP address from the secondary VNIC’s subnet, and you can configure the number of allocated IP addresses as needed.
```yaml
---
apiVersion: oci.oraclecloud.com/v1beta1
kind: OCINodeClass
metadata:
  name: my-ocinodeclass
spec:
  shapeConfigs:
    - ocpus: 2  
      memoryInGbs: 8
    - ocpus: 4
      memoryInGbs: 16
  volumeConfig:
    bootVolumeConfig:
      imageConfig:
        imageType: OKEImage
        imageFilter: 
          osFilter: "Oracle Linux"
          osVersionFilter: "8"  # see OCINodeClass docs for imageFilter behavior
  networkConfig:
    primaryVnicConfig:
      subnetConfig:
        subnetId: <subnet-ocid>
    secondaryVnicConfigs:
      - subnetConfig:
          subnetId: <subnet-ocid>  # pod subnet
        ipCount: 16
```

For more examples, see [Advanced Use Cases](advanced-use-cases.md).
