package rdb

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/user.md
var userDescription string

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		Description:   userDescription,
		CreateContext: ResourceUserCreate,
		ReadContext:   ResourceUserRead,
		UpdateContext: ResourceUserUpdate,
		DeleteContext: ResourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultInstanceTimeout),
			Read:    schema.DefaultTimeout(defaultInstanceTimeout),
			Update:  schema.DefaultTimeout(defaultInstanceTimeout),
			Delete:  schema.DefaultTimeout(defaultInstanceTimeout),
			Default: schema.DefaultTimeout(defaultInstanceTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    userSchema,
		CustomizeDiff: cdf.LocalityCheck("instance_id"),
	}
}

func userSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"instance_id": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			Description:      "Instance on which the user is created",
		},
		"name": {
			Type:        schema.TypeString,
			Description: "Database user name",
			Required:    true,
			ForceNew:    true,
		},
		"password": {
			Type:         schema.TypeString,
			Optional:     true,
			Sensitive:    true,
			Description:  "Database user password",
			ExactlyOneOf: []string{"password", "password_wo"},
		},
		"password_wo": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Database user password in [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) mode. Only one of `password` or `password_wo` should be specified. `password_wo` will not be set in the Terraform state. To update the `password_wo`, you must also update the `password_wo_version`.",
			WriteOnly:    true,
			ExactlyOneOf: []string{"password", "password_wo"},
			RequiredWith: []string{"password_wo_version"},
		},
		"password_wo_version": {
			Type:         schema.TypeInt,
			Optional:     true,
			Description:  "The version of the [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) password. To update the `password_wo`, you must also update the `password_wo_version`.",
			RequiredWith: []string{"password_wo"},
		},
		"is_admin": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Grant admin permissions to database user",
		},
		// Common
		"region": regional.Schema(),
	}
}

func ResourceUserCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	rdbAPI := newAPI(m)
	// resource depends on the instance locality
	regionalID := d.Get("instance_id").(string)

	region, instanceID, err := regional.ParseID(regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	ins, err := waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	var password string

	if p, exists := d.GetOk("password"); exists {
		password = p.(string)
	} else {
		password = d.GetRawConfig().GetAttr("password_wo").AsString()
	}

	createReq := &rdb.CreateUserRequest{
		Region:     region,
		InstanceID: ins.ID,
		Name:       d.Get("name").(string),
		Password:   password,
		IsAdmin:    d.Get("is_admin").(bool),
	}

	var user *rdb.User
	//  wrapper around StateChangeConf that will just retry write on database
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		currentUser, errCreateUser := rdbAPI.CreateUser(createReq, scw.WithContext(ctx))
		if errCreateUser != nil {
			if httperrors.Is409(errCreateUser) {
				_, errWait := waitForRDBInstance(ctx, rdbAPI, region, ins.ID, d.Timeout(schema.TimeoutCreate))
				if errWait != nil {
					return retry.NonRetryableError(errWait)
				}

				return retry.RetryableError(errCreateUser)
			}

			return retry.NonRetryableError(errCreateUser)
		}
		// set database information
		user = currentUser

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ResourceUserID(region, locality.ExpandID(instanceID), user.Name))

	return ResourceUserRead(ctx, d, m)
}

func ResourceUserRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	rdbAPI := newAPI(m)

	region, instanceID, userName, err := ResourceUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	res, err := rdbAPI.ListUsers(&rdb.ListUsersRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       &userName,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if len(res.Users) == 0 {
		tflog.Warn(ctx, fmt.Sprintf("couldn't find user with name: [%s]", userName))
		d.SetId("")

		return nil
	}

	user := res.Users[0]
	_ = d.Set("instance_id", regional.NewID(region, instanceID).String())
	_ = d.Set("name", user.Name)
	_ = d.Set("is_admin", user.IsAdmin)
	_ = d.Set("region", string(region))

	d.SetId(ResourceUserID(region, instanceID, user.Name))

	return nil
}

func ResourceUserUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	rdbAPI := newAPI(m)
	// resource depends on the instance locality
	region, instanceID, userName, err := ResourceUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &rdb.UpdateUserRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       userName,
	}

	if password, ok := d.GetOk("password"); ok {
		if d.HasChange("password") {
			// Check password field is being set (not just removed)
			req.Password = types.ExpandStringPtr(password)
		}
	} else if _, ok := d.GetOk("password_wo_version"); ok {
		if d.HasChange("password_wo_version") {
			req.Password = types.ExpandStringPtr(d.GetRawConfig().GetAttr("password_wo").AsString())
		}
	}

	if d.HasChange("is_admin") {
		req.IsAdmin = new(d.Get("is_admin").(bool))
	}

	_, err = rdbAPI.UpdateUser(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceUserRead(ctx, d, m)
}

func ResourceUserDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	rdbAPI := newAPI(m)
	// resource depends on the instance locality
	region, instanceID, userName, err := ResourceUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		errDeleteUser := rdbAPI.DeleteUser(&rdb.DeleteUserRequest{
			Region:     region,
			InstanceID: instanceID,
			Name:       userName,
		}, scw.WithContext(ctx))
		if errDeleteUser != nil {
			if httperrors.Is409(errDeleteUser) {
				_, errWait := waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutDelete))
				if errWait != nil {
					return retry.NonRetryableError(errWait)
				}

				return retry.RetryableError(errDeleteUser)
			}

			return retry.NonRetryableError(errDeleteUser)
		}
		// set database information
		return nil
	})

	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}

// ResourceUserID builds the resource identifier
// The resource identifier format is "Region/InstanceId/UserName"
func ResourceUserID(region scw.Region, instanceID string, userName string) (resourceID string) {
	return fmt.Sprintf("%s/%s/%s", region, instanceID, userName)
}

// ResourceUserParseID extracts instance ID and username from the resource identifier.
// The resource identifier format is "Region/InstanceId/UserName"
func ResourceUserParseID(resourceID string) (region scw.Region, instanceID string, userName string, err error) {
	idParts := strings.Split(resourceID, "/")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("can't parse user resource id: %s", resourceID)
	}

	return scw.Region(idParts[0]), idParts[1], idParts[2], nil
}
