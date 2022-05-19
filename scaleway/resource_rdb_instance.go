package scaleway

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultWaitRDBRetryInterval = 30 * time.Second
)

func resourceScalewayRdbInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayRdbInstanceCreate,
		ReadContext:   resourceScalewayRdbInstanceRead,
		UpdateContext: resourceScalewayRdbInstanceUpdate,
		DeleteContext: resourceScalewayRdbInstanceDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultRdbInstanceTimeout),
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
				DiffSuppressFunc: diffSuppressFuncIgnoreCase,
			},
			"engine": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Database's engine version id",
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
				Description: "Map of engine settings to be set.",
				Computed:    true,
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
				Type:     schema.TypeString,
				Default:  rdb.VolumeTypeLssd,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					rdb.VolumeTypeLssd.String(),
					rdb.VolumeTypeBssd.String(),
				}, false),
				Description: "Type of volume where data are stored",
			},
			"volume_size_in_gb": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Volume size (in GB) when volume_type is not lssd",
			},
			"private_network": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "List of private network to expose your database instance",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_net": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
							Description:  "The ip net of your private network",
						},
						"pn_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validationUUIDorUUIDWithLocality(),
							Description:  "The private network ID",
						},
						// Computed
						"endpoint_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The endpoint ID",
						},
						"ip": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsIPAddress,
							Description:  "The IP of your private service",
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
							Optional:    true,
							Computed:    true,
							Description: "The name of your private service",
						},
						"hostname": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The hostname of your endpoint",
						},
						"zone": zoneSchema(),
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
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Load balancer of the database instance",
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
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsIPAddress,
							Description:  "The IP of your load balancer service",
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsPortNumber,
							Description:  "The port of your load balancer service",
						},
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The name of your load balancer service",
						},
						"hostname": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The hostname of your endpoint",
						},
					},
				},
			},
			// Common
			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
	}
}

func resourceScalewayRdbInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, err := rdbAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &rdb.CreateInstanceRequest{
		Region:        region,
		ProjectID:     expandStringPtr(d.Get("project_id")),
		Name:          expandOrGenerateString(d.Get("name"), "rdb"),
		NodeType:      d.Get("node_type").(string),
		Engine:        d.Get("engine").(string),
		IsHaCluster:   d.Get("is_ha_cluster").(bool),
		DisableBackup: d.Get("disable_backup").(bool),
		UserName:      d.Get("user_name").(string),
		Password:      d.Get("password").(string),
		VolumeType:    rdb.VolumeType(d.Get("volume_type").(string)),
	}

	rawTag, tagExist := d.GetOk("tags")
	if tagExist {
		createReq.Tags = expandStrings(rawTag)
	}

	pn, pnExist := d.GetOk("private_network")
	if pnExist {
		createReq.InitEndpoints, err = expandPrivateNetwork(pn, pnExist)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		createReq.InitEndpoints = expandLoadBalancer()
	}

	if size, ok := d.GetOk("volume_size_in_gb"); ok {
		if createReq.VolumeType != rdb.VolumeTypeBssd {
			return diag.FromErr(fmt.Errorf("volume_size_in_gb should be used with volume_type %s only", rdb.VolumeTypeBssd.String()))
		}
		createReq.VolumeSize = scw.Size(uint64(size.(int)) * uint64(scw.GB))
	}

	res, err := rdbAPI.CreateInstance(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, res.ID))

	// Configure Schedule Backup
	// BackupScheduleFrequency and BackupScheduleRetention can only configure after instance creation
	if !d.Get("disable_backup").(bool) {
		updateReq := &rdb.UpdateInstanceRequest{
			Region:     region,
			InstanceID: res.ID,
		}

		backupSameRegion, backupSameRegionExist := d.GetOk("backup_same_region")
		if backupSameRegionExist {
			updateReq.BackupSameRegion = expandBoolPtr(backupSameRegion)
		}

		updateReq.IsBackupScheduleDisabled = scw.BoolPtr(d.Get("disable_backup").(bool))
		if backupScheduleFrequency, okFrequency := d.GetOk("backup_schedule_frequency"); okFrequency {
			updateReq.BackupScheduleFrequency = scw.Uint32Ptr(uint32(backupScheduleFrequency.(int)))
		}
		if backupScheduleRetention, okRetention := d.GetOk("backup_schedule_retention"); okRetention {
			updateReq.BackupScheduleRetention = scw.Uint32Ptr(uint32(backupScheduleRetention.(int)))
		}

		_, err = waitForRDBInstance(ctx, rdbAPI, region, res.ID, d.Timeout(schema.TimeoutCreate))
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
		res, err = waitForRDBInstance(ctx, rdbAPI, region, res.ID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err := rdbAPI.SetInstanceSettings(&rdb.SetInstanceSettingsRequest{
			InstanceID: res.ID,
			Region:     region,
			Settings:   expandInstanceSettings(settings),
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayRdbInstanceRead(ctx, d, meta)
}

func resourceScalewayRdbInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, ID, err := rdbAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// verify resource is ready
	res, err := waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutRead))
	if err != nil {
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
	_ = d.Set("user_name", d.Get("user_name").(string)) // user name and
	_ = d.Set("password", d.Get("password").(string))   // password are immutable
	if len(res.Tags) > 0 {
		_ = d.Set("tags", flattenSliceString(res.Tags))
	}
	if res.Endpoint != nil {
		_ = d.Set("endpoint_ip", flattenIPPtr(res.Endpoint.IP))
		_ = d.Set("endpoint_port", int(res.Endpoint.Port))
	} else {
		_ = d.Set("endpoint_ip", "")
		_ = d.Set("endpoint_port", 0)
	}
	if res.Volume != nil {
		_ = d.Set("volume_type", res.Volume.Type)
		_ = d.Set("volume_size_in_gb", int(res.Volume.Size/scw.GB))
	}
	_ = d.Set("read_replicas", flattenRdbInstanceReadReplicas(res.Endpoints))
	_ = d.Set("region", string(region))
	_ = d.Set("organization_id", res.OrganizationID)
	_ = d.Set("project_id", res.ProjectID)

	// set certificate
	cert, err := rdbAPI.GetInstanceCertificate(&rdb.GetInstanceCertificateRequest{
		Region:     region,
		InstanceID: ID,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	certContent, err := ioutil.ReadAll(cert.Content)
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("certificate", string(certContent))

	// set settings
	_ = d.Set("settings", flattenInstanceSettings(res.Settings))

	// set endpoints
	pnI, pnExist := flattenPrivateNetwork(res.Endpoints)
	if pnExist {
		_ = d.Set("private_network", pnI)
	}
	_ = d.Set("load_balancer", flattenLoadBalancer(res.Endpoints))

	return nil
}

//gocyclo:ignore
func resourceScalewayRdbInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, ID, err := rdbAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &rdb.UpdateInstanceRequest{
		Region:     region,
		InstanceID: ID,
	}

	if d.HasChange("name") {
		req.Name = expandStringPtr(d.Get("name"))
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
		req.BackupSameRegion = expandBoolPtr(d.Get("backup_same_region"))
	}
	if d.HasChange("tags") {
		req.Tags = scw.StringsPtr(expandStrings(d.Get("tags")))
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = rdbAPI.UpdateInstance(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	// Change settings
	if d.HasChange("settings") {
		_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil && !is404Error(err) {
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

	upgradeInstanceRequests := []rdb.UpgradeInstanceRequest(nil)
	if d.HasChanges("volume_type", "volume_size_in_gb") {
		volType := rdb.VolumeType(d.Get("volume_type").(string))

		switch volType {
		case rdb.VolumeTypeBssd:
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
					return diag.FromErr(fmt.Errorf("volume_size_in_gb cannot be decreased"))
				}

				if newSize%5 != 0 {
					return diag.FromErr(fmt.Errorf("volume_size_in_gb must be a multiple of 5"))
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
				return diag.FromErr(fmt.Errorf("volume_size_in_gb should be used with volume_type %s only", rdb.VolumeTypeBssd.String()))
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

	if d.HasChange("node_type") {
		upgradeInstanceRequests = append(upgradeInstanceRequests,
			rdb.UpgradeInstanceRequest{
				Region:     region,
				InstanceID: ID,
				NodeType:   expandStringPtr(d.Get("node_type")),
			})
	}

	if d.HasChange("is_ha_cluster") {
		upgradeInstanceRequests = append(upgradeInstanceRequests,
			rdb.UpgradeInstanceRequest{
				Region:     region,
				InstanceID: ID,
				EnableHa:   scw.BoolPtr(d.Get("is_ha_cluster").(bool)),
			})
	}
	for _, request := range upgradeInstanceRequests {
		_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}

		_, err = rdbAPI.UpgradeInstance(&request, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("password") {
		_, err := waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		req := &rdb.UpdateUserRequest{
			Region:     region,
			InstanceID: ID,
			Name:       d.Get("user_name").(string),
			Password:   expandStringPtr(d.Get("password")),
		}

		_, err = rdbAPI.UpdateUser(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("private_network") {
		// retrieve state
		res, err := waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		// get endpoints to detach. It will handle only private networks
		endPointsToRemove, err := endpointsToRemove(res.Endpoints, d.Get("private_network"))
		if err != nil {
			diag.FromErr(err)
		}
		for endPointID, remove := range endPointsToRemove {
			if remove {
				err := rdbAPI.DeleteEndpoint(
					&rdb.DeleteEndpointRequest{
						EndpointID: endPointID, Region: region},
					scw.WithContext(ctx))
				if err != nil {
					diag.FromErr(err)
				}
			}
		}

		// retrieve state
		_, err = waitForRDBInstance(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		// set new endpoints
		pn, pnExist := d.GetOk("private_network")
		if pnExist {
			privateEndpoints, err := expandPrivateNetwork(pn, pnExist)
			if err != nil {
				return diag.FromErr(err)
			}
			for _, e := range privateEndpoints {
				_, err := rdbAPI.CreateEndpoint(
					&rdb.CreateEndpointRequest{Region: region, InstanceID: ID, EndpointSpec: e},
					scw.WithContext(ctx))
				if err != nil {
					diag.FromErr(err)
				}
			}
		}
	}

	return resourceScalewayRdbInstanceRead(ctx, d, meta)
}

func resourceScalewayRdbInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, ID, err := rdbAPIWithRegionAndID(meta, d.Id())
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
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
