package controller

import (
	"cmp"
	"context"
	"slices"

	corev1 "k8s.io/api/core/v1"
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

const deviceAggregateName = "homelab-operator-anchor-device-based"

// DeviceReconciler reconciles namespace-scoped device/network aggregates into dnsmasq CRs.
type DeviceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config config.Config

	nsGate   reconcileGate
	triggers triggerRecorder
}

// +kubebuilder:rbac:groups=homelab.valdron.ca,resources=devices,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=homelab.valdron.ca,resources=devices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=homelab.valdron.ca,resources=networks,verbs=get;list;watch
// +kubebuilder:rbac:groups=homelab.valdron.ca,resources=networks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dnsmasq.kvaps.cf,resources=dnshosts;dhcphosts;dhcpoptions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=freeradius.valdron.ca,resources=devices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

func (r *DeviceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if !r.Config.EnableDeviceBasedHosts {
		return ctrl.Result{}, nil
	}
	return r.nsGate.run(req.Namespace, func() (ctrl.Result, error) {
		return r.reconcileNamespace(ctx, req)
	})
}

func (r *DeviceReconciler) reconcileNamespace(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	ns := req.Namespace
	if tr, ok := r.triggers.consume(ns); ok {
		logger = logger.WithValues(triggerLogValues(tr)...)
	} else if req.Name != "" {
		logger = logger.WithValues("triggerKind", "Network", "trigger", req.Name)
	}
	var devices homelabv1.DeviceList
	if err := r.List(ctx, &devices, client.InNamespace(ns)); err != nil {
		return ctrl.Result{}, err
	}
	var networks homelabv1.NetworkList
	if err := r.List(ctx, &networks, client.InNamespace(ns)); err != nil {
		return ctrl.Result{}, err
	}

	out := desired.BuildDeviceBased(devices.Items, networks.Items, r.Config)
	targetNS := r.Config.DnsmasqResourcesNamespace

	anchor, err := r.ensureAnchor(ctx, targetNS, deviceAggregateName, labels.Managed("device-based", ""))
	if err != nil {
		return ctrl.Result{}, err
	}

	var plans []ensure.ApplyPlan

	plan, err := r.applyDhcpOptions(ctx, targetNS, out.DhcpOptions, anchor)
	if err != nil {
		logger.Error(err, "dhcp options")
		r.updateDeviceStatuses(ctx, devices.Items, networks.Items, false, err.Error())
		return ctrl.Result{}, err
	}
	plans = append(plans, plan)

	desiredNames := map[string]struct{}{desired.DhcpOptionsName: {}}
	for name := range out.DhcpHosts {
		desiredNames[name] = struct{}{}
	}
	for name := range out.DnsHosts {
		desiredNames[name] = struct{}{}
	}

	for name, spec := range out.DhcpHosts {
		plan, err := r.applyDhcpHosts(ctx, targetNS, name, spec, anchor)
		if err != nil {
			return ctrl.Result{}, err
		}
		plans = append(plans, plan)
	}
	for name, spec := range out.DnsHosts {
		plan, err := r.applyDnsHosts(ctx, targetNS, name, spec, anchor)
		if err != nil {
			return ctrl.Result{}, err
		}
		plans = append(plans, plan)
	}

	if r.Config.EnableFreeRadiusDevices {
		frNS := r.Config.FreeRadiusDevicesNamespace
		freeRadiusOwner := client.Object(anchor)
		if frNS != targetNS {
			frAnchor, err := r.ensureAnchor(ctx, frNS, deviceAggregateName, labels.Managed("device-based", ""))
			if err != nil {
				return ctrl.Result{}, err
			}
			freeRadiusOwner = frAnchor
		}
		for name, spec := range out.FreeRadius {
			plan, err := r.applyFreeRadius(ctx, frNS, name, spec, freeRadiusOwner)
			if err != nil {
				return ctrl.Result{}, err
			}
			plans = append(plans, plan)
		}
	}

	if !r.Config.DryRun {
		if err := r.pruneManaged(ctx, targetNS, "device-based", desiredNames); err != nil {
			return ctrl.Result{}, err
		}
		if r.Config.EnableFreeRadiusDevices {
			keepFR := map[string]struct{}{}
			for name := range out.FreeRadius {
				keepFR[name] = struct{}{}
			}
			if err := r.pruneManaged(ctx, r.Config.FreeRadiusDevicesNamespace, "device-based", keepFR, dnsmasq.FreeRadiusDeviceGVK); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	r.updateDeviceStatuses(ctx, devices.Items, networks.Items, true, "")
	r.updateNetworkStatuses(ctx, networks.Items, true, "")

	logger = logger.WithValues(
		"devices", len(devices.Items),
		"networks", len(networks.Items),
	)
	logReconcileSummary(logger, r.Config.DryRun, "device-based", plans)

	return ctrl.Result{}, nil
}

func (r *DeviceReconciler) ensureAnchor(ctx context.Context, namespace, name string, lbl map[string]string) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	cm.Name = name
	cm.Namespace = namespace
	cm.Labels = lbl
	if cm.Data == nil {
		cm.Data = map[string]string{"purpose": "homelab-operator anchor"}
	}
	if r.Config.DryRun {
		return cm, nil
	}
	if _, err := ensure.Apply(ctx, r.Client, cm, false, func() error {
		cm.Labels = lbl
		if cm.Data == nil {
			cm.Data = map[string]string{"purpose": "homelab-operator anchor"}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	// Do not re-Get from the caching client here: parallel reconcilers can create the
	// ConfigMap via the API while this reconcile's Get still sees a stale cache miss.
	return cm, nil
}

func (r *DeviceReconciler) applyDhcpOptions(ctx context.Context, namespace string, spec dnsmasq.DhcpOptionsSpec, owner client.Object) (ensure.ApplyPlan, error) {
	obj := dnsmasq.NewDhcpOptions(desired.DhcpOptionsName, namespace, spec)
	return r.planAndApply(ctx, obj, labels.Managed("device-based", ""), owner)
}

func (r *DeviceReconciler) applyDhcpHosts(ctx context.Context, namespace, name string, spec dnsmasq.DhcpHostsSpec, owner client.Object) (ensure.ApplyPlan, error) {
	obj := dnsmasq.NewDhcpHosts(name, namespace, spec)
	net := networkFromResourceName(name)
	return r.planAndApply(ctx, obj, labels.Managed("device-based", net), owner)
}

func (r *DeviceReconciler) applyDnsHosts(ctx context.Context, namespace, name string, spec dnsmasq.DnsHostsSpec, owner client.Object) (ensure.ApplyPlan, error) {
	obj := dnsmasq.NewDnsHosts(name, namespace, spec)
	net := networkFromResourceName(name)
	return r.planAndApply(ctx, obj, labels.Managed("device-based", net), owner)
}

func (r *DeviceReconciler) planAndApply(ctx context.Context, obj *unstructured.Unstructured, lbl map[string]string, owner client.Object) (ensure.ApplyPlan, error) {
	plan, err := ensure.PlanUnstructured(ctx, r.Client, obj)
	if err != nil {
		return plan, err
	}
	if !r.Config.DryRun {
		_, err = ensure.ApplyUnstructured(ctx, r.Client, obj, false, lbl, owner)
	}
	return plan, err
}

func networkFromResourceName(name string) string {
	const p = "device-based-"
	if len(name) > len(p) && name[:len(p)] == p {
		return name[len(p):]
	}
	return ""
}

func (r *DeviceReconciler) applyFreeRadius(ctx context.Context, namespace, name string, spec map[string]any, owner client.Object) (ensure.ApplyPlan, error) {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(dnsmasq.FreeRadiusDeviceGVK)
	u.SetName(name)
	u.SetNamespace(namespace)
	u.Object["spec"] = spec
	return r.planAndApply(ctx, u, labels.Managed("device-based", ""), owner)
}

func (r *DeviceReconciler) pruneManaged(ctx context.Context, namespace, aggregate string, keep map[string]struct{}, gvks ...schema.GroupVersionKind) error {
	if len(gvks) == 0 {
		gvks = []schema.GroupVersionKind{dnsmasq.DhcpHostsGVK, dnsmasq.DnsHostsGVK}
	}
	for _, gvk := range gvks {
		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(schema.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind + "List"})
		if err := r.List(ctx, list, client.InNamespace(namespace), client.MatchingLabels{
			labels.AggregateKey: aggregate,
			labels.ManagedBy:    labels.ManagedByValue,
		}); err != nil {
			return err
		}
		for _, item := range list.Items {
			if _, ok := keep[item.GetName()]; ok {
				continue
			}
			if r.Config.DryRun {
				continue
			}
			if err := r.Delete(ctx, &item); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *DeviceReconciler) updateDeviceStatuses(ctx context.Context, devices []homelabv1.Device, networks []homelabv1.Network, ok bool, errMsg string) {
	netNames := map[string]struct{}{}
	for _, n := range networks {
		netNames[n.Name] = struct{}{}
	}
	for i := range devices {
		dev := &devices[i]
		origGen := dev.Status.ObservedGeneration
		origConds := append([]metav1.Condition(nil), dev.Status.Conditions...)
		apimeta.SetStatusCondition(&dev.Status.Conditions, metav1.Condition{
			Type:    homelabv1.ConditionNetworkAvailable,
			Status:  metav1.ConditionFalse,
			Reason:  "NetworkMissing",
			Message: "Referenced network does not exist",
		})
		if dev.Spec.Network != "" {
			if _, exists := netNames[dev.Spec.Network]; exists {
				apimeta.SetStatusCondition(&dev.Status.Conditions, metav1.Condition{
					Type:   homelabv1.ConditionNetworkAvailable,
					Status: metav1.ConditionTrue,
					Reason: "NetworkFound",
				})
			}
		} else {
			apimeta.SetStatusCondition(&dev.Status.Conditions, metav1.Condition{
				Type:    homelabv1.ConditionNetworkAvailable,
				Status:  metav1.ConditionFalse,
				Reason:  "NetworkNotSet",
				Message: "spec.network is empty",
			})
		}
		ready := false
		if ok && dev.Spec.Network != "" {
			if _, exists := netNames[dev.Spec.Network]; exists {
				ready = true
			}
		}
		if ready {
			apimeta.SetStatusCondition(&dev.Status.Conditions, metav1.Condition{
				Type:   homelabv1.ConditionReady,
				Status: metav1.ConditionTrue,
				Reason: ConditionReasonSynced,
			})
			now := metav1.Now()
			dev.Status.LastSyncTime = &now
		} else {
			msg := errMsg
			if msg == "" {
				msg = "Device not synced"
			}
			apimeta.SetStatusCondition(&dev.Status.Conditions, metav1.Condition{
				Type:    homelabv1.ConditionReady,
				Status:  metav1.ConditionFalse,
				Reason:  ConditionReasonNotSynced,
				Message: msg,
			})
		}
		dev.Status.ObservedGeneration = dev.Generation
		if origGen == dev.Status.ObservedGeneration && conditionSliceEqual(origConds, dev.Status.Conditions) {
			continue
		}
		if r.Config.DryRun {
			continue
		}
		_ = r.Status().Update(ctx, dev)
	}
}

func (r *DeviceReconciler) updateNetworkStatuses(ctx context.Context, networks []homelabv1.Network, ok bool, errMsg string) {
	for i := range networks {
		net := &networks[i]
		origGen := net.Status.ObservedGeneration
		origConds := append([]metav1.Condition(nil), net.Status.Conditions...)
		if ok {
			apimeta.SetStatusCondition(&net.Status.Conditions, metav1.Condition{
				Type:   homelabv1.ConditionReady,
				Status: metav1.ConditionTrue,
				Reason: ConditionReasonSynced,
			})
			now := metav1.Now()
			net.Status.LastSyncTime = &now
		} else {
			apimeta.SetStatusCondition(&net.Status.Conditions, metav1.Condition{
				Type:    homelabv1.ConditionReady,
				Status:  metav1.ConditionFalse,
				Reason:  ConditionReasonNotSynced,
				Message: errMsg,
			})
		}
		net.Status.ObservedGeneration = net.Generation
		if origGen == net.Status.ObservedGeneration && conditionSliceEqual(origConds, net.Status.Conditions) {
			continue
		}
		if r.Config.DryRun {
			continue
		}
		_ = r.Status().Update(ctx, net)
	}
}

func conditionSliceEqual(a, b []metav1.Condition) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Type != b[i].Type || a[i].Status != b[i].Status || a[i].Reason != b[i].Reason {
			return false
		}
	}
	return true
}

func (r *DeviceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if !r.Config.EnableDeviceBasedHosts {
		return nil
	}

	canonicalNetwork := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		n, ok := obj.(*homelabv1.Network)
		if !ok {
			return false
		}
		req := r.canonicalNetworkRequest(context.Background(), n.Namespace)
		return n.Name == req.Name
	})

	return ctrl.NewControllerManagedBy(mgr).
		Named("device").
		For(&homelabv1.Network{}, builder.WithPredicates(canonicalNetwork)).
		Watches(&homelabv1.Network{}, handler.EnqueueRequestsFromMapFunc(r.enqueueCanonicalNetwork)).
		Watches(&homelabv1.Device{}, handler.EnqueueRequestsFromMapFunc(r.mapDeviceToNetwork)).
		Complete(r)
}

func (r *DeviceReconciler) canonicalNetworkRequest(ctx context.Context, namespace string) reconcile.Request {
	var nets homelabv1.NetworkList
	if err := r.List(ctx, &nets, client.InNamespace(namespace)); err != nil || len(nets.Items) == 0 {
		return reconcile.Request{Namespace: namespace, Name: "default"}
	}
	slices.SortFunc(nets.Items, func(a, b homelabv1.Network) int { return cmp.Compare(a.Name, b.Name) })
	return reconcile.Request{Namespace: namespace, Name: nets.Items[0].Name}
}

func (r *DeviceReconciler) enqueueCanonicalNetwork(ctx context.Context, obj client.Object) []reconcile.Request {
	ns := obj.GetNamespace()
	r.triggers.note(ns, "Network", obj.GetName())
	return []reconcile.Request{r.canonicalNetworkRequest(ctx, ns)}
}

func (r *DeviceReconciler) mapDeviceToNetwork(ctx context.Context, obj client.Object) []reconcile.Request {
	dev, ok := obj.(*homelabv1.Device)
	if !ok {
		return nil
	}
	r.triggers.note(dev.Namespace, "Device", dev.Name)
	return []reconcile.Request{r.canonicalNetworkRequest(ctx, dev.Namespace)}
}
