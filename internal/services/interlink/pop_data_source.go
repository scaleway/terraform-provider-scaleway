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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/pop_data_source.md
var popDataSourceDescription string

func DataSourcePop() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourcePopRead,
		Description: popDataSourceDescription,
		Schema: map[string]*schema.Schema{
			"pop_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The ID of the PoP",
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				ConflictsWith:    []string{"name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "The name of the PoP to filter for",
				ConflictsWith: []string{"pop_id"},
			},
			"region": regional.Schema(),
			// Computed attributes
			"hosting_provider_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the PoP's hosting provider",
			},
			"address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Physical address of the PoP",
			},
			"city": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "City where the PoP is located",
			},
			"logo_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the PoP's logo",
			},
			"available_link_bandwidths_mbps": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Available bandwidth options in Mbps for hosted links",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Human-readable display name of the PoP",
			},
		},
	}
}

func DataSourcePopRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	popID, idExists := d.GetOk("pop_id")
	if idExists {
		return dataSourcePopReadByID(ctx, d, m, popID.(string))
	}

	return dataSourcePopReadByFilters(ctx, d, m)
}

func dataSourcePopReadByID(ctx context.Context, d *schema.ResourceData, m any, popID string) diag.Diagnostics {
	api, region, err := interlinkAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	idRegion, id, parseErr := regional.ParseID(popID)
	if parseErr == nil {
		region = idRegion
	} else {
		id = popID
	}

	pop, err := api.GetPop(&interlink.GetPopRequest{
		Region: region,
		PopID:  id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(pop.Region, pop.ID))

	return setPopState(d, pop)
}

func dataSourcePopReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := interlinkAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &interlink.ListPopsRequest{
		Region: region,
		Name:   types.ExpandStringPtr(d.Get("name")),
	}

	res, err := api.ListPops(req, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	if len(res.Pops) == 0 {
		return diag.FromErr(errors.New("no PoP found matching the specified filters"))
	}

	if len(res.Pops) > 1 {
		return diag.FromErr(fmt.Errorf("multiple PoPs (%d) found, please refine your filters or use pop_id", len(res.Pops)))
	}

	pop := res.Pops[0]
	d.SetId(regional.NewIDString(pop.Region, pop.ID))

	return setPopState(d, pop)
}

func setPopState(d *schema.ResourceData, pop *interlink.Pop) diag.Diagnostics {
	_ = d.Set("name", pop.Name)
	_ = d.Set("hosting_provider_name", pop.HostingProviderName)
	_ = d.Set("address", pop.Address)
	_ = d.Set("city", pop.City)
	_ = d.Set("logo_url", pop.LogoURL)
	_ = d.Set("display_name", pop.DisplayName)
	_ = d.Set("region", pop.Region.String())

	bandwidths := make([]int, len(pop.AvailableLinkBandwidthsMbps))
	for i, b := range pop.AvailableLinkBandwidthsMbps {
		bandwidths[i] = int(b)
	}

	_ = d.Set("available_link_bandwidths_mbps", bandwidths)

	return nil
}
