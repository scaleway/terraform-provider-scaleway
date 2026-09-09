package annotationstestfuncs

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	annotations "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_annotations_binding", &resource.Sweeper{
		Name: "scaleway_annotations_binding",
		F:    testSweepAnnotationsBinding,
	})
	resource.AddTestSweepers("scaleway_annotations_value", &resource.Sweeper{
		Name: "scaleway_annotations_value",
		F:    testSweepAnnotationsValue,
	})
	resource.AddTestSweepers("scaleway_annotations_key", &resource.Sweeper{
		Name: "scaleway_annotations_key",
		F:    testSweepAnnotationsKey,
	})
}

func testSweepAnnotationsBinding(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		api := annotations.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying test annotation bindings")

		orgID, exists := scwClient.GetDefaultOrganizationID()
		if !exists {
			logging.L.Warningf("sweeper: missing organizationID")

			return nil
		}

		bindingsResp, err := api.ListBindings(&annotations.ListBindingsRequest{
			OrganizationID: orgID,
		}, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("sweeper: failed to list bindings: %s", err)

			return nil
		}

		for _, binding := range bindingsResp.Bindings {
			err := api.DeleteBinding(&annotations.DeleteBindingRequest{
				BindingID: binding.ID,
			})
			if err != nil {
				logging.L.Warningf("sweeper: failed to delete binding %s: %s", binding.ID, err)
			} else {
				logging.L.Debugf("sweeper: deleted binding %s", binding.ID)
			}
		}

		return nil
	})
}

func testSweepAnnotationsValue(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		api := annotations.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying test annotation values")

		orgID, exists := scwClient.GetDefaultOrganizationID()
		if !exists {
			logging.L.Warningf("sweeper: missing organizationID")

			return nil
		}

		valuesResp, err := api.ListValues(&annotations.ListValuesRequest{
			OrganizationID: orgID,
		}, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("sweeper: failed to list annotation values: %s", err)

			return nil
		}

		for _, value := range valuesResp.Values {
			err := api.DeleteValue(&annotations.DeleteValueRequest{
				ValueID: value.ID,
			})
			if err != nil {
				logging.L.Warningf("sweeper: failed to delete value %s (%s): %s", value.Name, value.ID, err)
			} else {
				logging.L.Debugf("sweeper: deleted value %s (%s)", value.Name, value.ID)
			}
		}

		return nil
	})
}

func testSweepAnnotationsKey(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		api := annotations.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying test annotation keys")

		orgID, exists := scwClient.GetDefaultOrganizationID()
		if !exists {
			logging.L.Warningf("sweeper: missing organizationID")

			return nil
		}

		keysResp, err := api.ListKeys(&annotations.ListKeysRequest{
			OrganizationID: orgID,
		}, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("sweeper: failed to list annotation keys: %s", err)

			return nil
		}

		for _, key := range keysResp.Keys {
			err := api.DeleteKey(&annotations.DeleteKeyRequest{
				KeyID: key.ID,
			})
			if err != nil {
				logging.L.Warningf("sweeper: failed to delete key %s: %s", key.ID, err)
			} else {
				logging.L.Debugf("sweeper: deleted key %s", key.ID)
			}
		}

		return nil
	})
}
