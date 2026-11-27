package tem

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

//go:embed descriptions/tem_blocked_list_resource.md
var blockedListResourceDescription string

func ResourceBlockedList() *schema.Resource {
	return &schema.Resource{
		Description:   blockedListResourceDescription,
		CreateContext: ResourceBlockedListCreate,
		ReadContext:   ResourceBlockedListRead,
		DeleteContext: ResourceBlockedListDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: blockedListSchema,
		Identity:   identity.DefaultRegional(),
	}
}

func blockedListSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"domain_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The ID of the domain affected by the blocklist.",
		},
		"email": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Email address to block.",
		},
		"type": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			Description:  "Type of the blocked list. (mailbox_full or mailbox_not_found)",
			ValidateFunc: validation.StringInSlice([]string{"mailbox_full", "mailbox_not_found"}, false),
		},
		"reason": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "manual_block",
			ForceNew:    true,
			Description: "Reason for blocking the emails.",
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
	}
}

func ResourceBlockedListCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := temAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	_, domainID, err := regional.ParseID(d.Get("domain_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	emails := []string{d.Get("email").(string)}
	typeBlockedList := d.Get("type").(string)

	resp, err := api.BulkCreateBlocklists(&tem.BulkCreateBlocklistsRequest{
		Emails:   emails,
		Region:   region,
		DomainID: domainID,
		Type:     tem.BlocklistType(typeBlockedList),
		Reason:   new(d.Get("reason").(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := identity.SetRegionalIdentity(d, region, resp.Blocklists[0].ID); err != nil {
		return diag.FromErr(err)
	}

	return ResourceBlockedListRead(ctx, d, m)
}

func ResourceBlockedListRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &tem.ListBlocklistsRequest{
		Region: region,
	}

	// Narrow the search on refresh when domain/email are already in state.
	if domainIDAttr, ok := d.Get("domain_id").(string); ok && domainIDAttr != "" {
		_, domainID, parseErr := regional.ParseID(domainIDAttr)
		if parseErr != nil {
			return diag.FromErr(parseErr)
		}

		req.DomainID = domainID
	}

	if email, ok := d.Get("email").(string); ok && email != "" {
		req.Email = &email
	}

	blocklists, err := api.ListBlocklists(req, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	var found *tem.Blocklist

	for _, blocklist := range blocklists.Blocklists {
		if blocklist.ID == id {
			found = blocklist

			break
		}
	}

	if found == nil {
		d.SetId("")

		return nil
	}

	if err := identity.SetRegionalIdentity(d, region, found.ID); err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("email", found.Email)
	_ = d.Set("reason", found.Reason)
	_ = d.Set("domain_id", regional.NewIDString(region, found.DomainID))
	_ = d.Set("type", found.Type)

	return nil
}

func ResourceBlockedListDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteBlocklist(&tem.DeleteBlocklistRequest{
		Region:      region,
		BlocklistID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
