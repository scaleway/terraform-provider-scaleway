package tem

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceBlockedList() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceBlockedListCreate,
		ReadContext:   ResourceBlockedListRead,
		DeleteContext: ResourceBlockedListDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
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
		},
	}
}

func ResourceBlockedListCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := temAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	region, domainID, err := regional.ParseID(d.Get("domain_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	emails := []string{d.Get("email").(string)}
	reason := d.Get("reason").(string)
	typeBlockedList := d.Get("type").(string)

	_, err = api.BulkCreateBlocklists(&tem.BulkCreateBlocklistsRequest{
		Emails:   emails,
		Region:   region,
		DomainID: domainID,
		Type:     tem.BlocklistType(typeBlockedList),
		Reason:   &reason,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s-%s", region, domainID))

	return ResourceBlockedListRead(ctx, d, m)
}

func ResourceBlockedListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, _, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	blocklists, err := api.ListBlocklists(&tem.ListBlocklistsRequest{
		Region:   region,
		Email:    scw.StringPtr(d.Get("email").(string)),
		DomainID: d.Get("domain_id").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if len(blocklists.Blocklists) == 0 {
		d.SetId("")
		return nil
	}

	_ = d.Set("email", blocklists.Blocklists[0].Email)
	_ = d.Set("reason", blocklists.Blocklists[0].Reason)
	_ = d.Set("domain_id", blocklists.Blocklists[0].DomainID)
	_ = d.Set("type", blocklists.Blocklists[0].Type)
	return nil
}

func ResourceBlockedListDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	d.SetId("")
	return nil
}
