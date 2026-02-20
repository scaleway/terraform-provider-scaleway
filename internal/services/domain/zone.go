package domain

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceZone() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainZoneCreate,
		ReadContext:   resourceDomainZoneRead,
		UpdateContext: resourceZoneUpdate,
		DeleteContext: resourceZoneDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultDomainZoneTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    zoneSchema,
	}
}

func zoneSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"domain": {
			Type:        schema.TypeString,
			Description: "The domain where the DNS zone will be created.",
			Required:    true,
			ForceNew:    true,
		},
		"subdomain": {
			Type:        schema.TypeString,
			Description: "The subdomain of the DNS zone to create.",
			Required:    true,
		},
		"ns": {
			Type:        schema.TypeList,
			Description: "NameServer list for zone.",
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ns_default": {
			Type:        schema.TypeList,
			Description: "NameServer default list for zone.",
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ns_master": {
			Type:        schema.TypeList,
			Description: "NameServer master list for zone.",
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"status": {
			Type:        schema.TypeString,
			Description: "The domain zone status.",
			Computed:    true,
		},
		"message": {
			Type:        schema.TypeString,
			Description: "Message",
			Computed:    true,
		},
		"updated_at": {
			Type:        schema.TypeString,
			Description: "The date and time of the last update of the DNS zone.",
			Computed:    true,
		},
		"project_id": account.ProjectIDSchema(),
	}
}

func resourceDomainZoneCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	domainAPI := NewDomainAPI(m)

	domainName := strings.ToLower(d.Get("domain").(string))
	subdomainName := strings.ToLower(d.Get("subdomain").(string))
	zoneName := BuildZoneName(subdomainName, domainName)

	zones, err := domainAPI.ListDNSZones(&domain.ListDNSZonesRequest{
		ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		DNSZones:  []string{zoneName},
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	for i := range zones.DNSZones {
		if zones.DNSZones[i].Domain == domainName && zones.DNSZones[i].Subdomain == subdomainName {
			d.SetId(BuildZoneName(subdomainName, domainName))

			return resourceDomainZoneRead(ctx, d, m)
		}
	}

	var dnsZone *domain.DNSZone

	dnsZone, err = domainAPI.CreateDNSZone(&domain.CreateDNSZoneRequest{
		ProjectID: d.Get("project_id").(string),
		Domain:    domainName,
		Subdomain: subdomainName,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is409(err) {
			return resourceDomainZoneRead(ctx, d, m)
		}

		return diag.FromErr(err)
	}

	d.SetId(BuildZoneName(dnsZone.Subdomain, dnsZone.Domain))

	return resourceDomainZoneRead(ctx, d, m)
}

func resourceDomainZoneRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	domainAPI := NewDomainAPI(m)

	var zone *domain.DNSZone

	zones, err := domainAPI.ListDNSZones(&domain.ListDNSZonesRequest{
		ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		DNSZones:  []string{d.Id()},
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if len(zones.DNSZones) == 0 {
		return diag.FromErr(fmt.Errorf("no zone found with the name %s", d.Id()))
	}

	if len(zones.DNSZones) > 1 {
		return diag.FromErr(fmt.Errorf("%d zone found with the same name %s", len(zones.DNSZones), d.Id()))
	}

	zone = zones.DNSZones[0]

	_ = d.Set("subdomain", zone.Subdomain)
	_ = d.Set("domain", zone.Domain)
	_ = d.Set("ns", zone.Ns)
	_ = d.Set("ns_default", zone.NsDefault)
	_ = d.Set("ns_master", zone.NsMaster)
	_ = d.Set("status", zone.Status.String())
	_ = d.Set("message", zone.Message)
	_ = d.Set("updated_at", zone.UpdatedAt.String())
	_ = d.Set("project_id", zone.ProjectID)

	return nil
}

func resourceZoneUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	domainAPI := NewDomainAPI(m)

	if d.HasChangesExcept("subdomain") {
		_, err := domainAPI.UpdateDNSZone(&domain.UpdateDNSZoneRequest{
			ProjectID:  d.Get("project_id").(string),
			DNSZone:    d.Id(),
			NewDNSZone: new(d.Get("subdomain").(string)),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDomainZoneRead(ctx, d, m)
}

func resourceZoneDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	domainAPI := NewDomainAPI(m)

	_, err := waitForDNSZone(ctx, domainAPI, d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if httperrors.Is404(err) || httperrors.Is403(err) {
			return nil
		}

		return diag.FromErr(err)
	}

	_, err = domainAPI.DeleteDNSZone(&domain.DeleteDNSZoneRequest{
		ProjectID: d.Get("project_id").(string),
		DNSZone:   d.Id(),
	}, scw.WithContext(ctx))

	if err != nil && !httperrors.Is404(err) && !httperrors.Is403(err) {
		return diag.FromErr(err)
	}

	return nil
}
