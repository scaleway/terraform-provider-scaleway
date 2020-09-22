package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayRDBInstanceBeta() *schema.Resource {
	return &schema.Resource{
		Read: func(data *schema.ResourceData, i interface{}) error {
			api, region, err := rdbAPIWithRegion(data, i)
			if err != nil {
				return err
			}

			var instance *rdb.Instance
			instanceID, ok := data.GetOk("instance_id")
			if !ok {
				if data.Get("instance_id") != "" {
					instanceID = scw.StringPtr(expandID(data.Get("instance_id")))
				}
				res, err := api.ListInstances(&rdb.ListInstancesRequest{
					Region: region,
					Name:   String(data.Get("name").(string)),
				})
				if err != nil {
					return err
				}
				if len(res.Instances) == 0 {
					return fmt.Errorf("no instances found with the name %s", data.Get("name"))
				}
				if len(res.Instances) > 1 {
					return fmt.Errorf("%d instances found with the same name %s", len(res.Instances), data.Get("name"))
				}
				instance = res.Instances[0]
			} else {
				res, err := api.GetInstance(&rdb.GetInstanceRequest{
					Region:     region,
					InstanceID: expandID(instanceID),
				})
				if err != nil {
					return err
				}
				instance = res
			}

			data.SetId(datasourceNewRegionalizedID(instance.ID, region))
			_ = data.Set("instance_id", instance.ID)
			_ = data.Set("name", instance.Name)
			_ = data.Set("tags", instance.Tags)

			return nil
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The name of the rdb instance",
				ConflictsWith: []string{"instance_id"},
			},
			"instance_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The ID of the rdb instance",
				ValidateFunc:  validationUUIDorUUIDWithLocality(),
				ConflictsWith: []string{"name"},
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The tags associated with the rdb instance",
			},
			"node_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The type of database instance you want to create",
			},
			"engine": {
				Type:        schema.TypeString,
				Optional:    true,
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
			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
		},
	}
}
