package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	volumesnapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	"github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/apis/shahpratikr.dev/v1alpha1"
	snapshotInformer "github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/client/informers/externalversions/shahpratikr.dev/v1alpha1"
	snapshotLister "github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/client/listers/shahpratikr.dev/v1alpha1"
	"github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/utils"
)

type RestoreController struct {
	Clients        *utils.Clients
	SnapshotLister snapshotLister.SnapshotRestoreLister
	Cache          cache.InformerSynced
	Queue          workqueue.RateLimitingInterface
}

func NewRestoreController(clients *utils.Clients,
	snapshotInformer snapshotInformer.SnapshotRestoreInformer) *RestoreController {
	fmt.Println("starting restore controller")
	rc := &RestoreController{
		Clients:        clients,
		SnapshotLister: snapshotInformer.Lister(),
		Cache:          snapshotInformer.Informer().HasSynced,
		Queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
			"myqueue"),
	}
	snapshotInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: rc.handleAdd,
		},
	)
	return rc
}

func (rc *RestoreController) handleAdd(obj interface{}) {
	fmt.Println("handling create snapshot restore")
	rc.Queue.Add(obj)
}

func (rc *RestoreController) RunRestoreController(ch chan struct{}) error {
	if ok := cache.WaitForCacheSync(ch, rc.Cache); !ok {
		fmt.Println("cache was not synced")
	}
	go wait.Until(rc.restoreWorker, time.Second, ch)
	<-ch
	return nil
}

func (rc *RestoreController) restoreWorker() {
	for rc.processItems() {
	}
}

func (rc *RestoreController) processItems() bool {
	item, shutdown := rc.Queue.Get()
	if shutdown {
		return false
	}
	rc.Queue.Forget(item)
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
	err = rc.restoreBackup(bgContext, namespace, name)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func (rc *RestoreController) restoreBackup(bgContext context.Context,
	namespace, name string) error {
	snapshot, err := rc.Clients.SnapshotClientSet.ShahpratikrV1alpha1().SnapshotRestores(
		namespace).Get(bgContext, name, metav1.GetOptions{})
	if err != nil {
		rc.updateStatus(bgContext, FAILED, namespace, name, "", "")
		return err
	}
	rc.updateStatus(bgContext, INPROGRESS, namespace, name, "", "")
	backupCR, err := rc.Clients.SnapshotClientSet.ShahpratikrV1alpha1().SnapshotBackups(
		snapshot.Spec.BackupCRNamespace).Get(
		bgContext, snapshot.Spec.BackupCRName, metav1.GetOptions{})
	if err != nil {
		rc.updateStatus(bgContext, FAILED, namespace, name, "", "")
		return err
	}
	// Volume snapshot doesn't exists, return error
	volumeSnapshot, err := rc.Clients.ExternalSnapshotClientSet.SnapshotV1().VolumeSnapshots(
		backupCR.Status.VolumeSnapshotNamespace).Get(bgContext,
		backupCR.Status.VolumeSnapshotName, metav1.GetOptions{})
	if err != nil {
		rc.updateStatus(bgContext, FAILED, namespace, name, "", "")
		return err
	}
	err = rc.deletePVC()
	if err != nil {
		rc.updateStatus(bgContext, FAILED, namespace, name, "", "")
		return err
	}
	err = rc.createPVC(bgContext, volumeSnapshot, snapshot)
	if err != nil {
		rc.updateStatus(bgContext, FAILED, namespace, name, "", "")
		return err
	}
	rc.updateStatus(bgContext, COMPLETED, namespace, name,
		*volumeSnapshot.Spec.Source.PersistentVolumeClaimName, volumeSnapshot.Namespace)
	fmt.Printf("PVC %s created\n", *volumeSnapshot.Spec.Source.PersistentVolumeClaimName)
	return nil
}

func (rc *RestoreController) deletePVC() error {
	// ToDo: Add automation for recreating PVC via controller
	return nil
}

func (rc *RestoreController) createPVC(bgContext context.Context,
	volumeSnapshot *volumesnapshotv1.VolumeSnapshot, snapshot *v1alpha1.SnapshotRestore) error {
	apiGroup := APIGroup()
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *volumeSnapshot.Spec.Source.PersistentVolumeClaimName,
			Namespace: volumeSnapshot.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &snapshot.Spec.StorageClassname,
			DataSource: &corev1.TypedLocalObjectReference{
				Name:     volumeSnapshot.Name,
				Kind:     "VolumeSnapshot",
				APIGroup: &apiGroup,
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(
						volumeSnapshot.Status.RestoreSize.String()),
				},
			},
		},
	}
	_, err := rc.Clients.KubernetesClientSet.CoreV1().PersistentVolumeClaims(
		volumeSnapshot.Namespace).Create(bgContext, &pvc, metav1.CreateOptions{})
	return err
}

func (rc *RestoreController) updateStatus(bgContext context.Context,
	state, namespace, name, pvcName, pvcNamespace string) error {
	snapshotResource, err := rc.SnapshotLister.SnapshotRestores(namespace).Get(name)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	snapshotDeepcopy := snapshotResource.DeepCopy()
	snapshotDeepcopy.Status.Status = state
	snapshotDeepcopy.Status.PVCName = pvcName
	snapshotDeepcopy.Status.PVCNamespace = pvcNamespace
	_, err = rc.Clients.SnapshotClientSet.ShahpratikrV1alpha1().SnapshotRestores(
		namespace).UpdateStatus(bgContext, snapshotDeepcopy, metav1.UpdateOptions{})
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
