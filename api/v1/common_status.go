package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// HomelabStatus is shared observed state for homelab resources.
type HomelabStatus struct {
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`

	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

const (
	ConditionReady            = "Ready"
	ConditionNetworkAvailable = "NetworkAvailable"
)
