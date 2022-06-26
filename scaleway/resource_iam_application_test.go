package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_iam_application", &resource.Sweeper{
		Name: "scaleway_iam_application",
		F:    testSweepIamApplication,
	})
}

func testSweepIamApplication(_ string) error {
	return sweep(func(scwClient *scw.Client) error {
		api := iam.NewAPI(scwClient)

		listApps, err := api.ListApplications(&iam.ListApplicationsRequest{})
		if err != nil {
			return fmt.Errorf("failed to list applications: %w", err)
		}
		for _, app := range listApps.Applications {
			err = api.DeleteApplication(&iam.DeleteApplicationRequest{
				ApplicationID: app.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete application: %w", err)
			}
		}
		return nil
	})
}

func TestAccScalewayIamApplication_Basic(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIamApplicationDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iam_application" "main" {
							name = "tf_tests_app_basic"
							description = "a description"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamApplicationExists(tt, "scaleway_iam_application.main"),
					resource.TestCheckResourceAttr("scaleway_iam_application.main", "name", "tf_tests_app_basic"),
					resource.TestCheckResourceAttr("scaleway_iam_application.main", "description", "a description"),
				),
			},
			{
				Config: `
						resource "scaleway_iam_application" "main" {
							name = "tf_tests_app_basic_rename"
							description = "another description"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamApplicationExists(tt, "scaleway_iam_application.main"),
					resource.TestCheckResourceAttr("scaleway_iam_application.main", "name", "tf_tests_app_basic_rename"),
					resource.TestCheckResourceAttr("scaleway_iam_application.main", "description", "another description"),
				),
			},
		},
	})
}

func TestAccScalewayIamApplication_NoUpdate(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIamApplicationDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iam_application" "main" {
							name = "tf_tests_app_noupdate"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamApplicationExists(tt, "scaleway_iam_application.main"),
				),
			},
			{
				Config: `
						resource "scaleway_iam_application" "main" {
							name = "tf_tests_app_noupdate"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamApplicationExists(tt, "scaleway_iam_application.main"),
				),
			},
		},
	})
}

func testAccCheckScalewayIamApplicationExists(tt *TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		iamAPI := iamAPI(tt.Meta)

		_, err := iamAPI.GetApplication(&iam.GetApplicationRequest{
			ApplicationID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("could not find application: %w", err)
		}

		return nil
	}
}

func testAccCheckScalewayIamApplicationDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_iam_application" {
				continue
			}

			iamAPI := iamAPI(tt.Meta)

			_, err := iamAPI.GetApplication(&iam.GetApplicationRequest{
				ApplicationID: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("resource %s(%s) still exist", rs.Type, rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
