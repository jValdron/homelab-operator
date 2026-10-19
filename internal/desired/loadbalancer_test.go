package desired

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestBuildLoadBalancerBasedSortsByIP(t *testing.T) {
	services := []corev1.Service{
		lbService("high", "10.1.21.41"),
		lbService("low", "10.1.21.5"),
		lbService("mid", "10.1.21.15"),
	}
	spec := BuildLoadBalancerBased(services, ".int.example")
	if len(spec.Hosts) != 3 {
		t.Fatalf("expected 3 hosts, got %d", len(spec.Hosts))
	}
	want := []string{"10.1.21.5", "10.1.21.15", "10.1.21.41"}
	for i, ip := range want {
		if spec.Hosts[i].IP != ip {
			t.Fatalf("index %d: got %q want %q", i, spec.Hosts[i].IP, ip)
		}
	}
}

func lbService(name, ip string) corev1.Service {
	return corev1.Service{
		Name: name, Namespace: "ns",
		Spec: corev1.ServiceSpec{
			Type:            corev1.ServiceTypeLoadBalancer,
			LoadBalancerIP:  ip,
			ClusterIP:       "10.96.0.1",
			SessionAffinity: corev1.ServiceAffinityNone,
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{{IP: ip}},
			},
		},
	}
}
