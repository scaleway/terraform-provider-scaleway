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
		CheckDestroy:      testAccCheckScalewayMNQQueueDestroy(tt),
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

						message_max_size = 2048
						message_max_age  = 86400

						sqs {
							access_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].access_key
							secret_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].secret_key
							receive_wait_time_seconds = 0
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQQueueExists(tt, "scaleway_mnq_queue.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "name", "terraform-example-queue"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_queue.main", "message_max_size"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_queue.main", "message_max_age"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "sqs.#", "1"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "sqs.0.fifo_queue", "false"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "sqs.0.receive_wait_time_seconds", "0"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_queue.main", "sqs.0.url"),
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

						sqs {
							access_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].access_key
							secret_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].secret_key
							receive_wait_time_seconds = 0
							fifo_queue = true
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

						sqs {
							access_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].access_key
							secret_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].secret_key
							receive_wait_time_seconds = 0
							fifo_queue = true
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQQueueExists(tt, "scaleway_mnq_queue.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "name", "terraform-example-queue.fifo"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "sqs.0.fifo_queue", "true"),
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

						message_max_size = 4096

						sqs {
							access_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].access_key
							secret_key = scaleway_mnq_credential.main.sqs_sns_credentials[0].secret_key
							receive_wait_time_seconds = 0
							fifo_queue = true
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQQueueExists(tt, "scaleway_mnq_queue.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "message_max_size", "4096"),
				),
			},
		},
	})
}

func TestAccScalewayMNQQueue_BasicNATS(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping NATS tests because JetStream is not using HTTP calls thus cannot be recorded in cassettes")
	}

	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayMNQQueueDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mnq_namespace" "main" {
						name     = "test-mnq-nats-basic"
						protocol = "nats"
					}

					resource "scaleway_mnq_credential" "main" {
						name         = "test-queue-nats-basic"
						namespace_id = scaleway_mnq_namespace.main.id
					}

					resource "scaleway_mnq_queue" "main" {
						name         = "terraform-example-queue"
						namespace_id = scaleway_mnq_namespace.main.id

						nats {
							credentials = scaleway_mnq_credential.main.nats_credentials[0].content
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQQueueExists(tt, "scaleway_mnq_queue.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "name", "terraform-example-queue"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_queue.main", "message_max_age"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_queue.main", "message_max_size"),
				),
			},
			{
				Config: `
					resource "scaleway_mnq_namespace" "main" {
						name     = "test-mnq-nats-basic"
						protocol = "nats"
					}

					resource "scaleway_mnq_credential" "main" {
						name         = "test-queue-nats-basic"
						namespace_id = scaleway_mnq_namespace.main.id
					}

					resource "scaleway_mnq_queue" "main" {
						name         = "terraform-example-queue"
						namespace_id = scaleway_mnq_namespace.main.id

						message_max_age = 100
						message_max_size = 2048

						nats {
							credentials = scaleway_mnq_credential.main.nats_credentials[0].content
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQQueueExists(tt, "scaleway_mnq_queue.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "name", "terraform-example-queue"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "message_max_age", "100"),
					resource.TestCheckResourceAttr("scaleway_mnq_queue.main", "message_max_size", "2048"),
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

		namespaceRegion, namespaceID, queueName, err := decomposeMNQQueueID(rs.Primary.ID)
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
			return testAccCheckScalewayMNQQueueExistsSQS(tt, namespaceRegion, rs.Primary.Attributes["sqs.0.endpoint"], rs.Primary.Attributes["sqs.0.access_key"], rs.Primary.Attributes["sqs.0.secret_key"], queueName)
		case mnq.NamespaceProtocolNats:
			return testAccCheckScalewayMNQQueueExistsNATS(tt, namespaceRegion, rs.Primary.Attributes["nats.0.endpoint"], rs.Primary.Attributes["nats.0.credentials"], queueName)
		default:
			return fmt.Errorf("unknown protocol %s", namespace.Protocol)
		}
	}
}

func testAccCheckScalewayMNQQueueExistsSQS(tt *TestTools, region scw.Region, endpoint string, accessKey string, secretKey string, queueName string) error {
	sqsClient, err := newSQSClient(tt.Meta.httpClient, region.String(), endpoint, accessKey, secretKey)
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

func testAccCheckScalewayMNQQueueExistsNATS(_ *TestTools, region scw.Region, endpoint string, credentials string, queueName string) error {
	js, err := newNATSJetStreamClient(region.String(), endpoint, credentials)
	if err != nil {
		return err
	}

	_, err = js.StreamInfo(queueName)
	if err != nil {
		return err
	}

	return nil
}

func testAccCheckScalewayMNQQueueDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mnq_queue" {
				continue
			}

			namespaceRegion, namespaceID, queueName, err := decomposeMNQQueueID(rs.Primary.ID)
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
				return testAccCheckScalewayMNQQueueDestroySQS(tt, namespaceRegion, rs.Primary.Attributes["sqs.0.endpoint"], rs.Primary.Attributes["sqs.0.access_key"], rs.Primary.Attributes["sqs.0.secret_key"], queueName)
			case mnq.NamespaceProtocolNats:
				return testAccCheckScalewayMNQQueueDestroyNATS(tt, namespaceRegion, rs.Primary.Attributes["nats.0.endpoint"], rs.Primary.Attributes["nats.0.credentials"], queueName)
			default:
				return fmt.Errorf("unknown protocol %s", namespace.Protocol)
			}
		}

		return nil
	}
}

func testAccCheckScalewayMNQQueueDestroySQS(tt *TestTools, region scw.Region, endpoint string, accessKey string, secretKey string, queueName string) error {
	sqsClient, err := newSQSClient(tt.Meta.httpClient, region.String(), endpoint, accessKey, secretKey)
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

		return fmt.Errorf("failed to get queue url: %s", err)
	}

	return fmt.Errorf("queue still exists")
}

func testAccCheckScalewayMNQQueueDestroyNATS(_ *TestTools, region scw.Region, endpoint string, credentials string, queueName string) error {
	js, err := newNATSJetStreamClient(region.String(), endpoint, credentials)
	if err != nil {
		return err
	}

	_, err = js.StreamInfo(queueName)
	if err != nil {
		return err
	}

	return nil
}
