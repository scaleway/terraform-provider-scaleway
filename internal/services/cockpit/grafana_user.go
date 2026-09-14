package cockpit

import (
	"context"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceCockpitGrafanaUser() *schema.Resource {
	return &schema.Resource{
		CreateContext:      ResourceCockpitGrafanaUserCreate,
		ReadContext:        ResourceCockpitGrafanaUserRead,
		DeleteContext:      ResourceCockpitGrafanaUserDelete,
		DeprecationMessage: "This resource is deprecated and will be removed on January 1st, 2026. Grafana authentication is now managed through Scaleway IAM (Identity and Access Management). Use the 'scaleway_cockpit_grafana' data source to retrieve Grafana connection details. See https://www.scaleway.com/en/docs/observability/cockpit/",
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(DefaultCockpitTimeout),
			Read:    schema.DefaultTimeout(DefaultCockpitTimeout),
			Delete:  schema.DefaultTimeout(DefaultCockpitTimeout),
			Default: schema.DefaultTimeout(DefaultCockpitTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: cockpitGrafanaUserSchema,
		Identity:   cockpitGrafanaUserIdentity(),
	}
}

func cockpitGrafanaUserIdentity() *schema.ResourceIdentity {
	return identity.WrapSchemaMap(map[string]*schema.Schema{
		"project_id": identity.DefaultProjectIDAttribute(),
		"id": {
			Type:              schema.TypeString,
			Description:       "The ID of the Grafana user",
			RequiredForImport: true,
		},
	})
}

func cockpitGrafanaUserSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"login": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			Description:  "The login of the Grafana user",
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_-]{2,24}$`), "must have between 2 and 24 alphanumeric characters"),
		},
		"password": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The password of the Grafana user",
			Sensitive:   true,
		},
		"role": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			Description:      "The role of the Grafana user",
			ValidateDiagFunc: verify.ValidateEnum[cockpit.GrafanaUserRole](),
		},
		"grafana_url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The grafana URL",
		},
		"project_id": account.ProjectIDSchema(),
	}
}

func ResourceCockpitGrafanaUserCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, err := NewGlobalAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	login := d.Get("login").(string)
	role := cockpit.GrafanaUserRole(d.Get("role").(string))

	grafanaUser, err := retryOn403Value(ctx, func() (*cockpit.GrafanaUser, error) {
		return api.CreateGrafanaUser(&cockpit.GlobalAPICreateGrafanaUserRequest{ //nolint:staticcheck // legacy Grafana user resource uses deprecated API
			ProjectID: projectID,
			Login:     login,
			Role:      role,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("password", grafanaUser.Password)
	if err := setCockpitGrafanaUserIdentity(d, projectID, strconv.FormatUint(uint64(grafanaUser.ID), 10)); err != nil {
		return diag.FromErr(err)
	}

	return ResourceCockpitGrafanaUserRead(ctx, d, m)
}

func ResourceCockpitGrafanaUserRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, projectID, grafanaUserID, err := NewAPIGrafanaUserID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	grafanaUser, err := findGrafanaUser(ctx, api, projectID, grafanaUserID)
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if grafanaUser == nil {
		d.SetId("")

		return nil
	}

	grafana, err := retryOn403Value(ctx, func() (*cockpit.Grafana, error) {
		return api.GetGrafana(&cockpit.GlobalAPIGetGrafanaRequest{
			ProjectID: projectID,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("login", grafanaUser.Login)
	_ = d.Set("role", grafanaUser.Role)
	_ = d.Set("project_id", projectID)
	_ = d.Set("grafana_url", grafana.GrafanaURL)

	if err := setCockpitGrafanaUserIdentity(d, projectID, strconv.FormatUint(uint64(grafanaUser.ID), 10)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourceCockpitGrafanaUserDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, projectID, grafanaUserID, err := NewAPIGrafanaUserID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = retryOn403(ctx, func() error {
		return api.DeleteGrafanaUser(&cockpit.GlobalAPIDeleteGrafanaUserRequest{ //nolint:staticcheck // legacy Grafana user resource uses deprecated API
			ProjectID:     projectID,
			GrafanaUserID: grafanaUserID,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}

// findGrafanaUser waits until the Grafana user appears in ListGrafanaUsers. A freshly created user
// can be missing from the list while permissions and cockpit metadata propagate; clearing state in
// that window makes Terraform report an inconsistent result after apply.
func findGrafanaUser(ctx context.Context, api *cockpit.GlobalAPI, projectID string, grafanaUserID uint32) (*cockpit.GrafanaUser, error) {
	wait := transport.RetryOn403WaitTime
	if transport.DefaultWaitRetryInterval != nil {
		wait = *transport.DefaultWaitRetryInterval
	}

	deadline := time.Now().Add(transport.IAMPropagationTimeout)

	for {
		res, err := retryOn403Value(ctx, func() (*cockpit.ListGrafanaUsersResponse, error) {
			return api.ListGrafanaUsers(&cockpit.GlobalAPIListGrafanaUsersRequest{ //nolint:staticcheck // legacy Grafana user resource uses deprecated API
				ProjectID: projectID,
			}, scw.WithContext(ctx), scw.WithAllPages())
		})
		if err != nil {
			return nil, err
		}

		for _, user := range res.GrafanaUsers {
			if user.ID == grafanaUserID {
				return user, nil
			}
		}

		if time.Now().After(deadline) {
			return nil, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
}

func setCockpitGrafanaUserIdentity(d *schema.ResourceData, projectID, grafanaUserID string) error {
	resourceIdentity, err := d.Identity()
	if err != nil {
		return err
	}

	if err := resourceIdentity.Set("project_id", projectID); err != nil {
		return err
	}

	if err := resourceIdentity.Set("id", grafanaUserID); err != nil {
		return err
	}

	d.SetId(cockpitIDWithProjectID(projectID, grafanaUserID))

	return nil
}
