package account

import (
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func flattenProjects(projects []*accountSDK.Project) []map[string]any {
	flattenedProjects := make([]map[string]any, len(projects))
	for i, project := range projects {
		flattenedProjects[i] = map[string]any{
			"id":              project.ID,
			"name":            project.Name,
			"organization_id": project.OrganizationID,
			"created_at":      types.FlattenTime(project.CreatedAt),
			"updated_at":      types.FlattenTime(project.UpdatedAt),
			"description":     project.Description,
		}
	}

	return flattenedProjects
}
