package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayMNQQueue_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayMNQnNamespaceCreedDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_mnq_namespace" "main" {
				  name     = "test-mnq-sqs-basic"
				  protocol = "sqs_sns"
				}

				resource "scaleway_mnq_credential" "main" {
				  name         = "test-creed-sqs-basic"
				  namespace_id = scaleway_mnq_namespace.main.id
				  sqs_sns_credentials {
					permissions {
					  can_publish = true
					  can_receive = true
					  can_manage  = true
					}
				  }
				}

				resource "scaleway_mnq_queue" "main" {
				  name                      = "terraform-example-queue"
				  namespace_id = scaleway_mnq_namespace.main.id
				  access_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].access_key
				  secret_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].secret_key
				  delay_seconds             = 0
				  max_message_size          = 2048
				  message_retention_seconds = 86400
				  receive_wait_time_seconds = 0
				}
				`,
			},
		},
	})
}
