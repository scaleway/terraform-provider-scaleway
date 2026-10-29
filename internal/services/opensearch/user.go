package opensearch

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	searchdbapi "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
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
			DiffSuppressFunc: dsf.Locality,
			Description:      "Deployment on which the user is created",
		},
		"name": {
			Type:        schema.TypeString,
			Description: "OpenSearch user name",
			Required:    true,
			ForceNew:    true,
		},
		"password": {
			Type:         schema.TypeString,
			Optional:     true,
			Sensitive:    true,
			Description:  "OpenSearch user password. Only one of `password` or `password_wo` should be specified.",
			ExactlyOneOf: []string{"password", "password_wo"},
		},
		"password_wo": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "OpenSearch user password in [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) mode. Only one of `password` or `password_wo` should be specified. `password_wo` will not be set in the Terraform state. To update the `password_wo`, you must also update the `password_wo_version`.",
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

	deployment, err := waitForDeployment(ctx, api, region, deploymentID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	var password string

	if p, exists := d.GetOk("password"); exists {
		password = p.(string)
	} else {
		password = d.GetRawConfig().GetAttr("password_wo").AsString()
	}

	createReq := &searchdbapi.CreateUserRequest{
		Region:       region,
		DeploymentID: deployment.ID,
		Username:     d.Get("name").(string),
		Password:     password,
	}

	var user *searchdbapi.User

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		currentUser, errCreateUser := api.CreateUser(createReq, scw.WithContext(ctx))
		if errCreateUser != nil {
			if httperrors.Is409(errCreateUser) {
				_, errWait := waitForDeployment(ctx, api, region, deployment.ID, d.Timeout(schema.TimeoutCreate))
				if errWait != nil {
					return retry.NonRetryableError(errWait)
				}

				return retry.RetryableError(errCreateUser)
			}

			return retry.NonRetryableError(errCreateUser)
		}

		user = currentUser

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if err := identity.SetMultiPartIdentity(d, map[string]string{
		"region":        region.String(),
		"deployment_id": locality.ExpandID(deploymentID),
		"name":          user.Username,
	}, "region", "deployment_id", "name"); err != nil {
		return diag.FromErr(err)
	}

	return ResourceUserRead(ctx, d, m)
}

func ResourceUserRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)

	region, deploymentID, userName, err := ResourceUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDeployment(ctx, api, region, deploymentID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	res, err := api.ListUsers(&searchdbapi.ListUsersRequest{
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

	var found *searchdbapi.User

	for _, u := range res.Users {
		if u.Username == userName {
			found = u

			break
		}
	}

	if found == nil {
		tflog.Warn(ctx, "couldn't find user with name: ["+userName+"]")
		d.SetId("")

		return nil
	}

	_ = d.Set("deployment_id", regional.NewIDString(region, deploymentID))
	_ = d.Set("name", found.Username)
	_ = d.Set("region", string(region))

	if err := identity.SetMultiPartIdentity(d, map[string]string{
		"region":        region.String(),
		"deployment_id": deploymentID,
		"name":          found.Username,
	}, "region", "deployment_id", "name"); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourceUserUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)

	region, deploymentID, userName, err := ResourceUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDeployment(ctx, api, region, deploymentID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &searchdbapi.UpdateUserRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Username:     userName,
	}

	hasChanged := false

	if password, ok := d.GetOk("password"); ok {
		if d.HasChange("password") {
			req.Password = types.ExpandStringPtr(password)
			hasChanged = true
		}
	} else if _, ok := d.GetOk("password_wo_version"); ok {
		if d.HasChange("password_wo_version") {
			req.Password = types.ExpandStringPtr(d.GetRawConfig().GetAttr("password_wo").AsString())
			hasChanged = true
		}
	}

	if hasChanged {
		_, err = api.UpdateUser(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceUserRead(ctx, d, m)
}

func ResourceUserDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)

	region, deploymentID, userName, err := ResourceUserParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDeployment(ctx, api, region, deploymentID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}

		return diag.FromErr(err)
	}

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		errDeleteUser := api.DeleteUser(&searchdbapi.DeleteUserRequest{
			Region:       region,
			DeploymentID: deploymentID,
			Username:     userName,
		}, scw.WithContext(ctx))
		if errDeleteUser != nil {
			if httperrors.Is409(errDeleteUser) {
				_, errWait := waitForDeployment(ctx, api, region, deploymentID, d.Timeout(schema.TimeoutDelete))
				if errWait != nil {
					return retry.NonRetryableError(errWait)
				}

				return retry.RetryableError(errDeleteUser)
			}

			return retry.NonRetryableError(errDeleteUser)
		}

		return nil
	})

	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}

// ResourceUserParseID extracts region, deployment ID and username from the resource identifier.
// The resource identifier format is "Region/DeploymentId/UserName".
func ResourceUserParseID(resourceID string) (region scw.Region, deploymentID string, userName string, err error) {
	idParts := strings.Split(resourceID, "/")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("can't parse user resource id: %s", resourceID)
	}

	return scw.Region(idParts[0]), idParts[1], idParts[2], nil
}
