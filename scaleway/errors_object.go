package scaleway

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
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
	// ErrCodeObjectLockConfigurationNotFoundError object lock configuration not found
	ErrCodeObjectLockConfigurationNotFoundError = "ObjectLockConfigurationNotFoundError"
	// ErrCodeAuthorizationError authorization error
	ErrCodeAuthorizationError = "AuthorizationError"
	// ErrCodeInternalException internal exception
	ErrCodeInternalException = "InternalException"
	// ErrCodeInternalServiceError internal exception error
	ErrCodeInternalServiceError = "InternalServiceError"
	// ErrCodeInvalidAction invalid action
	ErrCodeInvalidAction = "InvalidAction"
	// ErrCodeInvalidParameterException invalid parameter exception
	ErrCodeInvalidParameterException = "InvalidParameterException"
	// ErrCodeInvalidParameterValue invalid parameter value
	ErrCodeInvalidParameterValue = "InvalidParameterValue"
	// ErrCodeInvalidRequest invalid request
	ErrCodeInvalidRequest = "InvalidRequest"
	// ErrCodeOperationDisabledException operation disabled exception
	ErrCodeOperationDisabledException = "OperationDisabledException"
	// ErrCodeOperationNotPermitted operation not permitted
	ErrCodeOperationNotPermitted = "OperationNotPermitted"
	// ErrCodeUnknownOperationException   unknown operation exception
	ErrCodeUnknownOperationException = "UnknownOperationException"
	// ErrCodeUnsupportedFeatureException = unsupported Feature exception
	ErrCodeUnsupportedFeatureException = "UnsupportedFeatureException"
	// ErrCodeUnsupportedOperation unsupported operation
	ErrCodeUnsupportedOperation = "UnsupportedOperation"
	// ErrCodeValidationError validation error
	ErrCodeValidationError = "ValidationError"
	// ErrCodeValidationException  validation exception
	ErrCodeValidationException = "ValidationException"
)

// errorISOUnsupported checks the partition and specific error to make
// an educated guess about whether the problem stems from a feature not being
// available in ISO (or non-standard partitions) that is normally available.
// true means that there is an error AND it suggests a feature is not supported
// in ISO. Be careful with false, which means either there is NO error or there
// is an error but not one that suggests an unsupported feature in ISO.
func errorISOUnsupported(partition string, err error) bool {
	if partition == endpoints.AwsPartitionID {
		return false
	}

	if err == nil { // not strictly necessary but make logic clearer
		return false
	}

	if tfawserr.ErrCodeContains(err, ErrCodeAccessDenied) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeAuthorizationError) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInternalException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInternalServiceError) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInvalidAction) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInvalidParameterException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInvalidParameterValue) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeInvalidRequest) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeOperationDisabledException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeOperationNotPermitted) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeUnknownOperationException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeUnsupportedFeatureException) {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeUnsupportedOperation) {
		return true
	}

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not support tagging") {
		return true
	}

	if tfawserr.ErrCodeContains(err, ErrCodeValidationException) {
		return true
	}

	return false
}
