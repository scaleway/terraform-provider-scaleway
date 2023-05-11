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
	// ErrCodeValidationException validation exception
	ErrCodeValidationException = "ValidationException"
)
