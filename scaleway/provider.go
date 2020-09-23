package scaleway

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	sdk "github.com/nicolai86/scaleway-sdk"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/validation"
)

var mu = sync.Mutex{}

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Scaleway access key.",
			},
			"secret_key": {
				Type:         schema.TypeString,
				Optional:     true, // To allow user to use deprecated `token`.
				Description:  "The Scaleway secret Key.",
				ValidateFunc: validationUUID(),
			},
			"organization_id": {
				Type:         schema.TypeString,
				Optional:     true, // To allow user to use deprecated `organization`.
				Description:  "The Scaleway organization ID.",
				ValidateFunc: validationUUID(),
			},
			"project_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The Scaleway project ID.",
				ValidateFunc: validationUUID(),
			},
			"region": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The Scaleway default region to use for your resources.",
				ValidateFunc: validationRegion(),
			},
			"zone": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The Scaleway default zone to use for your resources.",
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
				Optional:   true, // To allow user to use `organization_id`.
				Deprecated: "Use `organization_id` instead.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"scaleway_account_ssh_key":               resourceScalewayAccountSSKKey(),
			"scaleway_baremetal_server":              resourceScalewayBaremetalServer(),
			"scaleway_instance_ip":                   resourceScalewayInstanceIP(),
			"scaleway_instance_ip_reverse_dns":       resourceScalewayInstanceIPReverseDns(),
			"scaleway_instance_volume":               resourceScalewayInstanceVolume(),
			"scaleway_instance_security_group":       resourceScalewayInstanceSecurityGroup(),
			"scaleway_instance_security_group_rules": resourceScalewayInstanceSecurityGroupRules(),
			"scaleway_instance_server":               resourceScalewayInstanceServer(),
			"scaleway_instance_placement_group":      resourceScalewayInstancePlacementGroup(),
			"scaleway_k8s_cluster_beta":              resourceScalewayK8SClusterBeta(),
			"scaleway_k8s_pool_beta":                 resourceScalewayK8SPoolBeta(),
			"scaleway_lb_beta":                       resourceScalewayLbBeta(),
			"scaleway_lb_ip_beta":                    resourceScalewayLbIPBeta(),
			"scaleway_lb_backend_beta":               resourceScalewayLbBackendBeta(),
			"scaleway_lb_certificate_beta":           resourceScalewayLbCertificateBeta(),
			"scaleway_lb_frontend_beta":              resourceScalewayLbFrontendBeta(),
			"scaleway_registry_namespace_beta":       resourceScalewayRegistryNamespaceBeta(),
			"scaleway_rdb_instance_beta":             resourceScalewayRdbInstanceBeta(),
			"scaleway_rdb_user_beta":                 resourceScalewayRdbUserBeta(),
			"scaleway_object_bucket":                 resourceScalewayObjectBucket(),
			"scaleway_user_data":                     resourceScalewayUserData(),
			"scaleway_server":                        resourceScalewayServer(),
			"scaleway_token":                         resourceScalewayToken(),
			"scaleway_ssh_key":                       resourceScalewaySSHKey(),
			"scaleway_ip":                            resourceScalewayIP(),
			"scaleway_ip_reverse_dns":                resourceScalewayIPReverseDNS(),
			"scaleway_security_group":                resourceScalewaySecurityGroup(),
			"scaleway_security_group_rule":           resourceScalewaySecurityGroupRule(),
			"scaleway_volume":                        resourceScalewayVolume(),
			"scaleway_volume_attachment":             resourceScalewayVolumeAttachment(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"scaleway_bootscript":              dataSourceScalewayBootscript(),
			"scaleway_image":                   dataSourceScalewayImage(),
			"scaleway_security_group":          dataSourceScalewaySecurityGroup(),
			"scaleway_volume":                  dataSourceScalewayVolume(),
			"scaleway_account_ssh_key":         dataSourceScalewayAccountSSHKey(),
			"scaleway_instance_security_group": dataSourceScalewayInstanceSecurityGroup(),
			"scaleway_instance_server":         dataSourceScalewayInstanceServer(),
			"scaleway_instance_image":          dataSourceScalewayInstanceImage(),
			"scaleway_instance_volume":         dataSourceScalewayInstanceVolume(),
			"scaleway_baremetal_offer":         dataSourceScalewayBaremetalOffer(),
			"scaleway_lb_ip_beta":              dataSourceScalewayLbIPBeta(),
			"scaleway_marketplace_image_beta":  dataSourceScalewayMarketplaceImageBeta(),
			"scaleway_registry_namespace_beta": dataSourceScalewayRegistryNamespaceBeta(),
			"scaleway_registry_image_beta":     dataSourceScalewayRegistryImageBeta(),
		},
	}

	p.ConfigureFunc = func(data *schema.ResourceData) (i interface{}, e error) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return buildMeta(&metaConfig{
			providerSchema:   data,
			terraformVersion: terraformVersion,
		})
	}

	return p
}

type metaConfig struct {
	providerSchema   *schema.ResourceData
	terraformVersion string
	forceRegion      string
	forceZone        string
}

// providerConfigure creates the Meta object containing the SDK client.
func buildMeta(config *metaConfig) (*Meta, error) {

	httpClient := createRetryableHTTPClient(false)

	////
	// Load Profile
	////

	profile, err := loadProfile(config.providerSchema)
	if err != nil {
		return nil, err
	}
	if config.forceRegion != "" {
		profile.DefaultRegion = scw.StringPtr(config.forceRegion)
	}
	if config.forceZone != "" {
		profile.DefaultZone = scw.StringPtr(config.forceZone)
	}

	// TODO validated profile

	////
	// Create scaleway SDK client
	////

	opts := []scw.ClientOption{
		scw.WithUserAgent(fmt.Sprintf("terraform-provider/%s terraform/%s", version, config.terraformVersion)),
		scw.WithProfile(profile),
		scw.WithHTTPClient(httpClient),
	}

	scwClient, err := scw.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	////
	// Create legacy SDK client
	////

	var legacyRegions = map[string]string{
		scw.ZoneFrPar1.String(): "par1",
		scw.ZoneNlAms1.String(): "ams1",
	}

	var deprecatedClient *sdk.API
	if legacyRegion, exist := legacyRegions[*profile.DefaultZone]; exist {
		deprecatedClient, err = sdk.New(
			*profile.DefaultOrganizationID,
			*profile.SecretKey,
			legacyRegion,
			func(sdkApi *sdk.API) {
				sdkApi.Client = httpClient
			},
		)
	}

	// fetch known scaleway server types to support validation in r/server
	if len(commercialServerTypes) == 0 {
		instanceAPI := instance.NewAPI(scwClient)
		availabilityResp, err := instanceAPI.GetServerTypesAvailability(&instance.GetServerTypesAvailabilityRequest{}, scw.WithAllPages())
		if err == nil {
			for k := range availabilityResp.Servers {
				commercialServerTypes = append(commercialServerTypes, k)
			}
			sort.StringSlice(commercialServerTypes).Sort()
		}

		if os.Getenv("DISABLE_SCALEWAY_SERVER_TYPE_VALIDATION") != "" {
			commercialServerTypes = commercialServerTypes[:0]
		}
	}

	return &Meta{
		scwClient:        scwClient,
		deprecatedClient: deprecatedClient,
	}, nil

}

func loadProfile(d *schema.ResourceData) (*scw.Profile, error) {

	config, err := scw.LoadConfig()
	// If the config file do not exist, don't return an error as we may find config in ENV or flags.
	if _, isNotFoundError := err.(*scw.ConfigFileNotFoundError); isNotFoundError {
		config = &scw.Config{}
	} else if err != nil {
		return nil, err
	}

	// By default we set default zone and region to fr-par
	defaultZoneProfile := &scw.Profile{
		DefaultRegion: scw.StringPtr(scw.RegionFrPar.String()),
		DefaultZone:   scw.StringPtr(scw.ZoneFrPar1.String()),
	}

	activeProfile, err := config.GetActiveProfile()
	if err != nil {
		return nil, err
	}
	envProfile := scw.LoadEnvProfile()

	providerProfile := &scw.Profile{}

	if d != nil {
		if accessKey, exist := d.GetOk("access_key"); exist {
			providerProfile.AccessKey = scw.StringPtr(accessKey.(string))
		}
		if secretKey, exist := d.GetOk("secret_key"); exist {
			providerProfile.AccessKey = scw.StringPtr(secretKey.(string))
		}
		if organizationID, exist := d.GetOk("organization_id"); exist {
			providerProfile.DefaultOrganizationID = scw.StringPtr(organizationID.(string))
		}
		if projectID, exist := d.GetOk("project_id"); exist {
			providerProfile.DefaultProjectID = scw.StringPtr(projectID.(string))
		}
		if region, exist := d.GetOk("region"); exist {
			providerProfile.DefaultRegion = scw.StringPtr(region.(string))
		}
		if zone, exist := d.GetOk("zone"); exist {
			providerProfile.DefaultZone = scw.StringPtr(zone.(string))
		}
	}

	profile := scw.MergeProfiles(defaultZoneProfile, activeProfile, providerProfile, envProfile)

	// If profile have a defaultZone but no defaultRegion we set the defaultRegion
	// to the one of the defaultZone
	if profile.DefaultZone != nil && *profile.DefaultZone != "" &&
		(profile.DefaultRegion == nil || *profile.DefaultRegion == "") {
		zone := *profile.DefaultZone
		l.Debugf("guess region from %s zone", zone)
		region := zone[:len(zone)-2]
		if validation.IsRegion(region) {
			profile.DefaultRegion = scw.StringPtr(region)
		} else {
			l.Debugf("invalid guessed region '%s'", region)
		}
	}
	return profile, nil
}
