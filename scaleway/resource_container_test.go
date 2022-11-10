package scaleway

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var testDockerIMG = "docker.io/library/nginx:alpine"

func init() {
	resource.AddTestSweepers("scaleway_container", &resource.Sweeper{
		Name: "scaleway_container",
		F:    testSweepContainer,
	})
}

func testSweepContainer(_ string) error {
	return sweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		containerAPI := container.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the container in (%s)", region)
		listNamespaces, err := containerAPI.ListContainers(
			&container.ListContainersRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing containers in (%s) in sweeper: %s", region, err)
		}

		for _, cont := range listNamespaces.Containers {
			_, err := containerAPI.DeleteContainer(&container.DeleteContainerRequest{
				ContainerID: cont.ID,
				Region:      region,
			})
			if err != nil {
				l.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting container in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayContainer_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayContainerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerExists(tt, "scaleway_container.main"),
					testCheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					testCheckResourceAttrUUID("scaleway_container.main", "id"),
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
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerExists(tt, "scaleway_container.main"),
					testCheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "name", "my-container-tf"),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "8080"),
					resource.TestCheckResourceAttr("scaleway_container.main", "cpu_limit", "70"),
					resource.TestCheckResourceAttr("scaleway_container.main", "memory_limit", "128"),
					resource.TestCheckResourceAttr("scaleway_container.main", "min_scale", "0"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_scale", "20"),
					resource.TestCheckResourceAttr("scaleway_container.main", "timeout", "300"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_concurrency", "50"),
					resource.TestCheckResourceAttr("scaleway_container.main", "deploy", "false"),
					resource.TestCheckResourceAttr("scaleway_container.main", "privacy", container.ContainerPrivacyPublic.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "protocol", container.ContainerProtocolHTTP1.String()),
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
						memory_limit 	= 256
						cpu_limit		= 140
						deploy       	= false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerExists(tt, "scaleway_container.main"),
					testCheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "name", "my-container-tf"),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "5000"),
					resource.TestCheckResourceAttr("scaleway_container.main", "cpu_limit", "140"),
					resource.TestCheckResourceAttr("scaleway_container.main", "memory_limit", "256"),
					resource.TestCheckResourceAttr("scaleway_container.main", "min_scale", "1"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_scale", "2"),
					resource.TestCheckResourceAttr("scaleway_container.main", "timeout", "300"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_concurrency", "80"),
					resource.TestCheckResourceAttr("scaleway_container.main", "deploy", "false"),
					resource.TestCheckResourceAttr("scaleway_container.main", "protocol", container.ContainerProtocolHTTP1.String()),
				),
			},
		},
	})
}

func TestAccScalewayContainer_Env(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayContainerDestroy(tt),
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
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerExists(tt, "scaleway_container.main"),
					testCheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					testCheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "environment_variables.test", "test"),
					resource.TestCheckResourceAttr("scaleway_container.main", "secret_environment_variables.test_secret", "test_secret"),
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
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerExists(tt, "scaleway_container.main"),
					testCheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					testCheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "environment_variables.foo", "bar"),
					resource.TestCheckResourceAttr("scaleway_container.main", "secret_environment_variables.foo_secret", "bar_secret"),
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
					testAccCheckScalewayContainerExists(tt, "scaleway_container.main"),
					testCheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					testCheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "environment_variables.%"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "secret_environment_variables.%"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "environment_variables.foo"),
					resource.TestCheckNoResourceAttr("scaleway_container.main", "secret_environment_variables.foo_secret"),
				),
			},
		},
	})
}

func TestAccScalewayContainer_WithIMG(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	containerNamespace := "test-cr-ns-02"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayContainerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "%s"
						description = "test container"
					}
				`, containerNamespace),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "%s"
						description = "test container"
					}
				`, containerNamespace),
				Check: resource.ComposeTestCheckFunc(
					testConfigContainerNamespace(tt, "scaleway_container_namespace.main"),
				),
			},
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
						registry_image = "${scaleway_container_namespace.main.registry_endpoint}/nginx:test"
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
					testAccCheckScalewayContainerExists(tt, "scaleway_container.main"),
					testCheckResourceAttrUUID("scaleway_container.main", "id"),
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
					resource.TestCheckResourceAttr("scaleway_container.main", "privacy", container.ContainerPrivacyPrivate.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "protocol", container.ContainerProtocolH2c.String()),
				),
			},
		},
	})
}

func testAccCheckScalewayContainerExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource container not found: %s", n)
		}

		api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetContainer(&container.GetContainerRequest{
			ContainerID: id,
			Region:      region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayContainerDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_container_namespace" {
				continue
			}

			api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteContainer(&container.DeleteContainerRequest{
				ContainerID: id,
				Region:      region,
			})

			if err == nil {
				return fmt.Errorf("container (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}

func testConfigContainerNamespace(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Do not execute docker requests when running with cassettes
		if !*UpdateCassettes {
			return nil
		}

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		ns, err := api.WaitForNamespace(&container.WaitForNamespaceRequest{
			NamespaceID: id,
			Region:      region,
		})
		if err != nil {
			return fmt.Errorf("error waiting namespace: %v", err)
		}

		meta := tt.Meta
		var errorMessage ErrorRegistryMessage

		accessKey, _ := meta.scwClient.GetAccessKey()
		secretKey, _ := meta.scwClient.GetSecretKey()
		authConfig := types.AuthConfig{
			ServerAddress: ns.RegistryEndpoint,
			Username:      accessKey,
			Password:      secretKey,
		}

		cli, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
		if err != nil {
			return fmt.Errorf("could not connect to Docker: %v", err)
		}

		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			return fmt.Errorf("could not marshal auth config: %v", err)
		}

		ctx := context.Background()
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)

		out, err := cli.ImagePull(ctx, testDockerIMG, types.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("could not pull image: %v", err)
		}

		defer out.Close()

		buffIOReader := bufio.NewReader(out)
		for {
			streamBytes, errPull := buffIOReader.ReadBytes('\n')
			if errPull == io.EOF {
				break
			}
			err = json.Unmarshal(streamBytes, &errorMessage)
			if err != nil {
				return fmt.Errorf("could not unmarshal: %v", err)
			}

			if errorMessage.Error != "" {
				return fmt.Errorf(errorMessage.Error)
			}
		}

		imageTag := testDockerIMG
		scwTag := ns.RegistryEndpoint + "/nginx:test"

		err = cli.ImageTag(ctx, imageTag, scwTag)
		if err != nil {
			return fmt.Errorf("could not tag image: %v", err)
		}

		pusher, err := cli.ImagePush(ctx, scwTag, types.ImagePushOptions{RegistryAuth: authStr})
		if err != nil {
			return fmt.Errorf("could not push image: %v", err)
		}

		defer pusher.Close()

		buffIOReader = bufio.NewReader(pusher)
		for {
			streamBytes, errPush := buffIOReader.ReadBytes('\n')
			if errPush == io.EOF {
				break
			}
			err = json.Unmarshal(streamBytes, &errorMessage)
			if err != nil {
				return fmt.Errorf("could not unmarshal: %v", err)
			}

			if errorMessage.Error != "" {
				return fmt.Errorf(errorMessage.Error)
			}
		}

		_, err = api.WaitForNamespace(&container.WaitForNamespaceRequest{
			NamespaceID: id,
			Region:      region,
		})
		if err != nil {
			return fmt.Errorf("error waiting namespace: %v", err)
		}

		return nil
	}
}
