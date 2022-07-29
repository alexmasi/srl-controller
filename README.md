# SR Linux Controller for KNE

This is a k8s controller for running and managing SR Linux nodes launched from [openconfig/kne](https://github.com/openconfig/kne) topology.

## Install

To install the latest version of this controller on a cluster referenced in `~/.kube/config` issue the following command:

```bash
# latest version
kubectl apply -k https://github.com/srl-labs/srl-controller/config/default

# specific version
kubectl apply -k https://github.com/srl-labs/srl-controller/config/default?ref=v0.3.1
```

The resources of this controller will be scoped under `srlinux-controller` namespace.

```text
❯ kubectl get all -n srlinux-controller

NAME                                                        READY   STATUS    RESTARTS   AGE
pod/srlinux-controller-controller-manager-c7495dcc7-rbh7m   2/2     Running   0          6m5s

NAME                                                            TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
service/srlinux-controller-controller-manager-metrics-service   ClusterIP   10.96.34.86   <none>        8443/TCP   16m

NAME                                                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/srlinux-controller-controller-manager   1/1     1            1           16m

NAME                                                              DESIRED   CURRENT   READY   AGE
replicaset.apps/srlinux-controller-controller-manager-c7495dcc7   1         1         1       16m
```

### Installing from a repo

The controller can be installed with make directly from the repo:

```text
make deploy IMG=ghcr.io/srl-labs/srl-controller:0.3.1
```

Make sure to check which controller versions are [available](https://github.com/srl-labs/srl-controller/pkgs/container/srl-controller/versions).

## Uninstall

To uninstall the controller from the cluster:

```text
kubectl delete -k https://github.com/srl-labs/srl-controller/config/default
```

## Testing with `kne` & `kind`

To run this controller in a test cluster deployed with [`kne`](https://github.com/openconfig/kne) and [`kind`](https://kind.sigs.k8s.io/) follow the steps outlined in the [KNE repository](https://github.com/openconfig/kne/tree/main/docs).

Once the kne+kind cluster is created, a [demo topology with two SR Linux nodes](https://github.com/openconfig/kne/blob/db5fe5be01a1b6b65bd79e740e2c819c5aeb50b0/examples/srlinux/2node-srl-with-config.pbtxt) may be deployed as follows:

```bash
kne create examples/srlinux/2node-srl-with-config.pbtxt
```

This will deploy the SR Linux nodes and will create k8s services as per the topology configuration. The services will be exposed via MetalLB and can be queried as:

```text
❯ kubectl -n 3node-srlinux get svc
NAME         TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                                      AGE
service-r1   LoadBalancer   10.96.151.84    172.19.0.50   57400:30006/TCP,443:30004/TCP,22:30005/TCP   6m10s
service-r2   LoadBalancer   10.96.34.36     172.19.0.51   443:30010/TCP,22:30011/TCP,57400:30009/TCP   6m9s
```

To connect with SSH to the `r1` node, use `ssh admin@172.19.0.50` command.

## Controller operations

The controller is designed to manage the `Srlinux` custom resource defined with [the following CRD](https://doc.crds.dev/github.com/srl-labs/srl-controller).

The request to create/delete a resource of kind `Srlinux` is typically coming from `openconfig/kne` topology.

### Creation

When a request to create an `Srlinux` resource named `r1` in namespace `ns` comes in, the controller's reconcile loop does the following:

1. Checks if the pods exist within a namespace `ns` with a name `r1`
2. If the pod hasn't been found, then the controller first ensures that the necessary config maps exist in namespace `ns` and creates them otherwise.
3. When config maps are sorted out, the controller schedules a pod with the name `r1` and requeue the request
4. In a requeue run, the pod is now found and the controller updates the status of `Srlinux` resource with the image name that was used in the pod spec.

### Deletion

When a deletion happens on `Srlinux` resource, the reconcile loop does nothing.

### API access

This repo contains a clientset for API access to the `Srlinux` custom resource. Check [kne repo](https://github.com/openconfig/kne/blob/fc195a73035bcbf344791979ca3e067be47a249c/topo/node/srl/srl.go#L46) to see how this can be done.

## Building srl-controller container image

To build `srl-controller` container image execute:

```bash
# don't forget to set the correct tag
make docker-build IMG=ghcr.io/srl-labs/srl-controller:${tag}
```

Then upload the image to the registry.
