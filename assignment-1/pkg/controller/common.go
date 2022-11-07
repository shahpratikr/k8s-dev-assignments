package controller

import (
	"context"
	"fmt"

	volumesnapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	externalSnapshotClient "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
)

func SnapshotName(pvc *corev1.PersistentVolumeClaim) string {
	return pvc.Name + "-snapshot"
}

func getVolumeSnapshotClasses(bgContext context.Context,
	externalSnapshotClient externalSnapshotClient.Interface) (
	*volumesnapshotv1.VolumeSnapshotClassList, error) {
	return externalSnapshotClient.SnapshotV1().VolumeSnapshotClasses().List(
		bgContext, metav1.ListOptions{})
}

func getStorageClassProvisioner(bgContext context.Context, storageClassName string,
	storageClientSet *storagev1.StorageV1Client) (string, error) {
	storageClass, err := storageClientSet.StorageClasses().Get(bgContext, storageClassName,
		metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return storageClass.Provisioner, nil
}

func SnapshotClassname(bgContext context.Context, storageClassName string,
	externalSnapshotClient externalSnapshotClient.Interface,
	storageClientSet *storagev1.StorageV1Client) (string, error) {
	volumeSnapshotClasses, err := getVolumeSnapshotClasses(bgContext, externalSnapshotClient)
	if err != nil {
		return "", err
	}
	provisioner, err := getStorageClassProvisioner(bgContext, storageClassName, storageClientSet)
	if err != nil {
		return "", err
	}
	for _, snapshotClass := range volumeSnapshotClasses.Items {
		if snapshotClass.Driver == provisioner {
			return snapshotClass.GetObjectMeta().GetName(), nil
		}
	}
	return "", fmt.Errorf("volume snapshot class with %s storageclass not found", storageClassName)
}

func APIGroup() string {
	return "snapshot.storage.k8s.io"
}