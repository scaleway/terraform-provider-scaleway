package mnq

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceNatsAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceMNQNatsAccountCreate,
		ReadContext:   ResourceMNQNatsAccountRead,
		UpdateContext: ResourceMNQNatsAccountUpdate,
		DeleteContext: ResourceMNQNatsAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The nats account name",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint for interact with Nats",
			},
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),
		},
	}
}

func ResourceMNQNatsAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := newMNQNatsAPI(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	account, err := api.CreateNatsAccount(&mnq.NatsAPICreateNatsAccountRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
		Name:      types.ExpandOrGenerateString(d.Get("name").(string), "nats-account"),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, account.ID))

	return ResourceMNQNatsAccountRead(ctx, d, m)
}

func ResourceMNQNatsAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewNatsAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	account, err := api.GetNatsAccount(&mnq.NatsAPIGetNatsAccountRequest{
		Region:        region,
		NatsAccountID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", account.Name)
	_ = d.Set("region", account.Region)
	_ = d.Set("project_id", account.ProjectID)
	_ = d.Set("endpoint", account.Endpoint)

	return nil
}

func ResourceMNQNatsAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewNatsAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &mnq.NatsAPIUpdateNatsAccountRequest{
		Region:        region,
		NatsAccountID: id,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
	}

	if _, err := api.UpdateNatsAccount(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return ResourceMNQNatsAccountRead(ctx, d, m)
}

func ResourceMNQNatsAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewNatsAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteNatsAccount(&mnq.NatsAPIDeleteNatsAccountRequest{
		Region:        region,
		NatsAccountID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
