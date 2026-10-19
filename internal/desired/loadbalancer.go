package desired

import (
	"cmp"
	"slices"

	"github.com/jValdron/homelab-operator/internal/dnsmasq"
	corev1 "k8s.io/api/core/v1"
)

const LoadBalancerBasedName = "load-balancer-based"

func BuildLoadBalancerBased(services []corev1.Service, suffix string) dnsmasq.DnsHostsSpec {
	slices.SortFunc(services, func(a, b corev1.Service) int {
		return cmp.Compare(a.Namespace+"/"+a.Name, b.Namespace+"/"+b.Name)
	})

	byIP := map[string]*dnsmasq.DnsHostEntry{}
	for _, svc := range services {
		if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
			continue
		}
		ip := svc.Spec.LoadBalancerIP
		if len(svc.Status.LoadBalancer.Ingress) > 0 && svc.Status.LoadBalancer.Ingress[0].IP != "" {
			ip = svc.Status.LoadBalancer.Ingress[0].IP
		}
		if ip == "" {
			continue
		}
		entry, ok := byIP[ip]
		if !ok {
			entry = &dnsmasq.DnsHostEntry{IP: ip, Hostnames: []string{}}
			byIP[ip] = entry
		}
		entry.Hostnames = append(entry.Hostnames, svc.Name+suffix)
	}

	spec := dnsmasq.DnsHostsSpec{Hosts: []dnsmasq.DnsHostEntry{}}
	ips := make([]string, 0, len(byIP))
	for ip := range byIP {
		ips = append(ips, ip)
	}
	slices.Sort(ips)
	for _, ip := range ips {
		spec.Hosts = append(spec.Hosts, *byIP[ip])
	}
	return dnsmasq.NormalizeDnsHosts(spec)
}
