package container_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alexedwards/argon2id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	containerSDK "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/container"
)

func TestAccContainer_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						tags = ["tag1", "tag2"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					acctest.CheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "name"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "registry_image"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "domain_name"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "max_concurrency"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "domain_name"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "protocol"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "cpu_limit"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "timeout"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "memory_limit"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "max_scale"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "min_scale"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "privacy"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "local_storage_limit"),
					resource.TestCheckResourceAttr("scaleway_container.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_container.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "tags.1", "tag2"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
					}

					resource scaleway_container main {
						name = "my-container-tf"
						namespace_id = scaleway_container_namespace.main.id
						port = 8080
						cpu_limit = 70
						memory_limit = 128
						min_scale = 0
						max_scale = 20
						timeout = 300
						deploy = false
						local_storage_limit = 1000
						tags = ["tag"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					acctest.CheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "name", "my-container-tf"),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "8080"),
					resource.TestCheckResourceAttr("scaleway_container.main", "cpu_limit", "70"),
					resource.TestCheckResourceAttr("scaleway_container.main", "memory_limit", "128"),
					resource.TestCheckResourceAttr("scaleway_container.main", "min_scale", "0"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_scale", "20"),
					resource.TestCheckResourceAttr("scaleway_container.main", "timeout", "300"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_concurrency", "50"),
					resource.TestCheckResourceAttr("scaleway_container.main", "deploy", "false"),
					resource.TestCheckResourceAttr("scaleway_container.main", "privacy", containerSDK.ContainerPrivacyPublic.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "protocol", containerSDK.ContainerProtocolHTTP1.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "local_storage_limit", "1000"),
					resource.TestCheckResourceAttr("scaleway_container.main", "tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "tags.0", "tag"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
					}

					resource "scaleway_container" main {
						name 			= "my-container-tf"
						namespace_id	= scaleway_container_namespace.main.id
						port         	= 5000
						min_scale    	= 1
						max_scale    	= 2
						max_concurrency = 80
						memory_limit 	= 1120
						cpu_limit		= 280
						deploy       	= false
						local_storage_limit = 1500
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					acctest.CheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "name", "my-container-tf"),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "5000"),
					resource.TestCheckResourceAttr("scaleway_container.main", "cpu_limit", "280"),
					resource.TestCheckResourceAttr("scaleway_container.main", "memory_limit", "1120"),
					resource.TestCheckResourceAttr("scaleway_container.main", "min_scale", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_scale", "2"),
					resource.TestCheckResourceAttr("scaleway_container.main", "timeout", "300"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_concurrency", "80"),
					resource.TestCheckResourceAttr("scaleway_container.main", "deploy", "false"),
					resource.TestCheckResourceAttr("scaleway_container.main", "protocol", containerSDK.ContainerProtocolHTTP1.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "local_storage_limit", "1500"),
					resource.TestCheckResourceAttr("scaleway_container.main", "tags.#", "0"),
				),
			},
		},
	})
}

func TestAccContainer_Env(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						environment_variables = {
							"test" = "test"
						}
						secret_environment_variables = {
							"test_secret" = "test_secret"
							"first_secret" = "first_secret"
						}
					}
				`,
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
				Config: `
					resource scaleway_container_namespace main {
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						environment_variables = {
							"foo" = "bar"
						}
						secret_environment_variables = {
							"foo_secret" = "bar_secret"
							"test_secret" = "updated_secret"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					acctest.CheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "environment_variables.foo", "bar"),
					passwordMatchHash("scaleway_container.main", "secret_environment_variables.foo_secret", "bar_secret"),
					passwordMatchHash("scaleway_container.main", "secret_environment_variables.test_secret", "updated_secret"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "secret_environment_variables.first_secret"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						environment_variables = {}
						secret_environment_variables = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					acctest.CheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "environment_variables.%"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "secret_environment_variables.%"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "environment_variables.foo"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "secret_environment_variables.foo_secret"),
				),
			},
		},
	})
}

func TestAccContainer_WithIMG(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	containerNamespace := "test-cr-ns-02"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "%s"
						description = "test container"
					}

					resource scaleway_container main {
						name = "my-container-02"
						description = "environment variables test"
						namespace_id = scaleway_container_namespace.main.id
						registry_image = "nginx:latest"
						port = 80
						cpu_limit = 140
						memory_limit = 256
						min_scale = 3
						max_scale = 5
						timeout = 600
						max_concurrency = 80
						privacy = "private"
						protocol = "h2c"
						deploy = true

						environment_variables = {
							"foo" = "var"
						}
					}
				`, containerNamespace),
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					acctest.CheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "registry_image"),
					resource.TestCheckResourceAttr("scaleway_container.main", "name", "my-container-02"),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "80"),
					resource.TestCheckResourceAttr("scaleway_container.main", "cpu_limit", "140"),
					resource.TestCheckResourceAttr("scaleway_container.main", "memory_limit", "256"),
					resource.TestCheckResourceAttr("scaleway_container.main", "min_scale", "3"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_scale", "5"),
					resource.TestCheckResourceAttr("scaleway_container.main", "timeout", "600"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_concurrency", "80"),
					resource.TestCheckResourceAttr("scaleway_container.main", "deploy", "true"),
					resource.TestCheckResourceAttr("scaleway_container.main", "privacy", containerSDK.ContainerPrivacyPrivate.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "protocol", containerSDK.ContainerProtocolH2c.String()),
				),
			},
		},
	})
}

func TestAccContainer_HTTPOption(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						deploy = false
						http_option = "enabled"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "http_option", containerSDK.ContainerHTTPOptionEnabled.String()),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						deploy = false
						http_option = "redirected"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "http_option", containerSDK.ContainerHTTPOptionRedirected.String()),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						deploy = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "http_option", containerSDK.ContainerHTTPOptionEnabled.String()),
				),
			},
		},
	})
}

func TestAccContainer_Sandbox(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						deploy = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "sandbox"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						deploy = false
						sandbox = "v2"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "sandbox", containerSDK.ContainerSandboxV2.String()),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						deploy = false
						sandbox = "v1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "sandbox", containerSDK.ContainerSandboxV1.String()),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						sandbox = "v2"
					}
				`,
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						deploy = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					// Check default option returned by the API when you don't specify the health_check block.
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.failure_threshold", "30"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.interval", "10s"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						deploy = false

						health_check {
							http {
								path = "/test"
							}
							failure_threshold = 40
							interval = "12s"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.http.0.path", "/test"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.failure_threshold", "40"),
					resource.TestCheckResourceAttr("scaleway_container.main", "health_check.0.interval", "12s"),
				),
			},
		},
	})
}

func TestAccContainer_ScalingOption(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						deploy = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					// Check default option returned by the API when you don't specify the scaling_option block.
					resource.TestCheckResourceAttr("scaleway_container.main", "scaling_option.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "scaling_option.0.concurrent_requests_threshold", "50"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						deploy = false

						scaling_option {
							concurrent_requests_threshold = 15
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "scaling_option.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "scaling_option.0.concurrent_requests_threshold", "15"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						deploy = false

						min_scale = 1

						scaling_option {
							cpu_usage_threshold = 72
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "scaling_option.#", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "scaling_option.0.cpu_usage_threshold", "72"),
				),
			},

			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						deploy = false

						min_scale = 1

						scaling_option {
							memory_usage_threshold = 66
						}
					}
				`,
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isContainerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						command = [ "bash", "-c", "my-script.sh" ]
					}
				`,
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
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						command = [ "bash", "-c", "my-script.sh" ]
						args =    [ "some", "args" ]
					}
				`,
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
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						args =    [ "some", "args" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "command.#", "0"),
					resource.TestCheckResourceAttr("scaleway_container.main", "args.#", "2"),
					resource.TestCheckResourceAttr("scaleway_container.main", "args.0", "some"),
					resource.TestCheckResourceAttr("scaleway_container.main", "args.1", "args"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container.main", "command.#", "0"),
					resource.TestCheckResourceAttr("scaleway_container.main", "args.#", "0"),
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
