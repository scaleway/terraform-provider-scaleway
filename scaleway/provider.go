package scaleway

import (
	"log"
	"os"
	"sync"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/go-homedir"
	scwLogger "github.com/scaleway/scaleway-sdk-go/logger"
	"github.com/scaleway/scaleway-sdk-go/scwconfig"
	"github.com/scaleway/scaleway-sdk-go/utils"
)

var mu = sync.Mutex{}

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {

	// Init the SDK logger.
	scwLogger.SetLogger(logger{})

	// Init the Scaleway config.
	scwConfig, err := scwconfig.Load()
	if err != nil {
		log.Fatalf("error: cannot load configuration: %s", err)
	}

	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Scaleway access key.",
				DefaultFunc: schema.SchemaDefaultFunc(func() (interface{}, error) {
					// Keep the deprecated behavior
					if accessKey := os.Getenv("SCALEWAY_ACCESS_KEY"); accessKey != "" {
						log.Printf("[WARN] SCALEWAY_ACCESS_KEY is deprecated, please use SCW_ACCESS_KEY instead")
						return accessKey, nil
					}
					if accessKey, exist := scwConfig.GetAccessKey(); exist {
						return accessKey, nil
					}
					return nil, nil
				}),
			},
			"secret_key": {
				Type:        schema.TypeString,
				Optional:    true, // To allow user to use deprecated `token`.
				Description: "The Scaleway secret Key.",
				DefaultFunc: schema.SchemaDefaultFunc(func() (interface{}, error) {
					// No error is returned here to allow user to use deprecated `token`.
					if secretKey, exist := scwConfig.GetSecretKey(); exist {
						return secretKey, nil
					}
					return nil, nil
				}),
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true, // To allow user to use deprecated `organization`.
				Description: "The Scaleway organization ID.",
				DefaultFunc: schema.SchemaDefaultFunc(func() (interface{}, error) {
					if organizationID, exist := scwConfig.GetDefaultOrganizationID(); exist {
						return organizationID, nil
					}
					// No error is returned here to allow user to use deprecated `organization`.
					return nil, nil
				}),
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Scaleway default region to use for your resources.",
				DefaultFunc: schema.SchemaDefaultFunc(func() (interface{}, error) {
					// Keep the deprecated behavior
					// Note: The deprecated region format conversion is handled in `config.GetDeprecatedClient`.
					if region := os.Getenv("SCALEWAY_REGION"); region != "" {
						log.Printf("[WARN] SCALEWAY_REGION is deprecated, please use SCW_DEFAULT_REGION instead")
						return region, nil
					}
					if defaultRegion, exist := scwConfig.GetDefaultRegion(); exist {
						return string(defaultRegion), nil
					}
					return string(utils.RegionFrPar), nil
				}),
			},
			"zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Scaleway default zone to use for your resources.",
				DefaultFunc: schema.SchemaDefaultFunc(func() (interface{}, error) {
					if defaultZone, exist := scwConfig.GetDefaultZone(); exist {
						return string(defaultZone), nil
					}
					return nil, nil
				}),
			},

			// Deprecated values
			"token": {
				Type:       schema.TypeString,
				Optional:   true, // To allow user to use `secret_key`.
				Deprecated: "Use `secret_key` instead.",
				DefaultFunc: schema.SchemaDefaultFunc(func() (interface{}, error) {
					// Keep the deprecated behavior
					for _, k := range []string{"SCALEWAY_TOKEN", "SCALEWAY_ACCESS_KEY"} {
						if os.Getenv(k) != "" {
							log.Printf("[WARN] %s is deprecated, please use SCW_SECRET_KEY instead", k)
							return os.Getenv(k), nil
						}
					}
					if path, err := homedir.Expand("~/.scwrc"); err == nil {
						scwAPIKey, _, err := readDeprecatedScalewayConfig(path)
						if err != nil {
							// No error is returned here to allow user to use `secret_key`.
							log.Printf("[ERROR] cannot parse deprecated config file: %s", err)
							return nil, nil
						}
						// Depreciation log is already handled by scwconfig.
						return scwAPIKey, nil
					}
					// No error is returned here to allow user to use `secret_key`.
					return nil, nil
				}),
			},
			"organization": {
				Type:       schema.TypeString,
				Optional:   true, // To allow user to use `organization_id`.
				Deprecated: "Use `organization_id` instead.",
				DefaultFunc: schema.SchemaDefaultFunc(func() (interface{}, error) {
					// Keep the deprecated behavior
					if organization := os.Getenv("SCALEWAY_ORGANIZATION"); organization != "" {
						log.Printf("[WARN] SCALEWAY_ORGANIZATION is deprecated, please use SCW_DEFAULT_ORGANIZATION_ID instead")
						return organization, nil
					}
					if path, err := homedir.Expand("~/.scwrc"); err == nil {
						_, scwOrganization, err := readDeprecatedScalewayConfig(path)
						if err != nil {
							// No error is returned here to allow user to use `organization_id`.
							log.Printf("[ERROR] cannot parse deprecated config file: %s", err)
							return nil, nil
						}
						return scwOrganization, nil
					}
					// No error is returned here to allow user to use `organization_id`.
					return nil, nil
				}),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"scaleway_bucket":              resourceScalewayBucket(),
			"scaleway_user_data":           resourceScalewayUserData(),
			"scaleway_server":              resourceScalewayServer(),
			"scaleway_token":               resourceScalewayToken(),
			"scaleway_ssh_key":             resourceScalewaySSHKey(),
			"scaleway_ip":                  resourceScalewayIP(),
			"scaleway_ip_reverse_dns":      resourceScalewayIPReverseDNS(),
			"scaleway_security_group":      resourceScalewaySecurityGroup(),
			"scaleway_security_group_rule": resourceScalewaySecurityGroupRule(),
			"scaleway_volume":              resourceScalewayVolume(),
			"scaleway_volume_attachment":   resourceScalewayVolumeAttachment(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"scaleway_bootscript":     dataSourceScalewayBootscript(),
			"scaleway_image":          dataSourceScalewayImage(),
			"scaleway_security_group": dataSourceScalewaySecurityGroup(),
			"scaleway_volume":         dataSourceScalewayVolume(),
		},

		ConfigureFunc: providerConfigure,
	}
}

// providerConfigure creates the Meta object containing the SDK client.
func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := &Config{
		AccessKey:        d.Get("access_key").(string),
		SecretKey:        d.Get("secret_key").(string),
		DefaultProjectID: d.Get("organization_id").(string),
		DefaultRegion:    utils.Region(d.Get("region").(string)),
		DefaultZone:      utils.Zone(d.Get("zone").(string)),
	}

	// Handle deprecated values.
	// Deprecation messages will be handled by terraform.
	if config.SecretKey == "" {
		config.SecretKey = d.Get("token").(string)
	}
	if config.DefaultProjectID == "" {
		config.DefaultProjectID = d.Get("organization").(string)
	}
	if config.SecretKey == "" && config.AccessKey != "" {
		log.Println("[WARN] you seem to use the access_key instead of secret_key in the provider. This bogus behavior is deprecated, please use the `secret_key` field instead.")
		config.SecretKey = config.AccessKey
	}

	return config.Meta()
}
