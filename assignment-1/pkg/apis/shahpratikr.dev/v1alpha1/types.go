package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="VolumeSnapshotName",type=string,JSONPath=`.status.volumesnapshotname`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
type SnapshotBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BackupSpec   `json:"spec,omitempty"`
	Status            BackupStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SnapshotBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnapshotBackup `json:"items,omitempty"`
}

type BackupSpec struct {
	PVCName string `json:"pvcname"`
	// +kubebuilder:default=default
	PVCNamespace string `json:"pvcnamespace,omitempty"`
	// +kubebuilder:default=false
	WaitForVolumeSnapshot bool `json:"waitforvolumesnapshot,omitempty"`
}

type BackupStatus struct {
	Status                  string `json:"status,omitempty"`
	VolumeSnapshotName      string `json:"volumesnapshotname,omitempty"`
	VolumeSnapshotNamespace string `json:"volumesnapshotnamespace,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="PVCName",type=string,JSONPath=`.status.pvcname`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
type SnapshotRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RestoreSpec   `json:"spec,omitempty"`
	Status            RestoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SnapshotRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnapshotRestore `json:"items,omitempty"`
}

type RestoreSpec struct {
	// +kubebuilder:default=default
	BackupCRNamespace string `json:"backupcrnamespace,omitempty"`
	BackupCRName      string `json:"backupcrname"`
	StorageClassname  string `json:"storageclassname"`
}

type RestoreStatus struct {
	Status       string `json:"status,omitempty"`
	PVCName      string `json:"pvcname,omitempty"`
	PVCNamespace string `json:"pvcnamespace,omitempty"`
}
