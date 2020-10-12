package scaleway

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayRdbUserBeta() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayRdbUserBetaCreate,
		ReadContext:   resourceScalewayRdbUserBetaRead,
		UpdateContext: resourceScalewayRdbUserBetaUpdate,
		DeleteContext: resourceScalewayRdbUserBetaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "Instance on which the user is created",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Database user name",
				Required:    true,
				ForceNew:    true,
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Database user password",
			},
			"is_admin": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Grant admin permissions to database user",
			},
			// Common
			"region": regionSchema(),
		},
	}
}

func resourceScalewayRdbUserBetaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	instanceID := d.Get("instance_id").(string)
	createReq := &rdb.CreateUserRequest{
		Region:     region,
		InstanceID: expandID(instanceID),
		Name:       d.Get("name").(string),
		Password:   d.Get("password").(string),
		IsAdmin:    d.Get("is_admin").(bool),
	}

	res, err := rdbAPI.CreateUser(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resourceScalewayRdbUserBetaID(region, expandID(instanceID), res.Name))

	return resourceScalewayRdbUserBetaRead(ctx, d, m)
}

func resourceScalewayRdbUserBetaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID, userName, err := resourceScalewayRdbUserBetaParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	res, err := rdbAPI.ListUsers(&rdb.ListUsersRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       &userName,
	}, scw.WithContext(ctx))

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	var user = res.Users[0]
	_ = d.Set("instance_id", newRegionalID(region, instanceID))
	_ = d.Set("name", user.Name)
	_ = d.Set("is_admin", user.IsAdmin)

	d.SetId(resourceScalewayRdbUserBetaID(region, instanceID, user.Name))

	return nil
}

func resourceScalewayRdbUserBetaUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID, userName, err := resourceScalewayRdbUserBetaParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	req := &rdb.UpdateUserRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       userName,
	}

	if d.HasChange("password") {
		req.Password = expandStringPtr(d.Get("password"))
	}
	if d.HasChange("is_admin") {
		req.IsAdmin = scw.BoolPtr(d.Get("is_admin").(bool))
	}

	_, err = rdbAPI.UpdateUser(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayRdbUserBetaRead(ctx, d, m)
}

func resourceScalewayRdbUserBetaDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID, userName, err := resourceScalewayRdbUserBetaParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	err = rdbAPI.DeleteUser(&rdb.DeleteUserRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       userName,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}

// Build the resource identifier
// The resource identifier format is "Region/InstanceId/UserName"
func resourceScalewayRdbUserBetaID(region scw.Region, instanceID string, userName string) (resourceID string) {
	return fmt.Sprintf("%s/%s/%s", region, instanceID, userName)
}

// Extract instance ID and username from the resource identifier.
// The resource identifier format is "Region/InstanceId/UserName"
func resourceScalewayRdbUserBetaParseID(resourceID string) (instanceID string, userName string, err error) {
	idParts := strings.Split(resourceID, "/")
	if len(idParts) != 3 {
		return "", "", fmt.Errorf("can't parse user resource id: %s", resourceID)
	}
	return idParts[1], idParts[2], nil
}
