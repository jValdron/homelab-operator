package ensure

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplyPlan describes how desired state relates to the live object (spec comparison only).
type ApplyPlan struct {
	Action    string // create, update, unchanged
	GroupKind string
	Namespace string
	Name      string
	Summary   string
}

// PlanUnstructured compares desired spec with the cluster object. Labels and owner references are not compared.
func PlanUnstructured(ctx context.Context, c client.Client, desired *unstructured.Unstructured) (ApplyPlan, error) {
	gvk := desired.GroupVersionKind()
	plan := ApplyPlan{
		GroupKind: gvk.Kind + "." + gvk.Group,
		Namespace: desired.GetNamespace(),
		Name:      desired.GetName(),
		Summary:   summarizeSpec(desired),
	}

	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(gvk)
	if err := c.Get(ctx, client.ObjectKeyFromObject(desired), current); err != nil {
		if apierrors.IsNotFound(err) || meta.IsNoMatchError(err) {
			plan.Action = "create"
			return plan, nil
		}
		return plan, err
	}

	desiredSpec, dOK, err := unstructured.NestedMap(desired.Object, "spec")
	if err != nil {
		return plan, err
	}
	currentSpec, cOK, err := unstructured.NestedMap(current.Object, "spec")
	if err != nil {
		return plan, err
	}
	if !dOK {
		desiredSpec = nil
	}
	if !cOK {
		currentSpec = nil
	}
	if equality.Semantic.DeepEqual(desiredSpec, currentSpec) {
		plan.Action = "unchanged"
	} else {
		plan.Action = "update"
	}
	return plan, nil
}

func summarizeSpec(u *unstructured.Unstructured) string {
	if hosts, ok, _ := unstructured.NestedSlice(u.Object, "spec", "hosts"); ok {
		return summarizeHostEntries(hosts)
	}
	if options, ok, _ := unstructured.NestedSlice(u.Object, "spec", "options"); ok {
		return fmt.Sprintf("%d dhcp options", len(options))
	}
	if macs, ok, _ := unstructured.NestedStringSlice(u.Object, "spec", "macAddresses"); ok {
		vlan, _, _ := unstructured.NestedFieldNoCopy(u.Object, "spec", "vlan")
		return fmt.Sprintf("freeradius device vlan=%v macs=%d", vlan, len(macs))
	}
	return "empty spec"
}

func summarizeHostEntries(hosts []any) string {
	var names []string
	for _, h := range hosts {
		m, ok := h.(map[string]any)
		if !ok {
			continue
		}
		ip, _ := m["ip"].(string)
		hn, _ := m["hostnames"].([]any)
		if len(hn) > 0 {
			if s, ok := hn[0].(string); ok {
				names = append(names, s+"@"+ip)
				continue
			}
		}
		hostname, _ := m["hostname"].(string)
		if hostname != "" {
			names = append(names, hostname+"@"+ip)
		} else if ip != "" {
			names = append(names, ip)
		}
	}
	const maxShow = 8
	shown := names
	if len(shown) > maxShow {
		shown = shown[:maxShow]
	}
	suffix := ""
	if len(names) > maxShow {
		suffix = fmt.Sprintf(", …+%d", len(names)-maxShow)
	}
	if len(shown) == 0 {
		return fmt.Sprintf("%d entries", len(hosts))
	}
	return fmt.Sprintf("%d entries (%s%s)", len(hosts), strings.Join(shown, ", "), suffix)
}
