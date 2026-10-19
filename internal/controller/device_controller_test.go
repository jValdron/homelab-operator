package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	homelabv1 "github.com/jValdron/homelab-operator/api/v1"
	"github.com/jValdron/homelab-operator/internal/config"
)

var _ = Describe("Device Controller", func() {
	const (
		resourceName      = "test-net"
		resourceNamespace = "default"
	)

	ctx := context.Background()
	typeNamespacedName := types.NamespacedName{Name: resourceName, Namespace: resourceNamespace}

	BeforeEach(func() {
		net := &homelabv1.Network{}
		err := k8sClient.Get(ctx, typeNamespacedName, net)
		if err != nil && errors.IsNotFound(err) {
			Expect(k8sClient.Create(ctx, &homelabv1.Network{
				Name: resourceName, Namespace: resourceNamespace,
				Spec: homelabv1.NetworkSpec{
					VLAN: 100, CIDRBlock: "10.0.0.0/24", ServeDHCP: true, Router: "10.0.0.1",
				},
			})).To(Succeed())
		}
	})

	It("reconciles a namespace without error", func() {
		r := &DeviceReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
			Config: config.Config{
				EnableDeviceBasedHosts:    true,
				DnsmasqResourcesNamespace: "default",
				DryRun:                    true,
			},
		}
		_, err := r.Reconcile(ctx, reconcile.Request{Namespace: resourceNamespace, Name: resourceName})
		Expect(err).NotTo(HaveOccurred())
	})
})
