package mnq

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const natsCredentialsCreateRetryTimeout = 60 * time.Second

func ResourceNatsCredentials() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceMNQNatsCredentialsCreate,
		ReadContext:   ResourceMNQNatsCredentialsRead,
		DeleteContext: ResourceMNQNatsCredentialsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    natsCredentialsSchema,
		Identity:      identity.DefaultRegional(),
	}
}

func natsCredentialsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
			ForceNew:    true,
			Description: "The nats credentials name",
		},
		"file": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The credentials file",
			Sensitive:   true,
		},
		"region": regional.Schema(),
	}
}

func ResourceMNQNatsCredentialsCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newMNQNatsAPI(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &mnq.NatsAPICreateNatsCredentialsRequest{
		Region:        region,
		NatsAccountID: locality.ExpandID(d.Get("account_id").(string)),
		Name:          types.ExpandOrGenerateString(d.Get("name").(string), "nats-credentials"),
	}

	var credentials *mnq.NatsCredentials

	err = retry.RetryContext(ctx, natsCredentialsCreateRetryTimeout, func() *retry.RetryError {
		credentials, err = api.CreateNatsCredentials(req, scw.WithContext(ctx))
		if err == nil {
			return nil
		}

		if isMNQNamespaceReadRetryableError(err) {
			return retry.RetryableError(err)
		}

		return retry.NonRetryableError(err)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("file", credentials.Credentials.Content)

	if err := identity.SetRegionalIdentity(d, region, credentials.ID); err != nil {
		return diag.FromErr(err)
	}

	return ResourceMNQNatsCredentialsRead(ctx, d, m)
}

func ResourceMNQNatsCredentialsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewNatsAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	credentials, err := api.GetNatsCredentials(&mnq.NatsAPIGetNatsCredentialsRequest{
		Region:            region,
		NatsCredentialsID: id,
	}, scw.WithContext(ctx))
	if err != nil && isMNQNamespaceReadRetryableError(err) {
		err = retryMNQNamespaceRead(ctx, func() error {
			credentials, err = api.GetNatsCredentials(&mnq.NatsAPIGetNatsCredentialsRequest{
				Region:            region,
				NatsCredentialsID: id,
			}, scw.WithContext(ctx))

			return err
		})
	}

	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if err := identity.SetRegionalIdentity(d, region, id); err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("account_id", regional.NewIDString(region, credentials.NatsAccountID))
	_ = d.Set("name", credentials.Name)
	_ = d.Set("region", region)

	return nil
}

func ResourceMNQNatsCredentialsDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
