package scaleway

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func newS3Client(region, accessKey, secretKey string) (*s3.S3, error) {

	config := &aws.Config{}
	config.WithRegion(region)
	config.WithCredentials(credentials.NewStaticCredentials(accessKey, secretKey, ""))
	config.WithEndpoint("https://s3." + region + ".scw.cloud")

	s, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	return s3.New(s), nil
}

func newS3ClientFromMeta(meta *Meta) (*s3.S3, error) {
	region, _ := meta.scwClient.GetDefaultRegion()
	accessKey, _ := meta.scwClient.GetAccessKey()
	secretKey, _ := meta.scwClient.GetSecretKey()
	return newS3Client(region.String(), accessKey, secretKey)
}

// s3ClientWithRegion returns a new S3 client with the correct region extracted from the resource data.
func s3ClientWithRegion(d *schema.ResourceData, m interface{}) (*s3.S3, scw.Region, error) {
	meta := m.(*Meta)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	accessKey, _ := meta.scwClient.GetAccessKey()
	secretKey, _ := meta.scwClient.GetSecretKey()

	s3Client, err := newS3Client(region.String(), accessKey, secretKey)
	if err != nil {
		return nil, "", err
	}

	return s3Client, region, err
}

// s3ClientWithRegion returns a new S3 client with the correct region and name extracted from the resource data.
func s3ClientWithRegionAndName(m interface{}, name string) (*s3.S3, scw.Region, string, error) {
	meta := m.(*Meta)

	region, name, err := parseRegionalID(name)
	if err != nil {
		return nil, "", name, err
	}
	accessKey, _ := meta.scwClient.GetAccessKey()
	secretKey, _ := meta.scwClient.GetSecretKey()

	s3Client, err := newS3Client(region.String(), accessKey, secretKey)
	if err != nil {
		return nil, "", "", err
	}

	return s3Client, region, name, err
}
