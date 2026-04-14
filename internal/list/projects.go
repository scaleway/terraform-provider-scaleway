package list

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

type ProjectModel interface {
	GetProjects() types.List
}

// ExtractProjects determines project id to query, if "*" is passed, then all projects
// will be queried
func ExtractProjects(ctx context.Context, model ProjectModel, meta *meta.Meta) ([]string, error) {
	var projectsToQuery []string

	projectsList := model.GetProjects()
	if projectsList.IsNull() {
		return nil, nil
	}

	diags := projectsList.ElementsAs(ctx, &projectsToQuery, false)
	if diags.HasError() {
		return nil, fmt.Errorf("converting projects: %s", diags.Errors()[0].Detail())
	}

	var result []string

	for _, project := range projectsToQuery {
		if project == "*" {
			api := account.NewProjectAPI(meta)

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
