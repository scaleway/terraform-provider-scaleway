package mongodb

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceInstance() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceInstance().Schema)

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["name"].ConflictsWith = []string{"instance_id"}
	dsSchema["instance_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "instance id",
		ConflictsWith:    []string{"name"},
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: DataSourceInstanceRead,
		Schema:      dsSchema,
	}
}

func DataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	mongodbAPI, zone, region, err := newAPIWithZoneAndRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID, ok := d.GetOk("instance_id")
	if !ok {
		instanceName := d.Get("name").(string)

		res, err := mongodbAPI.ListInstances(&mongodb.ListInstancesRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(instanceName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundInstance, err := datasource.FindExact(
			res.Instances,
			func(s *mongodb.Instance) bool { return s.Name == instanceName },
			instanceName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		instanceID = foundInstance.ID
	}

	zonedID := datasource.NewZonedID(instanceID, zone)
	d.SetId(zonedID)

	err = d.Set("instance_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}

	getReq := &mongodb.GetInstanceRequest{
		Region:     region,
		InstanceID: locality.ExpandID(instanceID.(string)),
	}

	instance, err := mongodbAPI.GetInstance(getReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", instance.Name)
	_ = d.Set("version", instance.Version)
	_ = d.Set("node_number", int(instance.NodeNumber))
	_ = d.Set("node_type", instance.NodeType)
	_ = d.Set("project_id", instance.ProjectID)
	_ = d.Set("tags", instance.Tags)
	_ = d.Set("created_at", instance.CreatedAt.Format(time.RFC3339))
	_ = d.Set("region", instance.Region.String())

	if instance.Volume != nil {
		_ = d.Set("volume_type", instance.Volume.Type)
		_ = d.Set("volume_size_in_gb", int(instance.Volume.Size/scw.GB))
	}

	publicNetworkEndpoint, publicNetworkExists := flattenPublicNetwork(instance.Endpoints)
	if publicNetworkExists {
		_ = d.Set("public_network", publicNetworkEndpoint)
	}

	if len(instance.Settings) > 0 {
		settingsMap := make(map[string]string)
		for _, setting := range instance.Settings {
			settingsMap[setting.Name] = setting.Value
		}

		_ = d.Set("settings", settingsMap)
	}

	return nil
}
