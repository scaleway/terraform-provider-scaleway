package provider_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// TestSDKv2ProviderConfigSources tests that the SDKv2 provider
// can properly load credentials from different sources in the correct priority order:
// config file < provider config < environment variables

func TestSDKv2ProviderConfigSources_ActiveProfile(t *testing.T) {
	t.Run("Test config file loading", func(t *testing.T) {
		unsetEnv(false)

		tempDir := t.TempDir()
		configFile := tempDir + "/config.yaml"

		if err := os.WriteFile(configFile, []byte(configContent), 0o644); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		t.Setenv("SCW_CONFIG_PATH", configFile)

		// Create a minimal provider schema with no provider config overrides
		providerSchema := generateProviderSchema(t, map[string]any{})

		// Test with an empty provider config
		profile, credentialsSource, err := meta.LoadProfile(
			context.Background(),
			providerSchema,
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

func TestSDKv2ProviderConfigSources_ProviderConfig(t *testing.T) {
	t.Run("Test provider config overrides config file", func(t *testing.T) {
		unsetEnv(false)

		tempDir := t.TempDir()
		configFile := tempDir + "/config.yaml"

		if err := os.WriteFile(configFile, []byte(configContent), 0o644); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		t.Setenv("SCW_CONFIG_PATH", configFile)

		// Create provider schema with provider config that should override config file
		providerSchema := generateProviderSchema(t, map[string]any{
			"access_key": "override-access-key",
			"secret_key": "override-secret-key",
			"project_id": "override-project-id",
			"region":     "fr-par",
			"zone":       "fr-par-1",
		})

		profile, credentialsSource, err := meta.LoadProfile(
			context.Background(),
			providerSchema,
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

func TestSDKv2ProviderConfigSources_EnvConfig(t *testing.T) {
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

		// Create provider schema with provider config and config file, but env vars should take precedence
		providerSchema := generateProviderSchema(t, map[string]any{
			"access_key": "config-access-key",
			"secret_key": "config-secret-key",
			"project_id": "config-project-id",
			"region":     "fr-par",
			"zone":       "fr-par-1",
		})

		profile, credentialsSource, err := meta.LoadProfile(
			context.Background(),
			providerSchema,
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

func TestSDKv2ProviderConfigSources_NoConfig(t *testing.T) {
	t.Run("Test defaults when no config provided", func(t *testing.T) {
		unsetEnv(true)

		// Test with no config - should get defaults
		providerSchema := generateProviderSchema(t, map[string]any{})

		profile, _, err := meta.LoadProfile(
			context.Background(),
			providerSchema,
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

func TestSDKv2ProviderMetaInitialization(t *testing.T) {
	t.Run("Test that meta is properly initialized", func(t *testing.T) {
		unsetEnv(true)

		sdkv2Config := &meta.Config{}

		m, err := meta.NewMeta(t.Context(), sdkv2Config)
		if err != nil {
			t.Fatalf("NewMeta failed: %v", err)
		}

		if m == nil {
			t.Fatal("meta is nil - NewMeta returned nil")
		}

		if m.ScwClient() == nil {
			t.Fatal("meta.ScwClient() is nil")
		}

		if m.HTTPClient() == nil {
			t.Fatal("meta.HTTPClient() is nil")
		}

		if m.Endpoints() != nil {
			t.Fatal("meta.Endpoints() should be nil")
		}

		if m.S3UsePathStyle() {
			t.Fatal("meta.S3UsePathStyle() should be false")
		}
	})
	t.Run("Test that meta is properly initialized with filled config", func(t *testing.T) {
		unsetEnv(true)

		s3Endpoint := "https://my-s3-endpoint.com"

		sdkv2Config := &meta.Config{
			Endpoints: map[string]string{
				"s3": s3Endpoint,
			},
			S3UsePathStyle: true,
		}

		m, err := meta.NewMeta(t.Context(), sdkv2Config)
		if err != nil {
			t.Fatalf("NewMeta failed: %v", err)
		}

		if m == nil {
			t.Fatal("meta is nil - NewMeta returned nil")
		}

		if m.ScwClient() == nil {
			t.Fatal("meta.ScwClient() is nil")
		}

		if m.HTTPClient() == nil {
			t.Fatal("meta.HTTPClient() is nil")
		}

		if m.Endpoints() == nil {
			t.Fatal("meta.Endpoints() is nil")
		}

		if m.Endpoints()["s3"] != s3Endpoint {
			t.Fatalf("meta.Endpoints()[\"s3\"] is '%s', expected '%s'", m.Endpoints()["s3"], s3Endpoint)
		}

		if !m.S3UsePathStyle() {
			t.Fatal("meta.S3UsePathStyle() should be true")
		}
	})
}

func generateProviderSchema(t *testing.T, m map[string]any) *schema.ResourceData {
	t.Helper()

	return schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"access_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The Scaleway access key for testing",
		},
		"secret_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The Scaleway secret key for testing",
		},
		"profile": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The Scaleway profile for testing",
		},
		"project_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The Scaleway project ID for testing",
		},
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The Scaleway region for testing",
		},
		"zone": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The Scaleway zone for testing",
		},
		"api_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The Scaleway API URL for testing",
		},
		"endpoints": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"s3": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The Scaleway S3 endpoint for testing",
					},
				},
			},
		},
		"s3_use_path_style": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "The Scaleway S3 path style for testing",
		},
	}, m)
}
