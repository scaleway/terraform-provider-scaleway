package edgeservicestestfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	edge "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_edge_services_pipeline", &resource.Sweeper{
		Name: "scaleway_edge_services_pipeline",
		F:    testSweepPipeline,
	})
	resource.AddTestSweepers("scaleway_edge_services_backend_stage", &resource.Sweeper{
		Name: "scaleway_edge_services_backend_stage",
		F:    testSweepBackend,
	})
	resource.AddTestSweepers("scaleway_edge_services_tls_stage", &resource.Sweeper{
		Name: "scaleway_edge_services_tls_stage",
		F:    testSweepTLS,
	})
	resource.AddTestSweepers("scaleway_edge_services_dns_stage", &resource.Sweeper{
		Name: "scaleway_edge_services_dns_stage",
		F:    testSweepDNS,
	})
	resource.AddTestSweepers("scaleway_edge_services_cache_stage", &resource.Sweeper{
		Name: "scaleway_edge_services_cache_stage",
		F:    testSweepCache,
	})
	resource.AddTestSweepers("scaleway_edge_services_plan", &resource.Sweeper{
		Name: "scaleway_edge_services_plan",
		F:    testSweepPlan,
	})
	resource.AddTestSweepers("scaleway_edge_services_waf_stage", &resource.Sweeper{
		Name: "scaleway_edge_services_waf_stage",
		F:    testSweepWAF,
	})
	resource.AddTestSweepers("scaleway_edge_services_route_stage", &resource.Sweeper{
		Name: "scaleway_edge_services_route_stage",
		F:    testSweepRoute,
	})
}

func testSweepPipeline(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		edgeAPI := edge.NewAPI(scwClient)

		listPipelines, err := edgeAPI.ListPipelines(&edge.ListPipelinesRequest{})
		if err != nil {
			return fmt.Errorf("failed to list pipelines: %w", err)
		}

		for _, pipeline := range listPipelines.Pipelines {
			err = edgeAPI.DeletePipeline(&edge.DeletePipelineRequest{
				PipelineID: pipeline.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete pipeline: %w", err)
			}
		}

		return nil
	})
}

func testSweepDNS(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		edgeAPI := edge.NewAPI(scwClient)

		listPipelines, err := edgeAPI.ListPipelines(&edge.ListPipelinesRequest{})
		if err != nil {
			return fmt.Errorf("failed to list pipelines: %w", err)
		}

		for _, pipeline := range listPipelines.Pipelines {
			listDNS, err := edgeAPI.ListDNSStages(&edge.ListDNSStagesRequest{
				PipelineID: pipeline.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to list DNS stages: %w", err)
			}

			for _, stage := range listDNS.Stages {
				err = edgeAPI.DeleteDNSStage(&edge.DeleteDNSStageRequest{
					DNSStageID: stage.ID,
				})
				if err != nil {
					return fmt.Errorf("failed to delete DNS stage: %w", err)
				}
			}
		}

		return nil
	})
}

func testSweepTLS(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		edgeAPI := edge.NewAPI(scwClient)

		listPipelines, err := edgeAPI.ListPipelines(&edge.ListPipelinesRequest{})
		if err != nil {
			return fmt.Errorf("failed to list pipelines: %w", err)
		}

		for _, pipeline := range listPipelines.Pipelines {
			listTLS, err := edgeAPI.ListTLSStages(&edge.ListTLSStagesRequest{
				PipelineID: pipeline.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to list TLS stages: %w", err)
			}

			for _, stage := range listTLS.Stages {
				err = edgeAPI.DeleteTLSStage(&edge.DeleteTLSStageRequest{
					TLSStageID: stage.ID,
				})
				if err != nil {
					return fmt.Errorf("failed to delete TLS stage: %w", err)
				}
			}
		}

		return nil
	})
}

func testSweepCache(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		edgeAPI := edge.NewAPI(scwClient)

		listPipelines, err := edgeAPI.ListPipelines(&edge.ListPipelinesRequest{})
		if err != nil {
			return fmt.Errorf("failed to list pipelines: %w", err)
		}

		for _, pipeline := range listPipelines.Pipelines {
			listCaches, err := edgeAPI.ListCacheStages(&edge.ListCacheStagesRequest{
				PipelineID: pipeline.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to list cache stages: %w", err)
			}

			for _, stage := range listCaches.Stages {
				err = edgeAPI.DeleteCacheStage(&edge.DeleteCacheStageRequest{
					CacheStageID: stage.ID,
				})
				if err != nil {
					return fmt.Errorf("failed to delete cache stage: %w", err)
				}
			}
		}

		return nil
	})
}

func testSweepBackend(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		edgeAPI := edge.NewAPI(scwClient)

		listPipelines, err := edgeAPI.ListPipelines(&edge.ListPipelinesRequest{})
		if err != nil {
			return fmt.Errorf("failed to list pipelines: %w", err)
		}

		for _, pipeline := range listPipelines.Pipelines {
			listBackends, err := edgeAPI.ListBackendStages(&edge.ListBackendStagesRequest{
				PipelineID: pipeline.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to list backend stage: %w", err)
			}

			for _, stage := range listBackends.Stages {
				err = edgeAPI.DeleteBackendStage(&edge.DeleteBackendStageRequest{
					BackendStageID: stage.ID,
				})
				if err != nil {
					return fmt.Errorf("failed to delete backend stage: %w", err)
				}
			}
		}

		return nil
	})
}

func testSweepPlan(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		edgeAPI := edge.NewAPI(scwClient)

		listPipelines, err := edgeAPI.ListPipelines(&edge.ListPipelinesRequest{})
		if err != nil {
			return fmt.Errorf("failed to list pipelines when deleting plan: %w", err)
		}

		for _, pipeline := range listPipelines.Pipelines {
			err = edgeAPI.DeleteCurrentPlan(&edge.DeleteCurrentPlanRequest{
				ProjectID: pipeline.ProjectID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete current plan: %w", err)
			}
		}

		return nil
	})
}

func testSweepWAF(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		edgeAPI := edge.NewAPI(scwClient)

		listPipelines, err := edgeAPI.ListPipelines(&edge.ListPipelinesRequest{})
		if err != nil {
			return fmt.Errorf("failed to list pipelines: %w", err)
		}

		for _, pipeline := range listPipelines.Pipelines {
			listWAF, err := edgeAPI.ListWafStages(&edge.ListWafStagesRequest{})
			if err != nil {
				return fmt.Errorf("failed to list WAF stage: %w", err)
			}

			for _, stage := range listWAF.Stages {
				err = edgeAPI.DeleteWafStage(&edge.DeleteWafStageRequest{
					WafStageID: stage.ID,
				})
				if err != nil {
					return fmt.Errorf("failed to delete WAF stage: %w", err)
				}
			}
		}

		return nil
	})
}

func testSweepRoute(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		edgeAPI := edge.NewAPI(scwClient)

		listPipelines, err := edgeAPI.ListPipelines(&edge.ListPipelinesRequest{})
		if err != nil {
			return fmt.Errorf("failed to list pipelines: %w", err)
		}

		for _, pipeline := range listPipelines.Pipelines {
			listRoutes, err := edgeAPI.ListRouteStages(&edge.ListRouteStagesRequest{
				PipelineID: pipeline.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to list route stage: %w", err)
			}

			for _, stage := range listRoutes.Stages {
				err = edgeAPI.DeleteRouteStage(&edge.DeleteRouteStageRequest{
					RouteStageID: stage.ID,
				})
				if err != nil {
					return fmt.Errorf("failed to delete route stage: %w", err)
				}
			}
		}

		return nil
	})
}
