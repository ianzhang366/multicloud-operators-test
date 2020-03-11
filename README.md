# multicloud-operators-test
## Table of Contents
- [Overview](#overview)
- [Installation](#installation)
- [Examples](#examples)
    - [Standalone Example](#standalone-example)
    - [Multicloud Example](#multicloud-example)
- [Documentation](#documentation)

## Overview
This operator is intended to facilitate the testing of MCM operators (channel, subscription, placementrule). The operator will take in a list of resources along with their expected/desired status, and verify that the specified status matches the current state of the resource in a defined cluster. The operator can be extended to multiple uses such as automating longevity testing of operators, pods, or any resources. The operator could also be used to test multiple environments at once, as long as the kubeconfig of the clusters you want to test on are provided.

## Installation
1. Clone this repository
```bash
mkdir -p <project-directory>
cd <project-directory>
git clone https://github.ibm.com/steve-kim-ibm/multicloud-operators-test.git
cd multicloud-operators-test
```

2. a) *For multicloud environments:* Provide the operator with your managed clusters' (or any  clusters you want to test on) kubeconfigs in the [`configs`](./configs/) folder defined under the project repository. **Important:** Make sure the name of the config files match the name of your clusters. For instance, the config file from your cluster `french`, should be named `french` under the `configs` folder. You do not need to include the config of your hub cluster. Once you've included all the config files of your clusters with the approporiate names under the `configs` folder, run the following command to create a configmap called `kubeconfigs` to be mounted onto the operator:
```bash
kubectl create configmap kubeconfigs --from-file=configs/
```

2. b) *For standalone environments:* Create an empty configmap `kubeconfigs` to be mounted onto the operator.
```bash
kubectl create configmap kubeconfigs
```

3. Deploy the CRD and set up the operator
```bash
kubectl apply -f deploy/crds/app.ibm.com_apptests_crd.yaml
kubectl apply -f deploy
```

4. Check that the operator is running 
```bash
% kubectl get deploy
NAME                        READY   UP-TO-DATE   AVAILABLE   AGE
multicloud-operators-test   1/1     1            1           16s
```

## Examples
### Standalone Example
**Requirements**
- 1 Kubernetes Cluster with the `multicloud-operators-test` operator installed

This section will provide an example for testing the MCM subscription operator by following the standalone example outlined in the [multicloud-operators-subscription](https://github.com/IBM/multicloud-operators-subscription) repo.

**Steps**
1. Follow the [Quick Start](https://github.com/IBM/multicloud-operators-subscription/blob/master/README.md#quick-start) instructions provided in the multicloud-operators-subscription repo, applying the `helmrepo-channel` example.

2. Deploy the nginx CR. The test will check to see if there are three nginx backend pods running alongside the nginx-ingress controller. The CR can be seen [here.](./deploy/crs/nginx_cr.yaml)
```bash
kubectl apply -f deploy/crs/nginx_cr.yaml
```

3. Check the status of `nginx-test`
```bash
% kubectl get apptest
NAME         STATUS    AGE
nginx-test   Success   8s
```

**Example of a failed test**

The following is an example of when the test is unsuccessful. We can observe under `status.failedResources` which resources failed the test.
```bash
% kubectl get apptests
NAME         STATUS   AGE
nginx-test   Failed   20s

% kubectl describe apptest
Name:         nginx-test
Namespace:    default
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"app.ibm.com/v1alpha1","kind":"AppTest","metadata":{"annotations":{},"name":"nginx-test","namespace":"default"},"spec":{"res..."
API Version:  app.ibm.com/v1alpha1
Kind:         AppTest
Metadata:
  Creation Timestamp:  2020-01-27T20:21:28Z
  Generation:          2
  Resource Version:    8814408
  Self Link:           /apis/app.ibm.com/v1alpha1/namespaces/default/apptests/nginx-test
  UID:                 99850f01-4142-11ea-9296-00000a101b0f
Spec:
  Resources:
    API Version:  apps/v1
    Cluster:      /
    Desired Status:
      Available Replicas:  0
      Ready Replicas:      1
      Replicas:            1
      Updated Replicas:    1
    Kind:                  Deployment
    Metadata:
      Name:       nginx-ingress-controller
      Namespace:  default
    API Version:  apps/v1
    Cluster:      /
    Desired Status:
      Available Replicas:  3
      Ready Replicas:      3
      Replicas:            3
      Updated Replicas:    3
    Kind:                  Deployment
    Metadata:
      Name:       nginx-ingress-default-backend
      Namespace:  default
Status:
  Failed Resources:
    Cluster:  /
    Current Status:
      Available Replicas:  1
      Ready Replicas:      1
      Replicas:            1
      Updated Replicas:    1
    Desired Status:
      Available Replicas:  0
      Ready Replicas:      1
      Replicas:            1
      Updated Replicas:    1
    Messages:
      currentStatus and desiredStatus does not match.
    Metadata:
      Creation Timestamp:  <nil>
      Name:                nginx-ingress-controller
      Namespace:           default
  Test Status:             Failed
Events:                    <none>
```


### Multicloud Example
**Requirements**
- A multicloud environment with 1 hub cluster and 1 managed cluster with the `multicloud-operators-test`, and `mutlicloud-operators-subscription` operator installed
- **Note:** This example will be checking resources on the hub cluster and one of your managed clusters. Copy the config file of the managed cluster locally and rename it to be the name of the cluster (*see Step 2*)

This section will provide an example for testing MCM operators by checking if the guestbook application was successfully deployed via the `subscription` and `placementrule` operators with the intended results. *Note:* This example test is not an exhaustive test for testing the MCM operators.

**Steps**
1. Deploy the Guestbook CR. The test will check if the subscription is in `Propagated` status, and examine three deployments (the guestbook's frontend, redismaster, and redisslave) to check if they are all running as expected. The full CR can be seen [here](./deploy/crs/guestbook_cr.yaml), and the resources for the Guestbook application can be seen [here.](./deploy/guestbook/) Replace `<name-of-your-cluster>` with the name of your managed cluster with the commands shown below.

```bash
# For MacOS 
sed -i '' 's/MANAGED_CLUSTER/<name-of-your-cluster>/g' deploy/crs/guestbook_cr.yaml
sed -i '' 's/MANAGED_CLUSTER/<name-of-your-cluster>/g' deploy/guestbook/04-placement.yaml

# For Linux
sed -i 's/MANAGED_CLUSTER/<name-of-your-cluster>/g' deploy/crs/guestbook_cr.yaml
sed -i 's/MANAGED_CLUSTER/<name-of-your-cluster>/g' deploy/guestbook/04-placement.yaml

kubectl apply -f deploy/guestbook
kubectl apply -f deploy/crs/guestbook_cr.yaml
```

2. Check the status of  guestbook-test
```bash
% kubectl get apptests
NAME             STATUS    AGE
guestbook-test   Success   15m
```

## Documentation
### AppTest CR Structure
The following YAML structure shows the required fields for a apptest and some of the common optional fields. Your YAML structure needs to include some required fields and values. Depending on your apptest requirements or application management requirements, you might need to include other optional fields and values. The structure for an apptest is the same whether you are deploying to a single cluster or multiple clusters.

The following YAML structure shows the required fields for an application and some of the common optional fields. You can compose the YAML content with any tool.
```yaml
apiVersion: app.ibm.com/v1alpha1
kind: AppTest
metadata:
  name:
  namespace:
  annotations:
  labels:
spec:
  resources:
    - apiVersion: 
      kind: 
      metadata: 
        name:
        namespace:
        annotations:
        labels:
      cluster: 
      currentStatus:
      desiredStatus:
      messages:
status:
  failedResources:
  testStatus:        
```

The following table outlines each required and optional field:

| **Field**                        | **Description** |
|----------------------------------|-----------------|
| apiVersion                       | Required. Set the value to app.ibm.com/v1alpha1. | 
| kind                             | Required. Set the value to AppTest to indicate that the resource is a apptest. |
| metadata.name                    | Required. The name of the AppTest. |
| metadata.namespace               | Optional. The namespace for the AppTest. |
| metadata.annotations             | Optional. The annotations for the AppTest. |
| metadata.labels                  | Optional. The labels for the AppTest. |
| spec.resources                   | Required. An array of custom or Kubernetes resources that you want to check during testing.|
| spec.resources.metadata          | Name and Namespace required. The metadata of the resource that you want to check. |
| spec.resources.cluster           | Required. The cluster you expect the resource to be in. Define hub cluster with "/". |
| spec.resources.desiredStatus     | Required. The desired status of the resource you are testing. You will need to define the specific fields you want to use to compare between the expected and actual status. |
| spec.resources.currentStatus     | Optional. This is used for the operator to store the current status of the resource you defined, as to compare the expected and actual status of the resource being tested. This will be displayed under status.failedResources along with the desiredStatus if the two statuses for the defined fields differ. |
| spec.resources.messages          | Optional. Used by the operator to record messages/reasons for test failure (ex. no defined resource found in cluster, no access, etc) |
| status.failedResources           | Array of resources that failed; reason for failure for each resource is recorded under spec.resources.messages |
| status.testStatus                | Final testStatus (Success or Failed) |

## Connecting to your clusters
**Modifying connections to the clusters**
