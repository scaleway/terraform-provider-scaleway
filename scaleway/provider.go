package scaleway

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/go-homedir"
	scwLogger "github.com/scaleway/scaleway-sdk-go/logger"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var mu = sync.Mutex{}

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {

	// Init the SDK logger.
	scwLogger.SetLogger(l)

	// Try to migrate config
	_, err := scw.MigrateLegacyConfig()
	if err != nil {
		l.Errorf("cannot migrate configuration: %s", err)
		return nil
	}

	// Load active profile
	var activeProfile *scw.Profile
	scwConfig, err := scw.LoadConfig()
	if err != nil {
		l.Warningf("cannot load configuration: %s", err)
	} else {
		activeProfile, err = scwConfig.GetActiveProfile()
		if err != nil {
			l.Errorf("cannot load configuration: %s", err)
		}
	}

	// load env
	envProfile := scw.LoadEnvProfile()

	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Scaleway access key.",
				DefaultFunc: func() (interface{}, error) {
					// Keep the deprecated behavior
					if accessKey := os.Getenv("SCALEWAY_ACCESS_KEY"); accessKey != "" {
						l.Warningf("SCALEWAY_ACCESS_KEY is deprecated, please use SCW_ACCESS_KEY instead")
						return accessKey, nil
					}
					if envProfile.AccessKey != nil {
						return *envProfile.AccessKey, nil
					}
					if activeProfile != nil && activeProfile.AccessKey != nil {
						return *activeProfile.AccessKey, nil
					}
					return nil, nil
				},
			},
			"secret_key": {
				Type:        schema.TypeString,
				Optional:    true, // To allow user to use deprecated `token`.
				Description: "The Scaleway secret Key.",
				DefaultFunc: func() (interface{}, error) {
					// No error is returned here to allow user to use deprecated `token`.
					if envProfile.SecretKey != nil {
						return *envProfile.SecretKey, nil
					}
					if activeProfile != nil && activeProfile.SecretKey != nil {
						return *activeProfile.SecretKey, nil
					}

					// Keep the deprecated behavior from 'token'.
					for _, k := range []string{"SCALEWAY_TOKEN", "SCALEWAY_ACCESS_KEY"} {
						if os.Getenv(k) != "" {
							l.Warningf("%s is deprecated, please use SCW_SECRET_KEY instead", k)
							return os.Getenv(k), nil
						}
					}
					if path, err := homedir.Expand("~/.scwrc"); err == nil {
						scwAPIKey, _, err := readDeprecatedScalewayConfig(path)
						if err != nil {
							// No error is returned here to allow user to use `secret_key`.
							l.Errorf("cannot parse deprecated config file: %s", err)
							return nil, nil
						}
						// Depreciation log is already handled by scw config.
						return scwAPIKey, nil
					}
					// No error is returned here to allow user to use `secret_key`.
					return nil, nil
				},
				ValidateFunc: validationUUID(),
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true, // To allow user to use deprecated `organization`.
				Description: "The Scaleway project ID.",
				DefaultFunc: schema.SchemaDefaultFunc(func() (interface{}, error) {
					if envProfile.DefaultProjectID != nil {
						return *envProfile.DefaultProjectID, nil
					}
					if activeProfile != nil && activeProfile.DefaultProjectID != nil {
						return *activeProfile.DefaultProjectID, nil
					}

					// Keep the deprecated behavior of 'organization'.
					if organization := os.Getenv("SCALEWAY_ORGANIZATION"); organization != "" {
						l.Warningf("SCALEWAY_ORGANIZATION is deprecated, please use SCW_DEFAULT_PROJECT_ID instead")
						return organization, nil
					}
					if path, err := homedir.Expand("~/.scwrc"); err == nil {
						_, scwOrganization, err := readDeprecatedScalewayConfig(path)
						if err != nil {
							// No error is returned here to allow user to use `project_id`.
							l.Errorf("cannot parse deprecated config file: %s", err)
							return nil, nil
						}
						return scwOrganization, nil
					}
					// No error is returned here to allow user to use `project_id`.
					return nil, nil
				}),
				ValidateFunc: validationUUID(),
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Scaleway default region to use for your resources.",
				DefaultFunc: func() (interface{}, error) {
					// Keep the deprecated behavior
					// Note: The deprecated region format conversion is handled in `config.GetDeprecatedClient`.
					if region := os.Getenv("SCALEWAY_REGION"); region != "" {
						l.Warningf("SCALEWAY_REGION is deprecated, please use SCW_DEFAULT_REGION instead")
						return region, nil
					}
					if envProfile.DefaultRegion != nil {
						return *envProfile.DefaultRegion, nil
					}
					if activeProfile != nil && activeProfile.DefaultRegion != nil {
						return *activeProfile.DefaultRegion, nil
					}
					return string(scw.RegionFrPar), nil
				},
				ValidateFunc: validationRegion(),
			},
			"zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Scaleway default zone to use for your resources.",
				DefaultFunc: func() (interface{}, error) {
					if envProfile.DefaultZone != nil {
						return *envProfile.DefaultZone, nil
					}
					if activeProfile != nil && activeProfile.DefaultZone != nil {
						return *activeProfile.DefaultZone, nil
					}
					return nil, nil
				},
				ValidateFunc: validationZone(),
			},

			// Deprecated values
			"token": {
				Type:       schema.TypeString,
				Optional:   true, // To allow user to use `secret_key`.
				Deprecated: "Use `secret_key` instead.",
			},
			"organization": {
				Type:       schema.TypeString,
				Optional:   true, // To allow user to use `project_id`.
				Deprecated: "Use `project_id` instead.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"scaleway_bucket":                           resourceScalewayBucket(),
			"scaleway_instance_ip":                      resourceScalewayInstanceIP(),
			"scaleway_compute_instance_volume":          resourceScalewayComputeInstanceVolume(),
			"scaleway_compute_instance_security_group":  resourceScalewayComputeInstanceSecurityGroup(),
			"scaleway_compute_instance_server":          resourceScalewayComputeInstanceServer(),
			"scaleway_compute_instance_placement_group": resourceScalewayComputeInstancePlacementGroup(),
			"scaleway_storage_object_bucket":            resourceScalewayStorageObjectBucket(),
			"scaleway_user_data":                        resourceScalewayUserData(),
			"scaleway_server":                           resourceScalewayServer(),
			"scaleway_token":                            resourceScalewayToken(),
			"scaleway_ssh_key":                          resourceScalewaySSHKey(),
			"scaleway_ip":                               resourceScalewayIP(),
			"scaleway_ip_reverse_dns":                   resourceScalewayIPReverseDNS(),
			"scaleway_security_group":                   resourceScalewaySecurityGroup(),
			"scaleway_security_group_rule":              resourceScalewaySecurityGroupRule(),
			"scaleway_volume":                           resourceScalewayVolume(),
			"scaleway_volume_attachment":                resourceScalewayVolumeAttachment(),
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
	apiKey := ""
	if v, ok := d.Get("secret_key").(string); ok {
		apiKey = v
	} else if v, ok := d.Get("token").(string); ok {
		apiKey = v
	} else {
		if v, ok := d.Get("access_key").(string); ok {
			l.Warningf("you seem to use the access_key instead of secret_key in the provider. This bogus behavior is deprecated, please use the `secret_key` field instead.")
			apiKey = v
		}
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		projectID = d.Get("organization").(string)
	}

	if apiKey == "" {
		if path, err := homedir.Expand("~/.scwrc"); err == nil {
			scwAPIKey, scwOrganization, err := readDeprecatedScalewayConfig(path)
			if err != nil {
				return nil, fmt.Errorf("error loading credentials from SCW: %s", err)
			}
			apiKey = scwAPIKey
			projectID = scwOrganization
		}
	}

	rawRegion := d.Get("region").(string)
	region, err := scw.ParseRegion(rawRegion)
	if err != nil {
		return nil, err
	}

	rawZone := d.Get("zone").(string)
	zone, err := scw.ParseZone(rawZone)
	if err != nil {
		return nil, err
	}

	meta := &Meta{
		AccessKey:        d.Get("access_key").(string),
		SecretKey:        apiKey,
		DefaultProjectID: projectID,
		DefaultRegion:    region,
		DefaultZone:      zone,
	}

	meta.bootstrap()
	if err != nil {
		return nil, err
	}

	// fetch known scaleway server types to support validation in r/server
	if len(commercialServerTypes) == 0 {
		if availability, err := meta.deprecatedClient.GetServerAvailabilities(); err == nil {
			commercialServerTypes = availability.CommercialTypes()
			sort.StringSlice(commercialServerTypes).Sort()
		}
		if os.Getenv("DISABLE_SCALEWAY_SERVER_TYPE_VALIDATION") != "" {
			commercialServerTypes = commercialServerTypes[:0]
		}
	}

	return meta, nil
}
