## Steps

# Initialise application

- Create custom resources
```bash
$ kubectl create -f manifests/shahpratikr.dev_snapshotbackups.yaml -f manifests/shahpratikr.dev_snapshotrestores.yaml
```

- Build binary for the application
```bash
$ go mod tidy
$ go build
```

# Start application
```bash
$ ./assignment-1
```

# Verify application

Execute following steps in a separate terminal window to verify application.

- Create a Pod and a PVC
```bash
$ kubectl create -f manifests/pvc.yaml -f manifests/pod.yaml
```

- Add dummy data into volume
```bash
$ kubectl exec -it test -- bash
$ echo "testing is in progress" > /mnt/data/test.txt
$ exit
```

- Create CR to take a PVC snapshot
```bash
$ kubectl create -f manifests/backup-snapshot.yaml
```

- Verify volume snapshot is created and `ReadyToUse` field is set to `true`
```bash
$ kubectl get volumesnapshot
```

- Delete the Pod and PVC
```bash
$ kubectl delete -f manifests/pod.yaml -f manifests/pvc.yaml 
```

- Set `backupname` field in `./manifests/restore-snapshot.yaml` to name of volumesnapshot

- Create CR to restore volume snapshot
```bash
$ kubectl create -f manifests/restore-snapshot.yaml
```

- Re-create the pod
```bash
$ kubectl create -f manifests/pod.yaml
```

- Exec into the pod and verify file `/mnt/data/test.txt` is present with data
```bash
$ kubectl exec -it test -- bash
$ cat /mnt/data/test.txt
$ exit
```
