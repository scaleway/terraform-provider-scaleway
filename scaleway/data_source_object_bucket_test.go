package scaleway

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	accountV2 "github.com/scaleway/scaleway-sdk-go/api/account/v2"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestAccScalewayDataSourceObjectStorage_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket")
	// resourceName := "data.scaleway_object_bucket.main"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_object_bucket" "base-01" {
					name = "%s"
					tags = {
						foo = "bar"
					}
				}

				data "scaleway_object_bucket" "selected" {
					name = scaleway_object_bucket.base-01.name
				}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_bucket.selected", "name", bucketName),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceObjectStorage_ProjectIDAllowed(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket")

	project, iamAPIKey, terminateFakeSideProject := createFakeSideProject(tt)
	defer terminateFakeSideProject()

	ctx := context.Background()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: fakeSideProjectProviders(ctx, tt, project, iamAPIKey),
		CheckDestroy:      testAccCheckScalewayObjectDestroy(tt),
		Steps: []resource.TestStep{
			// Create a bucket from the main provider into the side project and read it from the side provider
			// The side provider should only be able to read the bucket from the side project
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base" {
						name = "%[1]s"
						project_id = "%[2]s"
					}

					data "scaleway_object_bucket" "selected" {
						name = scaleway_object_bucket.base.name
						provider = side
					}
				`,
					bucketName,
					project.ID,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_bucket.selected", "name", bucketName),
					resource.TestCheckResourceAttr("data.scaleway_object_bucket.selected", "project_id", project.ID),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceObjectStorage_ProjectIDForbidden(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket")

	project, iamAPIKey, terminateFakeSideProject := createFakeSideProject(tt)
	defer terminateFakeSideProject()

	ctx := context.Background()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: fakeSideProjectProviders(ctx, tt, project, iamAPIKey),
		CheckDestroy:      testAccCheckScalewayObjectDestroy(tt),
		Steps: []resource.TestStep{
			// The side provider should not be able to read the bucket from the main project
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base" {
						name = "%[1]s"
					}

					data "scaleway_object_bucket" "selected" {
						name = scaleway_object_bucket.base.name
						provider = side
					}
				`,
					bucketName,
					project.ID,
				),
				ExpectError: regexp.MustCompile("failed getting Object Storage bucket"),
			},
		},
	})
}

func createFakeSideProject(tt *TestTools) (*accountV2.Project, *iam.APIKey, func()) {
	t := tt.T

	terminateFunctions := []func(){}

	projectName := sdkacctest.RandomWithPrefix("test-acc-scaleway-project")
	iamApplicationName := sdkacctest.RandomWithPrefix("test-acc-scaleway-iam-app")
	iamPolicyName := sdkacctest.RandomWithPrefix("test-acc-scaleway-iam-policy")

	projectAPI := accountV2.NewAPI(tt.Meta.scwClient)
	project, err := projectAPI.CreateProject(&accountV2.CreateProjectRequest{
		Name: projectName,
	})
	require.NoError(t, err)
	terminateFunctions = append(terminateFunctions, func() {
		err := projectAPI.DeleteProject(&accountV2.DeleteProjectRequest{
			ProjectID: project.ID,
		})
		require.NoError(t, err)
	})

	iamAPI := iam.NewAPI(tt.Meta.scwClient)
	iamApplication, err := iamAPI.CreateApplication(&iam.CreateApplicationRequest{
		Name: iamApplicationName,
	})
	require.NoError(t, err)
	terminateFunctions = append(terminateFunctions, func() {
		err := iamAPI.DeleteApplication(&iam.DeleteApplicationRequest{
			ApplicationID: iamApplication.ID,
		})
		require.NoError(t, err)
	})

	iamPolicy, err := iamAPI.CreatePolicy(&iam.CreatePolicyRequest{
		Name:          iamPolicyName,
		ApplicationID: expandStringPtr(iamApplication.ID),
		Rules: []*iam.RuleSpecs{
			{
				ProjectIDs:         &[]string{project.ID},
				PermissionSetNames: &[]string{"ObjectStorageReadOnly", "ObjectStorageObjectsRead", "ObjectStorageBucketsRead"},
			},
		},
	})
	require.NoError(t, err)
	terminateFunctions = append(terminateFunctions, func() {
		err := iamAPI.DeletePolicy(&iam.DeletePolicyRequest{
			PolicyID: iamPolicy.ID,
		})
		require.NoError(t, err)
	})

	iamAPIKey, err := iamAPI.CreateAPIKey(&iam.CreateAPIKeyRequest{
		ApplicationID:    expandStringPtr(iamApplication.ID),
		DefaultProjectID: &project.ID,
	})
	require.NoError(t, err)
	terminateFunctions = append(terminateFunctions, func() {
		err := iamAPI.DeleteAPIKey(&iam.DeleteAPIKeyRequest{
			AccessKey: iamAPIKey.AccessKey,
		})
		require.NoError(t, err)
	})

	return project, iamAPIKey, func() {
		for i := len(terminateFunctions) - 1; i >= 0; i-- {
			terminateFunctions[i]()
		}
	}
}

func fakeSideProjectProviders(ctx context.Context, tt *TestTools, project *accountV2.Project, iamAPIKey *iam.APIKey) map[string]func() (*schema.Provider, error) {
	t := tt.T

	metaSide, err := buildMeta(ctx, &metaConfig{
		terraformVersion:    "terraform-tests",
		httpClient:          tt.Meta.httpClient,
		forceProjectID:      project.ID,
		forceOrganizationID: project.OrganizationID,
		forceAccessKey:      iamAPIKey.AccessKey,
		forceSecretKey:      *iamAPIKey.SecretKey,
	})
	require.NoError(t, err)

	providers := map[string]func() (*schema.Provider, error){
		"side": func() (*schema.Provider, error) {
			return Provider(&ProviderConfig{Meta: metaSide})(), nil
		},
	}

	for k, v := range tt.ProviderFactories {
		providers[k] = v
	}

	return providers
}
