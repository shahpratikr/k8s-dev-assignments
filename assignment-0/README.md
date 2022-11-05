# assignment-0

# Steps to execute application

- start minikube
```bash
$ minikube start
```

- build binary
```bash
$ go mod tidy
$ go build
```

- start controller
```bash
$ ./assignment-0
```

- create deployment
```bash
$ kubectl create -f deployment.yaml
```

- verify service is created
```bash
$ kubectl get service
```
Note the nodeport value from output

- verify nginx application is accessible
```bash
$ kubectl get nodes -o wide
```
Get internal IP from output
In browser, run IP:NodePort
