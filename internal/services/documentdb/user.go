package documentdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	documentdb "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDocumentDBUserCreate,
		ReadContext:   resourceDocumentDBUserRead,
		UpdateContext: resourceDocumentDBUpdate,
		DeleteContext: resourceDocumentDBUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultDocumentDBInstanceTimeout),
			Read:    schema.DefaultTimeout(defaultDocumentDBInstanceTimeout),
			Update:  schema.DefaultTimeout(defaultDocumentDBInstanceTimeout),
			Delete:  schema.DefaultTimeout(defaultDocumentDBInstanceTimeout),
			Default: schema.DefaultTimeout(defaultDocumentDBInstanceTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.IsUUIDorUUIDWithLocality(),
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
			"region": regional.Schema(),
		},
		CustomizeDiff: cdf.LocalityCheck("instance_id"),
	}
}

func resourceDocumentDBUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := documentDBAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	// resource depends on the instance locality
	regionalID := d.Get("instance_id").(string)
	region, instanceID, err := regional.ParseID(regionalID)
	if err != nil {
		diag.FromErr(err)
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	createUserReq := &documentdb.CreateUserRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       d.Get("name").(string),
		Password:   d.Get("password").(string),
		IsAdmin:    d.Get("is_admin").(bool),
	}

	var user *documentdb.User
	//  wrapper around StateChangeConf that will just retry write on database
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		currentUser, errCreateUser := api.CreateUser(createUserReq, scw.WithContext(ctx))
		if errCreateUser != nil {
			if httperrors.Is409(errCreateUser) {
				_, errWait := waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutCreate))
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

	d.SetId(resourceDocumentDBUserID(region, locality.ExpandID(instanceID), user.Name))

	return resourceDocumentDBUserRead(ctx, d, m)
}

func resourceDocumentDBUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, _, err := documentDBAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	region, instanceID, userName, err := ResourceDocumentDBUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.ListUsers(&documentdb.ListUsersRequest{
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
		tflog.Warn(ctx, fmt.Sprintf("couldn'd find documentDB user with name: [%s]", userName))
		d.SetId("")
		return nil
	}

	user := res.Users[0]
	_ = d.Set("instance_id", regional.NewID(region, instanceID).String())
	_ = d.Set("name", user.Name)
	_ = d.Set("is_admin", user.IsAdmin)
	_ = d.Set("region", string(region))

	d.SetId(resourceDocumentDBUserID(region, instanceID, user.Name))

	return nil
}

func resourceDocumentDBUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, _, err := documentDBAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	// resource depends on the instance locality
	region, instanceID, userName, err := ResourceDocumentDBUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &documentdb.UpdateUserRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       userName,
	}

	if d.HasChange("password") {
		req.Password = types.ExpandStringPtr(d.Get("password"))
	}
	if d.HasChange("is_admin") {
		req.IsAdmin = scw.BoolPtr(d.Get("is_admin").(bool))
	}

	_, err = api.UpdateUser(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDocumentDBUserRead(ctx, d, m)
}

func resourceDocumentDBUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := documentDBAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	// resource depends on the instance locality
	region, instanceID, userName, err := ResourceDocumentDBUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		errDeleteUser := api.DeleteUser(&documentdb.DeleteUserRequest{
			Region:     region,
			InstanceID: instanceID,
			Name:       userName,
		}, scw.WithContext(ctx))
		if errDeleteUser != nil {
			if httperrors.Is409(errDeleteUser) {
				_, errWait := waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutDelete))
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

// Build the resource identifier
// The resource identifier format is "Region/InstanceId/UserName"
func resourceDocumentDBUserID(region scw.Region, instanceID string, userName string) (resourceID string) {
	return fmt.Sprintf("%s/%s/%s", region, instanceID, userName)
}

// ResourceDocumentDBUserParseID extracts instance ID and username from the resource identifier.
// The resource identifier format is "Region/InstanceId/UserName"
func ResourceDocumentDBUserParseID(resourceID string) (region scw.Region, instanceID string, userName string, err error) {
	idParts := strings.Split(resourceID, "/")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("can't parse user resource id: %s", resourceID)
	}
	return scw.Region(idParts[0]), idParts[1], idParts[2], nil
}
