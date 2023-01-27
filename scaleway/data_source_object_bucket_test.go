package scaleway

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestAccScalewayDataSourceObjectStorage_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket")
	// resourceName := "data.scaleway_object_bucket.main"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_object_bucket" "base-01" {
					name = "%s"
					tags = {
						foo = "bar"
					}
				}

				data "scaleway_object_bucket" "selected" {
					name = scaleway_object_bucket.base-01.name
				}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_bucket.selected", "name", bucketName),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceObjectStorage_ProjectIDAllowed(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket")

	project, iamAPIKey, terminateFakeSideProject, err := createFakeSideProject(tt)
	require.NoError(t, err)

	ctx := context.Background()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: fakeSideProjectProviders(ctx, tt, project, iamAPIKey),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				return terminateFakeSideProject()
			},
			testAccCheckScalewayObjectDestroy(tt),
		),
		Steps: []resource.TestStep{
			// Create a bucket from the main provider into the side project and read it from the side provider
			// The side provider should only be able to read the bucket from the side project
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base" {
						name = "%[1]s"
						project_id = "%[2]s"
					}

					data "scaleway_object_bucket" "selected" {
						name = scaleway_object_bucket.base.name
						provider = side
					}
				`,
					bucketName,
					project.ID,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_bucket.selected", "name", bucketName),
					resource.TestCheckResourceAttr("data.scaleway_object_bucket.selected", "project_id", project.ID),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceObjectStorage_ProjectIDForbidden(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket")

	project, iamAPIKey, terminateFakeSideProject, err := createFakeSideProject(tt)
	require.NoError(t, err)

	ctx := context.Background()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: fakeSideProjectProviders(ctx, tt, project, iamAPIKey),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				return terminateFakeSideProject()
			},
			testAccCheckScalewayObjectDestroy(tt),
		),
		Steps: []resource.TestStep{
			// The side provider should not be able to read the bucket from the main project
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base" {
						name = "%[1]s"
					}

					data "scaleway_object_bucket" "selected" {
						name = scaleway_object_bucket.base.name
						provider = side
					}
				`,
					bucketName,
					project.ID,
				),
				ExpectError: regexp.MustCompile("failed getting Object Storage bucket"),
			},
		},
	})
}
