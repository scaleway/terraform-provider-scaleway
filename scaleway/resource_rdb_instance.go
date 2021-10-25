package scaleway

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	rdbV1 "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
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
				Default:  rdbV1.VolumeTypeLssd,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					rdbV1.VolumeTypeLssd.String(),
					rdbV1.VolumeTypeBssd.String(),
				}, false),
				Description: "Type of volume where data are stored",
			},
			"volume_size_in_gb": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Volume size (in GB) when volume_type is not lssd",
			},
			"load_balancer": {
				ConflictsWith: []string{"private_network"},
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				Description:   "Expose data base exposed by a load-balancer",
			},
			"private_network": {
				ConflictsWith: []string{"load_balancer"},
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				Description:   "List of private network to expose your database instance",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
							Description:  "The IP of your service",
						},
						"pn_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validationUUIDorUUIDWithLocality(),
							Description:  "The private network ID",
						},
					},
				},
			},
			// Computed
			"endpoint_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint IP of the database instance",
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

	createReq := &rdbV1.CreateInstanceRequest{
		Region:        region,
		ProjectID:     expandStringPtr(d.Get("project_id")),
		Name:          expandOrGenerateString(d.Get("name"), "rdb"),
		NodeType:      d.Get("node_type").(string),
		Engine:        d.Get("engine").(string),
		IsHaCluster:   d.Get("is_ha_cluster").(bool),
		DisableBackup: d.Get("disable_backup").(bool),
		UserName:      d.Get("user_name").(string),
		Password:      d.Get("password").(string),
		Tags:          expandStrings(d.Get("tags")),
		VolumeType:    rdbV1.VolumeType(d.Get("volume_type").(string)),
	}

	pn, pnExist := d.GetOk("private_network")
	if pnExist {
		createReq.InitEndpoints = expandPrivateNetwork(pn, pnExist)
	}

	lb, lbExist := d.GetOk("load_balancer")
	if lbExist {
		createReq.InitEndpoints = expandLoadBalancer(lb)
	}

	if size, ok := d.GetOk("volume_size_in_gb"); ok {
		if createReq.VolumeType != rdbV1.VolumeTypeBssd {
			return diag.FromErr(fmt.Errorf("volume_size_in_gb should be used with volume_type %s only", rdbV1.VolumeTypeBssd.String()))
		}
		createReq.VolumeSize = scw.Size(uint64(size.(int)) * uint64(scw.GB))
	}

	res, err := rdbAPI.CreateInstance(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, res.ID))

	retryInterval := defaultWaitRDBRetryInterval
	_, err = rdbAPI.WaitForInstance(&rdbV1.WaitForInstanceRequest{
		Region:        region,
		InstanceID:    res.ID,
		Timeout:       scw.TimeDurationPtr(defaultRdbInstanceTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	// Configure Schedule Backup
	// BackupScheduleFrequency and BackupScheduleRetention can only configure after instance creation
	if !d.Get("disable_backup").(bool) {
		updateReq := &rdbV1.UpdateInstanceRequest{
			Region:     region,
			InstanceID: res.ID,
		}
		updateReq.IsBackupScheduleDisabled = scw.BoolPtr(d.Get("disable_backup").(bool))
		if backupScheduleFrequency, okFrequency := d.GetOk("backup_schedule_frequency"); okFrequency {
			updateReq.BackupScheduleFrequency = scw.Uint32Ptr(uint32(backupScheduleFrequency.(int)))
		}
		if backupScheduleRetention, okRetention := d.GetOk("backup_schedule_retention"); okRetention {
			updateReq.BackupScheduleRetention = scw.Uint32Ptr(uint32(backupScheduleRetention.(int)))
		}

		_, err = rdbAPI.UpdateInstance(updateReq, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	// Configure Instance settings
	if settings, ok := d.GetOk("settings"); ok {
		_, err := rdbAPI.SetInstanceSettings(&rdbV1.SetInstanceSettingsRequest{
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

	res, err := rdbAPI.GetInstance(&rdbV1.GetInstanceRequest{
		Region:     region,
		InstanceID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
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
	_ = d.Set("user_name", d.Get("user_name").(string)) // username and
	_ = d.Set("password", d.Get("password").(string))   // password are immutable
	_ = d.Set("tags", res.Tags)
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
	_ = d.Set("read_replicas", flattenRdbInstanceReadReplicas(res.ReadReplicas))
	_ = d.Set("region", string(region))
	_ = d.Set("organization_id", res.OrganizationID)
	_ = d.Set("project_id", res.ProjectID)

	// set certificate
	cert, err := rdbAPI.GetInstanceCertificate(&rdbV1.GetInstanceCertificateRequest{
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

	lbBoolean := *expandBoolPtr(d.Get("load_balancer"))
	_ = d.Set("load_balancer", lbBoolean)

	// set private_network
	_ = d.Set("private_network", flattenInstancePrivateNetwork(res.Endpoints))

	return nil
}

func resourceScalewayRdbInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, ID, err := rdbAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &rdbV1.UpdateInstanceRequest{
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
	if d.HasChange("tags") {
		req.Tags = scw.StringsPtr(expandStrings(d.Get("tags")))
	}

	_, err = rdbAPI.UpdateInstance(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	// Change settings
	if d.HasChange("settings") {
		_, err := rdbAPI.SetInstanceSettings(&rdbV1.SetInstanceSettingsRequest{
			InstanceID: ID,
			Region:     region,
			Settings:   expandInstanceSettings(d.Get("settings")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	upgradeInstanceRequests := []rdbV1.UpgradeInstanceRequest(nil)
	if d.HasChanges("volume_type", "volume_size_in_gb") {
		volType := rdbV1.VolumeType(d.Get("volume_type").(string))

		switch volType {
		case rdbV1.VolumeTypeBssd:
			if d.HasChange("volume_type") {
				upgradeInstanceRequests = append(upgradeInstanceRequests,
					rdbV1.UpgradeInstanceRequest{
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
					rdbV1.UpgradeInstanceRequest{
						Region:     region,
						InstanceID: ID,
						VolumeSize: scw.Uint64Ptr(newSize * uint64(scw.GB)),
					})
			}
		case rdbV1.VolumeTypeLssd:
			_, ok := d.GetOk("volume_size_in_gb")
			if d.HasChange("volume_size_in_gb") && ok {
				return diag.FromErr(fmt.Errorf("volume_size_in_gb should be used with volume_type %s only", rdbV1.VolumeTypeBssd.String()))
			}
			if d.HasChange("volume_type") {
				upgradeInstanceRequests = append(upgradeInstanceRequests,
					rdbV1.UpgradeInstanceRequest{
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
			rdbV1.UpgradeInstanceRequest{
				Region:     region,
				InstanceID: ID,
				NodeType:   expandStringPtr(d.Get("node_type")),
			})
	}

	if d.HasChange("is_ha_cluster") {
		upgradeInstanceRequests = append(upgradeInstanceRequests,
			rdbV1.UpgradeInstanceRequest{
				Region:     region,
				InstanceID: ID,
				EnableHa:   scw.BoolPtr(d.Get("is_ha_cluster").(bool)),
			})
	}
	for _, request := range upgradeInstanceRequests {
		_, err = rdbAPI.UpgradeInstance(&request, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		retryInterval := defaultWaitRDBRetryInterval
		_, err = rdbAPI.WaitForInstance(&rdbV1.WaitForInstanceRequest{
			Region:        region,
			InstanceID:    ID,
			Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout * 3), // upgrade takes some time
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		// Wait for the instance to settle after upgrading
		time.Sleep(30 * time.Second) // lintignore:R018
	}

	if d.HasChange("password") {
		req := &rdbV1.UpdateUserRequest{
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

	return resourceScalewayRdbInstanceRead(ctx, d, meta)
}

func resourceScalewayRdbInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, ID, err := rdbAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	retryInterval := defaultWaitRDBRetryInterval
	// We first wait in case the instance is in a transient state
	_, err = rdbAPI.WaitForInstance(&rdbV1.WaitForInstanceRequest{
		InstanceID:    ID,
		Region:        region,
		Timeout:       scw.TimeDurationPtr(LbWaitForTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	_, err = rdbAPI.DeleteInstance(&rdbV1.DeleteInstanceRequest{
		Region:     region,
		InstanceID: ID,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}
	// Lastly wait in case the instance is in a transient state
	_, err = rdbAPI.WaitForInstance(&rdbV1.WaitForInstanceRequest{
		InstanceID:    ID,
		Region:        region,
		Timeout:       scw.TimeDurationPtr(LbWaitForTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
