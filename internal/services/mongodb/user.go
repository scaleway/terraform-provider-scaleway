package mongodb

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
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
			Create:  schema.DefaultTimeout(defaultMongodbInstanceTimeout),
			Read:    schema.DefaultTimeout(defaultMongodbInstanceTimeout),
			Update:  schema.DefaultTimeout(defaultMongodbInstanceTimeout),
			Delete:  schema.DefaultTimeout(defaultMongodbInstanceTimeout),
			Default: schema.DefaultTimeout(defaultMongodbInstanceTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    userSchema,
		CustomizeDiff: customdiff.All(
			cdf.LocalityCheck("instance_id"),
			func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
				if rolesRaw, ok := diff.GetOk("roles"); ok {
					roles := rolesRaw.(*schema.Set).List()
					for _, roleRaw := range roles {
						role := roleRaw.(map[string]any)
						databaseName := role["database_name"].(string)
						anyDatabase := role["any_database"].(bool)

						if databaseName != "" && anyDatabase {
							return errors.New("database_name and any_database are mutually exclusive")
						}

						if databaseName == "" && !anyDatabase {
							return errors.New("either database_name or any_database must be specified")
						}
					}
				}

				return nil
			},
		),
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
			Description: "MongoDB user name",
			Required:    true,
			ForceNew:    true,
		},
		"password": {
			Type:         schema.TypeString,
			Optional:     true,
			Sensitive:    true,
			Description:  "MongoDB user password. Only one of `password` or `password_wo` should be specified.",
			ExactlyOneOf: []string{"password", "password_wo"},
		},
		"password_wo": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "MongoDB user password in [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) mode. Only one of `password` or `password_wo` should be specified. `password_wo` will not be set in the Terraform state. To update the `password_wo`, you must also update the `password_wo_version`.",
			WriteOnly:    true,
			ExactlyOneOf: []string{"password", "password_wo"},
			RequiredWith: []string{
				"password_wo_version",
			},
		},
		"password_wo_version": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The version of the [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) password. To update the `password_wo`, you must also update the `password_wo_version`.",
			RequiredWith: []string{
				"password_wo",
			},
		},
		"roles": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "List of roles assigned to the user",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"role": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Role name (read, read_write, db_admin, sync)",
					},
					"database_name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Database name for the role",
					},
					"any_database": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Apply role to any database",
					},
				},
			},
		},
		// Common
		"region": regional.Schema(),
	}
}

func ResourceUserCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	mongodbAPI := newAPI(m)
	regionalID := d.Get("instance_id").(string)

	region, instanceID, err := regional.ParseID(regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	instance, err := waitForInstance(ctx, mongodbAPI, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	var password string
	if _, ok := d.GetOk("password_wo_version"); ok {
		password = d.GetRawConfig().GetAttr("password_wo").AsString()
	} else {
		// If `password` is not set, it will be set as the default empty string
		password = d.Get("password").(string)
	}

	createReq := &mongodb.CreateUserRequest{
		Region:     region,
		InstanceID: instance.ID,
		Name:       d.Get("name").(string),
		Password:   password,
	}

	user, err := mongodbAPI.CreateUser(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ResourceUserID(region, instanceID, user.Name))

	// Set user roles if provided
	if rolesSet, ok := d.GetOk("roles"); ok && rolesSet.(*schema.Set).Len() > 0 {
		roles := expandUserRoles(rolesSet.(*schema.Set))
		setRoleReq := &mongodb.SetUserRoleRequest{
			Region:     region,
			InstanceID: instanceID,
			UserName:   user.Name,
			Roles:      roles,
		}

		_, err = mongodbAPI.SetUserRole(setRoleReq, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceUserRead(ctx, d, m)
}

func ResourceUserRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	mongodbAPI := newAPI(m)

	region, instanceID, userName, err := ResourceUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForInstance(ctx, mongodbAPI, region, instanceID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	res, err := mongodbAPI.ListUsers(&mongodb.ListUsersRequest{
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
	_ = d.Set("region", string(region))

	if _, ok := d.GetOk("password_wo_version"); !ok {
		_ = d.Set("password", d.Get("password"))
	}

	// Set user roles
	if len(user.Roles) > 0 {
		_ = d.Set("roles", flattenUserRoles(user.Roles))
	}

	d.SetId(ResourceUserID(region, instanceID, user.Name))

	return nil
}

func ResourceUserUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	mongodbAPI := newAPI(m)

	region, instanceID, userName, err := ResourceUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForInstance(ctx, mongodbAPI, region, instanceID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	if password, ok := d.GetOk("password"); ok {
		if d.HasChange("password") {
			req := &mongodb.UpdateUserRequest{
				Region:     region,
				InstanceID: instanceID,
				Name:       userName,
				Password:   types.ExpandStringPtr(password.(string)),
			}

			_, err = mongodbAPI.UpdateUser(req, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	} else if _, ok := d.GetOk("password_wo_version"); ok {
		if d.HasChange("password_wo_version") {
			req := &mongodb.UpdateUserRequest{
				Region:     region,
				InstanceID: instanceID,
				Name:       userName,
				Password:   types.ExpandStringPtr(d.GetRawConfig().GetAttr("password_wo").AsString()),
			}

			_, err = mongodbAPI.UpdateUser(req, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Handle roles changes
	if d.HasChange("roles") {
		rolesSet := d.Get("roles").(*schema.Set)
		roles := expandUserRoles(rolesSet)

		setRoleReq := &mongodb.SetUserRoleRequest{
			Region:     region,
			InstanceID: instanceID,
			UserName:   userName,
			Roles:      roles,
		}

		_, err = mongodbAPI.SetUserRole(setRoleReq, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceUserRead(ctx, d, m)
}

func ResourceUserDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	mongodbAPI := newAPI(m)

	region, instanceID, userName, err := ResourceUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForInstance(ctx, mongodbAPI, region, instanceID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	err = mongodbAPI.DeleteUser(&mongodb.DeleteUserRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       userName,
	}, scw.WithContext(ctx))

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
