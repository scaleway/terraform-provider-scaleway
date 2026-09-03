package objecttestfuncs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_object_bucket", &resource.Sweeper{
		Name: "scaleway_object_bucket",
		F:    testSweepStorageObjectBucket,
	})
}

func testSweepStorageObjectBucket(_ string) error {
	return acctest.SweepRegions([]scw.Region{scw.RegionFrPar, scw.RegionNlAms, scw.RegionPlWaw}, func(scwClient *scw.Client, region scw.Region) error {
		// List projects
		accountAPI := accountSDK.NewProjectAPI(scwClient)
		req := &accountSDK.ProjectAPIListProjectsRequest{}

		listProjects, err := accountAPI.ListProjects(req, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("failed to list projects: %s", err)

			return nil
		}

		// For each project, delete all buckets
		for _, p := range listProjects.Projects {
			s3client, err := object.SharedS3ClientForRegionWithProjectID(region, p.ID)
			ctx := context.Background()

			if err != nil {
				logging.L.Warningf("error getting client: %s", err)

				return nil
			}

			listBucketResponse, err := s3client.ListBuckets(ctx, &s3.ListBucketsInput{})
			if err != nil {
				logging.L.Warningf("couldn't list buckets: %s", err)

				return nil
			}

			for _, bucket := range listBucketResponse.Buckets {
				logging.L.Debugf("Deleting %q bucket", *bucket.Name)

				if !acctest.IsTestResource(*bucket.Name) {
					continue
				}

				err := EmptyBucket(ctx, s3client, bucket.Name)
				if err != nil {
					// We do not "continue" here, to still attempt deleting the bucket
					logging.L.Warningf("error deleting bucket objects in Sweeper: %s", err)
				}

				_, err = s3client.DeleteBucket(ctx, &s3.DeleteBucketInput{
					Bucket: bucket.Name,
				})
				if err != nil {
					logging.L.Warningf("error deleting bucket in Sweeper: %s", err)
				}
			}
		}

		return nil
	})
}

// EmptyBucket deletes all objects in a standard (non-versioned) S3 bucket.
func EmptyBucket(ctx context.Context, client *s3.Client, bucketName *string) error {
	logging.L.Debugf("Deleting objects from %q bucket", *bucketName)

	// Create a paginator to handle buckets with more than 1000 objects
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: bucketName,
	})

	// Iterate through each page of objects
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get page: %w", err)
		}

		// If the page is empty, we're done
		if len(page.Contents) == 0 {
			continue
		}

		// Prepare the list of ObjectIdentifiers to delete
		var objectIds []types.ObjectIdentifier
		for _, obj := range page.Contents {
			objectIds = append(objectIds, types.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		// Batch deletion (max 1000 per request)
		_, err = client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: bucketName,
			Delete: &types.Delete{
				Objects: objectIds,
				Quiet:   aws.Bool(true), // Set to true to only return errors in the response
			},
		})
		if err != nil {
			return fmt.Errorf("failed to delete objects batch: %w", err)
		}

		logging.L.Debugf("Successfully deleted a batch of %d objects", len(objectIds))
	}

	return nil
}
