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

//go:embed descriptions/dedicated_connection_data_source.md
var dedicatedConnectionDataSourceDescription string

func DataSourceDedicatedConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceDedicatedConnectionRead,
		Description: dedicatedConnectionDataSourceDescription,
		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The ID of the dedicated connection",
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				ConflictsWith:    []string{"name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "The name of the dedicated connection to filter for",
				ConflictsWith: []string{"connection_id"},
			},
			"region": regional.Schema(),
			"project_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the project the dedicated connection belongs to",
			},
			// Computed attributes
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Organization ID",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the dedicated connection",
			},
			"tags": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of tags associated with the dedicated connection",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"pop_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the PoP where the dedicated connection is located",
			},
			"bandwidth_mbps": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Bandwidth size of the dedicated connection in Mbps",
			},
			"available_link_bandwidths": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Sizes of the links supported on this dedicated connection",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"demarcation_info": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Demarcation details required by the data center to set up the Cross Connect",
			},
			"vlan_range": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "VLAN range for self-hosted links on this dedicated connection",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Start of the VLAN range",
						},
						"end": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "End of the VLAN range",
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the dedicated connection",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last modification date of the dedicated connection",
			},
		},
	}
}

func DataSourceDedicatedConnectionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	connectionID, idExists := d.GetOk("connection_id")
	if idExists {
		return dataSourceDedicatedConnectionReadByID(ctx, d, m, connectionID.(string))
	}

	return dataSourceDedicatedConnectionReadByFilters(ctx, d, m)
}

func dataSourceDedicatedConnectionReadByID(ctx context.Context, d *schema.ResourceData, m any, connectionID string) diag.Diagnostics {
	api, region, err := interlinkAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	idRegion, id, parseErr := regional.ParseID(connectionID)
	if parseErr == nil {
		region = idRegion
	} else {
		id = connectionID
	}

	conn, err := api.GetDedicatedConnection(&interlink.GetDedicatedConnectionRequest{
		Region:       region,
		ConnectionID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(conn.Region, conn.ID))

	return setDedicatedConnectionState(d, conn)
}

func dataSourceDedicatedConnectionReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := interlinkAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &interlink.ListDedicatedConnectionsRequest{
		Region: region,
		Name:   types.ExpandStringPtr(d.Get("name")),
	}

	res, err := api.ListDedicatedConnections(req, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)

	var matches []*interlink.DedicatedConnection

	for _, c := range res.Connections {
		if name == "" || c.Name == name {
			matches = append(matches, c)
		}
	}

	if len(matches) == 0 {
		return diag.FromErr(errors.New("no dedicated connection found matching the specified filters"))
	}

	if len(matches) > 1 {
		return diag.FromErr(fmt.Errorf("multiple dedicated connections (%d) found, please refine your filters or use connection_id", len(matches)))
	}

	conn := matches[0]
	d.SetId(regional.NewIDString(conn.Region, conn.ID))

	return setDedicatedConnectionState(d, conn)
}

func setDedicatedConnectionState(d *schema.ResourceData, conn *interlink.DedicatedConnection) diag.Diagnostics {
	_ = d.Set("name", conn.Name)
	_ = d.Set("status", conn.Status.String())
	_ = d.Set("tags", conn.Tags)
	_ = d.Set("pop_id", regional.NewIDString(conn.Region, conn.PopID))
	_ = d.Set("bandwidth_mbps", int(conn.BandwidthMbps))
	_ = d.Set("project_id", conn.ProjectID)
	_ = d.Set("organization_id", conn.OrganizationID)
	_ = d.Set("demarcation_info", types.FlattenStringPtr(conn.DemarcationInfo))
	_ = d.Set("created_at", types.FlattenTime(conn.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(conn.UpdatedAt))
	_ = d.Set("region", conn.Region.String())

	bandwidths := make([]int, len(conn.AvailableLinkBandwidths))
	for i, b := range conn.AvailableLinkBandwidths {
		bandwidths[i] = int(b)
	}

	_ = d.Set("available_link_bandwidths", bandwidths)

	if conn.VlanRange != nil {
		_ = d.Set("vlan_range", []map[string]any{{
			"start": int(conn.VlanRange.Start),
			"end":   int(conn.VlanRange.End),
		}})
	}

	return nil
}
