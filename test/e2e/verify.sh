#!/usr/bin/env bash
set -euo pipefail

echo "Waiting for operator deployment..."
kubectl wait --for=condition=Available deployment/homelab-operator-controller-manager \
  -n homelab-operator-system --timeout=180s

echo "Applying sample CRs..."
kubectl apply -k config/samples

echo "Waiting for Device Ready..."
kubectl wait --for=condition=Ready device.homelab.valdron.ca/jpc0 -n default --timeout=120s

echo "Checking dnsmasq DnsHosts for device network..."
kubectl get dnshosts device-based-devices -n default -o yaml | grep -q '10.1.30.2'

echo "Deleting device..."
kubectl delete device.homelab.valdron.ca jpc0 -n default --wait=true

echo "E2E verify passed."
