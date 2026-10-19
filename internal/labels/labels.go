package labels

const (
	ManagedBy      = "app.kubernetes.io/managed-by"
	ManagedByValue = "homelab-operator"

	AggregateKey = "homelab.valdron.ca/aggregate"
	NetworkKey   = "homelab.valdron.ca/network"
)

func Managed(aggregate, network string) map[string]string {
	m := map[string]string{
		ManagedBy:    ManagedByValue,
		AggregateKey: aggregate,
	}
	if network != "" {
		m[NetworkKey] = network
	}
	return m
}
