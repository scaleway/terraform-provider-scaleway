package interlink

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	interlink "github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

//go:embed descriptions/link.md
var linkDescription string

func ResourceLink() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceLinkCreate,
		ReadContext:   ResourceLinkRead,
		UpdateContext: ResourceLinkUpdate,
		DeleteContext: ResourceLinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultLinkTimeout),
			Read:    schema.DefaultTimeout(defaultLinkTimeout),
			Update:  schema.DefaultTimeout(defaultLinkTimeout),
			Delete:  schema.DefaultTimeout(defaultLinkTimeout),
			Default: schema.DefaultTimeout(defaultLinkTimeout),
		},
		Description:   linkDescription,
		Identity:      identity.DefaultRegional(),
		SchemaVersion: 0,
		SchemaFunc:    linkSchema,
	}
}

func linkSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "Name of the link",
		},
		"tags": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "List of tags associated with the link",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"pop_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "PoP (location) where the link will be created",
		},
		"bandwidth_mbps": {
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
			Description: "Desired bandwidth for the link. Must be compatible with available link bandwidths and remaining bandwidth capacity of the connection",
		},
		"connection_id": {
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			Description:   "If set, creates a self-hosted link using this dedicated physical connection",
			ConflictsWith: []string{"partner_id"},
		},
		"partner_id": {
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			Description:   "If set, creates a hosted link on a partner's connection. Specify the ID of the chosen partner, who already has a shared connection with available bandwidth",
			ConflictsWith: []string{"connection_id"},
		},
		"vpc_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "ID of the Scaleway VPC to attach to the link",
			DiffSuppressFunc: dsf.Locality,
		},
		"enable_route_propagation": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Defines whether route propagation is enabled or not. Defaults to false",
		},
		"peer_asn": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			Description: "For self-hosted links, the peer AS Number to establish BGP session. If not given, a default one will be assigned",
		},
		"vlan": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
			Description: "For self-hosted links only, the VLAN ID. If the VLAN is not available (already taken or out of range), an error is returned",
		},
		"routing_policy_v4_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "If set, attaches this routing policy containing IPv4 prefixes to the link. A BGP IPv4 session will be created",
		},
		"routing_policy_v6_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "If set, attaches this routing policy containing IPv6 prefixes to the link. A BGP IPv6 session will be created",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Status of the link",
		},
		"bgp_v4_status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Status of the link's BGP IPv4 session",
		},
		"bgp_v6_status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Status of the link's BGP IPv6 session",
		},
		"pairing_key": {
			Type:        schema.TypeString,
			Computed:    true,
			Sensitive:   true,
			Description: "Used to identify a link from a user or partner's point of view",
		},
		"scw_bgp_config": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "BGP configuration on Scaleway's side",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"asn": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "AS Number of the BGP peer",
					},
					"ipv4": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "IPv4 address of the BGP peer",
					},
					"ipv6": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "IPv6 address of the BGP peer",
					},
				},
			},
		},
		"peer_bgp_config": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "BGP configuration on peer's side (on-premises or other hosting provider)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"asn": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "AS Number of the BGP peer",
					},
					"ipv4": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "IPv4 address of the BGP peer",
					},
					"ipv6": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "IPv6 address of the BGP peer",
					},
				},
			},
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Creation date of the link",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Last modification date of the link",
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
		"organization_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The Organization ID the link is associated with",
		},
	}
}

func ResourceLinkCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := interlinkAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &interlink.CreateLinkRequest{
		Region:        region,
		ProjectID:     d.Get("project_id").(string),
		Name:          types.ExpandOrGenerateString(d.Get("name").(string), "link"),
		Tags:          types.ExpandStrings(d.Get("tags")),
		PopID:         regional.ExpandID(d.Get("pop_id").(string)).ID,
		BandwidthMbps: uint64(d.Get("bandwidth_mbps").(int)),
	}

	if connectionID, ok := d.GetOk("connection_id"); ok {
		req.ConnectionID = types.ExpandStringPtr(regional.ExpandID(connectionID.(string)).ID)
	}

	if partnerID, ok := d.GetOk("partner_id"); ok {
		req.PartnerID = types.ExpandStringPtr(regional.ExpandID(partnerID.(string)).ID)
	}

	if peerASN, ok := d.GetOk("peer_asn"); ok {
		req.PeerAsn = new(uint32(peerASN.(int)))
	}

	if vlan, ok := d.GetOk("vlan"); ok {
		req.Vlan = new(uint32(vlan.(int)))
	}

	if rpV4, ok := d.GetOk("routing_policy_v4_id"); ok {
		req.RoutingPolicyV4ID = types.ExpandStringPtr(regional.ExpandID(rpV4.(string)).ID)
	}

	if rpV6, ok := d.GetOk("routing_policy_v6_id"); ok {
		req.RoutingPolicyV6ID = types.ExpandStringPtr(regional.ExpandID(rpV6.(string)).ID)
	}

	link, err := api.CreateLink(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetRegionalIdentity(d, link.Region, link.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForLink(ctx, api, region, link.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("enable_route_propagation").(bool) {
		_, err = api.EnableRoutePropagation(&interlink.EnableRoutePropagationRequest{
			Region: link.Region,
			LinkID: link.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForLink(ctx, api, link.Region, link.ID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if vpcID, ok := d.GetOk("vpc_id"); ok {
		_, err = api.AttachVpc(&interlink.AttachVpcRequest{
			Region: link.Region,
			LinkID: link.ID,
			VpcID:  regional.ExpandID(vpcID.(string)).ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForLink(ctx, api, link.Region, link.ID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceLinkRead(ctx, d, m)
}

func ResourceLinkRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	link, err := api.GetLink(&interlink.GetLinkRequest{
		LinkID: id,
		Region: region,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	diags := setLinkState(d, link)

	err = identity.SetRegionalIdentity(d, link.Region, link.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func setLinkState(d *schema.ResourceData, link *interlink.Link) diag.Diagnostics {
	_ = d.Set("name", link.Name)
	_ = d.Set("region", link.Region)
	_ = d.Set("project_id", link.ProjectID)
	_ = d.Set("organization_id", link.OrganizationID)
	_ = d.Set("tags", link.Tags)
	_ = d.Set("pop_id", regional.NewIDString(link.Region, link.PopID))
	_ = d.Set("bandwidth_mbps", int(link.BandwidthMbps))
	_ = d.Set("status", link.Status.String())
	_ = d.Set("bgp_v4_status", link.BgpV4Status.String())
	_ = d.Set("bgp_v6_status", link.BgpV6Status.String())
	_ = d.Set("enable_route_propagation", link.EnableRoutePropagation)
	_ = d.Set("vlan", int(link.Vlan))
	_ = d.Set("created_at", types.FlattenTime(link.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(link.UpdatedAt))

	if link.VpcID != nil {
		_ = d.Set("vpc_id", regional.NewIDString(link.Region, *link.VpcID))
	} else {
		_ = d.Set("vpc_id", "")
	}

	if link.RoutingPolicyV4ID != nil {
		_ = d.Set("routing_policy_v4_id", regional.NewIDString(link.Region, *link.RoutingPolicyV4ID))
	}

	if link.RoutingPolicyV6ID != nil {
		_ = d.Set("routing_policy_v6_id", regional.NewIDString(link.Region, *link.RoutingPolicyV6ID))
	}

	scwBgpConfig, err := flattenBgpConfig(link.ScwBgpConfig)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("scw_bgp_config", scwBgpConfig)

	peerBgpConfig, err := flattenBgpConfig(link.PeerBgpConfig)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("peer_bgp_config", peerBgpConfig)

	if link.Partner != nil {
		_ = d.Set("partner_id", regional.NewIDString(link.Region, link.Partner.PartnerID))
		_ = d.Set("pairing_key", link.Partner.PairingKey)
		_ = d.Set("connection_id", "")
	} else {
		_ = d.Set("partner_id", "")
		_ = d.Set("pairing_key", "")
	}

	if link.Self != nil {
		_ = d.Set("connection_id", regional.NewIDString(link.Region, link.Self.ConnectionID))
	} else if link.Partner == nil {
		_ = d.Set("connection_id", "")
	}

	return nil
}

func ResourceLinkUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	hasChanged := false

	req := &interlink.UpdateLinkRequest{
		Region: region,
		LinkID: id,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if d.HasChange("peer_asn") {
		req.PeerAsn = new(uint32(d.Get("peer_asn").(int)))
		hasChanged = true
	}

	if hasChanged {
		_, err = api.UpdateLink(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_, err = waitForLink(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("enable_route_propagation") {
		if d.Get("enable_route_propagation").(bool) {
			_, err = api.EnableRoutePropagation(&interlink.EnableRoutePropagationRequest{
				Region: region,
				LinkID: id,
			}, scw.WithContext(ctx))
		} else {
			_, err = api.DisableRoutePropagation(&interlink.DisableRoutePropagationRequest{
				Region: region,
				LinkID: id,
			}, scw.WithContext(ctx))
		}
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForLink(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("vpc_id") {
		oldRaw, newRaw := d.GetChange("vpc_id")
		oldVpcID := oldRaw.(string)
		newVpcID := newRaw.(string)

		if oldVpcID != "" {
			_, err = api.DetachVpc(&interlink.DetachVpcRequest{
				Region: region,
				LinkID: id,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			_, err = waitForLink(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if newVpcID != "" {
			_, err = api.AttachVpc(&interlink.AttachVpcRequest{
				Region: region,
				LinkID: id,
				VpcID:  regional.ExpandID(newVpcID).ID,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			_, err = waitForLink(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return ResourceLinkRead(ctx, d, m)
}

func ResourceLinkDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("enable_route_propagation").(bool) {
		_, err = api.DisableRoutePropagation(&interlink.DisableRoutePropagationRequest{
			Region: region,
			LinkID: id,
		}, scw.WithContext(ctx))
		if err != nil && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}

		_, err = waitForLink(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
		if err != nil && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	if vpcID := d.Get("vpc_id").(string); vpcID != "" {
		_, err = api.DetachVpc(&interlink.DetachVpcRequest{
			Region: region,
			LinkID: id,
		}, scw.WithContext(ctx))
		if err != nil && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}

		_, err = waitForLink(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
		if err != nil && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	_, err = api.DeleteLink(&interlink.DeleteLinkRequest{
		Region: region,
		LinkID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForLink(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}
