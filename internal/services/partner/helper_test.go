package partner_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/provider"
	"github.com/stretchr/testify/require"
)

const (
	partnerOrganizationIDEnv = "SCW_PARTNER_ORGANIZATION_ID"
	partnerAccessKeyEnv      = "SCW_PARTNER_ACCESS_KEY"
	partnerSecretKeyEnv      = "SCW_PARTNER_SECRET_KEY"
)

func newPartnerTestTools(t *testing.T) (*acctest.TestTools, string) {
	t.Helper()

	orgID := os.Getenv(partnerOrganizationIDEnv)
	accessKey := os.Getenv(partnerAccessKeyEnv)
	secretKey := os.Getenv(partnerSecretKeyEnv)

	if orgID == "" || accessKey == "" || secretKey == "" {
		t.Skip("partner organization acceptance tests require SCW_PARTNER_ORGANIZATION_ID, SCW_PARTNER_ACCESS_KEY and SCW_PARTNER_SECRET_KEY to be set")
	}

	tt := acctest.NewTestTools(t)

	ctx := t.Context()

	m, err := meta.NewMeta(ctx, &meta.Config{
		TerraformVersion:    "terraform-tests",
		HTTPClient:          tt.Meta.HTTPClient(),
		ForceOrganizationID: orgID,
		ForceAccessKey:      accessKey,
		ForceSecretKey:      secretKey,
	})
	require.NoError(t, err)

	tt.Meta = m
	tt.ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"scaleway": func() (tfprotov6.ProviderServer, error) {
			providers, err := provider.NewProviderList(ctx, &provider.Config{Meta: m})
			if err != nil {
				return nil, err
			}

			muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
			if err != nil {
				return nil, err
			}

			return muxServer.ProviderServer(), nil
		},
	}

	return tt, orgID
}
