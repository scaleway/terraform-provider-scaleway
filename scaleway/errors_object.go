package scaleway

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#pkg-constants

const (
	// ErrCodeNoSuchTagSet tag(s) not found
	ErrCodeNoSuchTagSet = "NoSuchTagSet"
	// ErrCodeNoSuchCORSConfiguration CORS configuration not found
	ErrCodeNoSuchCORSConfiguration = "NoSuchCORSConfiguration"
	// ErrCodeNoSuchLifecycleConfiguration lifecycle configuration rule not found
	ErrCodeNoSuchLifecycleConfiguration = "NoSuchLifecycleConfiguration"
	// ErrCodeAccessDenied action on resource is denied
	ErrCodeAccessDenied = "AccessDenied"
	// ErrCodeBucketNotEmpty bucket is not empty
	ErrCodeBucketNotEmpty = "BucketNotEmpty"
	// ErrCodeNoSuchBucketPolicy policy not found
	ErrCodeNoSuchBucketPolicy = "NoSuchBucketPolicy"
)
