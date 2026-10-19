package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// IngressRouterSpec defines the desired state of IngressRouter.
type IngressRouterSpec struct {
	IngressClass string `json:"ingressClass,omitempty"`
	IPAddress    string `json:"ipAddress"`
}

// IngressRouterStatus defines the observed state of IngressRouter.
type IngressRouterStatus struct {
	HomelabStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ir

// IngressRouter is the Schema for the ingressrouters API.
type IngressRouter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IngressRouterSpec   `json:"spec,omitempty"`
	Status IngressRouterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IngressRouterList contains a list of IngressRouter.
type IngressRouterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IngressRouter `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &IngressRouter{}, &IngressRouterList{})
		return nil
	})
}
