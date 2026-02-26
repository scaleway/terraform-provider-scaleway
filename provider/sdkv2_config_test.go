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

		// Create a minimal provider schema with no provider config overrides
		providerSchema := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
			"access_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"secret_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"profile": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"api_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
		}, map[string]interface{}{})

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

		// Create provider schema with provider config that should override config file
		providerSchema := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
			"access_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"secret_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"profile": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"api_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
		}, map[string]interface{}{
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
		providerSchema := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
			"access_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"secret_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"profile": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"api_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
		}, map[string]interface{}{
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
		_ = os.Unsetenv(scw.ScwAccessKeyEnv)
		_ = os.Unsetenv(scw.ScwSecretKeyEnv)
		_ = os.Unsetenv(scw.ScwDefaultProjectIDEnv)
		_ = os.Unsetenv(scw.ScwDefaultRegionEnv)
		_ = os.Unsetenv(scw.ScwDefaultZoneEnv)
		_ = os.Unsetenv("SCW_CONFIG_PATH")

		// Test with no config - should get defaults
		providerSchema := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
			"access_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"secret_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"profile": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"api_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
		}, map[string]interface{}{})

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
