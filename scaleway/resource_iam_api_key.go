package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"golang.org/x/net/context"
)

func resourceScalewayIamApiKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayIamApiKeyCreate,
		ReadContext:   resourceScalewayIamApiKeyRead,
		UpdateContext: resourceScalewayIamApiKeyUpdate,
		DeleteContext: resourceScalewayIamApiKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
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
			"expired_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the expiration of the iam api key",
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
			},
			"application_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "ID of the application attached to the api key",
				ConflictsWith: []string{"user_id"},
				ValidateFunc:  validationUUID(),
			},
			"user_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "ID of the user attached to the api key",
				ConflictsWith: []string{"application_id"},
				ValidateFunc:  validationUUID(),
			},
			"editable": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether or not the iam api key is editable",
			},
			"creation_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IP Address of the device which created the API key",
			},
			"default_project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayIamApiKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamAPI := iamAPI(meta)
	res, err := iamAPI.CreateAPIKey(&iam.CreateAPIKeyRequest{
		ApplicationID:    expandStringPtr(d.Get("application_id")),
		UserID:           expandStringPtr(d.Get("user_id")),
		ExpiresAt:        expandTimePtr(d.Get("expires_at")),
		DefaultProjectID: expandStringPtr(d.Get("project_id")),
		Description:      d.Get("description").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.AccessKey)

	return resourceScalewayIamApiKeyRead(ctx, d, meta)
}

func resourceScalewayIamApiKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)
	res, err := api.GetAPIKey(&iam.GetAPIKeyRequest{
		AccessKey: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	_ = d.Set("description", res.Description)
	_ = d.Set("created_at", flattenTime(res.CreatedAt))
	_ = d.Set("updated_at", flattenTime(res.UpdatedAt))
	_ = d.Set("expired_at", flattenTime(res.ExpiredAt))
	_ = d.Set("access_key", res.AccessKey)
	_ = d.Set("secret_key", res.SecretKey)

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

func resourceScalewayIamApiKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)

	req := &iam.UpdateAPIKeyRequest{
		AccessKey: d.Id(),
	}

	hasChanged := false

	if d.HasChange("description") {
		req.Description = expandStringPtr(d.Get("description"))
		hasChanged = true
	}

	if d.HasChange("default_project_id") {
		req.DefaultProjectID = expandStringPtr(d.Get("default_project_id"))
		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateAPIKey(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayIamApiKeyRead(ctx, d, meta)
}

func resourceScalewayIamApiKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)

	err := api.DeleteAPIKey(&iam.DeleteAPIKeyRequest{
		AccessKey: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
