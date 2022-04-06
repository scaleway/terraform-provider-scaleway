package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayRdbPrivilege() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayRdbPrivilegeCreate,
		ReadContext:   resourceScalewayRdbPrivilegeRead,
		DeleteContext: resourceScalewayRdbPrivilegeDelete,
		UpdateContext: resourceScalewayRdbPrivilegeUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultRdbInstanceTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validationUUIDWithLocality(),
				Description:  "Instance on which the database is created",
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
				Type:        schema.TypeString,
				Description: "Privilege",
				ValidateFunc: validation.StringInSlice([]string{
					rdb.PermissionReadonly.String(),
					rdb.PermissionReadwrite.String(),
					rdb.PermissionAll.String(),
					rdb.PermissionCustom.String(),
					rdb.PermissionNone.String(),
				}, false),
				Required: true,
			},
		},
	}
}

func resourceScalewayRdbPrivilegeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI := newRdbAPI(meta)

	region, instanceID, err := parseRegionalID(d.Get("instance_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	userName, _ := d.Get("user_name").(string)
	createReq := &rdb.SetPrivilegeRequest{
		Region:       region,
		InstanceID:   instanceID,
		DatabaseName: d.Get("database_name").(string),
		UserName:     userName,
		Permission:   rdb.Permission(d.Get("permission").(string)),
	}

	//  wrapper around StateChangeConf that will just retry  write on database
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, errSetPrivilege := rdbAPI.SetPrivilege(createReq, scw.WithContext(ctx))
		if errSetPrivilege != nil {
			if is409Error(errSetPrivilege) {
				_, errWait := waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutCreate))
				if errWait != nil {
					return resource.NonRetryableError(errWait)
				}
				return resource.RetryableError(errSetPrivilege)
			}
			return resource.NonRetryableError(errSetPrivilege)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, instanceID))
	return resourceScalewayRdbPrivilegeRead(ctx, d, meta)
}

func resourceScalewayRdbPrivilegeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI := newRdbAPI(meta)
	region, instanceID, err := parseRegionalID(d.Get("instance_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	dbName, _ := d.Get("database_name").(string)
	userName, _ := d.Get("user_name").(string)

	_, err = waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	listUsers, err := rdbAPI.ListUsers(&rdb.ListUsersRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       &userName,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if listUsers == nil || len(listUsers.Users) == 0 {
		d.SetId("")
		return nil
	}

	res, err := rdbAPI.ListPrivileges(&rdb.ListPrivilegesRequest{
		Region:       region,
		InstanceID:   instanceID,
		DatabaseName: &dbName,
		UserName:     &userName,
	}, scw.WithContext(ctx))

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if len(res.Privileges) == 0 {
		return diag.FromErr(fmt.Errorf("couldn't retrieve privileges for user[%s] on database [%s]", userName, dbName))
	}
	var privilege = res.Privileges[0]
	_ = d.Set("database_name", privilege.DatabaseName)
	_ = d.Set("user_name", privilege.UserName)
	_ = d.Set("permission", privilege.Permission)
	_ = d.Set("instance_id", newRegionalIDString(region, instanceID))

	return nil
}

func resourceScalewayRdbPrivilegeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI := newRdbAPI(meta)
	region, instanceID, err := parseRegionalID(d.Get("instance_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	userName, _ := d.Get("user_name").(string)
	listUsers, err := rdbAPI.ListUsers(&rdb.ListUsersRequest{
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

	if listUsers == nil || len(listUsers.Users) == 0 {
		d.SetId("")
		return nil
	}

	updateReq := &rdb.SetPrivilegeRequest{
		Region:       region,
		InstanceID:   instanceID,
		DatabaseName: d.Get("database_name").(string),
		UserName:     userName,
		Permission:   rdb.Permission(d.Get("permission").(string)),
	}

	//  wrapper around StateChangeConf that will just retry the database creation
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, errSet := rdbAPI.SetPrivilege(updateReq, scw.WithContext(ctx))
		if errSet != nil {
			if is409Error(errSet) {
				_, errWait := waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutUpdate))
				if errWait != nil {
					return resource.NonRetryableError(errWait)
				}
				return resource.RetryableError(errSet)
			}
			return resource.NonRetryableError(errSet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

//gocyclo:ignore
func resourceScalewayRdbPrivilegeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI := newRdbAPI(meta)
	region, instanceID, err := parseRegionalID(d.Get("instance_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("permission", rdb.PermissionNone)
	userName, _ := d.Get("user_name").(string)
	listUsers, err := rdbAPI.ListUsers(&rdb.ListUsersRequest{
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

	if listUsers != nil || len(listUsers.Users) == 0 {
		d.SetId("")
		return nil
	}

	updateReq := &rdb.SetPrivilegeRequest{
		Region:       region,
		InstanceID:   instanceID,
		DatabaseName: d.Get("database_name").(string),
		UserName:     userName,
		Permission:   rdb.PermissionNone,
	}

	//  wrapper around StateChangeConf that will just retry the database creation
	err = resource.RetryContext(ctx, defaultRdbInstanceTimeout, func() *resource.RetryError {
		// check if user exist on retry
		listUsers, errUserExist := rdbAPI.ListUsers(&rdb.ListUsersRequest{
			Region:     region,
			InstanceID: instanceID,
			Name:       &userName,
		}, scw.WithContext(ctx))
		if err != nil {
			if is404Error(err) {
				d.SetId("")
				return nil
			}
			return resource.NonRetryableError(errUserExist)
		}

		if listUsers != nil || len(listUsers.Users) == 0 {
			d.SetId("")
			return nil
		}
		_, errSet := rdbAPI.SetPrivilege(updateReq, scw.WithContext(ctx))
		if errSet != nil {
			if is409Error(errSet) {
				_, errWait := waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutDelete))
				if errWait != nil {
					return resource.NonRetryableError(errWait)
				}
				return resource.RetryableError(errSet)
			}
			return resource.NonRetryableError(errSet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
