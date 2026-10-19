package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NetworkSpec defines the desired state of Network.
type NetworkSpec struct {
	Role      string `json:"role,omitempty"`
	VLAN      int    `json:"vlan"`
	CIDRBlock string `json:"cidrBlock"`
	ServeDHCP bool   `json:"serveDhcp"`
	Router    string `json:"router,omitempty"`
	DNSServer string `json:"dnsServer,omitempty"`
}

// NetworkStatus defines the observed state of Network.
type NetworkStatus struct {
	HomelabStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=net

// Network is the Schema for the networks API.
type Network struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkSpec   `json:"spec,omitempty"`
	Status NetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkList contains a list of Network.
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Network `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &Network{}, &NetworkList{})
		return nil
	})
}
