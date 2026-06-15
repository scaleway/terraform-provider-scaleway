package opensearch

import (
	"context"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	searchdbapi "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
)

func deploymentSchemaVersion() int {
	return 1
}

func deploymentStateUpgraders() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    deploymentStateUpgradeV0SchemaType(),
			Version: 0,
			Upgrade: deploymentStateUpgradeV0ToV1,
		},
	}
}

func deploymentStateUpgradeV0SchemaType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"node_amount": cty.Number,
		"node_count":  cty.Number,
	})
}

func deploymentStateUpgradeV0ToV1(_ context.Context, rawState map[string]any, _ any) (map[string]any, error) {
	if _, exists := rawState["node_count"]; !exists {
		if value, ok := rawState["node_amount"]; ok {
			rawState["node_count"] = value
		}
	}

	delete(rawState, "node_amount")

	return rawState, nil
}

func deploymentNodeCountForState(d *schema.ResourceData, deployment *searchdbapi.Deployment) int {
	if deployment.NodeCount != 0 {
		return int(deployment.NodeCount)
	}

	// The API may return 0 for node_count while the deployment is ready (notably on shared tiers).
	// Preserve the configured value to avoid spurious ForceNew drift.
	for _, key := range []string{"node_count", "node_amount"} {
		if value, ok := d.GetOk(key); ok {
			if configured := value.(int); configured != 0 {
				return configured
			}
		}
	}

	return 0
}
