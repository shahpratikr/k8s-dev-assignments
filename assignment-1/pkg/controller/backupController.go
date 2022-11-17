package controller

import (
	"context"
	"fmt"
	"time"

	volumesnapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/apis/shahpratikr.dev/v1alpha1"
	snapshotInformer "github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/client/informers/externalversions/shahpratikr.dev/v1alpha1"
	snapshotLister "github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/client/listers/shahpratikr.dev/v1alpha1"
	"github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/utils"
)

type BackupController struct {
	Clients        *utils.Clients
	SnapshotLister snapshotLister.SnapshotBackupLister
	Cache          cache.InformerSynced
	Queue          workqueue.RateLimitingInterface
}

func NewBackupController(clients *utils.Clients,
	snapshotInformer snapshotInformer.SnapshotBackupInformer) *BackupController {
	fmt.Println("starting backup controller")
	bc := &BackupController{
		Clients:        clients,
		SnapshotLister: snapshotInformer.Lister(),
		Cache:          snapshotInformer.Informer().HasSynced,
		Queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
			"myqueue"),
	}
	snapshotInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: bc.handleAdd,
		},
	)
	return bc
}

func (bc *BackupController) handleAdd(obj interface{}) {
	fmt.Println("handling create snapshot backup")
	bc.Queue.Add(obj)
}

func (bc *BackupController) RunBackupController(ch chan struct{}) error {
	if ok := cache.WaitForCacheSync(ch, bc.Cache); !ok {
		fmt.Println("cache was not synced")
	}
	go wait.Until(bc.backupWorker, time.Second, ch)
	<-ch
	return nil
}

func (bc *BackupController) backupWorker() {
	for bc.processItems() {
	}
}

func (bc *BackupController) processItems() bool {
	item, shutdown := bc.Queue.Get()
	if shutdown {
		return false
	}
	bc.Queue.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	bgContext := context.Background()
	err = bc.createBackup(bgContext, namespace, name)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func (bc *BackupController) createBackup(bgContext context.Context, namespace, name string) error {
	snapshot, err := bc.SnapshotLister.SnapshotBackups(namespace).Get(name)
	if err != nil {
		return err
	}
	// PVC doesn't exists, return error
	pvc, err := bc.Clients.KubernetesClientSet.CoreV1().PersistentVolumeClaims(
		snapshot.Spec.PVCNamespace).Get(bgContext, snapshot.Spec.PVCName, metav1.GetOptions{})
	if err != nil {
		bc.updateStatus(bgContext, FAILED, snapshot)
		return err
	}
	bc.updateStatus(bgContext, INPROGRESS, snapshot)
	err = bc.createSnapshot(bgContext, pvc, snapshot)
	if err != nil {
		return err
	}
	return nil
}

func (bc *BackupController) createSnapshot(bgContext context.Context,
	pvc *corev1.PersistentVolumeClaim, snapshot *v1alpha1.SnapshotBackup) error {
	snapshotClassname, err := SnapshotClassname(bgContext, *pvc.Spec.StorageClassName,
		bc.Clients)
	if err != nil {
		bc.updateStatus(bgContext, FAILED, snapshot)
		return err
	}
	snapTemplate := &volumesnapshotv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: SnapshotName(pvc),
			Namespace:    snapshot.Spec.PVCNamespace,
		},
		Spec: volumesnapshotv1.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &snapshotClassname,
			Source: volumesnapshotv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvc.Name,
			},
		},
	}
	_, err = bc.Clients.ExternalSnapshotClientSet.SnapshotV1().VolumeSnapshots(
		pvc.Namespace).Create(bgContext, snapTemplate, metav1.CreateOptions{})
	if err != nil {
		bc.updateStatus(bgContext, FAILED, snapshot)
		return err
	}
	bc.updateStatus(bgContext, COMPLETED, snapshot)
	fmt.Println("snapshot created")
	return nil
}

func (bc *BackupController) updateStatus(bgContext context.Context, state string,
	snapshot *v1alpha1.SnapshotBackup) error {
	snapshotResource, err := bc.SnapshotLister.SnapshotBackups(snapshot.Namespace).Get(
		snapshot.Name)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	snapshotDeepcopy := snapshotResource.DeepCopy()
	snapshotDeepcopy.Status.Progress = state
	_, err = bc.Clients.SnapshotClientSet.ShahpratikrV1alpha1().SnapshotBackups(
		snapshot.Namespace).UpdateStatus(bgContext, snapshotDeepcopy, metav1.UpdateOptions{})
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
