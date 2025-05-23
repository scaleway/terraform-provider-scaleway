package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceToken() *schema.Resource {
	return &schema.Resource{
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
		CreateContext:                     ResourceCockpitTokenCreate,
		ReadContext:                       ResourceCockpitTokenRead,
		DeleteContext:                     ResourceCockpitTokenDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(DefaultCockpitTimeout),
			Read:    schema.DefaultTimeout(DefaultCockpitTimeout),
			Delete:  schema.DefaultTimeout(DefaultCockpitTimeout),
			Default: schema.DefaultTimeout(DefaultCockpitTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{Version: 0, Type: cockpitTokenUpgradeV1SchemaType(), Upgrade: cockpitTokenV1UpgradeFunc},
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the token",
			},
			"scopes": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "Endpoints",
				Elem:        resourceCockpitTokenScopes(),
			},
			"secret_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The secret key of the token",
				Sensitive:   true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the Cockpit Token (Format ISO 8601)",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the Cockpit Token (Format ISO 8601)",
			},
			"project_id": account.ProjectIDSchema(),
			"region":     regional.Schema(),
		},
	}
}

func resourceCockpitTokenScopes() *schema.Resource {
	return &schema.Resource{
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
		Schema: map[string]*schema.Schema{
			"query_metrics": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Query metrics",
			},
			"write_metrics": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
				Description: "Write metrics",
			},
			"setup_metrics_rules": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Setup metrics rules",
			},
			"query_logs": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Query logs",
			},
			"write_logs": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
				Description: "Write logs",
			},
			"setup_logs_rules": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Setup logs rules",
			},
			"setup_alerts": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Setup alerts",
			},
			"query_traces": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Query traces",
			},
			"write_traces": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Write traces",
			},
		},
	}
}

func ResourceCockpitTokenCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	name := d.Get("name").(string)
	rawScopes, scopesSet := d.GetOk("scopes")

	var scopes []cockpit.TokenScope

	if !scopesSet || len(rawScopes.([]interface{})) == 0 {
		schema := resourceCockpitTokenScopes().Schema
		for key, val := range schema {
			if defaultVal, ok := val.Default.(bool); ok && defaultVal {
				if scopeConst, found := scopeMapping[key]; found {
					scopes = append(scopes, scopeConst)
				}
			}
		}
	} else {
		scopes = expandCockpitTokenScopes(rawScopes)
	}

	logging.L.Debugf("Creating token %+v", scopes)

	res, err := api.CreateToken(&cockpit.RegionalAPICreateTokenRequest{
		Name:        name,
		TokenScopes: scopes,
		ProjectID:   projectID,
		Region:      region,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("secret_key", res.SecretKey)
	d.SetId(regional.NewIDString(region, res.ID))

	return ResourceCockpitTokenRead(ctx, d, m)
}

func ResourceCockpitTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.GetToken(&cockpit.RegionalAPIGetTokenRequest{
		Region:  region,
		TokenID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", res.Name)
	_ = d.Set("scopes", flattenCockpitTokenScopes(res.Scopes))
	_ = d.Set("created_at", types.FlattenTime(res.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(res.UpdatedAt))
	_ = d.Set("project_id", res.ProjectID)
	_ = d.Set("region", res.Region)

	return nil
}

func ResourceCockpitTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteToken(&cockpit.RegionalAPIDeleteTokenRequest{
		Region:  region,
		TokenID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
