package scaleway

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestAccScalewayMNQQueue_BasicSQS(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayMNQnNamespaceQueueDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mnq_namespace" "main" {
						name     = "test-mnq-sqs-basic"
						protocol = "sqs_sns"
					}

					resource "scaleway_mnq_credential" "main" {
						name         = "test-queue-sqs-basic"
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
						name         = "terraform-example-queue"
						namespace_id = scaleway_mnq_namespace.main.id

						sqs {
							access_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].access_key
							secret_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].secret_key
							max_message_size          = 2048
							message_retention_seconds = 86400
							receive_wait_time_seconds = 0
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQQueueExists(tt, "scaleway_mnq_queue.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "name", "terraform-example-queue"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "fifo_queue", "false"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "sqs.#", "1"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "sqs.0.max_message_size", "2048"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "sqs.0.message_retention_seconds", "86400"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "sqs.0.receive_wait_time_seconds", "0"),
				),
			},
			{
				Config: `
					resource "scaleway_mnq_namespace" "main" {
						name     = "test-mnq-sqs-basic"
						protocol = "sqs_sns"
					}

					resource "scaleway_mnq_credential" "main" {
						name         = "test-queue-sqs-basic"
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
						name         = "terraform-example-queue"
						namespace_id = scaleway_mnq_namespace.main.id

						fifo_queue = true

						sqs {
							access_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].access_key
							secret_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].secret_key
							max_message_size          = 2048
							message_retention_seconds = 86400
							receive_wait_time_seconds = 0
						}
					}
				`,
				ExpectError: regexp.MustCompile("invalid queue name"),
			},
			{
				Config: `
					resource "scaleway_mnq_namespace" "main" {
						name     = "test-mnq-sqs-basic"
						protocol = "sqs_sns"
					}

					resource "scaleway_mnq_credential" "main" {
						name         = "test-queue-sqs-basic"
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
						name         = "terraform-example-queue.fifo"
						namespace_id = scaleway_mnq_namespace.main.id

						fifo_queue = true

						sqs {
							access_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].access_key
							secret_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].secret_key
							max_message_size          = 2048
							message_retention_seconds = 86400
							receive_wait_time_seconds = 0
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQQueueExists(tt, "scaleway_mnq_queue.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "name", "terraform-example-queue.fifo"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "fifo_queue", "true"),
				),
			},
			{
				Config: `
					resource "scaleway_mnq_namespace" "main" {
						name     = "test-mnq-sqs-basic"
						protocol = "sqs_sns"
					}

					resource "scaleway_mnq_credential" "main" {
						name         = "test-queue-sqs-basic"
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
						name         = "terraform-example-queue.fifo"
						namespace_id = scaleway_mnq_namespace.main.id

						fifo_queue = true

						sqs {
							access_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].access_key
							secret_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].secret_key
							max_message_size          = 4096
							message_retention_seconds = 86400
							receive_wait_time_seconds = 0
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQQueueExists(tt, "scaleway_mnq_queue.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "sqs.0.max_message_size", "4096"),
				),
			},
		},
	})
}

func testAccCheckScalewayMNQQueueExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		namespaceRegion, namespaceID, queueName, err := decomposeMNQID(rs.Primary.ID)
		if err != nil {
			return err
		}

		api := mnq.NewAPI(tt.Meta.scwClient)
		namespace, err := api.GetNamespace(&mnq.GetNamespaceRequest{
			Region:      namespaceRegion,
			NamespaceID: namespaceID,
		})
		if err != nil {
			return err
		}

		switch namespace.Protocol {
		case mnq.NamespaceProtocolSqsSns:
			return testAccCheckScalewayMNQQueueExistsSQS(tt, namespaceRegion, rs.Primary.Attributes["sqs.0.access_key"], rs.Primary.Attributes["sqs.0.secret_key"], queueName)
		// case mnq.NamespaceProtocolNats:
		// 	return testAccCheckScalewayMNQQueueExistsNATS()
		default:
			return fmt.Errorf("unknown protocol %s", namespace.Protocol)
		}
	}
}

func testAccCheckScalewayMNQQueueExistsSQS(tt *TestTools, region scw.Region, accessKey string, secretKey string, queueName string) error {
	sqsClient, err := newSQSClient(tt.Meta.httpClient, region.String(), accessKey, secretKey)
	if err != nil {
		return err
	}

	_, err = sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return err
	}

	return nil
}

func testAccCheckScalewayMNQnNamespaceQueueDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mnq_queue" {
				continue
			}

			namespaceRegion, namespaceID, queueName, err := decomposeMNQID(rs.Primary.ID)
			if err != nil {
				return err
			}

			api := mnq.NewAPI(tt.Meta.scwClient)
			namespace, err := api.GetNamespace(&mnq.GetNamespaceRequest{
				Region:      namespaceRegion,
				NamespaceID: namespaceID,
			})
			if err != nil {
				if is404Error(err) {
					return nil
				}

				return err
			}

			switch namespace.Protocol {
			case mnq.NamespaceProtocolSqsSns:
				return testAccCheckScalewayMNQQueueDestroySQS(tt, namespaceRegion, rs.Primary.Attributes["sqs.0.access_key"], rs.Primary.Attributes["sqs.0.secret_key"], queueName)
			// case mnq.NamespaceProtocolNats:
			// 	return testAccCheckScalewayMNQQueueExistsNATS()
			default:
				return fmt.Errorf("unknown protocol %s", namespace.Protocol)
			}
		}

		return nil
	}
}

func testAccCheckScalewayMNQQueueDestroySQS(tt *TestTools, region scw.Region, accessKey string, secretKey string, queueName string) error {
	sqsClient, err := newSQSClient(tt.Meta.httpClient, region.String(), accessKey, secretKey)
	if err != nil {
		return err
	}

	_, err = sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
			return nil
		}

		return err
	}

	return fmt.Errorf("queue still exists")
}
