package object

import (
	"context"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// SharedS3ClientForRegion returns a common S3 client needed for the sweeper
func SharedS3ClientForRegion(region scw.Region) (*s3.S3, error) {
	ctx := context.Background()
	m, err := meta.NewMeta(ctx, &meta.Config{
		TerraformVersion: "terraform-tests",
		ForceZone:        region.GetZones()[0],
	})
	if err != nil {
		return nil, err
	}
	return NewS3ClientFromMeta(m, region.String())
}
