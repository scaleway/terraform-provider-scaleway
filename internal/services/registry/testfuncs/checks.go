package registrytestfuncs

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/image"
	dockerRegistrySDK "github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/registry"
)

var testDockerIMG = "docker.io/library/alpine:latest"

// PushImageToRegistry is a helper function that pulls an image, tags it for the Scaleway registry, and pushes it to the registry.
func PushImageToRegistry(tt *acctest.TestTools, registryEndpoint string, tagName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Do not execute Docker requests when running with cassettes
		if !*acctest.UpdateCassettes {
			return nil
		}

		meta := tt.Meta
		var errorMessage registry.ErrorRegistryMessage

		// Retrieve access and secret keys for authentication
		accessKey, _ := meta.ScwClient().GetAccessKey()
		secretKey, _ := meta.ScwClient().GetSecretKey()

		// Auth configuration for the registry
		authConfig := dockerRegistrySDK.AuthConfig{
			ServerAddress: registryEndpoint,
			Username:      accessKey,
			Password:      secretKey,
		}

		// Initialize Docker client
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return fmt.Errorf("could not connect to Docker: %v", err)
		}

		// Encode auth config into base64
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			return fmt.Errorf("could not marshal auth config: %v", err)
		}

		ctx := context.Background()
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)

		// Pull the Docker image
		out, err := cli.ImagePull(ctx, testDockerIMG, image.PullOptions{})
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
				return errors.New(errorMessage.Error)
			}
		}

		// Tag the image for Scaleway registry
		scwTag := registryEndpoint + "/alpine:" + tagName
		err = cli.ImageTag(ctx, testDockerIMG, scwTag)
		if err != nil {
			return fmt.Errorf("could not tag image: %v", err)
		}

		// Push the image to Scaleway registry
		pusher, err := cli.ImagePush(ctx, scwTag, image.PushOptions{RegistryAuth: authStr})
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
				return errors.New(errorMessage.Error)
			}
		}

		return nil
	}
}
