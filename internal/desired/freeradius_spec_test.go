package desired

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestFreeRadiusDeviceSpecJSONSafe(t *testing.T) {
	spec := freeRadiusDeviceSpec([]string{"aa:bb:cc:dd:ee:01"}, 40)
	if _, ok := spec["vlan"].(int64); !ok {
		t.Fatalf("vlan must be int64 for unstructured, got %T", spec["vlan"])
	}
	u := &unstructured.Unstructured{}
	u.Object = map[string]any{"spec": spec}
	if _, _, err := unstructured.NestedMap(u.Object, "spec"); err != nil {
		t.Fatalf("NestedMap spec: %v", err)
	}
}
