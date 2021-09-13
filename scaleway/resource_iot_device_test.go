package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iot "github.com/scaleway/scaleway-sdk-go/api/iot/v1"
)

const customDevCert = `-----BEGIN CERTIFICATE-----
MIIDkjCCAnqgAwIBAgIUJ/Xxw1ucfzPkzOTQFeojiTd9WDMwDQYJKoZIhvcNAQEL
BQAwRzELMAkGA1UEBhMCRlIxDjAMBgNVBAcTBVBhcmlzMREwDwYDVQQKEwhTY2Fs
ZXdheTEVMBMGA1UECxMMc2NhbGV3YXkuY29tMB4XDTIxMDkwOTEzMDYwMFoXDTMx
MDkwNzEzMDYwMFowRzELMAkGA1UEBhMCRlIxDjAMBgNVBAcTBVBhcmlzMREwDwYD
VQQKEwhTY2FsZXdheTEVMBMGA1UECxMMc2NhbGV3YXkuY29tMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvJ7LJtcEM9FCv3+4ZCIbB9p+0Nx0YbOZhGh1
X3cdALXoZ803uOvvO86rkYB73flHvRsSCgKyhD3uw2c4r3QJslRvljFYg2jAb5c0
A2rldJGsVxsv9FogY0i4sP8UA6ixCham9Tq5s0CD1VXJ+EaD92jp6FILhzJ7UBGD
PDLDPF73LcFTVjgNM8EQuQrkah38Et83j1Cqy/MLfrMWo6SY/oUyHTa1N9BZQHif
t6wxYCdV/i9JIRUmPL4w8TQRAURiMRjmAejnUyIekLhrrFm2W4R+p0WnVARMDrx3
THuxk5L3984l8n0ewwHRF+NGwxA09xqGrT+0ELcCmkhhOniS5wIDAQABo3YwdDAO
BgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYDVR0TAQH/BAIw
ADAdBgNVHQ4EFgQUr/KKPNxVmB/IZ/gGgRmpYcOKz+cwIAYDVR0RBBkwF4IVdGVz
dC10Zi5pb3Quc2N3LmNsb3VkMA0GCSqGSIb3DQEBCwUAA4IBAQAMxzVBFBw3U9fj
fUabLpk9+O/9iPlDJfW09c21P3iuI76CnxCLaOCAMctNtdMQSodaYnnA1w1A9+Oq
QH3B/ydZCHfVL/5FkUayYHG1uzsJir9IOtJ8QWEpeDaprO5XBqMtRfGzrz7fDB6x
uNolPpQkOwHtgALqnHJHpxnI49NoUDXM9ZHvk7YY4WE8gEfDwMN185k78qElf2d5
WuePu1khrEuTaXKyaiLD3pmxM86F/6Ho6V86mJpKXr/wmMU56TcKk9UURucVQZ1o
2m0aSh7KsPOWXiIKFpGQMXZZkdpHDOPFAeX2Z7PzLD7AwYSfgnhTmgxL1XqkoHG8
3Vu6zPcD
-----END CERTIFICATE-----
`

func TestAccScalewayIotDevice_Minimal(t *testing.T) {
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
						resource "scaleway_iot_device" "default" {
							name = "default"
							hub_id = scaleway_iot_hub.minimal.id
						}
						resource "scaleway_iot_hub" "minimal" {
							name = "minimal"
							product_plan = "plan_shared"
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotDeviceExists(tt, "scaleway_iot_device.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "hub_id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "name"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_insecure", "false"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_multiple_connections", "false"),
				),
			},
			{
				Config: `
						resource "scaleway_iot_device" "default" {
							name = "default"
							hub_id = scaleway_iot_hub.minimal.id
							allow_insecure = true
						}
						resource "scaleway_iot_hub" "minimal" {
							name = "minimal"
							product_plan = "plan_shared"
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotDeviceExists(tt, "scaleway_iot_device.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "hub_id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "name"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_insecure", "true"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_multiple_connections", "false"),
				),
			},
			{
				Config: `
						resource "scaleway_iot_device" "default" {
							name = "default"
							hub_id = scaleway_iot_hub.minimal.id
							allow_insecure = false
						}
						resource "scaleway_iot_hub" "minimal" {
							name = "minimal"
							product_plan = "plan_shared"
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotDeviceExists(tt, "scaleway_iot_device.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "hub_id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "name"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_insecure", "false"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_multiple_connections", "false"),
				),
			},
			{
				Config: `
						resource "scaleway_iot_device" "default" {
							name = "default"
							hub_id = scaleway_iot_hub.minimal.id
							message_filters {
								publish {
									policy = "reject"
									topics = ["1", "2", "3"]
								}
								subscribe {
									policy = "accept"
									topics = ["4", "5", "6"]
								}
							}
						}
						resource "scaleway_iot_device" "empty" {
							name = "empty"
							hub_id = scaleway_iot_hub.minimal.id
							message_filters {
								publish { }
								subscribe { }
							}
						}
						resource "scaleway_iot_hub" "minimal" {
							name = "minimal"
							product_plan = "plan_shared"
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotDeviceExists(tt, "scaleway_iot_device.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "hub_id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "name"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_insecure", "false"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_multiple_connections", "false"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "message_filters.0.publish.0.policy", "reject"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "message_filters.0.publish.0.topics.0", "1"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "message_filters.0.subscribe.0.policy", "accept"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "message_filters.0.subscribe.0.topics.0", "4"),
				),
			},
			{
				Config: `
						resource "scaleway_iot_device" "default" {
							name = "default"
							hub_id = scaleway_iot_hub.minimal.id
							allow_insecure = true
							certificate {
								crt = <<EOF
` + customDevCert + `EOF
							}
						}
						resource "scaleway_iot_hub" "minimal" {
							name = "minimal"
							product_plan = "plan_shared"
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotDeviceExists(tt, "scaleway_iot_device.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "hub_id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "name"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_insecure", "true"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_multiple_connections", "false"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "certificate.0.crt", customDevCert),
				),
			},
		},
	})
}

func testAccCheckScalewayIotDeviceExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		iotAPI, region, deviceID, err := iotAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = iotAPI.GetDevice(&iot.GetDeviceRequest{
			Region:   region,
			DeviceID: deviceID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
