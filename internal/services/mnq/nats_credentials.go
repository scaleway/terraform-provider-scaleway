package mnq

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceNatsCredentials() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceMNQNatsCredentialsCreate,
		ReadContext:   ResourceMNQNatsCredentialsRead,
		DeleteContext: ResourceMNQNatsCredentialsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "ID of the nats account",
				DiffSuppressFunc: dsf.Locality,
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The nats credentials name",
			},
			"file": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The credentials file",
			},
			"region": regional.Schema(),
		},
	}
}

func ResourceMNQNatsCredentialsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := newMNQNatsAPI(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	credentials, err := api.CreateNatsCredentials(&mnq.NatsAPICreateNatsCredentialsRequest{
		Region:        region,
		NatsAccountID: locality.ExpandID(d.Get("account_id").(string)),
		Name:          types.ExpandOrGenerateString(d.Get("name").(string), "nats-credentials"),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("file", credentials.Credentials.Content)

	d.SetId(regional.NewIDString(region, credentials.ID))

	return ResourceMNQNatsCredentialsRead(ctx, d, m)
}

func ResourceMNQNatsCredentialsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewNatsAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	credentials, err := api.GetNatsCredentials(&mnq.NatsAPIGetNatsCredentialsRequest{
		Region:            region,
		NatsCredentialsID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	natsAccountId := regional.NewIDString(region, credentials.NatsAccountID)

	_ = d.Set("account_id", natsAccountId)
	_ = d.Set("name", credentials.Name)
	_ = d.Set("region", region)

	return nil
}

func ResourceMNQNatsCredentialsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewNatsAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteNatsCredentials(&mnq.NatsAPIDeleteNatsCredentialsRequest{
		Region:            region,
		NatsCredentialsID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
