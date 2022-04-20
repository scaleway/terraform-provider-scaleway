package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayLbSubscriber_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_subscriber sub01 {
						name = "test-sub"
					  	email_config {
							email = "foo@bar.com"
					  	}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_subscriber.sub01", "name", "test-sub"),
					resource.TestCheckResourceAttr("scaleway_lb_subscriber.sub01", "email_config.#", "1"),
					resource.TestCheckResourceAttr("scaleway_lb_subscriber.sub01", "email_config.0.email", "foo@bar.com"),
				),
			},
			{
				Config: `
					resource scaleway_lb_subscriber sub01 {
						name = "test-sub"
					  	email_config {
							email = "zoo@bar.com"
					  	}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_subscriber.sub01", "name", "test-sub"),
					resource.TestCheckResourceAttr("scaleway_lb_subscriber.sub01", "email_config.#", "1"),
					resource.TestCheckResourceAttr("scaleway_lb_subscriber.sub01", "email_config.0.email", "zoo@bar.com"),
				),
			},
			{
				Config: `
					resource scaleway_lb_subscriber sub01 {
						name = "test-sub"
					  	webhook_config {
							uri = "http://www.google.com/"
					  	}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_subscriber.sub01", "name", "test-sub"),
					resource.TestCheckResourceAttr("scaleway_lb_subscriber.sub01", "webhook_config.#", "1"),
					resource.TestCheckResourceAttr("scaleway_lb_subscriber.sub01", "webhook_config.0.uri", "http://www.google.com/"),
				),
			},
		},
	})
}
