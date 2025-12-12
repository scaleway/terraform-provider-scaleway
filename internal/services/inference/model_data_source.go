package inference

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/inference/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceModel() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceModel().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "url", "name")
	dsSchema["name"].ConflictsWith = []string{"model_id"}
	dsSchema["model_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the model",
		ValidateDiagFunc: verify.IsUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		ReadContext: DataSourceModelRead,
		Schema:      dsSchema,
	}
}

func DataSourceModelRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	modelID, ok := d.GetOk("model_id")
	pageSize := uint32(1000)

	if !ok {
		modelName := d.Get("name").(string)

		modelList, err := api.ListModels(&inference.ListModelsRequest{
			Region:    region,
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
			PageSize:  &pageSize,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundModel, err := datasource.FindExact(
			modelList.Models,
			func(model *inference.Model) bool {
				return model.Name == modelName
			},
			modelName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		modelID = foundModel.ID
	}

	regionalID := datasource.NewRegionalID(modelID, region)
	d.SetId(regionalID)

	err = d.Set("model_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := ResourceModelRead(ctx, d, m)
	if diags != nil {
		return diags
	}

	if d.Id() == "" {
		return diag.FromErr(errors.New("model_id is empty"))
	}

	return nil
}
