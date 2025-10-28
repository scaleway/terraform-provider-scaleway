package rdb

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	ipamAPI "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/ipam"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceRdbInstanceCreate,
		ReadContext:   ResourceRdbInstanceRead,
		UpdateContext: ResourceRdbInstanceUpdate,
		DeleteContext: ResourceRdbInstanceDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultInstanceTimeout),
			Read:    schema.DefaultTimeout(defaultInstanceTimeout),
			Update:  schema.DefaultTimeout(defaultInstanceTimeout),
			Delete:  schema.DefaultTimeout(defaultInstanceTimeout),
			Default: schema.DefaultTimeout(defaultInstanceTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the database instance",
			},
			"node_type": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The type of database instance you want to create",
				DiffSuppressFunc: dsf.IgnoreCase,
			},
			"engine": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Database's engine version name (e.g., 'PostgreSQL-16', 'MySQL-8'). Changing this value triggers a blue/green upgrade using MajorUpgradeWorkflow with automatic endpoint migration",
				DiffSuppressFunc: dsf.IgnoreCase,
				ConflictsWith: []string{
					"snapshot_id",
				},
			},
			"snapshot_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID of an existing snapshot to create a new instance from. This allows restoring a database instance to the state captured in the specified snapshot. Conflicts with the `engine` attribute.",
				ConflictsWith: []string{
					"engine",
				},
			},
			"is_ha_cluster": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable or disable high availability for the database instance",
			},
			"disable_backup": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable automated backup for the database instance",
			},
			"backup_schedule_frequency": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Backup schedule frequency in hours",
			},
			"backup_schedule_retention": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Backup schedule retention in days",
			},
			"backup_same_region": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Boolean to store logical backups in the same region as the database instance",
			},
			"user_name": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
				Description: "Identifier for the first user of the database instance",
			},
			"password": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				Description: "Password for the first user of the database instance",
			},
			"settings": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Map of engine settings to be set on a running instance.",
				Computed:    true,
				Optional:    true,
			},
			"init_settings": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Map of engine settings to be set at database initialisation.",
				ForceNew:    true,
				Optional:    true,
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "List of tags [\"tag1\", \"tag2\", ...] attached to a database instance",
			},
			"volume_type": {
				Type:             schema.TypeString,
				Default:          rdb.VolumeTypeLssd,
				Optional:         true,
				ValidateDiagFunc: verify.ValidateEnum[rdb.VolumeType](),
				Description:      "Type of volume where data are stored",
			},
			"volume_size_in_gb": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Volume size (in GB) when volume_type is not lssd",
				ValidateFunc: validation.IntDivisibleBy(5),
			},
			"private_network": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "List of private network to expose your database instance",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pn_id": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
							DiffSuppressFunc: dsf.Locality,
							Description:      "The private network ID",
						},
						// Computed
						"endpoint_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The endpoint ID",
						},
						"ip_net": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsCIDR,
							Description:  "The IP with the given mask within the private subnet",
						},
						"ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IP of your Instance within the private service",
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsPortNumber,
							Description:  "The port of your private service",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of your private service",
						},
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The hostname of your endpoint",
						},
						"enable_ipam": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Whether or not the private network endpoint should be configured with IPAM",
						},
						"zone": zonal.Schema(),
					},
				},
			},
			// Computed
			"endpoint_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint IP of the database instance",
				Deprecated:  "Please use the private_network or the load_balancer attribute",
			},
			"endpoint_port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Endpoint port of the database instance",
				Deprecated:  "Please use the private_network or the load_balancer attribute",
			},
			"read_replicas": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Read replicas of the database instance",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP of the replica",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Port of the replica",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the replica",
						},
					},
				},
			},
			"certificate": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Certificate of the database instance",
			},
			"load_balancer": {
				Type:             schema.TypeList,
				Optional:         true,
				Computed:         true,
				MaxItems:         1,
				Description:      "Load balancer of the database instance",
				DiffSuppressFunc: LoadBalancerDiffSuppressFunc,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Computed
						"endpoint_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The endpoint ID",
						},
						"ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IP of your load balancer service",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The port of your load balancer service",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of your load balancer service",
						},
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The hostname of your endpoint",
						},
					},
				},
			},
			"logs_policy": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "Logs policy configuration",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Computed
						"max_age_retention": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "The max age (in days) of remote logs to keep on the Database Instance",
						},
						"total_disk_retention": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "The max disk size of remote logs to keep on the Database Instance.",
						},
					},
				},
			},
			"encryption_at_rest": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable or disable encryption at rest for the database instance",
			},
			"upgradable_versions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of available engine versions for upgrade",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Version ID for upgrade requests",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Engine name",
						},
						"version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Version string",
						},
						"minor_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Minor version string",
						},
					},
				},
			},
			"private_ip": {
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    true,
				Description: "The private IPv4 address associated with the resource",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the IPv4 address resource",
						},
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The private IPv4 address",
						},
					},
				},
			},
			// Common
			"region":          regional.Schema(),
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
		},
		CustomizeDiff: cdf.LocalityCheck("private_network.#.pn_id"),
	}
}

//gocyclo:ignore
func ResourceRdbInstanceCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	rdbAPI, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	var id string

	if regionalSnapshotID, ok := d.GetOk("snapshot_id"); ok {
		haCluster := d.Get("is_ha_cluster").(bool)
		nodeType := d.Get("node_type").(string)

		_, snapshotID, err := regional.ParseID(regionalSnapshotID.(string))
		if err != nil {
			return diag.FromErr(err)
		}

		createReqFromSnapshot := &rdb.CreateInstanceFromSnapshotRequest{
			SnapshotID:   snapshotID,
			Region:       region,
			InstanceName: types.ExpandOrGenerateString(d.Get("name"), "rdb"),
			IsHaCluster:  &haCluster,
			NodeType:     &nodeType,
		}

		res, err := rdbAPI.CreateInstanceFromSnapshot(createReqFromSnapshot, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForRDBInstance(ctx, rdbAPI, region, res.ID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}

		rawTag, tagExist := d.GetOk("tags")
		if tagExist {
			updateReq := &rdb.UpdateInstanceRequest{
				Region:     region,
				InstanceID: res.ID,
			}
			tags := types.ExpandStrings(rawTag)
			updateReq.Tags = &tags

			_, err = rdbAPI.UpdateInstance(updateReq, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// Configure endpoints after instance creation from snapshot
		if diags := createPrivateNetworkEndpoints(ctx, rdbAPI, region, res.ID, d); diags.HasError() {
			return diags
		}

		if diags := createLoadBalancerEndpoint(ctx, rdbAPI, region, res.ID, d); diags.HasError() {
			return diags
		}

		d.SetId(regional.NewIDString(region, res.ID))
		id = res.ID
	} else {
		createReq := &rdb.CreateInstanceRequest{
			Region:        region,
			ProjectID:     types.ExpandStringPtr(d.Get("project_id")),
			Name:          types.ExpandOrGenerateString(d.Get("name"), "rdb"),
			NodeType:      d.Get("node_type").(string),
			Engine:        d.Get("engine").(string),
			IsHaCluster:   d.Get("is_ha_cluster").(bool),
			DisableBackup: d.Get("disable_backup").(bool),
			UserName:      d.Get("user_name").(string),
			Password:      d.Get("password").(string),
			VolumeType:    rdb.VolumeType(d.Get("volume_type").(string)),
			Encryption: &rdb.EncryptionAtRest{
				Enabled: d.Get("encryption_at_rest").(bool),
			},
		}

		if initSettings, ok := d.GetOk("init_settings"); ok {
			createReq.InitSettings = expandInstanceSettings(initSettings)
		}

		rawTag, tagExist := d.GetOk("tags")
		if tagExist {
			createReq.Tags = types.ExpandStrings(rawTag)
		}

		// Init Endpoints
		endpoints, diags := collectEndpointSpecs(d)
		if diags.HasError() {
			return diags
		}

		for _, warning := range diags {
			tflog.Warn(ctx, warning.Detail)
		}

		createReq.InitEndpoints = endpoints

		if size, ok := d.GetOk("volume_size_in_gb"); ok {
			if createReq.VolumeType == rdb.VolumeTypeLssd {
				return diag.FromErr(fmt.Errorf("volume_size_in_gb should not be used with volume_type %s", rdb.VolumeTypeLssd.String()))
			}

			createReq.VolumeSize = scw.Size(uint64(size.(int)) * uint64(scw.GB))
		}

		res, err := rdbAPI.CreateInstance(createReq, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(regional.NewIDString(region, res.ID))
		id = res.ID
	}

	mustUpdate := false
	updateReq := &rdb.UpdateInstanceRequest{
		Region:     region,
		InstanceID: id,
	}
	// Configure Schedule Backup
	// BackupScheduleFrequency and BackupScheduleRetention can only configure after instance creation
	if !d.Get("disable_backup").(bool) {
		updateReq.BackupSameRegion = types.ExpandBoolPtr(d.Get("backup_same_region"))
		updateReq.IsBackupScheduleDisabled = scw.BoolPtr(d.Get("disable_backup").(bool))

		if backupScheduleFrequency, okFrequency := d.GetOk("backup_schedule_frequency"); okFrequency {
			updateReq.BackupScheduleFrequency = scw.Uint32Ptr(uint32(backupScheduleFrequency.(int)))
		}

		if backupScheduleRetention, okRetention := d.GetOk("backup_schedule_retention"); okRetention {
			updateReq.BackupScheduleRetention = scw.Uint32Ptr(uint32(backupScheduleRetention.(int)))
		}

		mustUpdate = true
	}

	policyRaw, exist := d.GetOk("logs_policy")
	if exist {
		updateReq.LogsPolicy = expandInstanceLogsPolicy(policyRaw)
		mustUpdate = true
	}

	if mustUpdate {
		_, err = waitForRDBInstance(ctx, rdbAPI, region, id, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = rdbAPI.UpdateInstance(updateReq, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	// Configure Instance settings
	if settings, ok := d.GetOk("settings"); ok {
		res, err := waitForRDBInstance(ctx, rdbAPI, region, id, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = rdbAPI.SetInstanceSettings(&rdb.SetInstanceSettingsRequest{
			InstanceID: res.ID,
			Region:     region,
			Settings:   expandInstanceSettings(settings),
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceRdbInstanceRead(ctx, d, m)
}

// createPrivateNetworkEndpoints creates private network endpoints for an instance
func createPrivateNetworkEndpoints(ctx context.Context, rdbAPI *rdb.API, region scw.Region, instanceID string, d *schema.ResourceData) diag.Diagnostics {
	if pn, pnExist := d.GetOk("private_network"); pnExist {
		ipamConfig, staticConfig := getIPConfigCreate(d, "ip_net")

		privateEndpoints, diags := expandPrivateNetwork(pn, pnExist, ipamConfig, staticConfig)
		if diags.HasError() {
			return diags
		}

		for _, warning := range diags {
			tflog.Warn(ctx, warning.Detail)
		}

		for _, e := range privateEndpoints {
			_, err := rdbAPI.CreateEndpoint(
				&rdb.CreateEndpointRequest{Region: region, InstanceID: instanceID, EndpointSpec: e},
				scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return nil
}

// createLoadBalancerEndpoint creates load balancer endpoint for an instance
func createLoadBalancerEndpoint(ctx context.Context, rdbAPI *rdb.API, region scw.Region, instanceID string, d *schema.ResourceData) diag.Diagnostics {
	if _, lbExists := d.GetOk("load_balancer"); lbExists {
		_, err := rdbAPI.CreateEndpoint(&rdb.CreateEndpointRequest{
			Region:       region,
			InstanceID:   instanceID,
			EndpointSpec: expandLoadBalancer(),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

// collectEndpointSpecs collects all endpoint specifications for instance creation
func collectEndpointSpecs(d *schema.ResourceData) ([]*rdb.EndpointSpec, diag.Diagnostics) {
	var endpoints []*rdb.EndpointSpec

	var diags diag.Diagnostics

	// Add private network endpoints
	if pn, pnExist := d.GetOk("private_network"); pnExist {
		ipamConfig, staticConfig := getIPConfigCreate(d, "ip_net")

		privateEndpoints, pnDiags := expandPrivateNetwork(pn, pnExist, ipamConfig, staticConfig)
		if pnDiags.HasError() {
			return nil, pnDiags
		}

		diags = append(diags, pnDiags...)

		endpoints = append(endpoints, privateEndpoints...)
	}

	// Add load balancer endpoint
	if _, lbExists := d.GetOk("load_balancer"); lbExists {
		endpoints = append(endpoints, expandLoadBalancer())
	}

	return endpoints, diags
}

func ResourceRdbInstanceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	rdbAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// verify resource is ready
	res, err := waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", res.Name)
	_ = d.Set("node_type", res.NodeType)
	_ = d.Set("engine", res.Engine)
	_ = d.Set("is_ha_cluster", res.IsHaCluster)
	_ = d.Set("disable_backup", res.BackupSchedule.Disabled)
	_ = d.Set("backup_schedule_frequency", int(res.BackupSchedule.Frequency))
	_ = d.Set("backup_schedule_retention", int(res.BackupSchedule.Retention))
	_ = d.Set("backup_same_region", res.BackupSameRegion)
	_ = d.Set("tags", types.FlattenSliceString(res.Tags))

	var loadBalancerEndpoint *rdb.Endpoint

	for _, endpoint := range res.Endpoints {
		if endpoint.LoadBalancer != nil {
			loadBalancerEndpoint = endpoint

			break
		}
	}

	if loadBalancerEndpoint != nil {
		switch {
		case loadBalancerEndpoint.IP != nil:
			_ = d.Set("endpoint_ip", types.FlattenIPPtr(loadBalancerEndpoint.IP))
		case loadBalancerEndpoint.Hostname != nil:
			_ = d.Set("endpoint_ip", loadBalancerEndpoint.Hostname)
		default:
			_ = d.Set("endpoint_ip", "")
		}

		_ = d.Set("endpoint_port", int(loadBalancerEndpoint.Port))
	} else {
		_ = d.Set("endpoint_ip", "")
		_ = d.Set("endpoint_port", 0)
	}

	if res.Volume != nil {
		_ = d.Set("volume_type", res.Volume.Type)
		_ = d.Set("volume_size_in_gb", int(res.Volume.Size/scw.GB))
	}

	_ = d.Set("read_replicas", []string{})
	_ = d.Set("region", string(region))
	_ = d.Set("organization_id", res.OrganizationID)
	_ = d.Set("project_id", res.ProjectID)

	if res.Encryption != nil {
		_ = d.Set("encryption_at_rest", res.Encryption.Enabled)
	}

	upgradableVersions := make([]map[string]any, len(res.UpgradableVersion))
	for i, version := range res.UpgradableVersion {
		upgradableVersions[i] = map[string]any{
			"id":            version.ID,
			"name":          version.Name,
			"version":       version.Version,
			"minor_version": version.MinorVersion,
		}
	}

	_ = d.Set("upgradable_versions", upgradableVersions)

	// set user and password
	if user, ok := d.GetOk("user_name"); ok {
		_ = d.Set("user_name", user.(string))
	} else {
		users, err := rdbAPI.ListUsers(&rdb.ListUsersRequest{
			Region:     region,
			InstanceID: res.ID,
		}, scw.WithContext(ctx), scw.WithAllPages())
		if err != nil {
			return diag.FromErr(err)
		}

		for _, u := range users.Users {
			if u.IsAdmin {
				_ = d.Set("user_name", u.Name)

				break
			}
		}
	}

	_ = d.Set("password", d.Get("password").(string))

	// set certificate
	cert, err := rdbAPI.GetInstanceCertificate(&rdb.GetInstanceCertificateRequest{
		Region:     region,
		InstanceID: ID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	certContent, err := io.ReadAll(cert.Content)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("certificate", string(certContent))

	// set settings
	_ = d.Set("settings", flattenInstanceSettings(res.Settings))
	_ = d.Set("init_settings", flattenInstanceSettings(res.InitSettings))

	// set logs policy
	_ = d.Set("logs_policy", flattenInstanceLogsPolicy(res.LogsPolicy))

	// set endpoints
	diags := diag.Diagnostics{}
	privateIPs := []map[string]any(nil)
	authorized := true

	if pnI, pnExist := flattenPrivateNetwork(res.Endpoints); pnExist {
		_ = d.Set("private_network", pnI)

		// Read private IP if possible
		for _, endpoint := range res.Endpoints {
			if endpoint.PrivateNetwork == nil || endpoint.PrivateNetwork.ProvisioningMode != rdb.EndpointPrivateNetworkDetailsProvisioningModeIpam {
				continue
			}

			resourceType := ipamAPI.ResourceTypeRdbInstance
			opts := &ipam.GetResourcePrivateIPsOptions{
				ResourceID:       &res.ID,
				ResourceType:     &resourceType,
				PrivateNetworkID: &endpoint.PrivateNetwork.PrivateNetworkID,
				ProjectID:        &res.ProjectID,
			}

			endpointPrivateIPs, err := ipam.GetResourcePrivateIPs(ctx, m, region, opts)

			switch {
			case err == nil:
				privateIPs = append(privateIPs, endpointPrivateIPs...)
			case httperrors.Is403(err):
				authorized = false

				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Warning,
					Summary:       "Unauthorized to read RDB Instance's private IP, please check your IAM permissions",
					Detail:        err.Error(),
					AttributePath: cty.GetAttrPath("private_ip"),
				})
			default:
				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Warning,
					Summary:       fmt.Sprintf("Unable to get private IP for instance %q", res.Name),
					Detail:        err.Error(),
					AttributePath: cty.GetAttrPath("private_ip"),
				})
			}

			if !authorized {
				break
			}
		}
	}

	if authorized {
		_ = d.Set("private_ip", privateIPs)
	}

	if lbI, lbExists := flattenLoadBalancer(res.Endpoints); lbExists {
		_ = d.Set("load_balancer", lbI)
	}

	return diags
}

//gocyclo:ignore
func ResourceRdbInstanceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	rdbAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	////////////////////
	// Upgrade instance
	////////////////////
	upgradeInstanceRequests := []rdb.UpgradeInstanceRequest(nil)

	rdbInstance, err := rdbAPI.GetInstance(&rdb.GetInstanceRequest{
		Region:     region,
		InstanceID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	diskIsFull := rdbInstance.Status == rdb.InstanceStatusDiskFull
	volType := rdb.VolumeType(d.Get("volume_type").(string))

	// Volume type and size
	if d.HasChanges("volume_type", "volume_size_in_gb") {
		switch volType {
		case rdb.VolumeTypeSbs5k, rdb.VolumeTypeSbs15k:
			if d.HasChange("volume_type") {
				upgradeInstanceRequests = append(upgradeInstanceRequests,
					rdb.UpgradeInstanceRequest{
						Region:     region,
						InstanceID: ID,
						VolumeType: &volType,
					})
			}

			if d.HasChange("volume_size_in_gb") {
				oldSizeInterface, newSizeInterface := d.GetChange("volume_size_in_gb")
				oldSize := uint64(oldSizeInterface.(int))
				newSize := uint64(newSizeInterface.(int))

				if newSize < oldSize {
					return diag.FromErr(errors.New("volume_size_in_gb cannot be decreased"))
				}

				if newSize%5 != 0 {
					return diag.FromErr(errors.New("volume_size_in_gb must be a multiple of 5"))
				}

				upgradeInstanceRequests = append(upgradeInstanceRequests,
					rdb.UpgradeInstanceRequest{
						Region:     region,
						InstanceID: ID,
						VolumeSize: scw.Uint64Ptr(newSize * uint64(scw.GB)),
					})
			}
		case rdb.VolumeTypeLssd:
			_, ok := d.GetOk("volume_size_in_gb")
			if d.HasChange("volume_size_in_gb") && ok {
				return diag.FromErr(fmt.Errorf("volume_size_in_gb should be used with volume_type %s or %s only", rdb.VolumeTypeSbs5k.String(), rdb.VolumeTypeSbs15k.String()))
			}

			if d.HasChange("volume_type") {
				upgradeInstanceRequests = append(upgradeInstanceRequests,
					rdb.UpgradeInstanceRequest{
						Region:     region,
						InstanceID: ID,
						VolumeType: &volType,
					})
			}
		default:
			return diag.FromErr(fmt.Errorf("unknown volume_type %s", volType.String()))
		}
	}

	// Node type
	if d.HasChange("node_type") {
		// Upgrading the node_type with block storage is not allowed when the disk is full, so if we are in this case,
		// we can only allow this action if an increase of the size of the volume is also scheduled before it.
		if !diskIsFull || len(upgradeInstanceRequests) > 0 {
			upgradeInstanceRequests = append(upgradeInstanceRequests,
				rdb.UpgradeInstanceRequest{
					Region:     region,
					InstanceID: ID,
					NodeType:   types.ExpandStringPtr(d.Get("node_type")),
				})
		} else {
			return diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Node type upgrade forbidden when disk is full",
				Detail:   "You cannot upgrade the node_type of an instance when it is in disk_full state. Please increase the volume_size_in_gb first.",
			}}
		}
	}

	// HA cluster
	if d.HasChange("is_ha_cluster") {
		upgradeInstanceRequests = append(upgradeInstanceRequests,
			rdb.UpgradeInstanceRequest{
				Region:     region,
				InstanceID: ID,
				EnableHa:   scw.BoolPtr(d.Get("is_ha_cluster").(bool)),
			})
	}

	// If we are switching to local storage, we have to make sure that the node_type upgrade is done first
	if d.HasChange("volume_type") {
		wantedVolumeType := d.Get("volume_type")
		if wantedVolumeType == rdb.VolumeTypeLssd.String() {
			for i, req := range upgradeInstanceRequests {
				if req.NodeType != nil && i != 0 {
					upgradeInstanceRequests[0], upgradeInstanceRequests[i] = upgradeInstanceRequests[i], upgradeInstanceRequests[0]
				}
			}
		}
	}

	if d.HasChange("encryption_at_rest") {
		oldValue, newValue := d.GetChange("encryption_at_rest")

		if oldValue.(bool) && !newValue.(bool) {
			return diag.FromErr(errors.New("disabling encryption_at_rest is not supported once it has been enabled"))
		}

		upgradeInstanceRequests = append(upgradeInstanceRequests,
			rdb.UpgradeInstanceRequest{
				Region:           region,
				InstanceID:       ID,
				EnableEncryption: scw.BoolPtr(newValue.(bool)),
			})
	}

	if d.HasChange("engine") {
		oldEngine, newEngine := d.GetChange("engine")
		newEngineStr := newEngine.(string)

		targetVersionID := ""

		var availableVersions []string
		for _, version := range rdbInstance.UpgradableVersion {
			availableVersions = append(availableVersions, version.Name)
			if version.Name == newEngineStr {
				targetVersionID = version.ID

				break
			}
		}

		if targetVersionID == "" {
			return diag.FromErr(fmt.Errorf("engine version %s is not available for upgrade from %s. Available versions: %v",
				newEngineStr, oldEngine.(string), availableVersions))
		}

		upgradeInstanceRequests = append(upgradeInstanceRequests,
			rdb.UpgradeInstanceRequest{
				Region:     region,
				InstanceID: ID,
				MajorUpgradeWorkflow: &rdb.UpgradeInstanceRequestMajorUpgradeWorkflow{
					UpgradableVersionID: targetVersionID,
					WithEndpoints:       true,
				},
			})
	}

	for i := range upgradeInstanceRequests {
		_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}

		upgradedInstance, err := rdbAPI.UpgradeInstance(&upgradeInstanceRequests[i], scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		if upgradeInstanceRequests[i].MajorUpgradeWorkflow != nil && upgradedInstance.ID != ID {
			tflog.Info(ctx, fmt.Sprintf("Engine upgrade created new instance, updating ID from %s to %s", ID, upgradedInstance.ID))
			oldInstanceID := ID
			ID = upgradedInstance.ID
			d.SetId(regional.NewIDString(region, ID))

			_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
			if err != nil && !httperrors.Is404(err) {
				return diag.FromErr(err)
			}

			_, err = waitForRDBInstance(ctx, rdbAPI, region, oldInstanceID, d.Timeout(schema.TimeoutUpdate))
			if err != nil && !httperrors.Is404(err) {
				tflog.Warn(ctx, fmt.Sprintf("Old instance %s not ready for deletion: %v", oldInstanceID, err))
			} else {
				_, err = rdbAPI.DeleteInstance(&rdb.DeleteInstanceRequest{
					Region:     region,
					InstanceID: oldInstanceID,
				}, scw.WithContext(ctx))
				if err != nil && !httperrors.Is404(err) {
					tflog.Warn(ctx, fmt.Sprintf("Failed to delete old instance %s: %v", oldInstanceID, err))
				} else {
					_, err = rdbAPI.WaitForInstance(&rdb.WaitForInstanceRequest{
						Region:     region,
						InstanceID: oldInstanceID,
						Timeout:    scw.TimeDurationPtr(d.Timeout(schema.TimeoutUpdate)),
					}, scw.WithContext(ctx))
					if err != nil && !httperrors.Is404(err) {
						tflog.Warn(ctx, fmt.Sprintf("Error waiting for old instance %s deletion: %v", oldInstanceID, err))
					}
				}
			}
		} else {
			_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
			if err != nil && !httperrors.Is404(err) {
				return diag.FromErr(err)
			}
		}
	}

	////////////////////
	// Update instance
	////////////////////
	req := &rdb.UpdateInstanceRequest{
		Region:     region,
		InstanceID: ID,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandStringPtr(d.Get("name"))
	}

	if d.HasChange("disable_backup") {
		req.IsBackupScheduleDisabled = scw.BoolPtr(d.Get("disable_backup").(bool))
	}

	if d.HasChange("backup_schedule_frequency") {
		req.BackupScheduleFrequency = scw.Uint32Ptr(uint32(d.Get("backup_schedule_frequency").(int)))
	}

	if d.HasChange("backup_schedule_retention") {
		req.BackupScheduleRetention = scw.Uint32Ptr(uint32(d.Get("backup_schedule_retention").(int)))
	}

	if d.HasChange("backup_same_region") {
		req.BackupSameRegion = types.ExpandBoolPtr(d.Get("backup_same_region"))
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChange("logs_policy") {
		policyRaw, exist := d.GetOk("logs_policy")
		if exist {
			req.LogsPolicy = expandInstanceLogsPolicy(policyRaw)
		}
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = rdbAPI.UpdateInstance(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	////////////////////
	// Change settings
	////////////////////
	if d.HasChange("settings") {
		_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}

		_, err := rdbAPI.SetInstanceSettings(&rdb.SetInstanceSettingsRequest{
			InstanceID: ID,
			Region:     region,
			Settings:   expandInstanceSettings(d.Get("settings")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	////////////////////
	// Update user
	////////////////////
	if d.HasChange("password") {
		_, err := waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		req := &rdb.UpdateUserRequest{
			Region:     region,
			InstanceID: ID,
			Name:       d.Get("user_name").(string),
			Password:   types.ExpandStringPtr(d.Get("password")),
		}

		_, err = rdbAPI.UpdateUser(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	////////////////////
	// Update endpoints
	////////////////////
	if d.HasChanges("private_network") {
		// retrieve state
		res, err := waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		// delete old endpoint
		for _, e := range res.Endpoints {
			if e.PrivateNetwork != nil {
				err := rdbAPI.DeleteEndpoint(
					&rdb.DeleteEndpointRequest{
						EndpointID: e.ID, Region: region,
					},
					scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
		// retrieve state
		_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
		// set new endpoint
		pn, pnExist := d.GetOk("private_network")
		if pnExist {
			ipamConfig, staticConfig := getIPConfigUpdate(d, "ip_net")

			privateEndpoints, diags := expandPrivateNetwork(pn, pnExist, ipamConfig, staticConfig)
			if diags.HasError() {
				return diags
			}

			for _, warning := range diags {
				tflog.Warn(ctx, warning.Detail)
			}

			for _, e := range privateEndpoints {
				_, err := rdbAPI.CreateEndpoint(
					&rdb.CreateEndpointRequest{Region: region, InstanceID: ID, EndpointSpec: e},
					scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	if d.HasChanges("load_balancer") {
		// retrieve state
		res, err := waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
		// delete old endpoint
		for _, e := range res.Endpoints {
			if e.LoadBalancer != nil {
				err := rdbAPI.DeleteEndpoint(&rdb.DeleteEndpointRequest{
					EndpointID: e.ID,
					Region:     region,
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
		// retrieve state
		_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
		// set new endpoint
		if _, lbExists := d.GetOk("load_balancer"); lbExists {
			_, err := rdbAPI.CreateEndpoint(&rdb.CreateEndpointRequest{
				Region:       region,
				InstanceID:   ID,
				EndpointSpec: expandLoadBalancer(),
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return ResourceRdbInstanceRead(ctx, d, m)
}

func ResourceRdbInstanceDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	rdbAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// We first wait in case the instance is in a transient state
	_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = rdbAPI.DeleteInstance(&rdb.DeleteInstanceRequest{
		Region:     region,
		InstanceID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	// Lastly wait in case the instance is in a transient state
	_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
