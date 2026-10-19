package desired

import (
	"cmp"
	"slices"

	homelabv1 "github.com/jValdron/homelab-operator/api/v1"
	"github.com/jValdron/homelab-operator/internal/dnsmasq"
	networkingv1 "k8s.io/api/networking/v1"
)

const IngressBasedName = "ingress-based"

type HTTPRouteLike struct {
	Gateway   string
	Hostnames []string
}

func BuildIngressBased(routers []homelabv1.IngressRouter, ingresses []networkingv1.Ingress, routes []HTTPRouteLike) dnsmasq.DnsHostsSpec {
	slices.SortFunc(routers, func(a, b homelabv1.IngressRouter) int { return cmp.Compare(a.Name, b.Name) })

	routerByClass := map[string]homelabv1.IngressRouter{}
	for _, ir := range routers {
		if ir.Spec.IngressClass != "" {
			routerByClass[ir.Spec.IngressClass] = ir
		}
	}

	hostsByRouter := map[string]*dnsmasq.DnsHostEntry{}

	getHost := func(routerName, ip string) *dnsmasq.DnsHostEntry {
		if hostsByRouter[routerName] == nil {
			hostsByRouter[routerName] = &dnsmasq.DnsHostEntry{IP: ip, Hostnames: []string{}}
		}
		return hostsByRouter[routerName]
	}

	slices.SortFunc(ingresses, func(a, b networkingv1.Ingress) int {
		return cmp.Compare(a.Namespace+"/"+a.Name, b.Namespace+"/"+b.Name)
	})
	for _, ing := range ingresses {
		class := ingressClass(ing)
		ir, ok := routerByClass[class]
		if !ok {
			continue
		}
		host := getHost(ir.Name, ir.Spec.IPAddress)
		for _, rule := range ing.Spec.Rules {
			if rule.Host != "" {
				host.Hostnames = append(host.Hostnames, rule.Host)
			}
		}
	}

	slices.SortFunc(routes, func(a, b HTTPRouteLike) int { return cmp.Compare(a.Gateway, b.Gateway) })
	for _, hr := range routes {
		ir, ok := routerByClass[hr.Gateway]
		if !ok {
			continue
		}
		host := getHost(ir.Name, ir.Spec.IPAddress)
		host.Hostnames = append(host.Hostnames, hr.Hostnames...)
	}

	spec := dnsmasq.DnsHostsSpec{Hosts: []dnsmasq.DnsHostEntry{}}
	keys := make([]string, 0, len(hostsByRouter))
	for k := range hostsByRouter {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		spec.Hosts = append(spec.Hosts, *hostsByRouter[k])
	}
	return dnsmasq.NormalizeDnsHosts(spec)
}

func ingressClass(ing networkingv1.Ingress) string {
	if ing.Spec.IngressClassName != nil {
		return *ing.Spec.IngressClassName
	}
	return ing.Annotations["kubernetes.io/ingress.class"]
}
