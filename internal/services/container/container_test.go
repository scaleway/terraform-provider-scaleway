package container_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alexedwards/argon2id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	containerSDK "github.com/scaleway/scaleway-sdk-go/api/container/v1"
	containerSDKBeta "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/container"
	vpcchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

const (
	defaultTestImage    = "nginx:1.29.4-alpine"
	helloWorldTestImage = "hello-world:latest"
)

func TestAccContainer_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-basic"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						description = "This container has a description"
						tags = [ "terraform-test", "scaleway_container", "basic" ]
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					acctest.CheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container_namespace.main", "id", "scaleway_container.main", "namespace_id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "image", defaultTestImage),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "80"),
					resource.TestCheckResourceAttr("scaleway_container.main", "description", "This container has a description"),
					resource.TestCheckResourceAttr("scaleway_container.main", "tags.#", "3"),
					resource.TestCheckResourceAttr("scaleway_container.main", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_container.main", "tags.1", "scaleway_container"),
					resource.TestCheckResourceAttr("scaleway_container.main", "tags.2", "basic"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "cpu_limit"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "domain_name"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "public_endpoint"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "local_storage_limit"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "local_storage_limit_bytes"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "memory_limit"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "memory_limit_bytes"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "min_scale"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "max_scale"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "name"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "privacy"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "protocol"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "timeout"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-basic"
					}

					resource scaleway_container main {
						name = "my-container-tf"
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 8080
						description = "new description"
						tags = [ "basic" ]

						cpu_limit = 70
						local_storage_limit_bytes = 1000000000
						memory_limit_bytes = 128000000
						min_scale = 0
						max_scale = 20
						privacy = "private"
						protocol = "http1"
						sandbox = "v2"
						timeout = 300
					}
				`, helloWorldTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					acctest.CheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "name", "my-container-tf"),
					resource.TestCheckResourceAttr("scaleway_container.main", "image", helloWorldTestImage),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "8080"),
					resource.TestCheckResourceAttr("scaleway_container.main", "description", "new description"),
					resource.TestCheckResourceAttr("scaleway_container.main", "tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "tags.0", "basic"),
					resource.TestCheckResourceAttr("scaleway_container.main", "cpu_limit", "70"),
					resource.TestCheckResourceAttr("scaleway_container.main", "local_storage_limit", "1000"),
					resource.TestCheckResourceAttr("scaleway_container.main", "local_storage_limit_bytes", "1000000000"),
					resource.TestCheckResourceAttr("scaleway_container.main", "memory_limit", "128"),
					resource.TestCheckResourceAttr("scaleway_container.main", "memory_limit_bytes", "128000000"),
					resource.TestCheckResourceAttr("scaleway_container.main", "min_scale", "0"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_scale", "20"),
					resource.TestCheckResourceAttr("scaleway_container.main", "privacy", containerSDK.ContainerPrivacyPrivate.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "protocol", containerSDK.ContainerProtocolHTTP1.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "sandbox", containerSDK.ContainerSandboxV2.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "timeout", "300"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-basic"
					}

					resource "scaleway_container" main {
						name 			= "my-container-was-renamed"
						namespace_id	= scaleway_container_namespace.main.id
						image           = "%s"
						port         	= 5000

						cpu_limit = 280
						local_storage_limit = 1500
						memory_limit = 1120
						min_scale = 1
						max_scale = 2
						privacy = "public"
						protocol = "h2c"
						sandbox = "v1"
						timeout = 360
					}
				`, helloWorldTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					acctest.CheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "name", "my-container-was-renamed"),
					resource.TestCheckResourceAttr("scaleway_container.main", "image", helloWorldTestImage),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "5000"),
					resource.TestCheckResourceAttr("scaleway_container.main", "description", ""),
					resource.TestCheckResourceAttr("scaleway_container.main", "tags.#", "0"),
					resource.TestCheckResourceAttr("scaleway_container.main", "cpu_limit", "280"),
					resource.TestCheckResourceAttr("scaleway_container.main", "local_storage_limit", "1500"),
					resource.TestCheckResourceAttr("scaleway_container.main", "local_storage_limit_bytes", "1500000000"),
					resource.TestCheckResourceAttr("scaleway_container.main", "memory_limit", "1120"),
					resource.TestCheckResourceAttr("scaleway_container.main", "memory_limit_bytes", "1120000000"),
					resource.TestCheckResourceAttr("scaleway_container.main", "min_scale", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_scale", "2"),
					resource.TestCheckResourceAttr("scaleway_container.main", "privacy", containerSDK.ContainerPrivacyPublic.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "protocol", containerSDK.ContainerProtocolH2c.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "sandbox", containerSDK.ContainerSandboxV1.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "timeout", "360"),
				),
			},
		},
	})
}

func TestAccContainer_Env(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-env"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						environment_variables = {
							"test" = "test"
						}
						secret_environment_variables = {
							"test_secret" = "test_secret"
							"first_secret" = "first_secret"
						}
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					acctest.CheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "environment_variables.test", "test"),
					passwordMatchHash("scaleway_container.main", "secret_environment_variables.test_secret", "test_secret"),
					passwordMatchHash("scaleway_container.main", "secret_environment_variables.first_secret", "first_secret"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-env"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						environment_variables = {
							"test" = "test"
							"foo" = "bar"
						}
						secret_environment_variables = {
							"foo_secret" = "bar_secret"
							"test_secret" = "updated_secret"
						}
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					acctest.CheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "environment_variables.test", "test"),
					resource.TestCheckResourceAttr("scaleway_container.main", "environment_variables.foo", "bar"),
					passwordMatchHash("scaleway_container.main", "secret_environment_variables.foo_secret", "bar_secret"),
					passwordMatchHash("scaleway_container.main", "secret_environment_variables.test_secret", "updated_secret"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "secret_environment_variables.first_secret"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-env"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						environment_variables = {}
						secret_environment_variables = {}
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					acctest.CheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "environment_variables.%"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "environment_variables.test"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "environment_variables.foo"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "secret_environment_variables.%"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "secret_environment_variables.foo_secret"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "secret_environment_variables.test_secret"),
				),
			},
		},
	})
}

func TestAccContainer_Image(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-image"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						registry_image = "%s"
						port = 80
					}
				`, helloWorldTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "registry_image", helloWorldTestImage),
					resource.TestCheckResourceAttr("scaleway_container.main", "image", helloWorldTestImage),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "80"),
					resource.TestCheckResourceAttr("scaleway_container.main", "status", containerSDK.ContainerStatusError.String()),
				),
			},
			// Updating the 'registry_image' field with an image listening on port 80 should fix the container's status
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-image"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						registry_image = "%s"
						port = 80
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "registry_image", defaultTestImage),
					resource.TestCheckResourceAttr("scaleway_container.main", "image", defaultTestImage),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "80"),
					resource.TestCheckResourceAttr("scaleway_container.main", "status", containerSDK.ContainerStatusReady.String()),
				),
			},
			// Back to initial setting (container in error state) with 'image' set instead of 'registry_image'
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-image"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
					}
				`, helloWorldTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "registry_image", helloWorldTestImage),
					resource.TestCheckResourceAttr("scaleway_container.main", "image", helloWorldTestImage),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "80"),
					resource.TestCheckResourceAttr("scaleway_container.main", "status", containerSDK.ContainerStatusError.String()),
				),
			},
			// Updating the 'image' field with an image listening on port 80 should fix the container's status
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-image"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "registry_image", defaultTestImage),
					resource.TestCheckResourceAttr("scaleway_container.main", "image", defaultTestImage),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "80"),
					resource.TestCheckResourceAttr("scaleway_container.main", "status", containerSDK.ContainerStatusReady.String()),
				),
			},
		},
	})
}

func TestAccContainer_HTTPOption(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-http-otion"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						registry_image = "%s"
						port = 80
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "http_option", containerSDKBeta.ContainerHTTPOptionEnabled.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "https_connections_only", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-http-otion"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						https_connections_only = true
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "https_connections_only", "true"),
					resource.TestCheckResourceAttr("scaleway_container.main", "http_option", containerSDKBeta.ContainerHTTPOptionRedirected.String()),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-http-otion"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						http_option = "enabled"
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "http_option", containerSDKBeta.ContainerHTTPOptionEnabled.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "https_connections_only", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-http-otion"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						http_option = "redirected"
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "http_option", containerSDKBeta.ContainerHTTPOptionRedirected.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "https_connections_only", "true"),
				),
			},
		},
	})
}

func TestAccContainer_Sandbox(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-sandbox"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						deploy = false
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "sandbox"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-sandbox"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						deploy = false
						sandbox = "v2"
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "sandbox", containerSDK.ContainerSandboxV2.String()),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-sandbox"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						deploy = false
						sandbox = "v1"
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "sandbox", containerSDK.ContainerSandboxV1.String()),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-sandbox"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						sandbox = "v2"
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "sandbox", containerSDK.ContainerSandboxV2.String()),
				),
			},
		},
	})
}

func TestAccContainer_HealthCheck(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-healthcheck"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						deploy = false
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					// Check default option returned by the API when you don't specify the health_check block.
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.failure_threshold", "30"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.interval", "10s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.failure_threshold", "30"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.interval", "10s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.timeout", "1s"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-healthcheck"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						deploy = false

						health_check {
							http {
								path = "/test"
							}
							failure_threshold = 40
							interval = "12s"
						}
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.http.0.path", "/test"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.failure_threshold", "40"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.interval", "12s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.http.0.path", "/test"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.failure_threshold", "40"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.interval", "12s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.timeout", "1s"),
				),
			},
		},
	})
}

func TestAccContainer_ScalingOption(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-scaling-option"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						deploy = false
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					// Check default option returned by the API when you don't specify the scaling_option block.
					resource.TestCheckResourceAttr("scaleway_container.main", "scaling_option.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "scaling_option.0.concurrent_requests_threshold", "50"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-scaling-option"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						deploy = false

						scaling_option {
							concurrent_requests_threshold = 15
						}
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "scaling_option.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "scaling_option.0.concurrent_requests_threshold", "15"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-scaling-option"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						deploy = false

						min_scale = 1

						scaling_option {
							cpu_usage_threshold = 72
						}
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "scaling_option.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "scaling_option.0.cpu_usage_threshold", "72"),
				),
			},

			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-scaling-option"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80

						min_scale = 1

						scaling_option {
							memory_usage_threshold = 66
						}
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "scaling_option.0.memory_usage_threshold", "66"),
				),
			},
		},
	})
}

func TestAccContainer_CommandAndArgs(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-command-and-args"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						command = [ "bash", "-c", "my-script.sh" ]
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "command.#", "3"),
					resource.TestCheckResourceAttr("scaleway_container.main", "command.0", "bash"),
					resource.TestCheckResourceAttr("scaleway_container.main", "command.1", "-c"),
					resource.TestCheckResourceAttr("scaleway_container.main", "command.2", "my-script.sh"),
					resource.TestCheckResourceAttr("scaleway_container.main", "args.#", "0"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-command-and-args"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						command = [ "bash", "-c", "my-script.sh" ]
						args =    [ "some", "args" ]
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "command.#", "3"),
					resource.TestCheckResourceAttr("scaleway_container.main", "command.0", "bash"),
					resource.TestCheckResourceAttr("scaleway_container.main", "command.1", "-c"),
					resource.TestCheckResourceAttr("scaleway_container.main", "command.2", "my-script.sh"),
					resource.TestCheckResourceAttr("scaleway_container.main", "args.#", "2"),
					resource.TestCheckResourceAttr("scaleway_container.main", "args.0", "some"),
					resource.TestCheckResourceAttr("scaleway_container.main", "args.1", "args"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-command-and-args"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						args =    [ "some", "args" ]
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "command.#", "0"),
					resource.TestCheckResourceAttr("scaleway_container.main", "args.#", "2"),
					resource.TestCheckResourceAttr("scaleway_container.main", "args.0", "some"),
					resource.TestCheckResourceAttr("scaleway_container.main", "args.1", "args"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-command-and-args"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "command.#", "0"),
					resource.TestCheckResourceAttr("scaleway_container.main", "args.#", "0"),
				),
			},
		},
	})
}

func TestAccContainer_PrivateNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isNamespaceDestroyed(tt),
			isContainerDestroyed(tt),
			vpcchecks.CheckPrivateNetworkDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_vpc main {
						name = "tf-acctest-private-network"
					}

					resource scaleway_vpc_private_network start {
						name = "test-acc-container-pn-start"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_container_namespace main {
						name = "tf-acctest-private-network"
					}

					resource scaleway_container to_stay {
						name = "test-acc-container-pn-to-stay"
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80
						private_network_id = scaleway_vpc_private_network.start.id
						sandbox = "v1"
						tags = [ "to-stay" ]
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.to_stay"),
					resource.TestCheckResourceAttr("scaleway_container.to_stay", "sandbox", "v1"),
					resource.TestCheckResourceAttrPair("scaleway_container.to_stay", "private_network_id", "scaleway_vpc_private_network.start", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_vpc main {
						name = "tf-acctest-private-network"
					}

					resource scaleway_vpc_private_network start {
						name = "test-acc-container-pn-start"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_container_namespace main {
						name = "tf-acctest-private-network"
					}

					resource scaleway_container to_stay {
						name = "test-acc-container-pn-to-stay"
						namespace_id = scaleway_container_namespace.main.id
						image = "%[1]s"
						port = 80
						private_network_id = scaleway_vpc_private_network.start.id
						sandbox = "v1"
						tags = [ "to-stay" ]
					}

					resource scaleway_container to_change {
						name = "test-acc-container-pn-to-change"
						namespace_id = scaleway_container_namespace.main.id
						image = "%[1]s"
						port = 80
						private_network_id = scaleway_vpc_private_network.start.id
						sandbox = "v1"
						tags = [ "to-change" ]
					}

					resource scaleway_container to_remove {
						name = "test-acc-container-pn-to-remove"
						namespace_id = scaleway_container_namespace.main.id
						image = "%[1]s"
						port = 80
						private_network_id = scaleway_vpc_private_network.start.id
						sandbox = "v1"
						tags = [ "to-remove" ]
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.to_stay"),
					isContainerPresent(tt, "scaleway_container.to_change"),
					isContainerPresent(tt, "scaleway_container.to_remove"),
					resource.TestCheckResourceAttr("scaleway_container.to_stay", "sandbox", "v1"),
					resource.TestCheckResourceAttr("scaleway_container.to_change", "sandbox", "v1"),
					resource.TestCheckResourceAttr("scaleway_container.to_remove", "sandbox", "v1"),
					resource.TestCheckResourceAttrPair("scaleway_container.to_stay", "private_network_id", "scaleway_vpc_private_network.start", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container.to_change", "private_network_id", "scaleway_vpc_private_network.start", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container.to_remove", "private_network_id", "scaleway_vpc_private_network.start", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_vpc main {
						name = "tf-acctest-private-network"
					}

					resource scaleway_vpc_private_network start {
						name = "test-acc-container-pn-start"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_vpc_private_network new {
						name = "test-acc-container-pn-new"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_container_namespace main {
						name = "tf-acctest-private-network"
					}

					resource scaleway_container to_stay {
						name = "test-acc-container-pn-to-stay"
						namespace_id = scaleway_container_namespace.main.id
						image = "%[1]s"
						port = 80
						private_network_id = scaleway_vpc_private_network.start.id
						sandbox = "v1"
						tags = [ "to-stay" ]
					}

					resource scaleway_container to_change {
						name = "test-acc-container-pn-to-change"
						namespace_id = scaleway_container_namespace.main.id
						image = "%[1]s"
						port = 80
						private_network_id = scaleway_vpc_private_network.new.id
						sandbox = "v1"
						tags = [ "to-change" ]
					}

					resource scaleway_container to_remove {
						name = "test-acc-container-pn-to-remove"
						namespace_id = scaleway_container_namespace.main.id
						image = "%[1]s"
						port = 80
						sandbox = "v1"
						tags = [ "to-remove" ]
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.to_stay"),
					isContainerPresent(tt, "scaleway_container.to_change"),
					isContainerPresent(tt, "scaleway_container.to_remove"),
					resource.TestCheckResourceAttrPair("scaleway_container.to_stay", "private_network_id", "scaleway_vpc_private_network.start", "id"),
					resource.TestCheckResourceAttrPair("scaleway_container.to_change", "private_network_id", "scaleway_vpc_private_network.new", "id"),
					resource.TestCheckResourceAttr("scaleway_container.to_remove", "private_network_id", ""),
				),
			},
			{
				Config: `
					resource scaleway_vpc main {
						name = "tf-acctest-private-network"
					}

					resource scaleway_vpc_private_network start {
						name = "test-acc-container-pn-start"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_vpc_private_network new {
						name = "test-acc-container-pn-new"
						vpc_id = scaleway_vpc.main.id
					}`,
			},
		},
	})
}

func TestAccContainer_Probes(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-probes"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80

						startup_probe {
							failure_threshold = 5
							interval = "15s"
							timeout = "1m"
							tcp = true
						}

						liveness_probe {
							failure_threshold = 3
							interval = "10s"
							timeout = "30s"
							http {
								path = "/test"
							}
						}
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "startup_probe.0.failure_threshold", "5"),
					resource.TestCheckResourceAttr("scaleway_container.main", "startup_probe.0.interval", "15s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "startup_probe.0.timeout", "1m0s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "startup_probe.0.tcp", "true"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "startup_probe.0.http.0"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.failure_threshold", "3"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.interval", "10s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.timeout", "30s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.http.0.path", "/test"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.tcp", "false"),
					// Check backward compatibility with health_check
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.failure_threshold", "3"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.interval", "10s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.http.0.path", "/test"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.tcp", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "tf-acctest-probes"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						image = "%s"
						port = 80

						startup_probe {
							failure_threshold = 10
							interval = "10s"
							timeout = "15s"
							http {
								path = "/test"
							}
						}

						liveness_probe {
							failure_threshold = 5
							interval = "10s"
							timeout = "30s"
							tcp = true
						}
					}
				`, defaultTestImage),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "startup_probe.0.failure_threshold", "10"),
					resource.TestCheckResourceAttr("scaleway_container.main", "startup_probe.0.interval", "10s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "startup_probe.0.timeout", "15s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "startup_probe.0.http.0.path", "/test"),
					resource.TestCheckResourceAttr("scaleway_container.main", "startup_probe.0.tcp", "false"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.failure_threshold", "5"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.interval", "10s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.timeout", "30s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "liveness_probe.0.tcp", "true"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "liveness_probe.0.http.0"),
					// Check backward compatibility with health_check
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.failure_threshold", "5"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.interval", "10s"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.tcp", "true"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "health_check.0.http.0"),
				),
			},
		},
	})
}

func isContainerPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource container not found: %s", n)
		}

		api, region, id, err := container.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetContainer(&containerSDK.GetContainerRequest{
			ContainerID: id,
			Region:      region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isContainerDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != containerNamespaceResource {
				continue
			}

			api, region, id, err := container.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteContainer(&containerSDK.DeleteContainerRequest{
				ContainerID: id,
				Region:      region,
			})
			if err == nil {
				return fmt.Errorf("container (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}

func passwordMatchHash(parent string, key string, password string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[parent]
		if !ok {
			return fmt.Errorf("resource container not found: %s", parent)
		}

		match, err := argon2id.ComparePasswordAndHash(password, rs.Primary.Attributes[key])
		if err != nil {
			return err
		}

		if !match {
			return errors.New("password and hash do not match")
		}

		return nil
	}
}
