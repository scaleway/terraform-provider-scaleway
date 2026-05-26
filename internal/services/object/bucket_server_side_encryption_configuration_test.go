package object_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object"
	objectchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object/testfuncs"
	"github.com/stretchr/testify/require"
)

func TestAccS3BucketServerSideEncryptionConfiguration_basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("sse-config-basic")
	resourceName := "scaleway_object_bucket_server_side_encryption_configuration.test"
	objectBucketTestDefaultProjectId, _ := tt.Meta.ScwClient().GetDefaultProjectID()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_basic(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(tt, resourceName, objectBucketTestDefaultProjectId),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "scaleway_object_bucket.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.bucket_key_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "project_id", objectBucketTestDefaultProjectId),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_basic_withKMS(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("sse-config-basic")
	resourceName := "scaleway_object_bucket_server_side_encryption_configuration.test"
	objectBucketTestDefaultProjectId, _ := tt.Meta.ScwClient().GetDefaultProjectID()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_basic_withKMS(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(tt, resourceName, objectBucketTestDefaultProjectId),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "scaleway_object_bucket.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", "the-key-id"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "aws:kms"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.bucket_key_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "project_id", objectBucketTestDefaultProjectId),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_KMS_withKey(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("sse-config-basic")
	resourceName := "scaleway_object_bucket_server_side_encryption_configuration.test"
	objectBucketTestDefaultProjectId, _ := tt.Meta.ScwClient().GetDefaultProjectID()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_KMS_withKey(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(tt, resourceName, objectBucketTestDefaultProjectId),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "scaleway_object_bucket.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", "my-kms-key-tf-test"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "aws:kms"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.bucket_key_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "project_id", objectBucketTestDefaultProjectId),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_wrongAlgo(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("sse-config-basic")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(
					bucketName, "hehehe-wait-i-dont-exist",
				),
				ExpectError: regexp.MustCompile(`to be one of`),
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_KeyID_withoutKMS(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("sse-config-basic")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_KeyID_withoutKMS(bucketName),
				ExpectError: regexp.MustCompile(
					`InvalidArgument: KMS master key id is only supported when using Server Side Encryption with KMS managed keys`,
				),
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_KMS_withoutBucketKey(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("sse-config-basic")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_KMS_withoutBucketKey(bucketName),
				ExpectError: regexp.MustCompile(
					`InvalidArgument: Bucket key is mandatory when using Server Side Encryption with KMS managed keys`,
				),
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySSEByDefault_AES256(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	rName := sdkacctest.RandomWithPrefix(ResourcePrefix)
	resourceName := "scaleway_object_bucket_server_side_encryption_configuration.test"
	objectBucketTestDefaultProjectId, _ := tt.Meta.ScwClient().GetDefaultProjectID()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, string(awstypes.ServerSideEncryptionAes256)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(tt, resourceName, objectBucketTestDefaultProjectId),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(awstypes.ServerSideEncryptionAes256)),
					resource.TestCheckResourceAttr(resourceName, "rule.0.bucket_key_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "project_id", objectBucketTestDefaultProjectId),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.bucket_key_enabled",
				},
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_sideProject(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("sse-config-basic")
	resourceName := "scaleway_object_bucket_server_side_encryption_configuration.test"

	project, iamAPIKey, terminateFakeSideProject, err := acctest.CreateFakeSideProject(
		tt,
		"ObjectStorageObjectsRead",
		"ObjectStorageBucketsRead",
		"ObjectStorageObjectsWrite",
		"ObjectStorageBucketsWrite",
	)
	require.NoError(t, err)

	ctx := t.Context()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.FakeSideProjectProviders(ctx, tt, project, iamAPIKey),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			func(_ *terraform.State) error {
				return terminateFakeSideProject()
			},
			objectchecks.IsBucketDestroyedFromProject(tt, project.ID),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_sideProject(bucketName, project.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(tt, resourceName, project.ID),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "scaleway_object_bucket.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "project_id", project.ID),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.bucket_key_enabled", "false"),
				),
			},

			// FIXME: the following block breaks the test with the following error
			/*
				Error: Cannot import non-existent remote object

				        While attempting to import an existing object to
				        "scaleway_object_bucket_server_side_encryption_configuration.test", the
				        provider detected that no object exists with the given id. Only pre-existing
				        objects can be imported; check that the id is correct and that it is
				        associated with the provider's configured region or endpoint, or use
				        "terraform apply" to create a new remote object for this resource.
			*/
			// {
			// 	ResourceName:      resourceName,
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// },
		},
	})
}

func testAccCheckBucketServerSideEncryptionConfigurationExists(tt *acctest.TestTools, n, projectId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		bucketRegion := rs.Primary.Attributes["region"]

		conn, err := object.NewS3ClientFromMetaWithProjectID(ctx, tt.Meta, bucketRegion, projectId)
		if err != nil {
			return err
		}

		_, err = findServerSideEncryptionConfiguration(ctx, conn, rs.Primary.Attributes["bucket"])

		return err
	}
}

func findServerSideEncryptionConfiguration(ctx context.Context, conn *s3.Client, bucketName string) (*awstypes.ServerSideEncryptionConfiguration, error) {
	input := s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucketName),
	}

	output, err := conn.GetBucketEncryption(ctx, &input)

	if tfawserr.ErrCodeEquals(err, object.ErrCodeNoSuchBucket, object.ErrCodeServerSideEncryptionConfigurationNotFound) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ServerSideEncryptionConfiguration == nil {
		return nil, fmt.Errorf("nil output for bucket %q", bucketName)
	}

	return output.ServerSideEncryptionConfiguration, nil
}

func testAccBucketServerSideEncryptionConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "scaleway_object_bucket" "test" {
  name = %[1]q
  region = "%[2]s"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name
  region = "%[2]s"

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
`, rName, objectTestsMainRegion)
}

func testAccBucketServerSideEncryptionConfigurationConfig_basic_withKMS(rName string) string {
	return fmt.Sprintf(`
resource "scaleway_object_bucket" "test" {
  name = %[1]q
  region = "%[2]s"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name
  region = "%[2]s"

  rule {
    apply_server_side_encryption_by_default {
	  kms_master_key_id = "the-key-id"
	  sse_algorithm = "aws:kms"
    }
	bucket_key_enabled = true
  }
}
`, rName, objectTestsMainRegion)
}

func testAccBucketServerSideEncryptionConfigurationConfig_KMS_withKey(rName string) string {
	return fmt.Sprintf(`
resource "scaleway_object_bucket" "test" {
  name = %[1]q
  region = "%[2]s"
}

resource "scaleway_key_manager_key" "mykmskey" {
  name        = "my-kms-key-tf-test"
  description = "This key is used to encrypt bucket objects"
  usage       = "asymmetric_encryption"
  algorithm   = "rsa_oaep_4096_sha256"
  unprotected = "true"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name
  region = "%[2]s"

  rule {
    apply_server_side_encryption_by_default {
	  kms_master_key_id = scaleway_key_manager_key.mykmskey.name
	  sse_algorithm = "aws:kms"
    }
	bucket_key_enabled = true
  }
}
`, rName, objectTestsMainRegion)
}

func testAccBucketServerSideEncryptionConfigurationConfig_KeyID_withoutKMS(rName string) string {
	return fmt.Sprintf(`
resource "scaleway_object_bucket" "test" {
  name = %[1]q
  region = "%[2]s"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name
  region = "%[2]s"

  rule {
    apply_server_side_encryption_by_default {
	  kms_master_key_id = "the-key-id"
	  sse_algorithm = "AES256"
    }
  }
}
`, rName, objectTestsMainRegion)
}

func testAccBucketServerSideEncryptionConfigurationConfig_KMS_withoutBucketKey(rName string) string {
	return fmt.Sprintf(`
resource "scaleway_object_bucket" "test" {
  name = %[1]q
  region = "%[2]s"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name
  region = "%[2]s"

  rule {
    apply_server_side_encryption_by_default {
	  kms_master_key_id = "the-key-id"
	  sse_algorithm = "aws:kms"
    }
  }
}
`, rName, objectTestsMainRegion)
}

func testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, sseAlgorithm string) string {
	return fmt.Sprintf(`
resource "scaleway_object_bucket" "test" {
  name = %[1]q
  region = "%[3]s"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name
  region = "%[3]s"

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = %[2]q
    }
  }
}
`, rName, sseAlgorithm, objectTestsMainRegion)
}

func testAccBucketServerSideEncryptionConfigurationConfig_sideProject(rName, projectId string) string {
	return fmt.Sprintf(`
resource "scaleway_object_bucket" "test" {
  name       = %[1]q
  region     = "%[2]s"
  project_id = "%[3]s"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name
  region     = "%[2]s"
  project_id = "%[3]s"

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
`, rName, objectTestsMainRegion, projectId)
}
