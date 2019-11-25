package scaleway

import (
	"io/ioutil"

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
	if res.Endpoint != nil {
		d.Set("endpoint_ip", flattenIPPtr(res.Endpoint.IP))
		d.Set("endpoint_port", int(res.Endpoint.Port))
	} else {
		d.Set("endpoint_ip", "")
		d.Set("endpoint_port", 0)
	}
	d.Set("read_replicas", flattenRdbInstanceReadReplicas(res.ReadReplicas))
	d.Set("region", string(region))
	d.Set("organization_id", res.OrganizationID)

	// set certificate
	cert, err := rdbAPI.GetInstanceCertificate(&rdb.GetInstanceCertificateRequest{
		Region:     region,
		InstanceID: ID,
	})
	if err != nil {
		return err
	}
	certContent, err := ioutil.ReadAll(cert.Content)
	if err != nil {
		return err
	}
	d.Set("certificate", string(certContent))

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

	if d.HasChange("tags") {
		req.Tags = scw.StringsPtr(expandStrings(d.Get("tags")))
	}

	_, err = rdbAPI.UpdateInstance(req)
	if err != nil {
		return err
	}

	if d.HasChange("node_type") {
		_, err = rdbAPI.UpgradeInstance(&rdb.UpgradeInstanceRequest{
			Region:     region,
			InstanceID: ID,
			NodeType:   d.Get("node_type").(string),
		})
		if err != nil {
			return err
		}

		_, err = rdbAPI.WaitForInstance(&rdb.WaitForInstanceRequest{
			Region:     region,
			InstanceID: ID,
			Timeout:    InstanceServerWaitForTimeout * 3, // upgrade takes some time
		})
		if err != nil {
			return err
		}
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
