package documentdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	documentdb "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourcePrivilege() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDocumentDBPrivilegeCreate,
		ReadContext:   resourceDocumentDBPrivilegeRead,
		DeleteContext: resourceDocumentDBPrivilegeDelete,
		UpdateContext: resourceDocumentDBPrivilegeUpdate,
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
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				Description:      "Instance on which the database is created",
			},
			"user_name": {
				Type:        schema.TypeString,
				Description: "User name",
				Required:    true,
			},
			"database_name": {
				Type:        schema.TypeString,
				Description: "Database name",
				Required:    true,
			},
			"permission": {
				Type:             schema.TypeString,
				Description:      "Privilege",
				ValidateDiagFunc: verify.ValidateEnum[documentdb.Permission](),
				Required:         true,
			},
			// Common
			"region": regional.Schema(),
		},
		CustomizeDiff: cdf.LocalityCheck("instance_id"),
	}
}

func resourceDocumentDBPrivilegeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := documentDBAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID := locality.ExpandID(d.Get("instance_id").(string))
	_, err = waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	userName, _ := d.Get("user_name").(string)
	databaseName, _ := d.Get("database_name").(string)
	createReq := &documentdb.SetPrivilegeRequest{
		Region:       region,
		InstanceID:   instanceID,
		DatabaseName: databaseName,
		UserName:     userName,
		Permission:   documentdb.Permission(d.Get("permission").(string)),
	}

	//  wrapper around StateChangeConf that will just retry  write on database
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		_, errSetPrivilege := api.SetPrivilege(createReq, scw.WithContext(ctx))
		if errSetPrivilege != nil {
			if httperrors.Is409(errSetPrivilege) {
				_, errWait := waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutCreate))
				if errWait != nil {
					return retry.NonRetryableError(errWait)
				}
				return retry.RetryableError(errSetPrivilege)
			}
			return retry.NonRetryableError(errSetPrivilege)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resourceDocumentDBUserPrivilegeID(region, locality.ExpandID(instanceID), databaseName, userName))

	return resourceDocumentDBPrivilegeRead(ctx, d, m)
}

func resourceDocumentDBPrivilegeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, _, err := documentDBAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	region, instanceID, databaseName, userName, err := resourceDocumentDBUserPrivilegeParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	listUsers, err := api.ListUsers(&documentdb.ListUsersRequest{
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

	if listUsers == nil || len(listUsers.Users) == 0 {
		d.SetId("")
		return nil
	}

	res, err := api.ListPrivileges(&documentdb.ListPrivilegesRequest{
		Region:       region,
		InstanceID:   instanceID,
		DatabaseName: &databaseName,
		UserName:     &userName,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if len(res.Privileges) == 0 {
		return diag.FromErr(fmt.Errorf("couldn't retrieve privileges for user[%s] on database [%s]", userName, databaseName))
	}
	privilege := res.Privileges[0]
	_ = d.Set("database_name", privilege.DatabaseName)
	_ = d.Set("user_name", privilege.UserName)
	_ = d.Set("permission", privilege.Permission)
	_ = d.Set("instance_id", regional.NewIDString(region, instanceID))
	_ = d.Set("region", region)

	return nil
}

func resourceDocumentDBPrivilegeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := documentDBAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	region, instanceID, databaseName, userName, err := resourceDocumentDBUserPrivilegeParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	listUsers, err := api.ListUsers(&documentdb.ListUsersRequest{
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

	if listUsers == nil || len(listUsers.Users) == 0 {
		d.SetId("")
		return nil
	}

	updateReq := &documentdb.SetPrivilegeRequest{
		Region:       region,
		InstanceID:   instanceID,
		DatabaseName: databaseName,
		UserName:     userName,
		Permission:   documentdb.Permission(d.Get("permission").(string)),
	}

	//  wrapper around StateChangeConf that will just retry the database creation
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *retry.RetryError {
		_, errSet := api.SetPrivilege(updateReq, scw.WithContext(ctx))
		if errSet != nil {
			if httperrors.Is409(errSet) {
				_, errWait := waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutUpdate))
				if errWait != nil {
					return retry.NonRetryableError(errWait)
				}
				return retry.RetryableError(errSet)
			}
			return retry.NonRetryableError(errSet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

//gocyclo:ignore
func resourceDocumentDBPrivilegeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := documentDBAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	region, instanceID, databaseName, userName, err := resourceDocumentDBUserPrivilegeParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("permission", documentdb.PermissionNone)
	listUsers, err := api.ListUsers(&documentdb.ListUsersRequest{
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

	if listUsers != nil && len(listUsers.Users) == 0 {
		d.SetId("")
		return nil
	}

	updateReq := &documentdb.SetPrivilegeRequest{
		Region:       region,
		InstanceID:   instanceID,
		DatabaseName: databaseName,
		UserName:     userName,
		Permission:   documentdb.PermissionNone,
	}

	//  wrapper around StateChangeConf that will just retry the database creation
	err = retry.RetryContext(ctx, defaultDocumentDBInstanceTimeout, func() *retry.RetryError {
		// check if user exist on retry
		listUsers, errUserExist := api.ListUsers(&documentdb.ListUsersRequest{
			Region:     region,
			InstanceID: instanceID,
			Name:       &userName,
		}, scw.WithContext(ctx))
		if err != nil {
			if httperrors.Is404(err) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(errUserExist)
		}

		if listUsers != nil && len(listUsers.Users) == 0 {
			d.SetId("")
			return nil
		}
		_, errSet := api.SetPrivilege(updateReq, scw.WithContext(ctx))
		if errSet != nil {
			if httperrors.Is409(errSet) {
				_, errWait := waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutDelete))
				if errWait != nil {
					return retry.NonRetryableError(errWait)
				}
				return retry.RetryableError(errSet)
			}
			return retry.NonRetryableError(errSet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// Build the resource identifier
// The resource identifier format is "Region/InstanceId/database/UserName"
func resourceDocumentDBUserPrivilegeID(region scw.Region, instanceID, database, userName string) (resourceID string) {
	return fmt.Sprintf("%s/%s/%s/%s", region, instanceID, database, userName)
}

// resourceDocumentDBUserPrivilegeParseID: The resource identifier format is "Region/InstanceId/DatabaseName/UserName"
func resourceDocumentDBUserPrivilegeParseID(resourceID string) (region scw.Region, instanceID, databaseName, userName string, err error) {
	idParts := strings.Split(resourceID, "/")
	if len(idParts) != 4 {
		return "", "", "", "", fmt.Errorf("can't parse user privilege resource id: %s", resourceID)
	}
	return scw.Region(idParts[0]), idParts[1], idParts[2], idParts[3], nil
}
