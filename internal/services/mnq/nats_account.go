package mnq

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const natsAccountReadAfterCreateRetryTimeout = 60 * time.Second

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
		SchemaFunc:    natsAccountSchema,
		Identity:      identity.DefaultRegional(),
	}
}

func natsAccountSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
	}
}

func ResourceMNQNatsAccountCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	if err := identity.SetRegionalIdentity(d, account.Region, account.ID); err != nil {
		return diag.FromErr(err)
	}

	err = retry.RetryContext(ctx, natsAccountReadAfterCreateRetryTimeout, func() *retry.RetryError {
		_, err = api.GetNatsAccount(&mnq.NatsAPIGetNatsAccountRequest{
			Region:        account.Region,
			NatsAccountID: account.ID,
		}, scw.WithContext(ctx))
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

	return ResourceMNQNatsAccountRead(ctx, d, m)
}

func ResourceMNQNatsAccountRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewNatsAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	account, err := api.GetNatsAccount(&mnq.NatsAPIGetNatsAccountRequest{
		Region:        region,
		NatsAccountID: id,
	}, scw.WithContext(ctx))
	if err != nil && isMNQNamespaceReadRetryableError(err) {
		err = retryMNQNamespaceRead(ctx, func() error {
			account, err = api.GetNatsAccount(&mnq.NatsAPIGetNatsAccountRequest{
				Region:        region,
				NatsAccountID: id,
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

	if err := identity.SetRegionalIdentity(d, account.Region, account.ID); err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", account.Name)
	_ = d.Set("region", account.Region)
	_ = d.Set("project_id", account.ProjectID)
	_ = d.Set("endpoint", account.Endpoint)

	return nil
}

func ResourceMNQNatsAccountUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

const natsAccountDeleteRetryTimeout = 60 * time.Second

func ResourceMNQNatsAccountDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewNatsAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &mnq.NatsAPIDeleteNatsAccountRequest{
		Region:        region,
		NatsAccountID: id,
	}

	err = retry.RetryContext(ctx, natsAccountDeleteRetryTimeout, func() *retry.RetryError {
		delErr := api.DeleteNatsAccount(req, scw.WithContext(ctx))
		if delErr == nil {
			return nil
		}
		// Transient: namespace routing unavailable, account still exists → retry
		if isMNQNamespaceReadRetryableError(delErr) {
			return retry.RetryableError(delErr)
		}
		// Account genuinely not found → treat as already deleted
		if httperrors.Is404(delErr) {
			return nil
		}
		return retry.NonRetryableError(delErr)
	})
	// If the retry timed out on a namespace error, assume the account is gone
	if err != nil && isMNQNamespaceReadRetryableError(err) {
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
