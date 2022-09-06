package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

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
	// ErrCodeNoSuchWebsiteConfiguration website configuration not found
	ErrCodeNoSuchWebsiteConfiguration = "NoSuchWebsiteConfiguration"
)

// RetryWhenAWSErrEqualsContext retries the specified function when it returns one of the specified AWS error.
func RetryWhenAWSErrEqualsContext(ctx context.Context, timeout time.Duration, f func() (interface{}, error), awsErrors ...error) (interface{}, error) { // nosemgrep:ci.aws-in-func-name
	return RetryWhenContext(ctx, timeout, f, func(err error) (bool, error) {
		for _, awsError := range awsErrors {
			if isS3Err(err, awsError) {
				return true, err
			}
		}
		return false, err
	})
}

// Retryable is a function that is used to decide if a function's error is retryable or not.
// The error argument can be `nil`.
// If the error is retryable, returns a bool value of `true` and an error (not necessarily the error passed as the argument).
// If the error is not retryable, returns a bool value of `false` and either no error (success state) or an error (not necessarily the error passed as the argument).
type Retryable func(error) (bool, error)

// RetryWhenContext retries the function `f` when the error it returns satisfies `predicate`.
// `f` is retried until `timeout` expires.
func RetryWhenContext(ctx context.Context, timeout time.Duration, f func() (interface{}, error), retryable Retryable) (interface{}, error) {
	var output interface{}

	err := resource.RetryContext(ctx, timeout, func() *resource.RetryError {
		var err error
		var retry bool

		output, err = f()
		retry, err = retryable(err)

		if retry {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if TimedOut(err) {
		output, err = f()
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
