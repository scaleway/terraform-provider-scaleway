package datawarehouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	datawarehouseapi "github.com/scaleway/scaleway-sdk-go/api/datawarehouse/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func ResourceDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseCreate,
		ReadContext:   resourceDatabaseRead,
		DeleteContext: resourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: databaseSchema,
	}
}

func databaseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"region": regional.Schema(),
		"deployment_id": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			Description:      "ID of the Datawarehouse deployment to which this database belongs.",
			DiffSuppressFunc: dsf.Locality,
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Name of the database.",
		},
		"size": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Size of the database (in GB).",
		},
	}
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := datawarehouseAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	deploymentID := locality.ExpandID(d.Get("deployment_id").(string))
	name := d.Get("name").(string)

	req := &datawarehouseapi.CreateDatabaseRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Name:         name,
	}

	_, err = api.CreateDatabase(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ResourceDatabaseID(region, deploymentID, name))

	return resourceDatabaseRead(ctx, d, meta)
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api := NewAPI(meta)

	region, deploymentID, name, err := ResourceDatabaseParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := api.ListDatabases(&datawarehouseapi.ListDatabasesRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Name:         scw.StringPtr(name),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	var found *datawarehouseapi.Database

	for _, db := range resp.Databases {
		if db.Name == name {
			found = db

			break
		}
	}

	if found == nil {
		d.SetId("")

		return nil
	}

	_ = d.Set("deployment_id", deploymentID)
	_ = d.Set("name", found.Name)
	_ = d.Set("size", int(found.Size))

	return nil
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api := NewAPI(meta)

	region, deploymentID, name, err := ResourceDatabaseParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteDatabase(&datawarehouseapi.DeleteDatabaseRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Name:         name,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func ResourceDatabaseID(region scw.Region, deploymentID, name string) string {
	return fmt.Sprintf("%s/%s/%s", region, deploymentID, name)
}

func ResourceDatabaseParseID(id string) (scw.Region, string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected region/deployment_id/name", id)
	}

	return scw.Region(parts[0]), parts[1], parts[2], nil
}
