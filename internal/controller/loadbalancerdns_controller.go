package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/jValdron/homelab-operator/internal/config"
	"github.com/jValdron/homelab-operator/internal/desired"
	"github.com/jValdron/homelab-operator/internal/dnsmasq"
	"github.com/jValdron/homelab-operator/internal/ensure"
	"github.com/jValdron/homelab-operator/internal/labels"
)

const lbAnchorName = "homelab-operator-anchor-loadbalancer-based"

// LoadBalancerDNSReconciler builds dnshosts/load-balancer-based from LoadBalancer services.
type LoadBalancerDNSReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config config.Config

	gate reconcileGate
}

// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups=dnsmasq.kvaps.cf,resources=dnshosts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch

func (r *LoadBalancerDNSReconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	if !r.Config.EnableLoadBalancerBasedDNSHosts {
		return ctrl.Result{}, nil
	}
	targetNS := r.Config.DnsmasqResourcesNamespace
	return r.gate.run(targetNS, func() (ctrl.Result, error) {
		return r.reconcileLoadBalancer(ctx, targetNS)
	})
}

func (r *LoadBalancerDNSReconciler) reconcileLoadBalancer(ctx context.Context, targetNS string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	rec := &DeviceReconciler{Client: r.Client, Scheme: r.Scheme, Config: r.Config}
	anchor, err := rec.ensureAnchor(ctx, targetNS, lbAnchorName, labels.Managed("load-balancer-based", ""))
	if err != nil {
		return ctrl.Result{}, err
	}

	var services corev1.ServiceList
	if err := r.List(ctx, &services); err != nil {
		return ctrl.Result{}, err
	}

	spec := desired.BuildLoadBalancerBased(services.Items, r.Config.LoadBalancerBasedSuffix)
	obj := dnsmasq.NewDnsHosts(desired.LoadBalancerBasedName, targetNS, spec)
	plan, err := ensure.PlanUnstructured(ctx, r.Client, obj)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !r.Config.DryRun {
		if _, err := ensure.ApplyUnstructured(ctx, r.Client, obj, false, labels.Managed("load-balancer-based", ""), anchor); err != nil {
			return ctrl.Result{}, err
		}
	}

	logReconcileSummary(logger, r.Config.DryRun, "load-balancer-based", []ensure.ApplyPlan{plan})
	return ctrl.Result{}, nil
}

func (r *LoadBalancerDNSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if !r.Config.EnableLoadBalancerBasedDNSHosts {
		return nil
	}

	lbOnly := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			s, ok := e.Object.(*corev1.Service)
			return ok && s.Spec.Type == corev1.ServiceTypeLoadBalancer
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			s, ok := e.ObjectNew.(*corev1.Service)
			return ok && s.Spec.Type == corev1.ServiceTypeLoadBalancer
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			s, ok := e.Object.(*corev1.Service)
			return ok && s.Spec.Type == corev1.ServiceTypeLoadBalancer
		},
		GenericFunc: func(event.GenericEvent) bool { return false },
	}

	anchorReq := reconcile.Request{Namespace: r.Config.DnsmasqResourcesNamespace, Name: lbAnchorName}
	anchorOnly := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		return obj.GetNamespace() == r.Config.DnsmasqResourcesNamespace && obj.GetName() == lbAnchorName
	})

	return ctrl.NewControllerManagedBy(mgr).
		Named("loadbalancerdns").
		For(&corev1.ConfigMap{}, builder.WithPredicates(anchorOnly)).
		Watches(&corev1.Service{}, handler.EnqueueRequestsFromMapFunc(func(context.Context, client.Object) []reconcile.Request {
			return []reconcile.Request{anchorReq}
		}), builder.WithPredicates(lbOnly)).
		Complete(r)
}
