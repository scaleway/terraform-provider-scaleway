package container_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	containerSDK "github.com/scaleway/scaleway-sdk-go/api/container/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/container"
)

// TestAccTrigger_SQS can only be recorded by making Terraform sleep for >=30s after projects' creation for the IAM
// permissions to be propagated. Also, since the scaleway_mnq_sqs resource has no other configurable attribute to
// differentiate the 1st and 2nd one, resources tend to get mixed up in the cassettes, therefore the test is flaky.
func TestAccTrigger_SQS(t *testing.T) {
	t.Skip("Test is flaky")

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	unchangedConfig := fmt.Sprintf(`
		resource "scaleway_account_project" "first" {
			name = "tf_tests_container_trigger_sqs_1"
			description = "Initial project for SQS trigger testing"
		}

		resource "scaleway_mnq_sqs" "main" {
			project_id = scaleway_account_project.first.id
		}

		resource "scaleway_mnq_sqs_credentials" "main" {
			project_id = scaleway_account_project.first.id
			permissions {
				can_publish = true
				can_receive = true
				can_manage  = true
			}
			depends_on = [ scaleway_mnq_sqs.main ]
		}

		resource "scaleway_mnq_sqs_queue" "main" {
			project_id = scaleway_account_project.first.id
			name = "TestQueue"
			sqs_endpoint = scaleway_mnq_sqs.main.endpoint
			access_key = scaleway_mnq_sqs_credentials.main.access_key
			secret_key = scaleway_mnq_sqs_credentials.main.secret_key
		}

		resource scaleway_container_namespace main {
			project_id = scaleway_account_project.first.id
			name = "tf-acctest-trigger-sqs"
		}

		resource scaleway_container main {
			namespace_id = scaleway_container_namespace.main.id
			image = "%s"
			port = 80
			privacy = "private"
			protocol = "http1"
			sandbox = "v2"
		}`, defaultTestImage)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isTriggerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: unchangedConfig + `
					resource scaleway_container_trigger main {
						container_id = scaleway_container.main.id
						name = "test-container-trigger-sqs"
						description = "This trigger was created with a description"
						destination_config {
							http_path = "/"
							http_method = "get"
						}
						sqs {
							endpoint = scaleway_mnq_sqs_queue.main.sqs_endpoint
							queue_url = scaleway_mnq_sqs_queue.main.url
							access_key = scaleway_mnq_sqs_credentials.main.access_key
							secret_key = scaleway_mnq_sqs_credentials.main.secret_key
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					isTriggerPresent(tt, "scaleway_container_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_trigger.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "name", "test-container-trigger-sqs"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "description", "This trigger was created with a description"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_path", "/"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_method", "get"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "container_id", "scaleway_container.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "sqs.0.access_key", "scaleway_mnq_sqs_credentials.main", "access_key"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "sqs.0.secret_key", "scaleway_mnq_sqs_credentials.main", "secret_key"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "sqs.0.queue_url", "scaleway_mnq_sqs_queue.main", "url"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "sqs.0.endpoint", "scaleway_mnq_sqs_queue.main", "sqs_endpoint"),
					isTriggerStatusReady(tt, "scaleway_container_trigger.main"),
				),
			},
			{
				Config: unchangedConfig + `
					resource "scaleway_account_project" "second" {
						name = "tf_tests_container_trigger_sqs_2"
						description = "Second project for SQS trigger testing"
					}
					resource "scaleway_mnq_sqs" "second" {
						project_id = scaleway_account_project.second.id
					}
					resource "scaleway_mnq_sqs_credentials" "second" {
						project_id = scaleway_account_project.second.id
						permissions {
							can_publish = false
							can_receive = true
							can_manage  = true
						}
						depends_on = [ scaleway_mnq_sqs.second ]
					}
					resource "scaleway_mnq_sqs_queue" "second" {
						project_id = scaleway_account_project.second.id
						name = "OtherQueue"
						sqs_endpoint = scaleway_mnq_sqs.second.endpoint
						access_key = scaleway_mnq_sqs_credentials.second.access_key
						secret_key = scaleway_mnq_sqs_credentials.second.secret_key
					}

					resource scaleway_container_trigger main {
						container_id = scaleway_container.main.id
						name = "test-container-trigger-sqs"
						destination_config {
							http_path = "/"
							http_method = "get"
						}
						sqs {
							endpoint = scaleway_mnq_sqs_queue.second.sqs_endpoint
							queue_url = scaleway_mnq_sqs_queue.second.url
							access_key = scaleway_mnq_sqs_credentials.second.access_key
							secret_key = scaleway_mnq_sqs_credentials.second.secret_key
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					isTriggerPresent(tt, "scaleway_container_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_trigger.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "name", "test-container-trigger-sqs"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "description", ""),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_path", "/"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_method", "get"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "container_id", "scaleway_container.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "sqs.0.access_key", "scaleway_mnq_sqs_credentials.second", "access_key"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "sqs.0.secret_key", "scaleway_mnq_sqs_credentials.second", "secret_key"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "sqs.0.queue_url", "scaleway_mnq_sqs_queue.second", "url"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "sqs.0.endpoint", "scaleway_mnq_sqs_queue.second", "sqs_endpoint"),
					isTriggerStatusReady(tt, "scaleway_container_trigger.main"),
				),
			},
			{
				Config: unchangedConfig + `
					resource scaleway_container_trigger main {
						container_id = scaleway_container.main.id
						name = "test-container-trigger-sqs"
						description = "Description added back"
						destination_config {
							http_path = "/"
							http_method = "get"
						}
						sqs {
							endpoint = scaleway_mnq_sqs_queue.main.sqs_endpoint
							queue_url = scaleway_mnq_sqs_queue.main.url
							access_key = scaleway_mnq_sqs_credentials.main.access_key
							secret_key = scaleway_mnq_sqs_credentials.main.secret_key
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					isTriggerPresent(tt, "scaleway_container_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_trigger.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "name", "test-container-trigger-sqs"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "description", "Description added back"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "container_id", "scaleway_container.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "sqs.0.access_key", "scaleway_mnq_sqs_credentials.main", "access_key"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "sqs.0.secret_key", "scaleway_mnq_sqs_credentials.main", "secret_key"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "sqs.0.queue_url", "scaleway_mnq_sqs_queue.main", "url"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "sqs.0.endpoint", "scaleway_mnq_sqs_queue.main", "sqs_endpoint"),
					isTriggerStatusReady(tt, "scaleway_container_trigger.main"),
				),
			},
		},
	})
}

func TestAccTrigger_Nats(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	triggerID := ""

	unchangedContainerConfig := fmt.Sprintf(`
		resource "scaleway_mnq_nats_account" "main" {}

		resource "scaleway_mnq_nats_credentials" "main" {
			account_id = scaleway_mnq_nats_account.main.id
		}

		resource scaleway_container_namespace main {
			name = "tf-acctest-trigger-nats"
		}

		resource scaleway_container main {
			namespace_id = scaleway_container_namespace.main.id
			image = "%s"
			port = 80
			privacy = "private"
			protocol = "http1"
			sandbox = "v2"
		}`, defaultTestImage)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isTriggerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: unchangedContainerConfig + `
					resource scaleway_container_trigger main {
						container_id = scaleway_container.main.id
						name = "test-container-trigger-nats"
						destination_config {
							http_path = "/endpoint"
							http_method = "post"
						}
						nats {
							subject = "TestSubject"
							server_urls = [ scaleway_mnq_nats_account.main.endpoint ]
							credentials_file_content = scaleway_mnq_nats_credentials.main.file
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					isTriggerPresent(tt, "scaleway_container_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_trigger.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "container_id", "scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "name", "test-container-trigger-nats"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_path", "/endpoint"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_method", "post"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "nats.0.subject", "TestSubject"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "nats.0.server_urls.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "nats.0.server_urls.0", "scaleway_mnq_nats_account.main", "endpoint"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "nats.0.credentials_file_content", "scaleway_mnq_nats_credentials.main", "file"),
					acctest.CheckResourceIDPersisted("scaleway_container_trigger.main", &triggerID),
					isTriggerStatusReady(tt, "scaleway_container_trigger.main"),
				),
			},
			{
				Config: unchangedContainerConfig + `
					resource "scaleway_mnq_nats_account" "second" {}

					resource "scaleway_mnq_nats_credentials" "second" {
						account_id = scaleway_mnq_nats_account.second.id
					}

					resource scaleway_container_trigger main {
						container_id = scaleway_container.main.id
						name = "test-container-trigger-nats-updated"
						tags = [ "add", "tags" ]
						destination_config {
							http_path = "/endpoint/has/changed"
							http_method = "post"
						}
						nats {
							subject = "ChangeSubject"
							server_urls = [ scaleway_mnq_nats_account.second.endpoint ]
							credentials_file_content = scaleway_mnq_nats_credentials.second.file
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					isTriggerPresent(tt, "scaleway_container_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_trigger.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "container_id", "scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "name", "test-container-trigger-nats-updated"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "tags.0", "add"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "tags.1", "tags"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_path", "/endpoint/has/changed"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_method", "post"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "nats.0.subject", "ChangeSubject"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "nats.0.server_urls.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "nats.0.server_urls.0", "scaleway_mnq_nats_account.second", "endpoint"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "nats.0.credentials_file_content", "scaleway_mnq_nats_credentials.second", "file"),
					acctest.CheckResourceIDPersisted("scaleway_container_trigger.main", &triggerID),
					isTriggerStatusReady(tt, "scaleway_container_trigger.main"),
				),
			},
			{
				Config: unchangedContainerConfig + fmt.Sprintf(`
					resource scaleway_container new {
						namespace_id = scaleway_container_namespace.main.id
						name = "new-container-for-trigger"
						image = "%s"
						port = 80
						privacy = "private"
						protocol = "http1"
						sandbox = "v2"
					}

					resource scaleway_container_trigger main {
						container_id = scaleway_container.new.id
						name = "test-container-trigger-nats-new"
						tags = [ "tags" ]
						destination_config {
							http_path = "/endpoint/has/changed"
							http_method = "put"
						}
						nats {
							subject = "ChangeSubject"
							server_urls = [ scaleway_mnq_nats_account.main.endpoint ]
							credentials_file_content = scaleway_mnq_nats_credentials.main.file
						}
					}`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isTriggerPresent(tt, "scaleway_container_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_trigger.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "container_id", "scaleway_container.new", "id"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "name", "test-container-trigger-nats-new"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "tags.0", "tags"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_path", "/endpoint/has/changed"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_method", "put"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "nats.0.subject", "ChangeSubject"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "nats.0.server_urls.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "nats.0.server_urls.0", "scaleway_mnq_nats_account.main", "endpoint"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "nats.0.credentials_file_content", "scaleway_mnq_nats_credentials.main", "file"),
					acctest.CheckResourceIDChanged("scaleway_container_trigger.main", &triggerID),
					isTriggerStatusReady(tt, "scaleway_container_trigger.main"),
				),
			},
		},
	})
}

func TestAccTrigger_Cron(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	unchangedContainerConfig := fmt.Sprintf(`
		resource scaleway_container_namespace main {
			name = "tf-acctest-trigger-cron"
		}

		resource scaleway_container main {
			namespace_id = scaleway_container_namespace.main.id
			image = "%s"
			privacy = "private"
			protocol = "http1"
			sandbox = "v2"
			port = 80
		}`, defaultTestImage)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isTriggerDestroyed(tt),
			isContainerDestroyed(tt),
			isNamespaceDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: unchangedContainerConfig + `
					resource scaleway_container_trigger main {
						container_id = scaleway_container.main.id
						name = "test-container-trigger-cron"
						tags = [ "test", "cron", "tags" ]
						destination_config {
							http_path = "/path"
							http_method = "get"
						}
						cron {
							schedule = "5 4 1 * *" #cron at 04:05 on day-of-month 1
							timezone = "Europe/Paris"
							body = "{\"message\": \"This is the content to send to the container.\"}"
							headers = {
								Content-Length = 45
								Content-Type = "application/json"
							}
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					isTriggerPresent(tt, "scaleway_container_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_trigger.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "container_id", "scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "name", "test-container-trigger-cron"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "tags.#", "3"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "tags.0", "test"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "tags.1", "cron"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "tags.2", "tags"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_path", "/path"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_method", "get"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.schedule", "5 4 1 * *"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.timezone", "Europe/Paris"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.body", "{\"message\": \"This is the content to send to the container.\"}"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.headers.Content-Length", "45"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.headers.Content-Type", "application/json"),
					isTriggerStatusReady(tt, "scaleway_container_trigger.main"),
				),
			},
			{
				Config: unchangedContainerConfig + `
					resource scaleway_container_trigger main {
						container_id = scaleway_container.main.id
						name = "test-container-trigger-cron"
						description = "This trigger verifies update operations"
						destination_config {
							http_path = "/path"
							http_method = "get"
						}
						cron {
							schedule = "0 0 * * *" #cron at 00:00 everyday
							timezone = "Asia/Seoul"
							body = "Content has changed!"
							headers = {
								Content-Length = 20
							}
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					isTriggerPresent(tt, "scaleway_container_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_trigger.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "container_id", "scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "name", "test-container-trigger-cron"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "description", "This trigger verifies update operations"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "tags.#", "0"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_path", "/path"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_method", "get"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.schedule", "0 0 * * *"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.timezone", "Asia/Seoul"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.body", "Content has changed!"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.headers.Content-Length", "20"),
					isTriggerStatusReady(tt, "scaleway_container_trigger.main"),
				),
			},
			{
				Config: unchangedContainerConfig + `
					resource scaleway_container_trigger main {
						container_id = scaleway_container.main.id
						name = "test-container-trigger-cron"
						tags = [ "test" ]
						destination_config {
							http_path = "/path-has-changed"
							http_method = "delete"
						}
						cron {
							schedule = "*/5 * * * SUN" #cron every 5min on Sundays
							timezone = "Asia/Seoul"
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					isTriggerPresent(tt, "scaleway_container_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_trigger.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "container_id", "scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "name", "test-container-trigger-cron"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "description", ""),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "tags.0", "test"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_path", "/path-has-changed"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_method", "delete"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.schedule", "*/5 * * * SUN"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.timezone", "Asia/Seoul"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.body", ""),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.headers.#", "0"),
					isTriggerStatusReady(tt, "scaleway_container_trigger.main"),
				),
			},
		},
	})
}

func TestAccTrigger_ChangeSource(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resourceID := ""

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isTriggerDestroyed(tt),
			isContainerDestroyed(tt),
			isNamespaceDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-trigger-change-source"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						privacy = "private"
						protocol = "http1"
						sandbox = "v2"
					}

					resource scaleway_container_trigger main {
						container_id = scaleway_container.main.id
						name = "test-container-trigger-change-cron"
						destination_config {
							http_path = "/patch.here"
							http_method = "patch"
						}
						cron {
							schedule = "5 4 1 * *" #cron at 04:05 on day-of-month 1
							timezone = "Europe/Paris"
						}
					}`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isTriggerPresent(tt, "scaleway_container_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_trigger.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "container_id", "scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "name", "test-container-trigger-change-cron"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_path", "/patch.here"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_method", "patch"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.schedule", "5 4 1 * *"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "cron.0.timezone", "Europe/Paris"),
					resource.TestCheckNoResourceAttr("scaleway_container_trigger.main", "nats.0"),
					resource.TestCheckNoResourceAttr("scaleway_container_trigger.main", "sqs.0"),
					isTriggerStatusReady(tt, "scaleway_container_trigger.main"),
					acctest.CheckResourceIDPersisted("scaleway_container_trigger.main", &resourceID),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-test-trigger-change-source"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						privacy = "private"
						protocol = "http1"
						sandbox = "v2"
					}

					resource "scaleway_mnq_nats_account" "main" {}

					resource "scaleway_mnq_nats_credentials" "main" {
						account_id = scaleway_mnq_nats_account.main.id
					}

					resource scaleway_container_trigger main {
						container_id = scaleway_container.main.id
						name = "test-container-trigger-change-nats"
						destination_config {
							http_path = "/patch.here"
							http_method = "patch"
						}
						nats {
							subject = "TestSubject"
							server_urls = [ scaleway_mnq_nats_account.main.endpoint ]
							credentials_file_content = scaleway_mnq_nats_credentials.main.file
						}
					}`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isTriggerPresent(tt, "scaleway_container_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_trigger.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "name", "test-container-trigger-change-nats"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_path", "/patch.here"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "destination_config.0.http_method", "patch"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "nats.0.subject", "TestSubject"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "nats.0.server_urls.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "nats.0.server_urls.0", "scaleway_mnq_nats_account.main", "endpoint"),
					resource.TestCheckResourceAttrPair("scaleway_container_trigger.main", "nats.0.credentials_file_content", "scaleway_mnq_nats_credentials.main", "file"),
					resource.TestCheckNoResourceAttr("scaleway_container_trigger.main", "cron.0"),
					resource.TestCheckNoResourceAttr("scaleway_container_trigger.main", "sqs.0"),
					isTriggerStatusReady(tt, "scaleway_container_trigger.main"),
					acctest.CheckResourceIDChanged("scaleway_container_trigger.main", &resourceID),
				),
			},
		},
	})
}

func isTriggerPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := container.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetTrigger(&containerSDK.GetTriggerRequest{
			TriggerID: id,
			Region:    region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isTriggerStatusReady(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := container.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		trigger, err := api.GetTrigger(&containerSDK.GetTriggerRequest{
			TriggerID: id,
			Region:    region,
		})
		if err != nil {
			return err
		}

		if trigger.Status != containerSDK.TriggerStatusReady {
			return fmt.Errorf("trigger status is %s, expected ready", trigger.Status)
		}

		return nil
	}
}

func isTriggerDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_container_trigger" {
				continue
			}

			api, region, id, err := container.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteTrigger(&containerSDK.DeleteTriggerRequest{
				TriggerID: id,
				Region:    region,
			})
			if err == nil {
				return fmt.Errorf("container trigger (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
