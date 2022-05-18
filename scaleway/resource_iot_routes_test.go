package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iot "github.com/scaleway/scaleway-sdk-go/api/iot/v1"
)

func TestAccScalewayIotRoute_RDB(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		// Destruction is done via the hub destruction.
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayIotHubDestroy(tt),
			testAccCheckScalewayRdbInstanceDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iot_hub" "minimal" {
							name         = "minimal"
							product_plan = "plan_shared"
						}

						resource "scaleway_rdb_instance" "minimal" {
							name           = "minimal"
							node_type      = "db-dev-s"
							engine         = "PostgreSQL-12"
							is_ha_cluster  = false
							disable_backup = true
							user_name      = "root"
							password       = "T3stP4ssw0rdD0N0tUs3!"
						}

						resource "scaleway_iot_route" "default" {
							name   = "default"
							hub_id = scaleway_iot_hub.minimal.id
							topic  = "#"
							database {
								query  = "SELECT NOW()"
								host   = scaleway_rdb_instance.minimal.endpoint_ip
								port   = scaleway_rdb_instance.minimal.endpoint_port
								dbname = "rdb"
								username = scaleway_rdb_instance.minimal.user_name
								password = scaleway_rdb_instance.minimal.password
							}
						}
						`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotRouteExists(tt, "scaleway_iot_route.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_route.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_route.default", "hub_id"),
					resource.TestCheckResourceAttr("scaleway_iot_route.default", "topic", "#"),
					resource.TestCheckResourceAttr("scaleway_iot_route.default", "database.0.query", "SELECT NOW()"),
					resource.TestCheckResourceAttrSet("scaleway_iot_route.default", "database.0.host"),
					resource.TestCheckResourceAttrSet("scaleway_iot_route.default", "database.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_iot_route.default", "database.0.username"),
					resource.TestCheckResourceAttr("scaleway_iot_route.default", "database.0.password", "T3stP4ssw0rdD0N0tUs3!"),
					resource.TestCheckResourceAttr("scaleway_iot_route.default", "database.0.dbname", "rdb"),
				),
			},
		},
	})
}

func TestAccScalewayIotRoute_S3(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := "test-acc-scaleway-iot-route-s3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		// Destruction is done via the hub destruction.
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckScalewayIotHubDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
						resource "scaleway_object_bucket" "minimal" {
							name = "%s"
						}

						resource "scaleway_iot_hub" "minimal" {
							name         = "minimal"
							product_plan = "plan_shared"
						}

						resource "scaleway_iot_route" "default" {
							name   = "default"
							hub_id = scaleway_iot_hub.minimal.id
							topic  = "#"

							s3 {
								bucket_region = scaleway_object_bucket.minimal.region
								bucket_name   = scaleway_object_bucket.minimal.name
								object_prefix = "foo"
								strategy      = "per_topic"
							}
							
							depends_on = [scaleway_object_bucket.minimal]
						}
						`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, bucketName),
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotRouteExists(tt, "scaleway_iot_route.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_route.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_route.default", "hub_id"),
					resource.TestCheckResourceAttr("scaleway_iot_route.default", "topic", "#"),
					resource.TestCheckResourceAttr("scaleway_iot_route.default", "s3.0.bucket_region", "fr-par"),
					resource.TestCheckResourceAttrSet("scaleway_iot_route.default", "s3.0.bucket_name"),
					resource.TestCheckResourceAttr("scaleway_iot_route.default", "s3.0.object_prefix", "foo"),
					resource.TestCheckResourceAttr("scaleway_iot_route.default", "s3.0.strategy", "per_topic"),
				),
			},
		},
	})
}

func TestAccScalewayIotRoute_REST(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		// Destruction is done via the hub destruction.
		CheckDestroy: testAccCheckScalewayIotHubDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iot_hub" "minimal" {
							name         = "minimal"
							product_plan = "plan_shared"
						}

						resource "scaleway_iot_route" "default" {
							name   = "default"
							hub_id = scaleway_iot_hub.minimal.id
							topic  = "#"

							rest {
								verb = "get"
								uri  = "http://scaleway.com"
								headers = {
									X-terraform-test = "inprogress"
								}
							}
						}
						`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotRouteExists(tt, "scaleway_iot_route.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_route.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_route.default", "hub_id"),
					resource.TestCheckResourceAttr("scaleway_iot_route.default", "topic", "#"),
					resource.TestCheckResourceAttr("scaleway_iot_route.default", "rest.0.verb", "get"),
					resource.TestCheckResourceAttr("scaleway_iot_route.default", "rest.0.uri", "http://scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_iot_route.default", "rest.0.headers.X-terraform-test", "inprogress"),
				),
			},
		},
	})
}

func testAccCheckScalewayIotRouteExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		iotAPI, region, routeID, err := iotAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = iotAPI.GetRoute(&iot.GetRouteRequest{
			Region:  region,
			RouteID: routeID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
