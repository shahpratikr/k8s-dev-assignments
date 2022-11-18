package main

import (
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	"github.com/shahpratikr/k8s-dev-assignments/assignment-0/pkg/controller"
	"github.com/shahpratikr/k8s-dev-assignments/assignment-0/pkg/utils"
)

func main() {
	kubernetesClientSet, err := utils.GetClient()
	if err != nil {
		log.Fatal(err)
	}
	startController(kubernetesClientSet)
}

func startController(kubernetesClientSet *kubernetes.Clientset) {
	informerFactory := informers.NewFilteredSharedInformerFactory(
		kubernetesClientSet, 10*time.Minute, "default", func(l0 *metav1.ListOptions) {
			l0.Kind = "deployments"
		})
	c := controller.NewController(kubernetesClientSet, informerFactory)
	informerFactory.Start(make(<-chan struct{}))
	c.Run(make(<-chan struct{}))
}
