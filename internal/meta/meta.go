package meta

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/version"
)

const (
	appendUserAgentEnvVar            = "TF_APPEND_USER_AGENT"
	CredentialsSourceEnvironment     = "Environment variable"
	CredentialsSourceDefault         = "Default"
	CredentialsSourceActiveProfile   = "Active Profile in config.yaml"
	CredentialsSourceProviderProfile = "Profile defined in provider{} block"
	CredentialsSourceInferred        = "CredentialsSourceInferred from default zone"
)

type CredentialsSource struct {
	Variables     map[string][]string
	AccessKey     string
	SecretKey     string
	ProjectID     string
	DefaultZone   string
	DefaultRegion string
}

// Meta contains config and SDK clients used by resources.
//
// This meta value is passed into all resources.
type Meta struct {
	// scwClient is the Scaleway SDK client.
	scwClient *scw.Client
	// httpClient can be either a regular http.Client used to make real HTTP requests
	// or it can be a http.Client used to record and replay cassettes which is useful
	// to replay recorded interactions with APIs locally
	httpClient *http.Client
	// credentialsSource stores information about the source (env, profile, etc.) of each credential
	credentialsSource *CredentialsSource
}

// NewMeta creates the Meta object containing the SDK client.
func NewMeta(ctx context.Context, config *Config) (*Meta, error) {
	////
	// Load Profile
	////
	profile, credentialsSource, err := loadProfile(ctx, config.ProviderSchema)
	if err != nil {
		return nil, err
	}

	if config.ForceZone != "" {
		region, err := config.ForceZone.Region()
		if err != nil {
			return nil, err
		}

		profile.DefaultRegion = scw.StringPtr(region.String())
		profile.DefaultZone = scw.StringPtr(config.ForceZone.String())
	}

	if config.ForceProjectID != "" {
		profile.DefaultProjectID = scw.StringPtr(config.ForceProjectID)
	}

	if config.ForceOrganizationID != "" {
		profile.DefaultOrganizationID = scw.StringPtr(config.ForceOrganizationID)
	}

	if config.ForceAccessKey != "" {
		profile.AccessKey = scw.StringPtr(config.ForceAccessKey)
	}

	if config.ForceSecretKey != "" {
		profile.SecretKey = scw.StringPtr(config.ForceSecretKey)
	}

	// TODO validated profile

	////
	// Create scaleway SDK client
	////
	opts := []scw.ClientOption{
		scw.WithUserAgent(customizeUserAgent(version.Version, config.TerraformVersion)),
		scw.WithProfile(profile),
	}

	httpClient := &http.Client{Transport: transport.NewRetryableTransport(http.DefaultTransport)}
	if config.HTTPClient != nil {
		httpClient = config.HTTPClient
	}

	opts = append(opts, scw.WithHTTPClient(httpClient))

	scwClient, err := scw.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &Meta{
		scwClient:         scwClient,
		httpClient:        httpClient,
		credentialsSource: credentialsSource,
	}, nil
}

func (m Meta) ScwClient() *scw.Client {
	return m.scwClient
}

func (m Meta) HTTPClient() *http.Client {
	return m.httpClient
}

func (m Meta) AccessKeySource() string {
	return m.credentialsSource.AccessKey
}

func (m Meta) SecretKeySource() string {
	return m.credentialsSource.SecretKey
}

func (m Meta) ProjectIDSource() string {
	return m.credentialsSource.ProjectID
}

func (m Meta) RegionSource() string {
	return m.credentialsSource.DefaultRegion
}

func (m Meta) ZoneSource() string {
	return m.credentialsSource.DefaultZone
}

// HasMultipleVariableSources return an informative message during the Provider initialization
// if there are multiple sources of configuration that could confuse the user
//
// Variable			AvailableSources									Using
// SCW_ACCESS_KEY	Active Profile in config.yaml, Environment variable	Environment variable
// SCW_SECRET_KEY	Active Profile in config.yaml, Environment variable	Environment variable
func (m Meta) HasMultipleVariableSources() (bool, string, error) {
	multiple := false

	variables := []string{scw.ScwAccessKeyEnv, scw.ScwSecretKeyEnv, scw.ScwDefaultProjectIDEnv, scw.ScwDefaultRegionEnv, scw.ScwDefaultZoneEnv}

	w := new(tabwriter.Writer)
	buf := &bytes.Buffer{}
	w.Init(buf, 0, 8, 0, '\t', 0)

	_, err := fmt.Fprintln(w, "Variable\tAvailableSources\tUsing")
	if err != nil {
		return false, "", err
	}

	for _, variable := range variables {
		values, ok := m.credentialsSource.Variables[variable]
		if ok {
			if len(values) > 1 {
				_, err := fmt.Fprintf(w, "%s\t%s\t%s\n", variable, strings.Join(values, ", "), values[len(values)-1])
				if err != nil {
					return false, "", err
				}

				multiple = true
			}
		}
	}

	if err := w.Flush(); err != nil {
		return false, "", err
	}

	return multiple, buf.String(), nil
}

type Config struct {
	ProviderSchema      *schema.ResourceData
	HTTPClient          *http.Client
	TerraformVersion    string
	ForceZone           scw.Zone
	ForceProjectID      string
	ForceOrganizationID string
	ForceAccessKey      string
	ForceSecretKey      string
}

func customizeUserAgent(providerVersion string, terraformVersion string) string {
	userAgent := fmt.Sprintf("terraform-provider/%s terraform/%s", providerVersion, terraformVersion)

	if appendUserAgent := os.Getenv(appendUserAgentEnvVar); appendUserAgent != "" {
		userAgent += " " + appendUserAgent
	}

	return userAgent
}

//gocyclo:ignore
func loadProfile(ctx context.Context, d *schema.ResourceData) (*scw.Profile, *CredentialsSource, error) {
	config, err := scw.LoadConfig()
	// If the config file do not exist, don't return an error as we may find config in ENV or flags.
	var configFileNotFoundError *scw.ConfigFileNotFoundError
	if errors.As(err, &configFileNotFoundError) {
		config = &scw.Config{}
	} else if err != nil {
		return nil, nil, err
	}

	// By default we set default zone and region to fr-par
	defaultZoneProfile := &scw.Profile{
		DefaultRegion: scw.StringPtr(scw.RegionFrPar.String()),
		DefaultZone:   scw.StringPtr(scw.ZoneFrPar1.String()),
	}

	activeProfile, err := config.GetActiveProfile()
	if err != nil {
		return nil, nil, err
	}

	envProfile := scw.LoadEnvProfile()

	providerProfile := &scw.Profile{}

	if d != nil {
		if profileName, exist := d.GetOk("profile"); exist {
			profileFromConfig, err := config.GetProfile(profileName.(string))
			if err == nil {
				providerProfile = profileFromConfig
			}
		}

		if accessKey, exist := d.GetOk("access_key"); exist {
			providerProfile.AccessKey = scw.StringPtr(accessKey.(string))
		}

		if secretKey, exist := d.GetOk("secret_key"); exist {
			providerProfile.SecretKey = scw.StringPtr(secretKey.(string))
		}

		if projectID, exist := d.GetOk("project_id"); exist {
			providerProfile.DefaultProjectID = scw.StringPtr(projectID.(string))
		}

		if orgID, exist := d.GetOk("organization_id"); exist {
			providerProfile.DefaultOrganizationID = scw.StringPtr(orgID.(string))
		}

		if region, exist := d.GetOk("region"); exist {
			providerProfile.DefaultRegion = scw.StringPtr(region.(string))
		}

		if zone, exist := d.GetOk("zone"); exist {
			providerProfile.DefaultZone = scw.StringPtr(zone.(string))
		}

		if apiURL, exist := d.GetOk("api_url"); exist {
			providerProfile.APIURL = scw.StringPtr(apiURL.(string))
		}
	}

	profile := scw.MergeProfiles(defaultZoneProfile, activeProfile, providerProfile, envProfile)
	credentialsSource := GetCredentialsSource(defaultZoneProfile, activeProfile, providerProfile, envProfile)

	// If profile have a defaultZone but no defaultRegion we set the defaultRegion
	// to the one of the defaultZone
	if profile.DefaultZone != nil && *profile.DefaultZone != "" &&
		(profile.DefaultRegion == nil || *profile.DefaultRegion == "") {
		zone := scw.Zone(*profile.DefaultZone)
		tflog.Debug(ctx, fmt.Sprintf("guess region from %s zone", zone))

		region, err := zone.Region()
		if err == nil {
			profile.DefaultRegion = scw.StringPtr(region.String())
			credentialsSource.DefaultRegion = CredentialsSourceInferred
		} else {
			tflog.Debug(ctx, "cannot guess region: "+err.Error())
		}
	}

	return profile, credentialsSource, nil
}

// GetCredentialsSource infers the source of the credentials based on the priority order of the different profiles
func GetCredentialsSource(defaultZoneProfile, activeProfile, providerProfile, envProfile *scw.Profile) *CredentialsSource {
	type SourceProfilePair struct {
		Profile *scw.Profile
		Source  string
	}

	profilesInOrder := []SourceProfilePair{
		{
			Source:  CredentialsSourceDefault,
			Profile: defaultZoneProfile,
		},
		{
			Source:  CredentialsSourceActiveProfile,
			Profile: activeProfile,
		},
		{
			Source:  CredentialsSourceProviderProfile,
			Profile: providerProfile,
		},
		{
			Source:  CredentialsSourceEnvironment,
			Profile: envProfile,
		},
	}
	credentialsSource := &CredentialsSource{}
	credentialsSource.Variables = map[string][]string{}

	for _, pair := range profilesInOrder {
		source := pair.Source
		profile := pair.Profile

		if profile.AccessKey != nil {
			credentialsSource.AccessKey = source
			credentialsSource.Variables[scw.ScwAccessKeyEnv] = append(credentialsSource.Variables[scw.ScwAccessKeyEnv], source)
		}

		if profile.SecretKey != nil {
			credentialsSource.SecretKey = source
			credentialsSource.Variables[scw.ScwSecretKeyEnv] = append(credentialsSource.Variables[scw.ScwSecretKeyEnv], source)
		}

		if profile.DefaultProjectID != nil {
			credentialsSource.ProjectID = source
			credentialsSource.Variables[scw.ScwDefaultProjectIDEnv] = append(credentialsSource.Variables[scw.ScwDefaultProjectIDEnv], source)
		}

		if profile.DefaultRegion != nil {
			credentialsSource.DefaultRegion = source
			if source != CredentialsSourceDefault {
				credentialsSource.Variables[scw.ScwDefaultRegionEnv] = append(credentialsSource.Variables[scw.ScwDefaultRegionEnv], source)
			}
		}

		if profile.DefaultZone != nil {
			credentialsSource.DefaultZone = source
			if source != CredentialsSourceDefault {
				credentialsSource.Variables[scw.ScwDefaultZoneEnv] = append(credentialsSource.Variables[scw.ScwDefaultZoneEnv], source)
			}
		}
	}

	return credentialsSource
}
