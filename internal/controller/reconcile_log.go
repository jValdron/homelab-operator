package controller

import (
	"github.com/go-logr/logr"

	"github.com/jValdron/homelab-operator/internal/ensure"
)

// logReconcileSummary logs one DEBUG line per reconcile and INFO for resources that changed.
func logReconcileSummary(logger logr.Logger, dryRun bool, aggregate string, plans []ensure.ApplyPlan) {
	plans = dedupeApplyPlans(plans)
	if len(plans) == 0 {
		logger.V(1).Info("No managed resources computed for aggregate", "aggregate", aggregate)
		return
	}

	var creates, updates, unchanged int
	for _, p := range plans {
		switch p.Action {
		case "create":
			creates++
		case "update":
			updates++
		default:
			unchanged++
		}
	}

	if dryRun {
		logger.V(1).Info("Dry-run reconcile summary", "aggregate", aggregate,
			"resources", len(plans), "wouldCreate", creates, "wouldUpdate", updates, "unchanged", unchanged)
	} else {
		logger.V(1).Info("Reconcile summary", "aggregate", aggregate,
			"resources", len(plans), "create", creates, "update", updates, "unchanged", unchanged)
	}

	for _, p := range plans {
		if p.Action == "unchanged" {
			continue
		}
		resource := p.GroupKind + "/" + p.Namespace + "/" + p.Name
		if dryRun {
			logger.Info("Dry-run: would change resource", "aggregate", aggregate, "resource", resource, "action", p.Action, "summary", p.Summary)
		} else {
			logger.Info("Changed resource", "aggregate", aggregate, "resource", resource, "action", p.Action, "summary", p.Summary)
		}
	}
}

func dedupeApplyPlans(plans []ensure.ApplyPlan) []ensure.ApplyPlan {
	byResource := map[string]ensure.ApplyPlan{}
	order := make([]string, 0, len(plans))
	for _, p := range plans {
		key := p.GroupKind + "/" + p.Namespace + "/" + p.Name
		if _, seen := byResource[key]; !seen {
			order = append(order, key)
		}
		byResource[key] = p
	}
	out := make([]ensure.ApplyPlan, 0, len(order))
	for _, key := range order {
		out = append(out, byResource[key])
	}
	return out
}
