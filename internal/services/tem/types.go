package tem

import (
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func flattenDomainReputation(reputation *tem.DomainReputation) any {
	if reputation == nil {
		return nil
	}

	return []map[string]any{
		{
			"status":             reputation.Status.String(),
			"score":              reputation.Score,
			"scored_at":          types.FlattenTime(reputation.ScoredAt),
			"previous_score":     types.FlattenUint32Ptr(reputation.PreviousScore),
			"previous_scored_at": types.FlattenTime(reputation.PreviousScoredAt),
		},
	}
}

func expandWebhookEventTypes(eventTypesInterface []any) []tem.WebhookEventType {
	eventTypes := make([]tem.WebhookEventType, len(eventTypesInterface))
	for i, v := range eventTypesInterface {
		eventTypes[i] = tem.WebhookEventType(v.(string))
	}

	return eventTypes
}
