package utils

import (
	"flag"

	externalSnapshotClient "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	snapshotClient "github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/client/clientset/versioned"
)

type Clients struct {
	SnapshotClientSet         *snapshotClient.Clientset
	KubernetesClientSet       *kubernetes.Clientset
	ExternalSnapshotClientSet *externalSnapshotClient.Clientset
	StorageClientSet          *storagev1.StorageV1Client
}

func getConfig() (*rest.Config, error) {
	var (
		config *rest.Config
		err    error
	)
	localKubeconfigPath := homedir.HomeDir() + "/.kube/config"
	kubeconfig := flag.String("kubeconfig", localKubeconfigPath, "location for kubeconfig file")
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

func GetClients() (*Clients, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	snapshotClientSet, err := snapshotClient.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	kubernetesClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	externalSnapshotClientSet, err := externalSnapshotClient.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	storageClientSet, err := storagev1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	c := &Clients{
		SnapshotClientSet:         snapshotClientSet,
		KubernetesClientSet:       kubernetesClientSet,
		ExternalSnapshotClientSet: externalSnapshotClientSet,
		StorageClientSet:          storageClientSet,
	}
	return c, nil
}
