package iam

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type IAMResourceHandler struct {
	iamAPI *iam.API
}

func NewIAMResourceHandler(iamAPI *iam.API) *IAMResourceHandler {
	return &IAMResourceHandler{
		iamAPI: iamAPI,
	}
}

func (f *IAMResourceHandler) FindByAnnotationValue(ctx context.Context, srn string) (*iam.APIKey, error) {
	accessKey := srn
	if idx := strings.LastIndex(srn, "/"); idx != -1 {
		accessKey = srn[idx+1:]
	}

	if accessKey == "" {
		return nil, fmt.Errorf("invalid SRN format: %s", srn)
	}

	apiKey, err := f.iamAPI.GetAPIKey(&iam.GetAPIKeyRequest{
		AccessKey: accessKey,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return apiKey, nil
}

func (f *IAMResourceHandler) FindByDescription(ctx context.Context, description string, organizationID string) (*iam.APIKey, error) {
	apiKeys, err := f.iamAPI.ListAPIKeys(&iam.ListAPIKeysRequest{
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

func (f *IAMResourceHandler) GetByID(ctx context.Context, id string) (*iam.APIKey, error) {
	apiKey, err := f.iamAPI.GetAPIKey(&iam.GetAPIKeyRequest{
		AccessKey: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (f *IAMResourceHandler) DeleteByID(ctx context.Context, id string) error {
	return f.iamAPI.DeleteAPIKey(&iam.DeleteAPIKeyRequest{
		AccessKey: id,
	}, scw.WithContext(ctx))
}

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
