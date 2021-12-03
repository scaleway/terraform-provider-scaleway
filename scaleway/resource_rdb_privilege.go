package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
				ValidateFunc: validationUUIDorUUIDWithLocality(),
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

	_, err = waitInstance(ctx, rdbAPI, region, instanceID)
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &rdb.SetPrivilegeRequest{
		Region:       region,
		InstanceID:   instanceID,
		DatabaseName: d.Get("database_name").(string),
		UserName:     d.Get("user_name").(string),
		Permission:   rdb.Permission(d.Get("permission").(string)),
	}

	_, err = rdbAPI.SetPrivilege(createReq, scw.WithContext(ctx))
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

	_, err = waitInstance(ctx, rdbAPI, region, instanceID)
	if err != nil {
		return diag.FromErr(err)
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

	_, err = waitInstance(ctx, rdbAPI, region, instanceID)
	if err != nil {
		return diag.FromErr(err)
	}

	updateReq := &rdb.SetPrivilegeRequest{
		Region:       region,
		InstanceID:   instanceID,
		DatabaseName: d.Get("database_name").(string),
		UserName:     d.Get("user_name").(string),
		Permission:   rdb.Permission(d.Get("permission").(string)),
	}
	_, err = rdbAPI.SetPrivilege(updateReq, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	_, err = waitInstance(ctx, rdbAPI, region, instanceID)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceScalewayRdbPrivilegeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_ = d.Set("permission", rdb.PermissionNone)
	return resourceScalewayRdbPrivilegeUpdate(ctx, d, meta)
}
