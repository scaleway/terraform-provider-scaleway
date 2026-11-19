package object

import (
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// This file exposes unexported helpers to the object_test (external test)
// package. It only compiles during `go test`.

func SetBucketState(ctx context.Context, d *schema.ResourceData, m any, bucketName string, region scw.Region, s3Client *s3.Client) diag.Diagnostics {
	return setBucketState(ctx, d, m, bucketName, region, s3Client)
}

func BucketSchema() map[string]*schema.Schema {
	return bucketSchema()
}

func NewS3Client(ctx context.Context, region, accessKey, secretKey, customEndpoint string, usePathStyle bool, httpClient *http.Client) (*s3.Client, error) {
	return newS3Client(ctx, region, accessKey, secretKey, customEndpoint, usePathStyle, httpClient)
}
