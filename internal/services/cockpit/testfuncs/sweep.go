package cockpittestfuncs

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
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
	resource.AddTestSweepers("scaleway_cockpit_source", &resource.Sweeper{
		Name: "scaleway_cockpit_source",
		F:    testSweepCockpitSource,
	})
}

func testSweepCockpitToken(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		accountAPI := accountSDK.NewProjectAPI(scwClient)
		cockpitAPI := cockpit.NewRegionalAPI(scwClient)

		listProjects, err := accountAPI.ListProjects(&accountSDK.ProjectAPIListProjectsRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		for _, project := range listProjects.Projects {
			if !strings.HasPrefix(project.Name, "tf_tests") {
				continue
			}

			listTokens, err := cockpitAPI.ListTokens(&cockpit.RegionalAPIListTokensRequest{
				ProjectID: project.ID,
			}, scw.WithAllPages())
			if err != nil {
				if httperrors.Is404(err) {
					continue
				}

				return fmt.Errorf("failed to list tokens: %w", err)
			}

			for _, token := range listTokens.Tokens {
				err = cockpitAPI.DeleteToken(&cockpit.RegionalAPIDeleteTokenRequest{
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
		cockpitAPI := cockpit.NewGlobalAPI(scwClient)

		listProjects, err := accountAPI.ListProjects(&accountSDK.ProjectAPIListProjectsRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		for _, project := range listProjects.Projects {
			if !strings.HasPrefix(project.Name, "tf_tests") {
				continue
			}

			listGrafanaUsers, err := cockpitAPI.ListGrafanaUsers(&cockpit.GlobalAPIListGrafanaUsersRequest{ //nolint:staticcheck // legacy Grafana user resource uses deprecated API
				ProjectID: project.ID,
			}, scw.WithAllPages())
			if err != nil {
				if httperrors.Is404(err) {
					continue
				}

				return fmt.Errorf("failed to list grafana users: %w", err)
			}

			for _, grafanaUser := range listGrafanaUsers.GrafanaUsers {
				err = cockpitAPI.DeleteGrafanaUser(&cockpit.GlobalAPIDeleteGrafanaUserRequest{ //nolint:staticcheck // legacy Grafana user resource uses deprecated API
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
	// Cockpit resource doesn't require explicit deactivation.
	// Sources, tokens, and other resources are cleaned up by their respective sweepers.
	return nil
}

func testSweepCockpitSource(_ string) error {
	return acctest.SweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		accountAPI := accountSDK.NewProjectAPI(scwClient)
		cockpitAPI := cockpit.NewRegionalAPI(scwClient)

		listProjects, err := accountAPI.ListProjects(&accountSDK.ProjectAPIListProjectsRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		for _, project := range listProjects.Projects {
			if !strings.HasPrefix(project.Name, "tf_tests") {
				continue
			}

			// Collect all sources by trying without filter and with all possible origins
			// Some sources may only appear when filtering by origin
			allDataSources := make(map[string]*cockpit.DataSource)

			// List of possible origins to try
			origins := []cockpit.DataSourceOrigin{
				cockpit.DataSourceOriginUnknownOrigin,
				cockpit.DataSourceOriginCustom,
				cockpit.DataSourceOriginScaleway,
				cockpit.DataSourceOriginExternal,
			}

			// First, try without any origin filter
			listDatasources, err := cockpitAPI.ListDataSources(&cockpit.RegionalAPIListDataSourcesRequest{
				ProjectID: project.ID,
				Region:    region,
			}, scw.WithAllPages())
			if err != nil {
				if !httperrors.Is404(err) {
					return fmt.Errorf("failed to list sources: %w", err)
				}
			} else {
				// Collect all sources from the unfiltered list
				for _, datasource := range listDatasources.DataSources {
					allDataSources[datasource.ID] = datasource
				}
			}

			// Always try with each origin to ensure we catch all sources
			// Some sources may only appear when filtering by specific origin
			for _, origin := range origins {
				listDatasources, err := cockpitAPI.ListDataSources(&cockpit.RegionalAPIListDataSourcesRequest{
					ProjectID: project.ID,
					Region:    region,
					Origin:    origin,
				}, scw.WithAllPages())
				if err != nil {
					if httperrors.Is404(err) {
						continue
					}

					// Don't return error here, just continue with next origin
					continue
				}

				// Collect all unique sources
				for _, datasource := range listDatasources.DataSources {
					allDataSources[datasource.ID] = datasource
				}
			}

			// Delete all collected sources
			for _, datasource := range allDataSources {
				err = cockpitAPI.DeleteDataSource(&cockpit.RegionalAPIDeleteDataSourceRequest{
					DataSourceID: datasource.ID,
					Region:       region,
				})
				if err != nil {
					if !httperrors.Is404(err) {
						return fmt.Errorf("failed to delete cockpit source: %w", err)
					}
				}
			}
		}

		return nil
	})
}
