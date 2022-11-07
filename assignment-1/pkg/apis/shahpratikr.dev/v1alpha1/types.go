package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Snapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SnapshotSpec   `json:"spec,omitempty"`
	Status            SnapshotStatus `json:"status,omitempty"`
}

type SnapshotSpec struct {
	PVCName   string `json:"pvcname,omitempty"`
	Operation string `json:"operation,omitempty"`
}

type SnapshotStatus struct {
	VolumeSnapshotName string `json:"volumesnapshotname,omitempty"`
	Storage            string `json:"storage,omitempty"`
	StorageClassname   string `json:"storageclassname,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Snapshot `json:"items,omitempty"`
}
