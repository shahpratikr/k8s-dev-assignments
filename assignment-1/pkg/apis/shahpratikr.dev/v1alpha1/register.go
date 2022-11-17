package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/scale/scheme"
)

var SchemeGroupVersion = schema.GroupVersion{
	Group:   "shahpratikr.dev",
	Version: "v1alpha1",
}

var (
	schemeBuilder = scheme.SchemeBuilder
	AddToScheme   = schemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func init() {
	schemeBuilder.Register(addKnownTypes)
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, &SnapshotBackup{}, &SnapshotBackupList{},
		&SnapshotRestore{}, &SnapshotRestoreList{})
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
