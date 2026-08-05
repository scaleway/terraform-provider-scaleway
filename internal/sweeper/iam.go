package sweeper

import (
	"context"
	"errors"
	"fmt"

	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// iamAPIKeySweeper sweeps scaleway_iam_api_key resources.
//
// API keys are global (organization-scoped) and are not taggable, so tag
// filtering is not supported for this type.
type iamAPIKeySweeper struct{}

func (s *iamAPIKeySweeper) TypeName() string        { return "scaleway_iam_api_key" }
func (s *iamAPIKeySweeper) Locality() Locality      { return LocalityGlobal }
func (s *iamAPIKeySweeper) SupportsTagFilter() bool { return false }

func (s *iamAPIKeySweeper) List(ctx context.Context, m *meta.Meta, filter Filter) ([]Resource, error) {
	api := iam.NewAPI(m.ScwClient())

	orgID := filter.OrganizationID
	if orgID == "" {
		defaultOrgID, exists := m.ScwClient().GetDefaultOrganizationID()
		if !exists {
			return nil, errors.New("no organization_id provided and no default organization configured")
		}

		orgID = defaultOrgID
	}

	resp, err := api.ListAPIKeys(&iam.ListAPIKeysRequest{
		OrganizationID: &orgID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list iam api keys: %w", err)
	}

	resources := make([]Resource, 0, len(resp.APIKeys))
	for _, key := range resp.APIKeys {
		if !key.Deletable {
			continue
		}

		resources = append(resources, Resource{
			ID:          key.AccessKey,
			DisplayName: key.Description,
		})
	}

	return resources, nil
}

func (s *iamAPIKeySweeper) Delete(ctx context.Context, m *meta.Meta, resource Resource) error {
	api := iam.NewAPI(m.ScwClient())

	err := api.DeleteAPIKey(&iam.DeleteAPIKeyRequest{
		AccessKey: resource.ID,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return err
	}

	return nil
}

// iamScimTokenSweeper sweeps scaleway_iam_scim_token resources.
//
// SCIM tokens are global (organization-scoped) and are not taggable, so tag
// filtering is not supported for this type. Tokens are listed via the SCIM
// configuration of the organization.
type iamScimTokenSweeper struct{}

func (s *iamScimTokenSweeper) TypeName() string        { return "scaleway_iam_scim_token" }
func (s *iamScimTokenSweeper) Locality() Locality      { return LocalityGlobal }
func (s *iamScimTokenSweeper) SupportsTagFilter() bool { return false }

func (s *iamScimTokenSweeper) List(ctx context.Context, m *meta.Meta, filter Filter) ([]Resource, error) {
	api := iam.NewAPI(m.ScwClient())

	orgID := filter.OrganizationID
	if orgID == "" {
		defaultOrgID, exists := m.ScwClient().GetDefaultOrganizationID()
		if !exists {
			return nil, errors.New("no organization_id provided and no default organization configured")
		}

		orgID = defaultOrgID
	}

	scimConfig, err := api.GetOrganizationScim(&iam.GetOrganizationScimRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get scim configuration: %w", err)
	}

	resp, err := api.ListScimTokens(&iam.ListScimTokensRequest{
		ScimID: scimConfig.ID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list scim tokens: %w", err)
	}

	resources := make([]Resource, 0, len(resp.ScimTokens))
	for _, token := range resp.ScimTokens {
		resources = append(resources, Resource{
			ID:          token.ID,
			DisplayName: token.ID,
		})
	}

	return resources, nil
}

func (s *iamScimTokenSweeper) Delete(ctx context.Context, m *meta.Meta, resource Resource) error {
	api := iam.NewAPI(m.ScwClient())

	err := api.DeleteScimToken(&iam.DeleteScimTokenRequest{
		TokenID: resource.ID,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return err
	}

	return nil
}
