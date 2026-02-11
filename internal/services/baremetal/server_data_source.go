package baremetal

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	baremetalV3 "github.com/scaleway/scaleway-sdk-go/api/baremetal/v3"
	ipamAPI "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/ipam"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceServer() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceServer().SchemaFunc())

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "zone", "project_id")

	dsSchema["name"].ConflictsWith = []string{"server_id"}
	dsSchema["server_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the server",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		ReadContext: DataSourceServerRead,
		Schema:      dsSchema,
	}
}

func DataSourceServerRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	serverID, ok := d.GetOk("server_id")
	if !ok { // Get server by zone and name.
		serverName := d.Get("name").(string)

		res, err := api.ListServers(&baremetal.ListServersRequest{
			Zone:      zone,
			Name:      scw.StringPtr(serverName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundServer, err := datasource.FindExact(
			res.Servers,
			func(s *baremetal.Server) bool { return s.Name == serverName },
			serverName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		serverID = foundServer.ID
	}

	zoneID := datasource.NewZonedID(serverID, zone)
	d.SetId(zoneID)

	err = d.Set("server_id", zoneID)
	if err != nil {
		return diag.FromErr(err)
	}

	server, err := api.GetServer(&baremetal.GetServerRequest{
		Zone:     zone,
		ServerID: serverID.(string),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	offer, err := api.GetOffer(&baremetal.GetOfferRequest{
		Zone:    server.Zone,
		OfferID: server.OfferID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	var os *baremetal.OS
	if server.Install != nil {
		os, err = api.GetOS(&baremetal.GetOSRequest{
			Zone: server.Zone,
			OsID: server.Install.OsID,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	privateNetworkAPI, _, err := newPrivateNetworkAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	listPrivateNetworks, err := privateNetworkAPI.ListServerPrivateNetworks(&baremetalV3.PrivateNetworkAPIListServerPrivateNetworksRequest{
		Zone:     server.Zone,
		ServerID: &server.ID,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to list server's private networks: %w", err))
	}

	privateNetworkIDs := make([]string, 0, listPrivateNetworks.TotalCount)
	for _, pn := range listPrivateNetworks.ServerPrivateNetworks {
		privateNetworkIDs = append(privateNetworkIDs, pn.PrivateNetworkID)
	}

	// Read private IPs if possible
	allPrivateIPs := make([]map[string]any, 0, listPrivateNetworks.TotalCount)
	diags := diag.Diagnostics{}

	pnRegion, err := server.Zone.Region()
	if err != nil {
		return diag.FromErr(err)
	}

	for _, privateNetworkID := range privateNetworkIDs {
		resourceType := ipamAPI.ResourceTypeBaremetalPrivateNic
		opts := &ipam.GetResourcePrivateIPsOptions{
			ResourceType:     &resourceType,
			PrivateNetworkID: &privateNetworkID,
			ProjectID:        &server.ProjectID,
		}

		privateIPs, err := ipam.GetResourcePrivateIPs(ctx, m, pnRegion, opts)

		switch {
		case err == nil:
			allPrivateIPs = append(allPrivateIPs, privateIPs...)
		case httperrors.Is403(err):
			return append(diags, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       "Unauthorized to read server's private IPs, please check your IAM permissions",
				Detail:        err.Error(),
				AttributePath: cty.GetAttrPath("private_ips"),
			})
		default:
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       fmt.Sprintf("Unable to get private IPs for server %s (pn_id: %s)", server.ID, privateNetworkID),
				Detail:        err.Error(),
				AttributePath: cty.GetAttrPath("private_ips"),
			})
		}
	}

	diags = setServerState(d, server, offer, os, listPrivateNetworks, allPrivateIPs)
	if diags != nil {
		return diags
	}

	if d.Id() == "" {
		return diag.Errorf("baremetal server (%s) not found", zoneID)
	}

	return nil
}
