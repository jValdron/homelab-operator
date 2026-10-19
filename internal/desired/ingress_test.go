package desired

import (
	"testing"

	homelabv1 "github.com/jValdron/homelab-operator/api/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

func TestBuildIngressBasedSortsHostnames(t *testing.T) {
	class := "external"
	routers := []homelabv1.IngressRouter{{
		Name: "external",
		Spec: homelabv1.IngressRouterSpec{IngressClass: class, IPAddress: "10.1.21.30"},
	}}
	ingresses := []networkingv1.Ingress{
		{
			Name: "b", Namespace: "ns",
			Spec: networkingv1.IngressSpec{
				IngressClassName: &class,
				Rules:            []networkingv1.IngressRule{{Host: "z.valdron.ca"}, {Host: "a.valdron.ca"}},
			},
		},
		{
			Name: "a", Namespace: "ns",
			Spec: networkingv1.IngressSpec{
				IngressClassName: &class,
				Rules:            []networkingv1.IngressRule{{Host: "m.valdron.ca"}},
			},
		},
	}

	spec := BuildIngressBased(routers, ingresses, nil)
	if len(spec.Hosts) != 1 || len(spec.Hosts[0].Hostnames) != 3 {
		t.Fatalf("unexpected spec: %+v", spec.Hosts)
	}
	want := []string{"a.valdron.ca", "m.valdron.ca", "z.valdron.ca"}
	for i := range want {
		if spec.Hosts[0].Hostnames[i] != want[i] {
			t.Fatalf("hostnames: got %v want %v", spec.Hosts[0].Hostnames, want)
		}
	}
}
