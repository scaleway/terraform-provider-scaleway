package provider_test

import (
	"context"
	"os"
	"testing"

	providerFramework "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/provider"
)

const configContent = `
profiles:
  test-profile:
    access_key: profile-access-key
    secret_key: profile-secret-key
    default_project_id: profile-project-id
    default_region: fr-par
    default_zone: fr-par-1
active_profile: test-profile
`

// TestFrameworkProviderConfigSources tests test that the framework provider
// can properly load credentials from different sources in the correct priority order:
// config file < provider config < environment variables

func TestFrameworkProviderConfigSources_ActiveProfile(t *testing.T) {
	t.Run("Test config file loading", func(t *testing.T) {
		_ = os.Unsetenv(scw.ScwAccessKeyEnv)
		_ = os.Unsetenv(scw.ScwSecretKeyEnv)
		_ = os.Unsetenv(scw.ScwDefaultProjectIDEnv)
		_ = os.Unsetenv(scw.ScwDefaultRegionEnv)
		_ = os.Unsetenv(scw.ScwDefaultZoneEnv)

		tempDir := t.TempDir()
		configFile := tempDir + "/config.yaml"

		if err := os.WriteFile(configFile, []byte(configContent), 0o644); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		t.Setenv("SCW_CONFIG_PATH", configFile)

		// Test with an empty provider config
		profile, credentialsSource, err := meta.LoadProfileFromFrameworkConfig(
			context.Background(),
			&meta.FrameworkProviderConfig{},
		)
		if err != nil {
			t.Fatalf("Failed to load profile: %v", err)
		}

		if credentialsSource.AccessKey != meta.CredentialsSourceActiveProfile {
			t.Errorf("Expected access key source to be active profile, got: %s", credentialsSource.AccessKey)
		}

		if credentialsSource.SecretKey != meta.CredentialsSourceActiveProfile {
			t.Errorf("Expected secret key source to be active profile, got: %s", credentialsSource.SecretKey)
		}

		if credentialsSource.ProjectID != meta.CredentialsSourceActiveProfile {
			t.Errorf("Expected project ID source to be active profile, got: %s", credentialsSource.ProjectID)
		}

		if profile.AccessKey == nil || *profile.AccessKey != "profile-access-key" {
			t.Errorf("Expected access key to be 'profile-access-key', got: %v", profile.AccessKey)
		}

		if profile.SecretKey == nil || *profile.SecretKey != "profile-secret-key" {
			t.Errorf("Expected secret key to be 'profile-secret-key', got: %v", profile.SecretKey)
		}

		if profile.DefaultProjectID == nil || *profile.DefaultProjectID != "profile-project-id" {
			t.Errorf("Expected project ID to be 'profile-project-id', got: %v", profile.DefaultProjectID)
		}
	})
}

func TestFrameworkProviderConfigSources_ProviderConfig(t *testing.T) {
	t.Run("Test provider config overrides config file", func(t *testing.T) {
		_ = os.Unsetenv(scw.ScwAccessKeyEnv)
		_ = os.Unsetenv(scw.ScwSecretKeyEnv)
		_ = os.Unsetenv(scw.ScwDefaultProjectIDEnv)
		_ = os.Unsetenv(scw.ScwDefaultRegionEnv)
		_ = os.Unsetenv(scw.ScwDefaultZoneEnv)

		tempDir := t.TempDir()
		configFile := tempDir + "/config.yaml"

		if err := os.WriteFile(configFile, []byte(configContent), 0o644); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		t.Setenv("SCW_CONFIG_PATH", configFile)

		// Test with provider config that should override config file
		profile, credentialsSource, err := meta.LoadProfileFromFrameworkConfig(
			context.Background(),
			&meta.FrameworkProviderConfig{
				AccessKey: "override-access-key",
				SecretKey: "override-secret-key",
				ProjectID: "override-project-id",
				Region:    "fr-par",
				Zone:      "fr-par-1",
			},
		)
		if err != nil {
			t.Fatalf("Failed to load profile: %v", err)
		}

		if credentialsSource.AccessKey != meta.CredentialsSourceProviderProfile {
			t.Errorf("Expected access key source to be provider profile, got: %s", credentialsSource.AccessKey)
		}

		if credentialsSource.SecretKey != meta.CredentialsSourceProviderProfile {
			t.Errorf("Expected secret key source to be provider profile, got: %s", credentialsSource.SecretKey)
		}

		if credentialsSource.ProjectID != meta.CredentialsSourceProviderProfile {
			t.Errorf("Expected project ID source to be provider profile, got: %s", credentialsSource.ProjectID)
		}

		if profile.AccessKey == nil || *profile.AccessKey != "override-access-key" {
			t.Errorf("Expected access key to be 'override-access-key', got: %v", profile.AccessKey)
		}

		if profile.SecretKey == nil || *profile.SecretKey != "override-secret-key" {
			t.Errorf("Expected secret key to be 'override-secret-key', got: %v", profile.SecretKey)
		}

		if profile.DefaultProjectID == nil || *profile.DefaultProjectID != "override-project-id" {
			t.Errorf("Expected project ID to be 'override-project-id', got: %v", profile.DefaultProjectID)
		}
	})
}

func TestFrameworkProviderConfigSources_EnvConfig(t *testing.T) {
	t.Run("Test environment variables override everything", func(t *testing.T) {
		t.Setenv(scw.ScwAccessKeyEnv, "env-access-key")
		t.Setenv(scw.ScwSecretKeyEnv, "env-secret-key")
		t.Setenv(scw.ScwDefaultProjectIDEnv, "env-project-id")
		t.Setenv(scw.ScwDefaultRegionEnv, scw.RegionFrPar.String())
		t.Setenv(scw.ScwDefaultZoneEnv, scw.ZoneFrPar1.String())

		tempDir := t.TempDir()
		configFile := tempDir + "/config.yaml"

		if err := os.WriteFile(configFile, []byte(configContent), 0o644); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		t.Setenv("SCW_CONFIG_PATH", configFile)

		// Test with provider config and config file, but env vars should take precedence
		profile, credentialsSource, err := meta.LoadProfileFromFrameworkConfig(
			context.Background(),
			&meta.FrameworkProviderConfig{
				AccessKey: "config-access-key",
				SecretKey: "config-secret-key",
				ProjectID: "config-project-id",
				Region:    "fr-par",
				Zone:      "fr-par-1",
			},
		)
		if err != nil {
			t.Fatalf("Failed to load profile: %v", err)
		}

		if credentialsSource.AccessKey != meta.CredentialsSourceEnvironment {
			t.Errorf("Expected access key source to be environment, got: %s", credentialsSource.AccessKey)
		}

		if credentialsSource.SecretKey != meta.CredentialsSourceEnvironment {
			t.Errorf("Expected secret key source to be environment, got: %s", credentialsSource.SecretKey)
		}

		if credentialsSource.ProjectID != meta.CredentialsSourceEnvironment {
			t.Errorf("Expected project ID source to be environment, got: %s", credentialsSource.ProjectID)
		}

		if profile.AccessKey == nil || *profile.AccessKey != "env-access-key" {
			t.Errorf("Expected access key to be 'env-access-key', got: %v", profile.AccessKey)
		}

		if profile.SecretKey == nil || *profile.SecretKey != "env-secret-key" {
			t.Errorf("Expected secret key to be 'env-secret-key', got: %v", profile.SecretKey)
		}

		if profile.DefaultProjectID == nil || *profile.DefaultProjectID != "env-project-id" {
			t.Errorf("Expected project ID to be 'env-project-id', got: %v", profile.DefaultProjectID)
		}
	})
}

func TestFrameworkProviderConfigSources_NoConfig(t *testing.T) {
	t.Run("Test defaults when no config provided", func(t *testing.T) {
		_ = os.Unsetenv(scw.ScwAccessKeyEnv)
		_ = os.Unsetenv(scw.ScwSecretKeyEnv)
		_ = os.Unsetenv(scw.ScwDefaultProjectIDEnv)
		_ = os.Unsetenv(scw.ScwDefaultRegionEnv)
		_ = os.Unsetenv(scw.ScwDefaultZoneEnv)
		_ = os.Unsetenv("SCW_CONFIG_PATH")

		// Test with no config - should get defaults
		profile, _, err := meta.LoadProfileFromFrameworkConfig(
			context.Background(),
			&meta.FrameworkProviderConfig{},
		)
		if err != nil {
			t.Fatalf("Failed to load profile: %v", err)
		}

		if profile.DefaultRegion == nil || *profile.DefaultRegion != scw.RegionFrPar.String() {
			t.Errorf("Expected default region to be 'fr-par', got: %v", profile.DefaultRegion)
		}

		if profile.DefaultZone == nil || *profile.DefaultZone != scw.ZoneFrPar1.String() {
			t.Errorf("Expected default zone to be 'fr-par-1', got: %v", profile.DefaultZone)
		}

		if profile.AccessKey != nil {
			t.Errorf("Expected access key to be nil, got: %v", profile.AccessKey)
		}

		if profile.SecretKey != nil {
			t.Errorf("Expected secret key to be nil, got: %v", profile.SecretKey)
		}

		if profile.DefaultProjectID != nil {
			t.Errorf("Expected project ID to be nil, got: %v", profile.DefaultProjectID)
		}
	})
}

func TestFrameworkProviderMetaInitialization(t *testing.T) {
	t.Run("Test that meta is properly initialized", func(t *testing.T) {
		_ = os.Unsetenv(scw.ScwAccessKeyEnv)
		_ = os.Unsetenv(scw.ScwSecretKeyEnv)
		_ = os.Unsetenv(scw.ScwDefaultProjectIDEnv)
		_ = os.Unsetenv(scw.ScwDefaultRegionEnv)
		_ = os.Unsetenv(scw.ScwDefaultZoneEnv)
		_ = os.Unsetenv("SCW_CONFIG_PATH")

		frameworkConfig := &meta.FrameworkProviderConfig{}

		m, err := meta.NewMetaFromFrameworkConfig(context.Background(), frameworkConfig, "1.0.0")
		if err != nil {
			t.Fatalf("NewMetaFromFrameworkConfig failed: %v", err)
		}

		if m == nil {
			t.Fatal("meta is nil - NewMetaFromFrameworkConfig returned nil")
		}

		if m.ScwClient() == nil {
			t.Fatal("meta.ScwClient() is nil")
		}

		if m.HTTPClient() == nil {
			t.Fatal("meta.HTTPClient() is nil")
		}
	})
}

func TestFrameworkProviderConfigure(t *testing.T) {
	t.Run("Test Configure properly assigns meta", func(t *testing.T) {
		_ = os.Unsetenv(scw.ScwAccessKeyEnv)
		_ = os.Unsetenv(scw.ScwSecretKeyEnv)
		_ = os.Unsetenv(scw.ScwDefaultProjectIDEnv)
		_ = os.Unsetenv(scw.ScwDefaultRegionEnv)
		_ = os.Unsetenv(scw.ScwDefaultZoneEnv)
		_ = os.Unsetenv("SCW_CONFIG_PATH")

		p := provider.NewFrameworkProvider(nil)()

		configValue := tfsdk.Config{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"access_key": schema.StringAttribute{
						Optional: true,
					},
					"secret_key": schema.StringAttribute{
						Optional: true,
					},
					"profile": schema.StringAttribute{
						Optional: true,
					},
					"project_id": schema.StringAttribute{
						Optional: true,
					},
					"organization_id": schema.StringAttribute{
						Optional: true,
					},
					"api_url": schema.StringAttribute{
						Optional: true,
					},
					"region": schema.StringAttribute{
						Optional: true,
					},
					"zone": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		}

		req := providerFramework.ConfigureRequest{
			Config: configValue,
		}

		var resp providerFramework.ConfigureResponse
		p.Configure(t.Context(), req, &resp)

		if resp.Diagnostics.HasError() {
			for _, diag := range resp.Diagnostics {
				t.Logf("Diagnostic: %s: %s", diag.Severity(), diag.Detail())
			}

			t.Fatalf("Configure failed")
		}

		if resp.ResourceData == nil {
			t.Fatal("resp.ResourceData is nil - meta was not properly assigned.")
		}

		if resp.DataSourceData == nil {
			t.Fatal("resp.DataSourceData is nil - meta was not properly assigned")
		}

		metaObj, ok := resp.ResourceData.(*meta.Meta)
		if !ok {
			t.Fatalf("ResourceData is not of type *meta.Meta, got: %T", resp.ResourceData)
		}

		if metaObj.ScwClient() == nil {
			t.Fatal("meta.ScwClient() is nil")
		}

		if metaObj.HTTPClient() == nil {
			t.Fatal("meta.HTTPClient() is nil")
		}
	})
}
