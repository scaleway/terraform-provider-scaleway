package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayRdbInstanceBeta() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayRdbInstanceBetaCreate,
		Read:   resourceScalewayRdbInstanceBetaRead,
		Update: resourceScalewayRdbInstanceBetaUpdate,
		Delete: resourceScalewayRdbInstanceBetaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The type of database instance you want to create",
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
				ForceNew:    true,
				Default:     false,
				Description: "Enable or disable high availability for the database instance",
			},
			"disable_backup": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable automated backup for the database instance",
			},
			"user_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Identifier for the first user of the database instance",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Password for the first user of the database instance",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "List of tags [\"tag1\", \"tag2\", ...] attached to a database instance",
			},

			// TODO: computed (endpoint_ip,endpoint_port,read_replicas,certificate,backup_schedule)

			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayRdbInstanceBetaCreate(d *schema.ResourceData, m interface{}) error {
	rdbAPI, region, err := getRdbAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	createReq := &rdb.CreateInstanceRequest{
		Region:         region,
		OrganizationID: d.Get("organization_id").(string),
		Name:           expandOrGenerateString(d.Get("name"), "rdb"),
		NodeType:       d.Get("node_type").(string),
		Engine:         d.Get("engine").(string),
		IsHaCluster:    d.Get("is_ha_cluster").(bool),
		DisableBackup:  d.Get("disable_backup").(bool),
		UserName:       d.Get("user_name").(string),
		Password:       d.Get("password").(string),
		Tags:           expandStrings(d.Get("tags")),
	}

	res, err := rdbAPI.CreateInstance(createReq)
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(region, res.ID))

	_, err = rdbAPI.WaitForInstance(&rdb.WaitForInstanceRequest{
		Region:     region,
		InstanceID: res.ID,
		Timeout:    InstanceServerWaitForTimeout,
	})
	if err != nil {
		return err
	}

	return resourceScalewayRdbInstanceBetaRead(d, m)
}

func resourceScalewayRdbInstanceBetaRead(d *schema.ResourceData, m interface{}) error {
	rdbAPI, region, ID, err := getRdbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := rdbAPI.GetInstance(&rdb.GetInstanceRequest{
		Region:     region,
		InstanceID: ID,
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", res.Name)
	d.Set("node_type", res.NodeType)
	d.Set("engine", res.Engine)
	d.Set("is_ha_cluster", res.IsHaCluster)
	d.Set("disable_backup", res.BackupSchedule.Disabled)
	d.Set("user_name", d.Get("user_name").(string)) // user name and
	d.Set("password", d.Get("password").(string))   // password are immutable
	d.Set("tags", res.Tags)
	d.Set("region", string(region))
	d.Set("organization_id", res.OrganizationID)

	return nil
}

func resourceScalewayRdbInstanceBetaUpdate(d *schema.ResourceData, m interface{}) error {
	rdbAPI, region, ID, err := getRdbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	req := &rdb.UpdateInstanceRequest{
		Region:     region,
		InstanceID: ID,
	}

	if d.HasChange("name") {
		req.Name = scw.StringPtr(d.Get("name").(string))
	}
	if d.HasChange("disable_backup") {
		req.IsBackupScheduleDisabled = scw.BoolPtr(d.Get("disable_backup").(bool))
	}

	//if d.HasChange("tags") {
	req.Tags = scw.StringsPtr(StringSliceFromState(d.Get("tags").([]interface{}))) // due to a bug in the API Tags must always be sent for now
	//}

	// TODO: handle engine upgrade
	_, err = rdbAPI.UpdateInstance(req)
	if err != nil {
		return err
	}

	return resourceScalewayRdbInstanceBetaRead(d, m)
}

func resourceScalewayRdbInstanceBetaDelete(d *schema.ResourceData, m interface{}) error {
	rdbAPI, region, ID, err := getRdbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	_, err = rdbAPI.DeleteInstance(&rdb.DeleteInstanceRequest{
		Region:     region,
		InstanceID: ID,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	_, err = rdbAPI.WaitForInstance(&rdb.WaitForInstanceRequest{
		InstanceID: ID,
		Region:     region,
		Timeout:    LbWaitForTimeout,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
