## assignment-1

# Problem statement

The objective of this assignment is to write a custom controller to create a snapshot of PVC and restore snapshot using Kubernetes Snapshot APIs.

# Pre-requisites

To execute the assignment, we need following packages to be present on system.

- Minikube ([Install Minikube](https://linuxhint.com/install-minikube-ubuntu/#b5))
- Kubectl ([Install Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/#install-kubectl-binary-with-curl-on-linux))

# Setup minikube

To create a volume snapshot, we need to enable `volumesnapshots` and `csi-hostpath-driver` addons in minikube. Run following commands in a terminal to enable these addons.

```bash
$ minikube addons enable volumesnapshots
$ minikube addons enable csi-hostpath-driver
```

Verify volumesnapshotclass is created once these addons are enabled.
```bash
$ kubectl get volumesnapshotclasses
NAME                     DRIVER                DELETIONPOLICY   AGE
csi-hostpath-snapclass   hostpath.csi.k8s.io   Delete           18m
```

Follow steps mentioned in the [doc](docs.md#Steps) to execute the application and verify backup and restore of PVC.