package iam

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	annotations "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// TODO: make these functions resource agnostic and move them to a global helper

// getIdentifierKey returns the ID of the annotation key "iam_terraform_identifier".
// Returns empty string if not found.
func (r *ApiKeyEphemeralResource) getIdentifierKey(ctx context.Context, organizationID string) (string, error) {
	keysResp, err := r.annotationsAPI.ListKeys(&annotations.ListKeysRequest{
		OrganizationID: organizationID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return "", fmt.Errorf("failed to list annotation keys: %w", err)
	}

	for _, key := range keysResp.Keys {
		if key.Name == iamTerraformIdentifierKey {
			return key.ID, nil
		}
	}

	return "", nil
}

// getOrCreateIdentifierKey returns the ID of the annotation key "iam_terraform_identifier",
// creating it if it doesn't exist.
func (r *ApiKeyEphemeralResource) getOrCreateIdentifierKey(ctx context.Context, organizationID string) (string, error) {
	keyID, err := r.getIdentifierKey(ctx, organizationID)
	if err != nil {
		return "", err
	}

	if keyID != "" {
		return keyID, nil
	}

	newKey, err := r.annotationsAPI.CreateKey(&annotations.CreateKeyRequest{
		OrganizationID: organizationID,
		Name:           iamTerraformIdentifierKey,
		Description:    "Identity key for Terraform-managed resources",
	}, scw.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("failed to create identity key: %w", err)
	}

	return newKey.ID, nil
}

// getIdentifierValue returns the ID of the annotation value with the given name under the specified key.
// Returns empty string if not found.
func (r *ApiKeyEphemeralResource) getIdentifierValue(ctx context.Context, keyID, annotationValue string) (string, error) {
	valuesResp, err := r.annotationsAPI.ListValues(&annotations.ListValuesRequest{
		KeyID: &keyID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return "", fmt.Errorf("failed to list annotation values: %w", err)
	}

	for _, value := range valuesResp.Values {
		if value.Name == annotationValue {
			return value.ID, nil
		}
	}

	return "", nil
}

// getOrCreateIdentifierValue returns the ID of the annotation value with the given name,
// creating it if it doesn't exist.
func (r *ApiKeyEphemeralResource) getOrCreateIdentifierValue(ctx context.Context, keyID, annotationValue string) (string, error) {
	identityValueID, err := r.getIdentifierValue(ctx, keyID, annotationValue)
	if err != nil {
		return "", err
	}

	if identityValueID != "" {
		return identityValueID, nil
	}

	newValue, err := r.annotationsAPI.CreateValue(&annotations.CreateValueRequest{
		KeyID:       keyID,
		Name:        annotationValue,
		Description: "Identity: " + annotationValue,
	}, scw.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("failed to create identity value: %w", err)
	}

	return newValue.ID, nil
}

// getAPIKeyFromBinding retrieves an API key from a binding's SRN.
func getAPIKeyFromBinding(ctx context.Context, iamAPI *iam.API, binding *annotations.Binding) (*iam.APIKey, error) {
	// Extract access key from SRN (take everything after the last /)
	accessKey := binding.Srn
	if idx := strings.LastIndex(binding.Srn, "/"); idx != -1 {
		accessKey = binding.Srn[idx+1:]
	}

	if accessKey == "" {
		return nil, fmt.Errorf("invalid SRN format: %s", binding.Srn)
	}

	apiKey, err := iamAPI.GetAPIKey(&iam.GetAPIKeyRequest{
		AccessKey: accessKey,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return apiKey, nil
}

// getUniqueBinding retrieves the unique binding for a given value ID.
// Returns nil if no binding exists, or an error if multiple bindings are found.
func (r *ApiKeyEphemeralResource) getUniqueBinding(ctx context.Context, organizationID, valueID string) (*annotations.Binding, error) {
	bindingsResp, err := r.annotationsAPI.ListBindings(&annotations.ListBindingsRequest{
		OrganizationID: organizationID,
		ValueID:        &valueID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list bindings: %w", err)
	}

	if len(bindingsResp.Bindings) == 0 {
		return nil, nil
	}

	if len(bindingsResp.Bindings) > 1 {
		return nil, fmt.Errorf("found %d bindings for identity value, expected at most 1", len(bindingsResp.Bindings))
	}

	return bindingsResp.Bindings[0], nil
}

// findExistingAPIKeyByAnnotations looks for an existing API key that has a binding
// to the annotation value specified by use_annotations_identifier.
// It returns the API key if exactly one is found, nil if none is found, or an error if multiple are found.
//
// The identifier is stored using a special annotation key "iam_terraform_identifier" with
// the value being the string provided in use_annotations_identifier.
func (r *ApiKeyEphemeralResource) findExistingAPIKeyByAnnotations(ctx context.Context, annotationValue, organizationID string) (*iam.APIKey, error) {
	identifierKeyID, err := r.getIdentifierKey(ctx, organizationID)
	if identifierKeyID == "" || err != nil {
		return nil, err
	}

	matchingValueID, err := r.getIdentifierValue(ctx, identifierKeyID, annotationValue)
	if matchingValueID == "" || err != nil {
		return nil, err
	}

	binding, err := r.getUniqueBinding(ctx, organizationID, matchingValueID)
	if binding == nil || err != nil {
		return nil, err
	}

	return getAPIKeyFromBinding(ctx, r.iamAPI, binding)
}

// getOrCreateBinding ensures a binding exists between the given SRN and value ID.
// Returns nil if the binding already exists or was successfully created.
// Errors if more than one binding is found.
func (r *ApiKeyEphemeralResource) getOrCreateBinding(ctx context.Context, apiKeySRN, identityValueID, organizationID string) error {
	existingBindingsResp, err := r.annotationsAPI.ListBindings(&annotations.ListBindingsRequest{
		OrganizationID: organizationID,
		ValueID:        &identityValueID,
		Srn:            &apiKeySRN,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return fmt.Errorf("failed to list existing bindings: %w", err)
	}

	if len(existingBindingsResp.Bindings) > 1 {
		return fmt.Errorf("found %d bindings for the given SRN and value ID, expected at most 1", len(existingBindingsResp.Bindings))
	}

	if len(existingBindingsResp.Bindings) == 1 {
		return nil
	}

	_, err = r.annotationsAPI.CreateBinding(&annotations.CreateBindingRequest{
		Srn:     apiKeySRN,
		ValueID: identityValueID,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to create binding: %w", err)
	}

	return nil
}

// getOrCreateIdentityAnnotations ensures that the annotation key, value, and binding
// exist for the given API key SRN and annotation value. It creates them if they don't exist.
func (r *ApiKeyEphemeralResource) getOrCreateIdentityAnnotations(ctx context.Context, apiKeySRN, annotationValue, organizationID string) error {
	identifierKeyID, err := r.getOrCreateIdentifierKey(ctx, organizationID)
	if identifierKeyID == "" || err != nil {
		return err
	}

	identityValueID, err := r.getOrCreateIdentifierValue(ctx, identifierKeyID, annotationValue)
	if identityValueID == "" || err != nil {
		return err
	}

	return r.getOrCreateBinding(ctx, apiKeySRN, identityValueID, organizationID)
}

// findExistingAPIKeyByDescription looks for an existing API key with the given description.
// It returns the API key if exactly one is found, nil if none is found, or an error if multiple are found.
func (r *ApiKeyEphemeralResource) findExistingAPIKeyByDescription(ctx context.Context, description, organizationID string) (*iam.APIKey, error) {
	apiKeys, err := r.iamAPI.ListAPIKeys(&iam.ListAPIKeysRequest{
		OrganizationID: &organizationID,
		Description:    &description,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	if len(apiKeys.APIKeys) == 0 {
		return nil, nil
	}

	if len(apiKeys.APIKeys) > 1 {
		return nil, fmt.Errorf("found %d API keys with description %q, expected at most 1", len(apiKeys.APIKeys), description)
	}

	return apiKeys.APIKeys[0], nil
}

// setApiKeyData sets the resource data from an API key.
// If includeSecretKey is true, the secret key will be set; otherwise it will be null.
func setApiKeyData(data *ApiKeyEphemeralResourceModel, apiKey *iam.APIKey) {
	data.CreatedAt = types.StringValue(apiKey.CreatedAt.Format(time.RFC3339))
	data.UpdatedAt = types.StringValue(apiKey.UpdatedAt.Format(time.RFC3339))

	data.AccessKey = types.StringValue(apiKey.AccessKey)
	if apiKey.SecretKey != nil {
		data.SecretKey = types.StringValue(*apiKey.SecretKey)
	} else {
		data.SecretKey = types.StringNull()
	}

	data.ExpiresAt = types.StringValue(apiKey.ExpiresAt.Format(time.RFC3339))
	data.CreationIP = types.StringValue(apiKey.CreationIP)
	data.DefaultProjectID = types.StringValue(apiKey.DefaultProjectID)
}
