package edgeservicestestfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	edge "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/edgeservices"
)

func CheckEdgeServicesPipelineDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_edge_services_pipeline" {
				continue
			}

			edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

			err := edgeAPI.DeletePipeline(&edge.DeletePipelineRequest{
				PipelineID: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("pipeline (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !httperrors.Is404(err) && !httperrors.Is403(err) {
				return err
			}
		}

		return nil
	}
}

func CheckEdgeServicesBackendDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_edge_services_backend_stage" {
				continue
			}

			edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

			err := edgeAPI.DeleteBackendStage(&edge.DeleteBackendStageRequest{
				BackendStageID: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("backend stage (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !httperrors.Is404(err) && !httperrors.Is403(err) {
				return err
			}
		}

		return nil
	}
}

func CheckEdgeServicesDNSDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_edge_services_dns_stage" {
				continue
			}

			edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

			err := edgeAPI.DeleteDNSStage(&edge.DeleteDNSStageRequest{
				DNSStageID: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("DNS stage (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !httperrors.Is404(err) && !httperrors.Is403(err) {
				return err
			}
		}

		return nil
	}
}

func CheckEdgeServicesTLSDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_edge_services_tls_stage" {
				continue
			}

			edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

			err := edgeAPI.DeleteTLSStage(&edge.DeleteTLSStageRequest{
				TLSStageID: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("TLS stage (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !httperrors.Is404(err) && !httperrors.Is403(err) {
				return err
			}
		}

		return nil
	}
}

func CheckEdgeServicesCacheDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_edge_services_cache_stage" {
				continue
			}

			edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

			err := edgeAPI.DeleteCacheStage(&edge.DeleteCacheStageRequest{
				CacheStageID: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("cache stage (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !httperrors.Is404(err) && !httperrors.Is403(err) {
				return err
			}
		}

		return nil
	}
}

func CheckEdgeServicesWAFDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_edge_services_waf_stage" {
				continue
			}

			edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

			err := edgeAPI.DeleteWafStage(&edge.DeleteWafStageRequest{
				WafStageID: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("WAF stage (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !httperrors.Is404(err) && !httperrors.Is403(err) {
				return err
			}
		}

		return nil
	}
}

func CheckEdgeServicesRouteDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_edge_services_route_stage" {
				continue
			}

			edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

			err := edgeAPI.DeleteRouteStage(&edge.DeleteRouteStageRequest{
				RouteStageID: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("route stage (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !httperrors.Is404(err) && !httperrors.Is403(err) {
				return err
			}
		}

		return nil
	}
}

func CheckEdgeServicesPipelineExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

		_, err := edgeAPI.GetPipeline(&edge.GetPipelineRequest{
			PipelineID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func CheckEdgeServicesBackendExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

		_, err := edgeAPI.GetBackendStage(&edge.GetBackendStageRequest{
			BackendStageID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func CheckEdgeServicesCacheExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

		_, err := edgeAPI.GetCacheStage(&edge.GetCacheStageRequest{
			CacheStageID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func CheckEdgeServicesDNSExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

		_, err := edgeAPI.GetDNSStage(&edge.GetDNSStageRequest{
			DNSStageID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func CheckEdgeServicesTLSExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

		_, err := edgeAPI.GetTLSStage(&edge.GetTLSStageRequest{
			TLSStageID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func CheckEdgeServicesWAFExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

		_, err := edgeAPI.GetWafStage(&edge.GetWafStageRequest{
			WafStageID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func CheckEdgeServicesRouteExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		edgeAPI := edgeservices.NewEdgeServicesAPI(tt.Meta)

		_, err := edgeAPI.GetRouteStage(&edge.GetRouteStageRequest{
			RouteStageID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
