package interlink

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	interlink "github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/partner_data_source.md
var partnerDataSourceDescription string

func DataSourcePartner() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourcePartnerRead,
		Description: partnerDataSourceDescription,
		Schema: map[string]*schema.Schema{
			"partner_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The ID of the partner",
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				ConflictsWith:    []string{"name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "The name of the partner to filter for",
				ConflictsWith: []string{"partner_id"},
			},
			"region": regional.Schema(),
			// Computed attributes
			"contact_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Contact email address of the partner",
			},
			"logo_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the partner's logo",
			},
			"portal_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the partner's portal",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the partner",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update date of the partner",
			},
		},
	}
}

func DataSourcePartnerRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	partnerID, idExists := d.GetOk("partner_id")
	if idExists {
		return dataSourcePartnerReadByID(ctx, d, m, partnerID.(string))
	}

	return dataSourcePartnerReadByFilters(ctx, d, m)
}

func dataSourcePartnerReadByID(ctx context.Context, d *schema.ResourceData, m any, partnerID string) diag.Diagnostics {
	api, region, err := interlinkAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	partner, err := api.GetPartner(&interlink.GetPartnerRequest{
		Region:    region,
		PartnerID: locality.ExpandID(partnerID),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, partner.ID))

	return setPartnerState(d, partner)
}

func dataSourcePartnerReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := interlinkAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &interlink.ListPartnersRequest{
		Region: region,
	}

	res, err := api.ListPartners(req, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)
	var matches []*interlink.Partner

	for _, p := range res.Partners {
		if name == "" || p.Name == name {
			matches = append(matches, p)
		}
	}

	if len(matches) == 0 {
		return diag.FromErr(errors.New("no partner found matching the specified filters"))
	}

	if len(matches) > 1 {
		return diag.FromErr(fmt.Errorf("multiple partners (%d) found, please refine your filters or use partner_id", len(matches)))
	}

	partner := matches[0]
	d.SetId(regional.NewIDString(region, partner.ID))

	return setPartnerState(d, partner)
}

func setPartnerState(d *schema.ResourceData, partner *interlink.Partner) diag.Diagnostics {
	_ = d.Set("name", partner.Name)
	_ = d.Set("contact_email", partner.ContactEmail)
	_ = d.Set("logo_url", partner.LogoURL)
	_ = d.Set("portal_url", partner.PortalURL)
	_ = d.Set("created_at", types.FlattenTime(partner.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(partner.UpdatedAt))

	return nil
}
