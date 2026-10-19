# homelab-operator

Kubernetes operator that tracks homelab **Devices** and **Networks**, and keeps [dnsmasq-controller](https://github.com/kvaps/dnsmasq-controller) (and optionally [freeradius-operator](https://github.com/jvaldron/freeradius-operator)) configuration in sync.

It also builds cluster-wide DNS entries from **Ingresses** / **IngressRouters** (and optional Gateway API **HTTPRoutes**) and from **LoadBalancer** Services.

## Notice

This comes as-is, with no support and no warranties.

## Background

Built for my homelab, not as a product. The first version was Node/TypeScript: Devices and Networks in the cluster instead of spreadsheets, dnsmasq CRs for DHCP/DNS, optional FreeRADIUS sync. It worked here for a long time; the Go rewrite (kubebuilder / controller-runtime) is the same idea with cleaner watches, leader election, and fewer creaky dependencies. Same `homelab.valdron.ca/v1` CRs for the most part. I used AI heavily on the port, expect rough edges.

## Requirements

- Go 1.27+
- Kubernetes cluster with homelab CRDs installed (`make install`)
- dnsmasq-controller CRDs already present in the cluster (kind/e2e installs them from [config/deps](config/deps))

## Configuration

Each setting can be set with an environment variable (typical in the manager Deployment) or a command-line flag (typical for `make run` / `bin/manager`). Flags default from env when present.

| Environment | CLI flag | Default | Description |
|-------------|----------|---------|-------------|
| `DRY_RUN` | `--dry-run` | false | Run reconcilers (list/compute) but skip all API writes (dnsmasq, FreeRADIUS, status, prune) |
| `DNSMASQ_RESOURCES_NAMESPACE` | `--dnsmasq-resources-namespace` | `default` | Namespace for dnsmasq CRs |
| `DNSMASQ_DNS_SERVER` | `--dnsmasq-dns-server` | — | Default DHCP DNS server option |
| `DNSMASQ_DOMAIN_NAME` | `--dnsmasq-domain-name` | — | DHCP domain-name option |
| `DNSMASQ_DOMAIN_SEARCH` | `--dnsmasq-domain-search` | — | Comma-separated domain-search values; if unset, derived from domain-name plus load-balancer and device suffixes when those are set |
| `ENABLE_DEVICE_BASED_HOSTS` | `--enable-device-based-hosts` | true | Device/Network → dnsmasq controller |
| `ENABLE_DEVICE_SHORT_NAME` | `--enable-device-short-name` | true | Use device name as DNS hostname |
| `DEVICE_SUFFIX` | `--device-suffix` | — | Extra hostname suffix per device |
| `ENABLE_FREERADIUS_DEVICES` | `--enable-freeradius-devices` | false | Sync FreeRADIUS Device CRs |
| `FREERADIUS_DEVICES_NAMESPACE` | `--freeradius-devices-namespace` | dnsmasq NS | Target namespace for FreeRADIUS devices |
| `FREERADIUS_DEVICES_WIRELESS_ONLY` | `--freeradius-devices-wireless-only` | true | Only wireless devices for FreeRADIUS |
| `ENABLE_INGRESS_BASED_DNS_HOSTS` | `--enable-ingress-based-dns-hosts` | true | Ingress → `dnshosts/ingress-based` |
| `USE_HTTP_ROUTES_FOR_INGRESS` | `--use-http-routes-for-ingress` | true | Include Gateway API HTTPRoutes |
| `ENABLE_LOAD_BALANCER_BASED_DNS_HOSTS` | `--enable-load-balancer-based-dns-hosts` | true | LB Services → `dnshosts/load-balancer-based` |
| `LOAD_BALANCER_BASED_SUFFIX` | `--load-balancer-based-suffix` | — | Suffix on LB DNS names |
| `HOMELAB_WATCH_NAMESPACES` | `--homelab-watch-namespaces` | — | Comma-separated; empty = all namespaces |

## Resources

### Device / Network / IngressRouter

API group: `homelab.valdron.ca/v1`. See [config/samples](config/samples/) for examples.

**IngressRouter** uses `spec.ipAddress` (DNS target for that ingress class) and `spec.ingressClass`.

Status includes `Ready`, `NetworkAvailable` (devices), `lastSyncTime`, and `observedGeneration`.

### Generated dnsmasq objects

| Source | dnsmasq resources |
|--------|-------------------|
| Devices/Networks | `dhcpoptions/device-based-hosts`, `device-based-<network>` DhcpHosts/DnsHosts |
| Ingresses | `dnshosts/ingress-based` |
| LoadBalancer Services | `dnshosts/load-balancer-based` |

## Development

```bash
make install          # homelab CRDs into current cluster
make run              # local controller (uses --leader-election-namespace=homelab-operator-system)
                      # override: make run LEADER_ELECTION_NAMESPACE=other; or --leader-elect=false
make test             # unit + envtest
make docker-build IMG=controller:dev
```

### Kind + Skaffold

Requires [kind](https://kind.sigs.k8s.io/) and [skaffold](https://skaffold.dev/).

```bash
make skaffold-dev      # kind cluster + skaffold dev -p dev
make test-e2e          # skaffold run -p e2e + test/e2e/verify.sh
make cleanup-kind      # delete kind cluster (optional)
```

Local kind/e2e dependencies under [config/deps](config/deps): [dnsmasq-controller](https://github.com/jValdron/dnsmasq-controller) CRDs, FreeRADIUS device CRD, and [Gateway API](https://gateway-api.sigs.k8s.io/) Standard Channel CRDs (remote kustomize bases; network required at deploy time).

## Deploy

Published images: `ghcr.io/jvaldron/homelab-operator:latest` (and version tags on git releases).

```bash
kubectl apply -k config/default
```

Image reference is set in [config/manager/kustomization.yaml](config/manager/kustomization.yaml). For a private fork, change `newName` / `newTag` or run:

```bash
cd config/manager && kustomize edit set image controller=ghcr.io/<owner>/homelab-operator:<tag>
```

## License

MIT — see [LICENSE](LICENSE).
