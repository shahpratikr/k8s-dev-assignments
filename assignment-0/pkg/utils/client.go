package utils

import (
	"flag"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getConfig() (*rest.Config, error) {
	var (
		config     *rest.Config
		err        error
		clusterErr error
	)
	localKubeconfigPath := homedir.HomeDir() + "/.kube/config"
	kubeconfig := flag.String("kubeconfig", localKubeconfigPath, "location for kubeconfig file")
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		config, clusterErr = rest.InClusterConfig()
		if clusterErr != nil {
			return nil, clusterErr
		}
	}
	return config, nil
}
func GetClient() (*kubernetes.Clientset, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
