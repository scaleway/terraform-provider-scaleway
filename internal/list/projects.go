package list

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	providerMeta "github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

type ProjectModel interface {
	GetProjects() types.List
}

// ExtractProjects determines project id to query.
// If projects is null, returns the default project ID from the provider config.
// If "*" is passed, all projects will be queried.
func ExtractProjects(ctx context.Context, model ProjectModel, meta *providerMeta.Meta) ([]string, error) {
	var projectsToQuery []string

	projectsList := model.GetProjects()
	if projectsList.IsNull() {
		defaultProjectID, exists := meta.ScwClient().GetDefaultProjectID()
		if !exists {
			return nil, errors.New("no project_ids specified and no default project ID configured")
		}

		return []string{defaultProjectID}, nil
	}

	diags := projectsList.ElementsAs(ctx, &projectsToQuery, false)
	if diags.HasError() {
		return nil, fmt.Errorf("converting projects: %s", diags.Errors()[0].Detail())
	}

	var result []string

	for _, project := range projectsToQuery {
		if project == "*" {
			api := accountSDK.NewProjectAPI(providerMeta.ExtractScwClient(meta))

			res, err := api.ListProjects(new(accountSDK.ProjectAPIListProjectsRequest{}), scw.WithContext(ctx), scw.WithAllPages())
			if err != nil {
				return nil, err
			}

			for _, p := range res.Projects {
				result = append(result, p.ID)
			}

			return result, nil
		}

		result = append(result, project)
	}

	return result, nil
}
