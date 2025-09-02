package autoscalingtestfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	autoscaling "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_autoscaling_instance_group", &resource.Sweeper{
		Name: "scaleway_autoscaling_instance_group",
		F:    testSweepInstanceGroup,
	})

	resource.AddTestSweepers("scaleway_autoscaling_instance_template", &resource.Sweeper{
		Name: "scaleway_autoscaling_instance_template",
		F:    testSweepInstanceTemplate,
	})

	resource.AddTestSweepers("scaleway_autoscaling_instance_policy", &resource.Sweeper{
		Name: "scaleway_autoscaling_instance_policy",
		F:    testSweepInstancePolicy,
	})
}

func testSweepInstanceGroup(_ string) error {
	return acctest.SweepZones((&autoscaling.API{}).Zones(), func(scwClient *scw.Client, zone scw.Zone) error {
		autoscalingAPI := autoscaling.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the autoscaling instance groups in (%s)", zone)

		listInstanceGroups, err := autoscalingAPI.ListInstanceGroups(
			&autoscaling.ListInstanceGroupsRequest{
				Zone: zone,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing instancegroup in (%s) in sweeper: %w", zone, err)
		}

		for _, instanceGroup := range listInstanceGroups.InstanceGroups {
			err = autoscalingAPI.DeleteInstanceGroup(&autoscaling.DeleteInstanceGroupRequest{
				InstanceGroupID: instanceGroup.ID,
				Zone:            zone,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%w)", err)

				return fmt.Errorf("error deleting instance group in sweeper: %w", err)
			}
		}

		return nil
	})
}

func testSweepInstanceTemplate(_ string) error {
	return acctest.SweepZones((&autoscaling.API{}).Zones(), func(scwClient *scw.Client, zone scw.Zone) error {
		autoscalingAPI := autoscaling.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the autoscaling instance templates in (%s)", zone)

		listInstanceTemplates, err := autoscalingAPI.ListInstanceTemplates(
			&autoscaling.ListInstanceTemplatesRequest{
				Zone: zone,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing instance templates in (%s) in sweeper: %w", zone, err)
		}

		for _, instanceTemplate := range listInstanceTemplates.InstanceTemplates {
			err = autoscalingAPI.DeleteInstanceTemplate(&autoscaling.DeleteInstanceTemplateRequest{
				TemplateID: instanceTemplate.ID,
				Zone:       zone,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%w)", err)

				return fmt.Errorf("error deleting instance template in sweeper: %w", err)
			}
		}

		return nil
	})
}

func testSweepInstancePolicy(_ string) error {
	return acctest.SweepZones((&autoscaling.API{}).Zones(), func(scwClient *scw.Client, zone scw.Zone) error {
		autoscalingAPI := autoscaling.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the autoscaling instance policies in (%s)", zone)

		listInstancePolicies, err := autoscalingAPI.ListInstancePolicies(&autoscaling.ListInstancePoliciesRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing instance policy in (%s) in sweeper: %w", zone, err)
		}

		for _, instancePolicy := range listInstancePolicies.Policies {
			err = autoscalingAPI.DeleteInstancePolicy(&autoscaling.DeleteInstancePolicyRequest{
				PolicyID: instancePolicy.ID,
				Zone:     zone,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%w)", err)

				return fmt.Errorf("error deleting instance policy in sweeper: %w", err)
			}
		}

		return nil
	})
}
