package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	homelabv1 "github.com/jValdron/homelab-operator/api/v1"
	"github.com/jValdron/homelab-operator/internal/config"
	"github.com/jValdron/homelab-operator/internal/desired"
	"github.com/jValdron/homelab-operator/internal/dnsmasq"
	"github.com/jValdron/homelab-operator/internal/ensure"
	"github.com/jValdron/homelab-operator/internal/labels"
)

const ingressAnchorName = "homelab-operator-anchor-ingress-based"

// IngressDNSReconciler builds dnshosts/ingress-based from Ingresses and IngressRouters.
type IngressDNSReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config config.Config

	gate reconcileGate
}

// +kubebuilder:rbac:groups=homelab.valdron.ca,resources=ingressrouters,verbs=get;list;watch
// +kubebuilder:rbac:groups=homelab.valdron.ca,resources=ingressrouters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;list;watch
// +kubebuilder:rbac:groups=dnsmasq.kvaps.cf,resources=dnshosts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch

func (r *IngressDNSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if !r.Config.EnableIngressBasedDNSHosts {
		return ctrl.Result{}, nil
	}
	targetNS := r.Config.DnsmasqResourcesNamespace
	return r.gate.run(targetNS, func() (ctrl.Result, error) {
		return r.reconcileIngress(ctx, targetNS)
	})
}

func (r *IngressDNSReconciler) reconcileIngress(ctx context.Context, targetNS string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	anchor, err := r.ensureIngressAnchor(ctx, targetNS)
	if err != nil {
		return ctrl.Result{}, err
	}

	var routers homelabv1.IngressRouterList
	if err := r.List(ctx, &routers); err != nil {
		return ctrl.Result{}, err
	}

	var ingresses networkingv1.IngressList
	if err := r.List(ctx, &ingresses); err != nil {
		return ctrl.Result{}, err
	}

	var routes []desired.HTTPRouteLike
	if r.Config.UseHTTPRoutesForIngress {
		routes, err = r.listHTTPRoutes(ctx)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	spec := desired.BuildIngressBased(routers.Items, ingresses.Items, routes)
	obj := dnsmasq.NewDnsHosts(desired.IngressBasedName, targetNS, spec)
	plan, err := ensure.PlanUnstructured(ctx, r.Client, obj)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !r.Config.DryRun {
		if _, err := ensure.ApplyUnstructured(ctx, r.Client, obj, false, labels.Managed("ingress-based", ""), anchor); err != nil {
			return ctrl.Result{}, err
		}
	}

	r.updateRouterStatuses(ctx, routers.Items, true, "")
	logReconcileSummary(logger, r.Config.DryRun, "ingress-based", []ensure.ApplyPlan{plan})
	return ctrl.Result{}, nil
}

func (r *IngressDNSReconciler) ensureIngressAnchor(ctx context.Context, namespace string) (*corev1.ConfigMap, error) {
	rec := &DeviceReconciler{Client: r.Client, Scheme: r.Scheme, Config: r.Config}
	return rec.ensureAnchor(ctx, namespace, ingressAnchorName, labels.Managed("ingress-based", ""))
}

func (r *IngressDNSReconciler) listHTTPRoutes(ctx context.Context) ([]desired.HTTPRouteLike, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRouteList"})
	if err := r.List(ctx, list); err != nil {
		return nil, err
	}
	var out []desired.HTTPRouteLike
	for _, item := range list.Items {
		parents, _, _ := unstructured.NestedSlice(item.Object, "spec", "parentRefs")
		hosts, _, _ := unstructured.NestedStringSlice(item.Object, "spec", "hostnames")
		if len(parents) == 0 {
			continue
		}
		p0, _ := parents[0].(map[string]any)
		gw, _ := p0["name"].(string)
		out = append(out, desired.HTTPRouteLike{Gateway: gw, Hostnames: hosts})
	}
	return out, nil
}

func (r *IngressDNSReconciler) updateRouterStatuses(ctx context.Context, routers []homelabv1.IngressRouter, ok bool, errMsg string) {
	for i := range routers {
		ir := &routers[i]
		origConds := append([]metav1.Condition(nil), ir.Status.Conditions...)
		if ok {
			apimeta.SetStatusCondition(&ir.Status.Conditions, metav1.Condition{
				Type: homelabv1.ConditionReady, Status: metav1.ConditionTrue, Reason: ConditionReasonSynced,
			})
			now := metav1.Now()
			ir.Status.LastSyncTime = &now
		} else {
			apimeta.SetStatusCondition(&ir.Status.Conditions, metav1.Condition{
				Type: homelabv1.ConditionReady, Status: metav1.ConditionFalse, Reason: ConditionReasonNotSynced, Message: errMsg,
			})
		}
		ir.Status.ObservedGeneration = ir.Generation
		if conditionSliceEqual(origConds, ir.Status.Conditions) {
			continue
		}
		if r.Config.DryRun {
			continue
		}
		_ = r.Status().Update(ctx, ir)
	}
}

func (r *IngressDNSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if !r.Config.EnableIngressBasedDNSHosts {
		return nil
	}

	anchorReq := reconcile.Request{Namespace: r.Config.DnsmasqResourcesNamespace, Name: ingressAnchorName}
	enqueueIngress := handler.EnqueueRequestsFromMapFunc(func(context.Context, client.Object) []reconcile.Request {
		return []reconcile.Request{anchorReq}
	})
	anchorOnly := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		return obj.GetNamespace() == r.Config.DnsmasqResourcesNamespace && obj.GetName() == ingressAnchorName
	})

	httpRoute := &unstructured.Unstructured{}
	httpRoute.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})

	return ctrl.NewControllerManagedBy(mgr).
		Named("ingressdns").
		For(&corev1.ConfigMap{}, builder.WithPredicates(anchorOnly)).
		Watches(&homelabv1.IngressRouter{}, enqueueIngress).
		Watches(&networkingv1.Ingress{}, enqueueIngress).
		Watches(httpRoute, enqueueIngress).
		Complete(r)
}
