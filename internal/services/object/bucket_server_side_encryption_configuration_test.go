package object_test

import (
	"context"
	"fmt"
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
)

func TestAccS3BucketServerSideEncryptionConfiguration_basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("sse-config-basic")
	resourceName := "scaleway_object_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_basic(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(tt, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "scaleway_object_bucket.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "AES256"),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
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

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySEEByDefault_AES256(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	rName := sdkacctest.RandomWithPrefix(ResourcePrefix)
	resourceName := "scaleway_object_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, string(awstypes.ServerSideEncryptionAes256)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(tt, resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(awstypes.ServerSideEncryptionAes256)),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
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

func testAccCheckBucketServerSideEncryptionConfigurationExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		bucketRegion := rs.Primary.Attributes["region"]

		conn, err := object.NewS3ClientFromMeta(ctx, tt.Meta, bucketRegion)
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
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
`, rName)
}

func testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, sseAlgorithm string) string {
	return fmt.Sprintf(`
resource "scaleway_object_bucket" "test" {
  name = %[1]q
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = %[2]q
    }
  }
}
`, rName, sseAlgorithm)
}
