package scaleway

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/utils"
)

// getS3ClientWithRegion returns a new S3 client with the correct region extracted from the resource data.
func getS3ClientWithRegion(d *schema.ResourceData, m interface{}) (*s3.S3, utils.Region, error) {
	meta := m.(*Meta)

	region, err := getRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	if region != meta.DefaultRegion {
		newS3Client, err := meta.createS3ClientForRegion(region)
		if err != nil {
			return nil, "", err
		}
		return newS3Client, region, nil
	}

	return meta.s3Client, region, err
}

// getS3ClientWithRegion returns a new S3 client with the correct region and id  extracted from the resource data.
func getS3ClientWithRegionAndID(d *schema.ResourceData, m interface{}) (*s3.S3, utils.Region, string, error) {
	meta := m.(*Meta)

	region, id, err := parseRegionalID(d.Id())
	if err != nil {
		return nil, "", id, err
	}

	if region != meta.DefaultRegion {
		newS3Client, err := meta.createS3ClientForRegion(region)
		if err != nil {
			return nil, "", id, err
		}
		return newS3Client, region, id, nil
	}

	return meta.s3Client, region, id, err

}
