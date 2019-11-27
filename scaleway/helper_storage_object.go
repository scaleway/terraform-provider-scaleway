package scaleway

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// getS3ClientWithRegion returns a new S3 client with the correct region extracted from the resource data.
func getS3ClientWithRegion(d *schema.ResourceData, m interface{}) (*s3.S3, scw.Region, error) {
	meta := m.(*Meta)

	region, err := getRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	if region != meta.DefaultRegion {
		// if the region is not the same as the default region:
		// we have to clone the meta object with the new region and create a new S3 client.
		newMeta := *meta
		newMeta.DefaultRegion = region

		err := newMeta.bootstrapS3Client()
		if err != nil {
			return nil, "", err
		}
		return newMeta.s3Client, region, nil
	}

	return meta.s3Client, region, err
}

// getS3ClientWithRegion returns a new S3 client with the correct region and name extracted from the resource data.
func getS3ClientWithRegionAndName(m interface{}, name string) (*s3.S3, scw.Region, string, error) {
	meta := m.(*Meta)

	region, name, err := parseRegionalID(name)
	if err != nil {
		return nil, "", name, err
	}

	if region != meta.DefaultRegion {
		// if the region is not the same as the default region:
		// we have to clone the meta object with the new region and create a new S3 client.
		newMeta := *meta
		newMeta.DefaultRegion = region

		err := newMeta.bootstrapS3Client()
		if err != nil {
			return nil, "", name, err
		}
		return newMeta.s3Client, region, name, nil
	}

	return meta.s3Client, region, name, err

}
