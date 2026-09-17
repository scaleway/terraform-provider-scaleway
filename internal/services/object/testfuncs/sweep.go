package objecttestfuncs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
	return acctest.SweepRegions([]scw.Region{scw.RegionFrPar, scw.RegionNlAms, scw.RegionPlWaw}, func(_ *scw.Client, region scw.Region) error {
		s3client, err := object.SharedS3ClientForRegion(region)
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

			if acctest.IsTestResource(*bucket.Name) {
				_, err := s3client.DeleteBucket(ctx, &s3.DeleteBucketInput{
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
