package accounttestfuncs

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_account_project", &resource.Sweeper{
		Name: "scaleway_account_project",
		F:    testSweepAccountProject,
	})
}

func testSweepAccountProject(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		accountAPI := accountSDK.NewProjectAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the project")

		req := &accountSDK.ProjectAPIListProjectsRequest{}

		listProjects, err := accountAPI.ListProjects(req, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("failed to list projects: %s", err)

			return nil
		}

		for _, project := range listProjects.Projects {
			// Do not delete default project
			if project.ID == req.OrganizationID || !acctest.IsTestResource(project.Name) {
				continue
			}

			err := accountAPI.DeleteProject(&accountSDK.ProjectAPIDeleteProjectRequest{
				ProjectID: project.ID,
			})
			if err == nil {
				continue
			}

			// If the project still contains resources the API answers 412
			// (resource_still_in_use): fall back to deleting it with its
			// resources.
			if !account.IsProjectStillInUseError(err) {
				logging.L.Debugf("sweeper: failed to delete project %s: %s", project.ID, err)

				continue
			}

			logging.L.Debugf("sweeper: project %s still in use, deleting it with its resources", project.ID)

			_, err = accountAPI.DeleteProjectWithResources(&accountSDK.ProjectAPIDeleteProjectWithResourcesRequest{
				ProjectID:   project.ID,
				ProjectName: project.Name,
			})
			if err != nil {
				logging.L.Debugf("sweeper: failed to delete project %s with resources: %s", project.ID, err)

				continue
			}
		}

		return nil
	})
}
