package controller

import (
	"context"
	"fmt"
	"time"

	volumesnapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	externalSnapshotClient "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/apis/shahpratikr.dev/v1alpha1"
	snapshotClient "github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/client/clientset/versioned"
	snapshotInformer "github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/client/informers/externalversions/shahpratikr.dev/v1alpha1"
	snapshotLister "github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/client/listers/shahpratikr.dev/v1alpha1"
	"github.com/shahpratikr/k8s-dev-assignments/assignment-1/pkg/utils"
)

type Controller struct {
	KubernetesClientSet    *kubernetes.Clientset
	SnapshotClientSet      snapshotClient.Interface
	ExternalSnapshotClient externalSnapshotClient.Interface
	StorageClientSet       *storagev1.StorageV1Client
	CacheSynced            cache.InformerSynced
	SnapshotLister         snapshotLister.SnapshotLister
	SnapshotQueue          workqueue.RateLimitingInterface
}

func NewController(clients *utils.Clients,
	snapshotInformer snapshotInformer.SnapshotInformer) *Controller {
	fmt.Println("starting controller")
	c := &Controller{
		KubernetesClientSet:    clients.KubernetesClientSet,
		SnapshotClientSet:      clients.SnapshotClientSet,
		ExternalSnapshotClient: clients.ExternalSnapshotClientSet,
		StorageClientSet:       clients.StorageClientSet,
		CacheSynced:            snapshotInformer.Informer().HasSynced,
		SnapshotLister:         snapshotInformer.Lister(),
		SnapshotQueue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(), "myqueue"),
	}
	snapshotInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			UpdateFunc: c.handleUpdate,
		},
	)
	return c
}

func (c *Controller) handleAdd(obj interface{}) {
	fmt.Println("handling create snapshot")
	c.SnapshotQueue.Add(obj)
}

func (c *Controller) handleUpdate(oldObj interface{}, newObj interface{}) {
	fmt.Println("handling update snapshot")
	c.SnapshotQueue.Add(newObj)
}

func (c *Controller) Run(ch chan struct{}) error {
	if ok := cache.WaitForCacheSync(ch, c.CacheSynced); !ok {
		fmt.Println("cache was not synced")
	}
	go wait.Until(c.worker, time.Second, ch)
	<-ch
	return nil
}

func (c *Controller) worker() {
	for c.processItems() {
	}
}

func (c *Controller) processItems() bool {
	item, shutdown := c.SnapshotQueue.Get()
	if shutdown {
		return false
	}
	c.SnapshotQueue.Forget(item)
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
	snapshot, err := c.SnapshotLister.Snapshots(namespace).Get(name)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	bgContext := context.Background()
	if snapshot.Spec.Operation == "backup" {
		err := c.createBackup(bgContext, snapshot)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	} else if snapshot.Spec.Operation == "restore" {
		err := c.restoreBackup(bgContext, snapshot)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	}
	return true
}

func (c *Controller) createBackup(bgContext context.Context, snapshot *v1alpha1.Snapshot) error {
	// PVC doesn't exists, return error
	pvc, err := c.KubernetesClientSet.CoreV1().PersistentVolumeClaims(
		snapshot.Namespace).Get(bgContext, snapshot.Spec.PVCName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// if volumesnapshot exists, don't create
	_, err = c.ExternalSnapshotClient.SnapshotV1().VolumeSnapshots(snapshot.Namespace).Get(
		bgContext, SnapshotName(pvc), metav1.GetOptions{})
	if err == nil {
		return err
	}
	err = c.createSnapshot(bgContext, pvc, snapshot)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) createSnapshot(bgContext context.Context, pvc *corev1.PersistentVolumeClaim,
	snapshot *v1alpha1.Snapshot) error {
	snapshotName := SnapshotName(pvc)
	snapshotClassname, err := SnapshotClassname(bgContext, *pvc.Spec.StorageClassName,
		c.ExternalSnapshotClient, c.StorageClientSet)
	if err != nil {
		return err
	}
	snapTemplate := &volumesnapshotv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: snapshotName,
		},
		Spec: volumesnapshotv1.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &snapshotClassname,
			Source: volumesnapshotv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvc.Name,
			},
		},
	}
	_, err = c.ExternalSnapshotClient.SnapshotV1().VolumeSnapshots(pvc.Namespace).Create(
		bgContext, snapTemplate, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	err = c.updateStatus(bgContext, snapshotName, pvc, snapshot)
	if err != nil {
		return err
	}
	fmt.Printf("snapshot %s created\n", snapshotName)
	return nil
}

func (c *Controller) updateStatus(bgContext context.Context, snapshotName string,
	pvc *corev1.PersistentVolumeClaim, snapshot *v1alpha1.Snapshot) error {
	snapshotDeepcopy := snapshot.DeepCopy()
	snapshotDeepcopy.Status.VolumeSnapshotName = snapshotName
	snapshotDeepcopy.Status.Storage = pvc.Spec.Resources.Requests.Storage().String()
	snapshotDeepcopy.Status.StorageClassname = *pvc.Spec.StorageClassName
	_, err := c.SnapshotClientSet.ShahpratikrV1alpha1().Snapshots(pvc.Namespace).Update(
		bgContext, snapshotDeepcopy, metav1.UpdateOptions{})
	return err
}

func (c *Controller) restoreBackup(bgContext context.Context, snapshot *v1alpha1.Snapshot) error {
	// Volume snapshot doesn't exists, return error
	volumeSnapshot, err := c.ExternalSnapshotClient.SnapshotV1().VolumeSnapshots(
		snapshot.Namespace).Get(bgContext, snapshot.Status.VolumeSnapshotName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	err = c.createPVC(bgContext, volumeSnapshot, snapshot)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) createPVC(bgContext context.Context,
	volumeSnapshot *volumesnapshotv1.VolumeSnapshot, snapshot *v1alpha1.Snapshot) error {
	apiGroup := APIGroup()
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: *volumeSnapshot.Spec.Source.PersistentVolumeClaimName,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &snapshot.Status.StorageClassname,
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
					corev1.ResourceStorage: resource.MustParse(snapshot.Status.Storage),
				},
			},
		},
	}
	_, err := c.KubernetesClientSet.CoreV1().PersistentVolumeClaims(
		volumeSnapshot.Namespace).Create(bgContext, &pvc, metav1.CreateOptions{})
	return err
}
