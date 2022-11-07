# assignment-0

# Steps to execute application

- start minikube
```bash
$ minikube start
```

- enable volumesnapshots and csi-hostpath-driver addons
```bash
minikube addons enable volumesnapshots
minikube addons enable csi-hostpath-driver
```

- verify volumesnapshotclass is created
```bash
kubectl get volumesnapshotclasses
```

- create CRD
```bash
kubectl create -f manifests/shahpratikr.dev_snapshots.yaml
```

- create PVC and Pod
```bash
kubectl create -f manifests/pvc.yaml -f manifests/pod.yaml 
```

- write dummy data into volume
```bash
kubectl exec -it test -- bash
echo "testing is in progress" > /mnt/data/test.txt
exit
```

- build binary
```bash
go mod tidy
go build
```

- start controller
```bash
./assignment-1
```

- create snapshot instance with CRD
```bash
kubectl create -f manifests/backup-snapshot.yaml 
```

- verify volume snapshot is created and `ReadyToUse` field is set to `true`
```bash
kubectl get volumesnapshot
```

- delete the Pod and PVC
```bash
kubectl delete -f manifests/pod.yaml -f manifests/pvc.yaml 
```

- create CRD with restore operation
```bash
kubectl apply -f manifests/restore-snapshot.yaml
```

- create pod again
```bash
kubectl create -f manifests/pod.yaml
```

- exec into the pod and verify file `/mnt/data/test.txt` is present with data
```bash
kubectl exec -it test -- bash
cat /mnt/data/test.txt
exit
```