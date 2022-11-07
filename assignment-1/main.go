package main

import (
	"log"

	"time"

	snapshotInformer "github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/client/informers/externalversions"
	"github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/controller"
	"github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/utils"
)

func main() {
	clients, err := utils.GetClients()
	if err != nil {
		log.Fatal(err)
	}
	err = startController(clients)
	if err != nil {
		log.Fatal(err)
	}
}

func startController(clients *utils.Clients) error {
	ch := make(chan struct{})
	informerFactory := snapshotInformer.NewSharedInformerFactory(clients.SnapshotClientSet,
		10*time.Minute)
	c := controller.NewController(clients, informerFactory.Shahpratikr().V1alpha1().Snapshots())
	informerFactory.Start(ch)
	return c.Run(ch)
}
