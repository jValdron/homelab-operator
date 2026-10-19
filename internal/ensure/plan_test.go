package ensure

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestSummarizeSpecFreeRadiusDevice(t *testing.T) {
	u := &unstructured.Unstructured{}
	u.Object = map[string]any{
		"spec": map[string]any{
			"macAddresses": []any{"aa:bb:cc:dd:ee:01"},
			"vlan":         int64(40),
		},
	}
	got := summarizeSpec(u)
	if got != "freeradius device vlan=40 macs=1" {
		t.Fatalf("unexpected summary: %q", got)
	}
}
