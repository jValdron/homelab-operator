package dnsmasq

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewDnsHosts(name, namespace string, spec DnsHostsSpec) *unstructured.Unstructured {
	return newObject(DnsHostsGVK, name, namespace, spec)
}

func NewDhcpHosts(name, namespace string, spec DhcpHostsSpec) *unstructured.Unstructured {
	return newObject(DhcpHostsGVK, name, namespace, spec)
}

func NewDhcpOptions(name, namespace string, spec DhcpOptionsSpec) *unstructured.Unstructured {
	return newObject(DhcpOptionsGVK, name, namespace, spec)
}

func newObject(gvk schema.GroupVersionKind, name, namespace string, spec any) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	u.SetName(name)
	u.SetNamespace(namespace)
	b, _ := json.Marshal(spec)
	var specMap map[string]any
	_ = json.Unmarshal(b, &specMap)
	_ = unstructured.SetNestedMap(u.Object, specMap, "spec")
	return u
}

func ParseDnsHostsSpec(u *unstructured.Unstructured) (DnsHostsSpec, error) {
	return parseSpec[DnsHostsSpec](u)
}

func ParseDhcpHostsSpec(u *unstructured.Unstructured) (DhcpHostsSpec, error) {
	return parseSpec[DhcpHostsSpec](u)
}

func ParseDhcpOptionsSpec(u *unstructured.Unstructured) (DhcpOptionsSpec, error) {
	return parseSpec[DhcpOptionsSpec](u)
}

func parseSpec[T any](u *unstructured.Unstructured) (T, error) {
	var out T
	m, found, err := unstructured.NestedMap(u.Object, "spec")
	if err != nil || !found {
		return out, err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(b, &out)
	return out, err
}
