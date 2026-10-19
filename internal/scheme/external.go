package scheme

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/jValdron/homelab-operator/internal/dnsmasq"
)

// AddExternal registers GVKs for third-party CRDs accessed as unstructured objects.
func AddExternal(s *runtime.Scheme) error {
	gvks := []schema.GroupVersionKind{
		dnsmasq.DnsHostsGVK,
		dnsmasq.DhcpHostsGVK,
		dnsmasq.DhcpOptionsGVK,
		dnsmasq.FreeRadiusDeviceGVK,
		{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"},
	}
	for _, gvk := range gvks {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		s.AddKnownTypeWithName(gvk, u)
		listGVK := gvk.GroupVersion().WithKind(gvk.Kind + "List")
		ul := &unstructured.UnstructuredList{}
		ul.SetGroupVersionKind(listGVK)
		s.AddKnownTypeWithName(listGVK, ul)
	}
	return nil
}
