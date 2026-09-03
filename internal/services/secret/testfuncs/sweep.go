package secrettestfuncs

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	secretSDK "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_secret", &resource.Sweeper{
		Name: "scaleway_secret",
		F:    testSweepSecret,
	})
}

func testSweepSecret(_ string) error {
	return acctest.SweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		secretAPI := secretSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: deleting the secrets in (%s)", region)

		var projectID *string
		if defaultProjectID, exists := scwClient.GetDefaultProjectID(); exists {
			projectID = &defaultProjectID
		}

		listSecrets, err := secretAPI.ListSecrets(&secretSDK.ListSecretsRequest{
			Region:    region,
			ProjectID: projectID,
		}, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("error listing secrets in (%s) in sweeper: %s", region, err)

			return nil
		}

		for _, se := range listSecrets.Secrets {
			err := secretAPI.DeleteSecret(&secretSDK.DeleteSecretRequest{
				SecretID: se.ID,
				Region:   region,
			})
			if err != nil {
				logging.L.Warningf("error deleting secret in sweeper: %s", err)
			}
		}

		return nil
	})
}
