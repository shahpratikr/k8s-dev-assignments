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
	startControllers(clients)
}

func startControllers(clients *utils.Clients) {
	ch := make(chan struct{})
	informerFactory := snapshotInformer.NewSharedInformerFactory(clients.SnapshotClientSet,
		10*time.Minute)
	backupController := controller.NewBackupController(clients,
		informerFactory.Shahpratikr().V1alpha1().SnapshotBackups())
	restoreController := controller.NewRestoreController(clients,
		informerFactory.Shahpratikr().V1alpha1().SnapshotRestores())
	go informerFactory.Start(ch)
	go backupController.RunBackupController(ch)
	restoreController.RunRestoreController(ch)
}
