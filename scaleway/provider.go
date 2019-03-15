package scaleway

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	homedir "github.com/mitchellh/go-homedir"
)

var mu = sync.Mutex{}

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALEWAY_ACCESS_KEY", nil),
				Deprecated:  "Use `token` instead.",
				Description: "The API key for Scaleway API operations.",
			},
			"token": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.SchemaDefaultFunc(func() (interface{}, error) {
					for _, k := range []string{"SCALEWAY_TOKEN", "SCALEWAY_ACCESS_KEY"} {
						if os.Getenv(k) != "" {
							return os.Getenv(k), nil
						}
					}
					if path, err := homedir.Expand("~/.scwrc"); err == nil {
						scwAPIKey, _, err := readScalewayConfig(path)
						if err != nil {
							return nil, err
						}
						return scwAPIKey, nil
					}
					return nil, errors.New("No token found")
				}),
				Description: "The API key for Scaleway API operations.",
			},
			"organization": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.SchemaDefaultFunc(func() (interface{}, error) {
					for _, k := range []string{"SCALEWAY_ORGANIZATION"} {
						if os.Getenv(k) != "" {
							return k, nil
						}
					}
					if path, err := homedir.Expand("~/.scwrc"); err == nil {
						_, scwOrganization, err := readScalewayConfig(path)
						if err != nil {
							return nil, err
						}
						return scwOrganization, nil
					}
					return nil, errors.New("No token found")
				}),
				Description: "The Organization ID (a.k.a. 'access key') for Scaleway API operations.",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALEWAY_REGION", "par1"),
				Description: "The Scaleway API region to use.",
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

type scalewayConfig struct {
	Organization string `json:"organization"`
	Token        string `json:"token"`
	Version      string `json:"version"`
}

func readScalewayConfig(path string) (string, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	var data scalewayConfig
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return "", "", err
	}
	return data.Token, data.Organization, nil
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	apiKey := ""
	if v, ok := d.Get("token").(string); ok {
		apiKey = v
	} else {
		if v, ok := d.Get("access_key").(string); ok {
			apiKey = v
		}
	}

	organization := d.Get("organization").(string)

	if apiKey == "" {
		if path, err := homedir.Expand("~/.scwrc"); err == nil {
			scwAPIKey, scwOrganization, err := readScalewayConfig(path)
			if err != nil {
				return nil, fmt.Errorf("Error loading credentials from SCW: %s", err)
			}
			apiKey = scwAPIKey
			organization = scwOrganization
		}
	}

	config := Config{
		Organization: organization,
		APIKey:       apiKey,
		Region:       d.Get("region").(string),
	}

	return config.Client()
}
