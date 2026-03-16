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
		CustomizeDiff: resourceZoneCustomizeDiff,
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

	// Check if a zone with the same name already exists in the project
	zones, err := domainAPI.ListDNSZones(&domain.ListDNSZonesRequest{
		ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		DNSZones:  []string{zoneName},
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	// If zone already exists, throw an error to prevent duplicate creation
	for i := range zones.DNSZones {
		if zones.DNSZones[i].Domain == domainName && zones.DNSZones[i].Subdomain == subdomainName {
			// Root zones (subdomain "") are auto-created with the domain — allow adoption
			if subdomainName == "" {
				d.SetId(BuildZoneName(subdomainName, domainName))

				return resourceDomainZoneRead(ctx, d, m)
			}

			// Zone already exists - throw error instead of managing existing resource
			return diag.FromErr(fmt.Errorf("a zone with domain '%s' and subdomain '%s' already exists in this project", domainName, subdomainName))
		}
	}

	// Proceed with zone creation only if no existing zone found with same name
	var dnsZone *domain.DNSZone

	dnsZone, err = domainAPI.CreateDNSZone(&domain.CreateDNSZoneRequest{
		ProjectID: d.Get("project_id").(string),
		Domain:    domainName,
		Subdomain: subdomainName,
	}, scw.WithContext(ctx))
	if err != nil {
		// Handle case where zone was already created by another process (409 conflict)
		if httperrors.Is409(err) {
			// Root zones are auto-created with the domain — allow adoption
			if subdomainName == "" {
				d.SetId(BuildZoneName(subdomainName, domainName))

				return resourceDomainZoneRead(ctx, d, m)
			}

			// Zone was created by another process - throw error instead of managing it
			return diag.FromErr(fmt.Errorf("a zone with domain '%s' and subdomain '%s' already exists (HTTP 409). This means either another process is creating the same zone, or it already exists in another project within your Scaleway Organization", domainName, subdomainName))
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

func resourceZoneCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, m any) error {
	// Only check during creation (when ID is not set)
	if diff.Id() != "" {
		return nil
	}

	// Check if domain or subdomain are being created/changed
	if !diff.HasChanges("domain", "subdomain") {
		return nil
	}

	domainAPI := NewDomainAPI(m)

	domainName := strings.ToLower(diff.Get("domain").(string))
	subdomainName := strings.ToLower(diff.Get("subdomain").(string))
	zoneName := BuildZoneName(subdomainName, domainName)

	// Check if a zone with the same name already exists in the project
	zones, err := domainAPI.ListDNSZones(&domain.ListDNSZonesRequest{
		ProjectID: types.ExpandStringPtr(diff.Get("project_id")),
		DNSZones:  []string{zoneName},
	}, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	// If zone already exists, add an error to prevent duplicate creation
	for i := range zones.DNSZones {
		if zones.DNSZones[i].Domain == domainName && zones.DNSZones[i].Subdomain == subdomainName {
			return fmt.Errorf("a zone with domain '%s' and subdomain '%s' already exists in this project", domainName, subdomainName)
		}
	}

	return nil
}
