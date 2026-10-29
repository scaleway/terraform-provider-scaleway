package opensearch

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	searchdbapi "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceDeployment() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceDeployment().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["name"].ConflictsWith = []string{"deployment_id"}
	dsSchema["deployment_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the OpenSearch deployment",
		ConflictsWith:    []string{"name"},
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: DataSourceDeploymentRead,
		Schema:      dsSchema,
	}
}

func DataSourceDeploymentRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	var deploymentID string

	if id, ok := d.GetOk("deployment_id"); ok {
		parsedRegion, parsedID, parseErr := regional.ParseID(id.(string))
		if parseErr != nil {
			deploymentID = locality.ExpandID(id.(string))
		} else {
			region = parsedRegion
			deploymentID = parsedID
		}
	} else {
		deploymentName := d.Get("name").(string)

		res, err := api.ListDeployments(&searchdbapi.ListDeploymentsRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(deploymentName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundDeployment, err := datasource.FindExact(
			res.Deployments,
			func(s *searchdbapi.Deployment) bool { return s.Name == deploymentName },
			deploymentName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		deploymentID = foundDeployment.ID
	}

	d.SetId(regional.NewIDString(region, deploymentID))
	_ = d.Set("deployment_id", regional.NewIDString(region, deploymentID))

	deployment, err := waitForDeployment(ctx, api, region, deploymentID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	return setDeploymentState(d, deployment)
}
