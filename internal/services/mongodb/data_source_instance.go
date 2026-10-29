package mongodb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceInstance() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceInstance().SchemaFunc())

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

func DataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	mongodbAPI, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	var instanceID string

	if id, ok := d.GetOk("instance_id"); ok {
		parsedRegion, parsedID, parseErr := regional.ParseID(id.(string))
		if parseErr != nil {
			instanceID = locality.ExpandID(id.(string))
		} else {
			region = parsedRegion
			instanceID = parsedID
		}
	} else {
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

	d.SetId(regional.NewIDString(region, instanceID))
	_ = d.Set("instance_id", regional.NewIDString(region, instanceID))

	instance, err := mongodbAPI.GetInstance(&mongodb.GetInstanceRequest{
		Region:     region,
		InstanceID: instanceID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	return setInstanceState(ctx, d, m, mongodbAPI, region, instance)
}
