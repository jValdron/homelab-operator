package dnsmasq

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewDhcpOptionsPreservesDomainSearch(t *testing.T) {
	spec := DhcpOptionsSpec{Options: []DhcpOption{
		{Key: "option:domain-name", Values: []string{"valdron.ca"}},
		{Key: "option:domain-search", Values: []string{"valdron.ca", "int.valdron.ca", "dev.valdron.ca"}},
	}}
	u := NewDhcpOptions("device-based-hosts", "kube-system", spec)
	if _, _, err := unstructured.NestedMap(u.Object, "spec"); err != nil {
		t.Fatalf("NestedMap spec: %v", err)
	}
	opts, ok, err := unstructured.NestedSlice(u.Object, "spec", "options")
	if err != nil || !ok || len(opts) != 2 {
		t.Fatalf("options: ok=%v err=%v len=%d", ok, err, len(opts))
	}
	second, ok := opts[1].(map[string]any)
	if !ok || second["key"] != "option:domain-search" {
		t.Fatalf("second option: %#v", opts[1])
	}
}
