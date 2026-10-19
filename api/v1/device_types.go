package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeviceSpec defines the desired state of Device.
type DeviceSpec struct {
	Network             string   `json:"network,omitempty"`
	IP                  string   `json:"ip,omitempty"`
	MACAddresses        []string `json:"macAddresses,omitempty"`
	DHCP                bool     `json:"dhcp,omitempty"`
	AdditionalHostnames []string `json:"additionalHostnames,omitempty"`

	Trunked      bool   `json:"trunked,omitempty"`
	Wired        bool   `json:"wired,omitempty"`
	Wireless     bool   `json:"wireless,omitempty"`
	USB          bool   `json:"usb,omitempty"`
	Location     string `json:"location,omitempty"`
	Position     int    `json:"position,omitempty"`
	PDU          string `json:"pdu,omitempty"`
	Switch       string `json:"switch,omitempty"`
	Side         string `json:"side,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty"`
	SerialNumber string `json:"serialNumber,omitempty"`
}

// DeviceStatus defines the observed state of Device.
type DeviceStatus struct {
	HomelabStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=dev
// +kubebuilder:printcolumn:name="Network",type=string,JSONPath=`.spec.network`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// Device is the Schema for the devices API.
type Device struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceSpec   `json:"spec,omitempty"`
	Status DeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DeviceList contains a list of Device.
type DeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Device `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &Device{}, &DeviceList{})
		return nil
	})
}
