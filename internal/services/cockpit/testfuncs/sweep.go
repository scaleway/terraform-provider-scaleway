package cockpittestfuncs

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	cockpitSDK "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/cockpit"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_cockpit", &resource.Sweeper{
		Name: "scaleway_cockpit",
		F:    testSweepCockpit,
	})
	resource.AddTestSweepers("scaleway_cockpit_grafana_user", &resource.Sweeper{
		Name: "scaleway_cockpit_grafana_user",
		F:    testSweepCockpitGrafanaUser,
	})
	resource.AddTestSweepers("scaleway_cockpit_token", &resource.Sweeper{
		Name: "scaleway_cockpit_token",
		F:    testSweepCockpitToken,
	})
}

func testSweepCockpitToken(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		accountAPI := accountSDK.NewProjectAPI(scwClient)
		cockpitAPI := cockpitSDK.NewAPI(scwClient)

		listProjects, err := accountAPI.ListProjects(&accountSDK.ProjectAPIListProjectsRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		for _, project := range listProjects.Projects {
			if !strings.HasPrefix(project.Name, "tf_tests") {
				continue
			}

			listTokens, err := cockpitAPI.ListTokens(&cockpitSDK.ListTokensRequest{
				ProjectID: project.ID,
			}, scw.WithAllPages())
			if err != nil {
				if httperrors.Is404(err) {
					return nil
				}

				return fmt.Errorf("failed to list tokens: %w", err)
			}

			for _, token := range listTokens.Tokens {
				err = cockpitAPI.DeleteToken(&cockpitSDK.DeleteTokenRequest{
					TokenID: token.ID,
				})
				if err != nil {
					if !httperrors.Is404(err) {
						return fmt.Errorf("failed to delete token: %w", err)
					}
				}
			}
		}

		return nil
	})
}

func testSweepCockpitGrafanaUser(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		accountAPI := accountSDK.NewProjectAPI(scwClient)
		cockpitAPI := cockpitSDK.NewAPI(scwClient)

		listProjects, err := accountAPI.ListProjects(&accountSDK.ProjectAPIListProjectsRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		for _, project := range listProjects.Projects {
			if !strings.HasPrefix(project.Name, "tf_tests") {
				continue
			}

			listGrafanaUsers, err := cockpitAPI.ListGrafanaUsers(&cockpitSDK.ListGrafanaUsersRequest{
				ProjectID: project.ID,
			}, scw.WithAllPages())
			if err != nil {
				if httperrors.Is404(err) {
					return nil
				}

				return fmt.Errorf("failed to list grafana users: %w", err)
			}

			for _, grafanaUser := range listGrafanaUsers.GrafanaUsers {
				err = cockpitAPI.DeleteGrafanaUser(&cockpitSDK.DeleteGrafanaUserRequest{
					ProjectID:     project.ID,
					GrafanaUserID: grafanaUser.ID,
				})
				if err != nil {
					if !httperrors.Is404(err) {
						return fmt.Errorf("failed to delete grafana user: %w", err)
					}
				}
			}
		}

		return nil
	})
}

func testSweepCockpit(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		accountAPI := accountSDK.NewProjectAPI(scwClient)
		cockpitAPI := cockpitSDK.NewAPI(scwClient)

		listProjects, err := accountAPI.ListProjects(&accountSDK.ProjectAPIListProjectsRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		for _, project := range listProjects.Projects {
			if !strings.HasPrefix(project.Name, "tf_tests") {
				continue
			}

			_, err = cockpitAPI.WaitForCockpit(&cockpitSDK.WaitForCockpitRequest{
				ProjectID: project.ID,
				Timeout:   scw.TimeDurationPtr(cockpit.DefaultCockpitTimeout),
			})
			if err != nil {
				if !httperrors.Is404(err) {
					return fmt.Errorf("failed to deactivate cockpit: %w", err)
				}
			}

			_, err = cockpitAPI.DeactivateCockpit(&cockpitSDK.DeactivateCockpitRequest{
				ProjectID: project.ID,
			})
			if err != nil {
				if !httperrors.Is404(err) {
					return fmt.Errorf("failed to deactivate cockpit: %w", err)
				}
			}
		}

		return nil
	})
}
