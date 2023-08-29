package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	document_db "github.com/scaleway/scaleway-sdk-go/api/document_db/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayDocumentDBInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayDocumentDBInstanceCreate,
		ReadContext:   resourceScalewayDocumentDBInstanceRead,
		UpdateContext: resourceScalewayDocumentDBInstanceUpdate,
		DeleteContext: resourceScalewayDocumentDBInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultDocumentDBInstanceTimeout),
			Read:    schema.DefaultTimeout(defaultDocumentDBInstanceTimeout),
			Update:  schema.DefaultTimeout(defaultDocumentDBInstanceTimeout),
			Delete:  schema.DefaultTimeout(defaultDocumentDBInstanceTimeout),
			Default: schema.DefaultTimeout(defaultDocumentDBInstanceTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The document db instance name",
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
				ForceNew:    true,
				Description: "Password for the first user of the database instance",
			},
			"volume_type": {
				Type:     schema.TypeString,
				Default:  document_db.VolumeTypeBssd,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					document_db.VolumeTypeLssd.String(),
					document_db.VolumeTypeBssd.String(),
				}, false),
				Description: "Type of volume where data are stored",
			},
			"volume_size_in_gb": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Volume size (in GB) when volume_type is not lssd",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "List of tags [\"tag1\", \"tag2\", ...] attached to a database instance",
			},
			"telemetry_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable telemetry, support FerretDB, an open-source project",
			},
			"region":     regionSchema(),
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayDocumentDBInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := documentDBAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &document_db.CreateInstanceRequest{
		Region:        region,
		ProjectID:     expandStringPtr(d.Get("project_id")),
		Name:          expandOrGenerateString(d.Get("name").(string), "document-instance"),
		NodeType:      d.Get("node_type").(string),
		Engine:        d.Get("engine").(string),
		IsHaCluster:   d.Get("is_ha_cluster").(bool),
		UserName:      d.Get("user_name").(string),
		Password:      d.Get("password").(string),
		Tags:          expandStrings(d.Get("tags")),
		VolumeType:    document_db.VolumeType(d.Get("volume_type").(string)),
		InitEndpoints: nil, // TODO
	}

	if size, ok := d.GetOk("volume_size_in_gb"); ok {
		if createReq.VolumeType != document_db.VolumeTypeBssd {
			return diag.FromErr(fmt.Errorf("volume_size_in_gb should be used with volume_type %s only", document_db.VolumeTypeBssd.String()))
		}
		createReq.VolumeSize = scw.Size(uint64(size.(int)) * uint64(scw.GB))
	}

	if d.Get("telemetry_enabled").(bool) {
		createReq.InitSettings = append(createReq.InitSettings, &document_db.InstanceSetting{
			Name:  "telemetry_reporting",
			Value: "true",
		})
	}

	instance, err := api.CreateInstance(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, instance.ID))

	_, err = waitForDocumentDBInstance(ctx, api, region, instance.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayDocumentDBInstanceRead(ctx, d, meta)
}

func resourceScalewayDocumentDBInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := documentDBAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	instance, err := waitForDocumentDBInstance(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", instance.Name)
	_ = d.Set("node_type", instance.NodeType)
	_ = d.Set("engine", instance.Engine)
	_ = d.Set("is_ha_cluster", instance.IsHaCluster)
	_ = d.Set("region", instance.Region)
	_ = d.Set("project_id", instance.ProjectID)
	_ = d.Set("tags", instance.Tags)

	if instance.Volume != nil {
		_ = d.Set("volume_type", instance.Volume.Type)
		_ = d.Set("volume_size_in_gb", int(instance.Volume.Size/scw.GB))
	}

	return nil
}

func resourceScalewayDocumentDBInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := documentDBAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	instance, err := waitForDocumentDBInstance(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	req := &document_db.UpdateInstanceRequest{
		Region:     region,
		InstanceID: instance.ID,
	}

	if d.HasChange("name") {
		req.Name = expandUpdatedStringPtr(d.Get("name"))
	}

	if d.HasChange("tags") {
		req.Tags = expandUpdatedStringsPtr(d.Get("tags"))
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := api.UpdateInstance(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	upgradeRequests := []*document_db.UpgradeInstanceRequest(nil)

	if d.HasChanges("volume_type", "volume_size_in_gb") {
		volType := document_db.VolumeType(d.Get("volume_type").(string))

		switch volType {
		case document_db.VolumeTypeBssd:
			if d.HasChange("volume_type") {
				upgradeRequests = append(upgradeRequests,
					&document_db.UpgradeInstanceRequest{
						Region:     region,
						InstanceID: id,
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

				upgradeRequests = append(upgradeRequests,
					&document_db.UpgradeInstanceRequest{
						Region:     region,
						InstanceID: id,
						VolumeSize: scw.Uint64Ptr(newSize * uint64(scw.GB)),
					})
			}
		case document_db.VolumeTypeLssd:
			_, ok := d.GetOk("volume_size_in_gb")
			if d.HasChange("volume_size_in_gb") && ok {
				return diag.FromErr(fmt.Errorf("volume_size_in_gb should be used with volume_type %s only", document_db.VolumeTypeBssd.String()))
			}
			if d.HasChange("volume_type") {
				upgradeRequests = append(upgradeRequests,
					&document_db.UpgradeInstanceRequest{
						Region:     region,
						InstanceID: id,
						VolumeType: &volType,
					})
			}
		default:
			return diag.FromErr(fmt.Errorf("unknown volume_type %s", volType.String()))
		}

		if d.HasChanges("node_type") {
			upgradeRequests = append(upgradeRequests, &document_db.UpgradeInstanceRequest{
				Region:     region,
				InstanceID: id,
				NodeType:   expandStringPtr(d.Get("node_type")),
			})
		}

		if d.HasChange("is_ha_cluster") {
			upgradeRequests = append(upgradeRequests, &document_db.UpgradeInstanceRequest{
				Region:     region,
				InstanceID: id,
				EnableHa:   expandBoolPtr(d.Get("is_ha_cluster")),
			})
		}
	}

	for _, upgradeRequest := range upgradeRequests {
		_, err = waitForDocumentDBInstance(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = api.UpgradeInstance(upgradeRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForDocumentDBInstance(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayDocumentDBInstanceRead(ctx, d, meta)
}

func resourceScalewayDocumentDBInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := documentDBAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteInstance(&document_db.DeleteInstanceRequest{
		Region:     region,
		InstanceID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
