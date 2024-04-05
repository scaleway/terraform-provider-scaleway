package instancetestfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_instance_image", &resource.Sweeper{
		Name:         "scaleway_instance_image",
		Dependencies: []string{"scaleway_instance_server"},
		F:            testSweepImage,
	})
}

func testSweepImage(_ string) error {
	return acctest.SweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		api := instanceSDK.NewAPI(scwClient)
		logging.L.Debugf("sweeper: destroying instance images in (%+v)", zone)

		listImagesResponse, err := api.ListImages(&instanceSDK.ListImagesRequest{
			Zone:   zone,
			Public: scw.BoolPtr(false),
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing instance images in sweeper: %w", err)
		}

		for _, image := range listImagesResponse.Images {
			err := api.DeleteImage(&instanceSDK.DeleteImageRequest{
				Zone:    zone,
				ImageID: image.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting instance image in sweeper: %w", err)
			}
		}

		return nil
	})
}
