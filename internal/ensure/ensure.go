package ensure

import (
	"context"
	"maps"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Apply creates or updates obj using CreateOrUpdate. When dryRun is true, no API writes occur.
func Apply(ctx context.Context, c client.Client, obj client.Object, dryRun bool, mutate func() error) (controllerutil.OperationResult, error) {
	if dryRun {
		return controllerutil.OperationResultNone, nil
	}
	return controllerutil.CreateOrUpdate(ctx, c, obj, mutate)
}

// ApplyUnstructured applies an unstructured resource with labels and owner refs.
func ApplyUnstructured(ctx context.Context, c client.Client, desired *unstructured.Unstructured, dryRun bool, extraLabels map[string]string, owner metav1.Object) (controllerutil.OperationResult, error) {
	if dryRun {
		return controllerutil.OperationResultNone, nil
	}

	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(desired.GroupVersionKind())
	current.SetName(desired.GetName())
	current.SetNamespace(desired.GetNamespace())

	result, err := controllerutil.CreateOrUpdate(ctx, c, current, func() error {
		spec, found, err := unstructured.NestedMap(desired.Object, "spec")
		if err != nil {
			return err
		}
		if found {
			current.Object["spec"] = spec
		}
		labels := current.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		maps.Copy(labels, extraLabels)
		current.SetLabels(labels)
		if owner != nil {
			if err := controllerutil.SetOwnerReference(owner, current, c.Scheme()); err != nil {
				return err
			}
		}
		return nil
	})
	return result, err
}
