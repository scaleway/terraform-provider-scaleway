package messageq

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
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
		Importer:      identity.CompositeRegionalImporter("region", "deployment_id", "name"),
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultDeploymentTimeout),
			Read:    schema.DefaultTimeout(defaultDeploymentTimeout),
			Update:  schema.DefaultTimeout(defaultDeploymentTimeout),
			Delete:  schema.DefaultTimeout(defaultDeploymentTimeout),
			Default: schema.DefaultTimeout(defaultDeploymentTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    userSchema,
		CustomizeDiff: cdf.LocalityCheck("deployment_id"),
		Identity:      identity.CompositeRegionalIdentity("deployment_id", "name"),
	}
}

func userSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"deployment_id": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			Description:      "Deployment on which the user is created",
		},
		"name": {
			Type:        schema.TypeString,
			Description: "MessageQ user name",
			Required:    true,
			ForceNew:    true,
		},
		"password": {
			Type:         schema.TypeString,
			Optional:     true,
			Sensitive:    true,
			Description:  "MessageQ user password. Only one of `password` or `password_wo` should be specified.",
			ExactlyOneOf: []string{"password", "password_wo"},
		},
		"password_wo": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "MessageQ user password in [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) mode. Only one of `password` or `password_wo` should be specified. `password_wo` will not be set in the Terraform state. To update the `password_wo`, you must also update the `password_wo_version`.",
			WriteOnly:    true,
			ExactlyOneOf: []string{"password", "password_wo"},
			RequiredWith: []string{"password_wo_version"},
		},
		"password_wo_version": {
			Type:         schema.TypeInt,
			Optional:     true,
			Description:  "The version of the [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) password. To update the `password_wo`, you must also update the `password_wo_version`.",
			RequiredWith: []string{"password_wo"},
		},
		"region": regional.Schema(),
	}
}

func ResourceUserCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)
	regionalID := d.Get("deployment_id").(string)

	region, deploymentID, err := regional.ParseID(regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDeployment(ctx, api, region, deploymentID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	var password string
	if _, ok := d.GetOk("password_wo_version"); ok {
		password = d.GetRawConfig().GetAttr("password_wo").AsString()
	} else {
		password = d.Get("password").(string)
	}

	user, err := api.CreateUser(&messageqapi.CreateUserRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Username:     d.Get("name").(string),
		Password:     password,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := identity.SetMultiPartIdentity(d, map[string]string{
		"region":        string(region),
		"deployment_id": deploymentID,
		"name":          user.Username,
	}, "region", "deployment_id", "name"); err != nil {
		return diag.FromErr(err)
	}

	return ResourceUserRead(ctx, d, m)
}

func ResourceUserRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)

	idParts := identity.ParseMultiPartID(d.Id(), "region", "deployment_id", "name")
	region := scw.Region(idParts["region"])
	deploymentID := idParts["deployment_id"]
	userName := idParts["name"]

	_, err := waitForDeployment(ctx, api, region, deploymentID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	res, err := api.ListUsers(&messageqapi.ListUsersRequest{
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

	if len(res.Users) == 0 {
		tflog.Warn(ctx, fmt.Sprintf("couldn't find user with name: [%s]", userName))
		d.SetId("")

		return nil
	}

	user := res.Users[0]

	if err := identity.SetMultiPartIdentity(d, map[string]string{
		"region":        string(region),
		"deployment_id": deploymentID,
		"name":          user.Username,
	}, "region", "deployment_id", "name"); err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("deployment_id", regional.NewID(region, deploymentID).String())
	_ = d.Set("name", user.Username)
	_ = d.Set("region", string(region))

	if _, ok := d.GetOk("password_wo_version"); !ok {
		_ = d.Set("password", d.Get("password"))
	}

	return nil
}

func ResourceUserUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)

	idParts := identity.ParseMultiPartID(d.Id(), "region", "deployment_id", "name")
	region := scw.Region(idParts["region"])
	deploymentID := idParts["deployment_id"]
	userName := idParts["name"]

	_, err := waitForDeployment(ctx, api, region, deploymentID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	if password, ok := d.GetOk("password"); ok {
		if d.HasChange("password") {
			_, err = api.UpdateUser(&messageqapi.UpdateUserRequest{
				Region:       region,
				DeploymentID: deploymentID,
				Username:     userName,
				Password:     types.ExpandStringPtr(password.(string)),
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	} else if _, ok := d.GetOk("password_wo_version"); ok {
		if d.HasChange("password_wo_version") {
			_, err = api.UpdateUser(&messageqapi.UpdateUserRequest{
				Region:       region,
				DeploymentID: deploymentID,
				Username:     userName,
				Password:     types.ExpandStringPtr(d.GetRawConfig().GetAttr("password_wo").AsString()),
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return ResourceUserRead(ctx, d, m)
}

func ResourceUserDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)

	idParts := identity.ParseMultiPartID(d.Id(), "region", "deployment_id", "name")
	region := scw.Region(idParts["region"])
	deploymentID := idParts["deployment_id"]
	userName := idParts["name"]

	_, err := waitForDeployment(ctx, api, region, deploymentID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}

		return diag.FromErr(err)
	}

	err = api.DeleteUser(&messageqapi.DeleteUserRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Username:     userName,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
