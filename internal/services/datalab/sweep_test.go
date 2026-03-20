package datalab_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	datalabSDK "github.com/scaleway/scaleway-sdk-go/api/datalab/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func init() {
	resource.AddTestSweepers("scaleway_datalab", &resource.Sweeper{
		Name: "scaleway_datalab",
		F:    testSweepDatalab,
	})
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepDatalab(_ string) error {
	return acctest.SweepRegions([]scw.Region{scw.RegionFrPar}, func(scwClient *scw.Client, region scw.Region) error {
		api := datalabSDK.NewAPI(scwClient)

		listResp, err := api.ListDatalabs(&datalabSDK.ListDatalabsRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list datalabs: %w", err)
		}

		for _, dl := range listResp.Datalabs {
			if !acctest.IsTestResource(dl.Name) {
				continue
			}

			_, err := api.DeleteDatalab(&datalabSDK.DeleteDatalabRequest{
				Region:    region,
				DatalabID: dl.ID,
			})
			if err != nil {
				if !httperrors.Is404(err) {
					logging.L.Warningf("sweeper: failed to delete datalab %s: %s", dl.ID, err)
				}
			}
		}

		return nil
	})
}
