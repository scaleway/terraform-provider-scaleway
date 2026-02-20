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

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: userSchema,
	}
}

func userSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"region": regional.Schema(),
		"deployment_id": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			Description:      "ID of the Datawarehouse deployment to which this user belongs.",
			DiffSuppressFunc: dsf.Locality,
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Name of the ClickHouse user.",
		},
		"password": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "Password for the ClickHouse user.",
		},
		"is_admin": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether the user has administrator privileges.",
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := datawarehouseAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	deploymentID := locality.ExpandID(d.Get("deployment_id").(string))
	name := d.Get("name").(string)
	password := d.Get("password").(string)
	isAdmin := d.Get("is_admin").(bool)

	req := &datawarehouseapi.CreateUserRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Name:         name,
		Password:     password,
		IsAdmin:      isAdmin,
	}

	_, err = api.CreateUser(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ResourceUserID(region, deploymentID, name))

	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api := NewAPI(meta)

	region, deploymentID, userName, err := ResourceUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := api.ListUsers(&datawarehouseapi.ListUsersRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Name:         &userName,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	var found *datawarehouseapi.User

	for _, u := range resp.Users {
		if u.Name == userName {
			found = u

			break
		}
	}

	if found == nil {
		d.SetId("")

		return nil
	}

	_ = d.Set("deployment_id", deploymentID)
	_ = d.Set("name", found.Name)
	_ = d.Set("is_admin", found.IsAdmin)

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api := NewAPI(meta)

	region, deploymentID, userName, err := ResourceUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	// Build the update request only for changed fields
	req := &datawarehouseapi.UpdateUserRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Name:         userName,
	}
	changed := false

	if d.HasChange("password") {
		req.Password = new(d.Get("password").(string))
		changed = true
	}

	if d.HasChange("is_admin") {
		req.IsAdmin = new(d.Get("is_admin").(bool))
		changed = true
	}

	if changed {
		_, err := api.UpdateUser(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api := NewAPI(meta)

	region, deploymentID, userName, err := ResourceUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteUser(&datawarehouseapi.DeleteUserRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Name:         userName,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func ResourceUserID(region scw.Region, deploymentID string, userName string) (resourceID string) {
	return fmt.Sprintf("%s/%s/%s", region, deploymentID, userName)
}

func ResourceUserParseID(resourceID string) (region scw.Region, deploymentID string, userName string, err error) {
	idParts := strings.Split(resourceID, "/")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("can't parse user resource id: %s", resourceID)
	}

	return scw.Region(idParts[0]), idParts[1], idParts[2], nil
}
