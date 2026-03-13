# Configure IAM Policies for KPO to manage OCI resources

KPO requires specific permissions assigned to its **identity principal** in order to manage OCI resources.

When KPO is installed using the Helm chart (running as a Kubernetes deployment), it runs in a Kubernetes namespace and uses **workload identity** to communicate with OCI services. OCI IAM policies are used to grant the necessary permissions to this workload identity principal.

A typical IAM policy statement for KPO uses the following structure:

```text
Allow any-user to <verb> <resource> in <location> where all {
  request.principal.type = 'workload',
  request.principal.namespace = '<namespace-name>',
  request.principal.service_account = '<service-account-name>',
  request.principal.cluster_id = '<cluster-ocid>'
}
```
Where:
- **verb**: The action to permit (varies by OCI resource)
- **resource**: The OCI resource type being authorized
- **location**: The resource compartment, or tenancy for all compartments
- **namespace-name**: The Kubernetes namespace where KPO is deployed (default: default)
- **service-account-name**: The service account used by KPO pods (default: karpenter)
- **cluster-ocid**: The OCID of the OKE cluster

### Basic Policies Required for KPO Operation
```text
Allow any-user to manage instance-family in compartment <compartment-name> where all { ... }
Allow any-user to manage volumes in compartment <compartment-name> where all { ... }
Allow any-user to manage volume-attachments in compartment <compartment-name> where all { ... }
Allow any-user to manage virtual-network-family in compartment <compartment-name> where all { ... }
Allow any-user to inspect compartments in compartment <compartment-name> where all { ... }
```
### Node Registration Policy for KPO-Launched Nodes
Nodes launched by KPO also need `CLUSTER_JOIN` permission to register with the cluster, using the same permission model as self-managed nodes. To enable this, create a dynamic group that matches all instances in the compartment(s) where nodes are launched. Assign the necessary policies for registration based on Oracle's best practices.
For detailed steps, refer to Oracle’s documentation: [Dynamic Group Policy for Self-Managed Nodes](https://docs.oracle.com/en-us/iaas/Content/ContEng/Tasks/contengdynamicgrouppolicyforselfmanagednodes.htm).
- Create a dynamic group to match instances in a compartment that KPO will launch nodes.
    ```text
    ALL {instance.compartment.id = '<compartment-ocid>'}
    ```
- Add IAM policy for instances in the dynamic group to register to the cluster
    ```text
    Allow dynamic-group <domain-name>/<dynamic-group-name> to {CLUSTER_JOIN} in compartment <compartment-name>
    ```
### Optional Policies (Feature-Specific)
Add the following only if specific features are enabled in OCINodeClass:
- Capacity Reservation
    ```text
    Allow any-user to use compute-capacity-reservations in compartment <compartment-name> where all { ... }
    ```
- Compute Cluster
    ```text
    Allow any-user to use compute-clusters in compartment <compartment-name> where all { ... }
    ```
- Cluster Placement Group
    ```text
    Allow any-user to use cluster-placement-groups in compartment <compartment-name> where all { ... }
    ```
- Defined Tags
    ```text
    Allow any-user to use tag-namespaces in compartment <compartment-name> where all {...}
    ```
