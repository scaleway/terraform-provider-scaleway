package iam

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceAPIKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIamAPIKeyCreate,
		ReadContext:   resourceIamAPIKeyRead,
		UpdateContext: resourceIamAPIKeyUpdate,
		DeleteContext: resourceIamAPIKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    apiKeySchema,
		Identity: identity.WrapSchemaMap(map[string]*schema.Schema{
			"access_key": {
				Type:              schema.TypeString,
				Description:       "Access Key of your API Key",
				RequiredForImport: true,
			},
		}),
	}
}

func apiKeySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The description of the iam api key",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the iam api key",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the iam api key",
		},
		"expires_at": {
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			Description:      "The date and time of the expiration of the iam api key. Cannot be changed afterwards",
			ValidateDiagFunc: verify.IsDate(),
			DiffSuppressFunc: dsf.TimeRFC3339,
		},
		"access_key": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The access key of the iam api key",
		},
		"secret_key": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The secret Key of the iam api key",
			Sensitive:   true,
		},
		"application_id": {
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			Description:      "ID of the application attached to the api key",
			ConflictsWith:    []string{"user_id"},
			ValidateDiagFunc: verify.IsUUID(),
		},
		"user_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "ID of the user attached to the api key",
			ConflictsWith:    []string{"application_id"},
			ValidateDiagFunc: verify.IsUUID(),
		},
		"editable": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether or not the iam api key is editable",
		},
		"creation_ip": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The IPv4 Address of the device which created the API key",
		},
		"default_project_id": account.ProjectIDSchema(),
	}
}

func resourceIamAPIKeyCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)

	res, err := api.CreateAPIKey(&iam.CreateAPIKeyRequest{
		ApplicationID:    types.ExpandStringPtr(d.Get("application_id")),
		UserID:           types.ExpandStringPtr(d.Get("user_id")),
		ExpiresAt:        types.ExpandTimePtr(d.Get("expires_at")),
		DefaultProjectID: types.ExpandStringPtr(d.Get("default_project_id")),
		Description:      d.Get("description").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("secret_key", res.SecretKey)

	err = identity.SetFlatIdentity(d, res.AccessKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceIamAPIKeyRead(ctx, d, m)
}

func resourceIamAPIKeyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)

	res, err := api.GetAPIKey(&iam.GetAPIKeyRequest{
		AccessKey: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("description", res.Description)
	_ = d.Set("created_at", types.FlattenTime(res.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(res.UpdatedAt))
	_ = d.Set("expires_at", types.FlattenTime(res.ExpiresAt))
	_ = d.Set("access_key", res.AccessKey)

	if res.ApplicationID != nil {
		_ = d.Set("application_id", res.ApplicationID)
	}

	if res.UserID != nil {
		_ = d.Set("user_id", res.UserID)
	}

	_ = d.Set("editable", res.Editable)
	_ = d.Set("creation_ip", res.CreationIP)
	_ = d.Set("default_project_id", res.DefaultProjectID)

	return nil
}

func resourceIamAPIKeyUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)

	req := &iam.UpdateAPIKeyRequest{
		AccessKey: d.Id(),
	}

	hasChanged := false

	if d.HasChange("description") {
		req.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
		hasChanged = true
	}

	if d.HasChange("default_project_id") {
		req.DefaultProjectID = types.ExpandStringPtr(d.Get("default_project_id"))
		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateAPIKey(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceIamAPIKeyRead(ctx, d, m)
}

func resourceIamAPIKeyDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)

	err := api.DeleteAPIKey(&iam.DeleteAPIKeyRequest{
		AccessKey: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
