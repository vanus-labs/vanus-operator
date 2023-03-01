# Vanus Operator
[![License](https://img.shields.io/badge/License-Apache_2.0-green.svg)](https://github.com/vanus-labs/vanus/blob/main/LICENSE)
[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://golang.org/)

## Table of Contents
- [Overview](#overview)
- [Quick Start](#quick-start)
  - [Deploy Vanus Operator](#deploy-vanus-operator)
  - [Define Your Vanus Cluster](#define-your-vanus-cluster)
  - [Create Vanus Cluster](#create-vanus-cluster)
  - [Delete Vanus Cluster](#delete-vanus-cluster)
  - [UnDeploy Vanus Operator](#undeploy-vanus-operator)

## Overview

Vanus Operator is to manage Vanus service instances deployed on the Kubernetes cluster.
It is built using the [Operator SDK](https://github.com/operator-framework/operator-sdk), which is part of the [Operator Framework](https://github.com/operator-framework/).

### Deploy Vanus Operator

1. To deploy the Vanus Operator on your Kubernetes cluster, please run the following command:

```shell
$ kubectl apply -f https://download.linkall.com/vanus/operator/latest.yml
```

**NOTE:** Default deployment is under the `vanus` namespace. 

2. Use command ```kubectl get pods``` to check the Vanus Operator deploy status like:

```
$ kubectl get pods
NAME                            READY   STATUS    RESTARTS   AGE
vanus-operator-c6dd5f54-zss8t   1/1     Running   0          9s
```

If you find that pod image is not found, run the following command to build a new one locally,
the image tag is specified by the `IMG` parameter.

```shell
$ make docker-build IMG=public.ecr.aws/vanus/operator:latest
```

Now you can use the CRDs provided by Vanus Operator to deploy your Vanus cluster.

### Define Your Vanus Cluster

Vanus Operator provides several CRDs to allow users define their Vanus cluster or Connector, which includes the Vanus, Connector.

1. Check the file ```vanus_v1alpha1_cluster.yaml``` in the ```example``` directory which we put these CR together:
```
apiVersion: core.vanus.ai/v1alpha1
kind: Vanus
metadata:
  name: vanus-cluster
  namespace: vanus
spec:
  # replicas is the number of controllers.
  replicas:
    controller: 3
    store: 3
    trigger: 1
    timer: 2
    gateway: 1
  version: v0.6.0
  # imagePullPolicy is the image pull policy
  imagePullPolicy: IfNotPresent
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 10Gi
```

The yaml defines the Vanus cluster minimization configuration. By default, Controller and Store is three replicas to ensure leader election.

### Create Vanus Cluster

1. Create the Vanus cluster by running:

```shell
$ kubectl apply -f example/vanus_v1alpha1_cluster.yaml
```

**NOTE:** Default creation is under the `vanus` namespace.

Check the status:

```
$ kubectl get pods -owide
NAME                            READY   STATUS    RESTARTS   AGE     IP            NODE       NOMINATED NODE   READINESS GATES
vanus-controller-0              1/1     Running   0          24s     172.17.0.9    minikube   <none>           <none>
vanus-controller-1              1/1     Running   0          18s     172.17.0.11   minikube   <none>           <none>
vanus-controller-2              1/1     Running   0          14s     172.17.0.13   minikube   <none>           <none>
vanus-gateway-95ccb5d95-5c6kc   1/1     Running   0          24s     172.17.0.7    minikube   <none>           <none>
vanus-operator-c6dd5f54-zss8t   1/1     Running   0          2m30s   172.17.0.3    minikube   <none>           <none>
vanus-store-0                   1/1     Running   0          24s     172.17.0.8    minikube   <none>           <none>
vanus-store-1                   1/1     Running   0          18s     172.17.0.10   minikube   <none>           <none>
vanus-store-2                   1/1     Running   0          15s     172.17.0.12   minikube   <none>           <none>
vanus-timer-684646fbcd-87s69    1/1     Running   0          24s     172.17.0.5    minikube   <none>           <none>
vanus-timer-684646fbcd-ck8qr    1/1     Running   0          24s     172.17.0.4    minikube   <none>           <none>
vanus-trigger-98b9c5846-7qcmq   1/1     Running   0          24s     172.17.0.6    minikube   <none>           <none>
```

Using the default yaml, we can see that there are 3 controller Pods and 3 store Pods running on the k8s cluster. In addition, there are 1 trigger Pod, 2 timer Pods and 1 gateway Pod.

2. Install `vsctl` and visit the Vanus cluster

**NOTE:** `vsctl` is a command line tool for Vanus

**Firstly**, download **vsctl**, the command line tool of Vanus.

```shell
curl -O https://download.linkall.com/vsctl/v0.6.0/linux-amd64/vsctl
chmod ug+x vsctl
sudo mv vsctl /usr/local/bin
```

**Secondly**, use `vsctl version` to check the installation

```shell
+-----------+---------------------------------+
|  Version  | v0.6.0                          |
|  Platform | linux/amd64                     |
| GitCommit | 0e2d371                         |
| BuildDate | 2022-11-01_03:47:49+0000        |
| GoVersion | go version go1.18.7 linux/amd64 |
+-----------+---------------------------------+
```

**Then**, set Vanus endpoint

use `minikube service list -n ${VANUS_NAMESPACE}` to get **Vanus Gateway**'s endpoint

```shell
|-----------|------------------|-----------------|---------------------------|
| NAMESPACE |       NAME       |   TARGET PORT   |            URL            |
|-----------|------------------|-----------------|---------------------------|
| vanus     | vanus-controller | No node port    |
| vanus     | vanus-gateway    | put/8080        | http://192.168.49.2:30001 |
|           |                  | get/8081        | http://192.168.49.2:30002 |
|           |                  | ctrl-proxy/8082 | http://192.168.49.2:30003 |
|-----------|------------------|-----------------|---------------------------|
```

**NOTE:** if you are using a normal k8s cluster, just use `kubectl get svc -n ${VANUS_NAMESPACE}` to find the endpoint

use evn variable to tell the endpoint to vsctl

```shell
export VANUS_GATEWAY=192.168.49.2:30001
```

**Finally**, validating if it has connected to Vanus

```shell
vsctl cluster controller topology
```

output should look like

```shell
+-------------------+--------+----------------------------------------------------+
|        NAME       | LEADER |                      ENDPOINT                      |
+-------------------+--------+----------------------------------------------------+
| Leader-controller |  TRUE  | vanus-controller-1.vanus-controller.vanus.svc:2048 |
+-------------------+--------+----------------------------------------------------+
|      Gateway      |    -   |                                                    |
+-------------------+--------+----------------------------------------------------+
```

More information can be found via the [Quick Start](https://docs.linkall.com/getting-started/quick-start)

Congratulations! You have successfully deployed your Vanus cluster by Vanus Operator.

### Delete Vanus Cluster

Delete the Vanus cluster by running:

```shell
$ kubectl delete -f example/vanus_v1alpha1_cluster.yaml
```

### UnDeploy Vanus Operator

To undeploy the Vanus Operator on your Kubernetes cluster, please run the following command:

```shell
$ kubectl delete -f https://download.linkall.com/vanus/operator/latest.yml
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

## License

Copyright 2023 Linkall Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

