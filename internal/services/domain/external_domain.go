package domain

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceExternalDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExternalDomainCreate,
		ReadContext:   resourceExternalDomainRead,
		DeleteContext: resourceExternalDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    externalDomainSchema,
		Identity:      identity.DefaultGlobal(),
	}
}

func externalDomainSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"domain": {
			Type:        schema.TypeString,
			Description: "The domain name to be managed.",
			Required:    true,
			ForceNew:    true,
		},
		"project_id":      account.ProjectIDSchema(),
		"organization_id": account.OrganizationIDSchema(),
		"ns_servers": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of NS servers for the domain.",
			Computed:    true,
		},
		"status": {
			Type:        schema.TypeString,
			Description: "The status of the domain.",
			Computed:    true,
		},
		"validation_token": {
			Type:        schema.TypeString,
			Description: "The validation token for the domain.",
			Computed:    true,
		},
	}
}

func resourceExternalDomainCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)

	projectID, _, err := meta.ExtractProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	domainName := d.Get("domain").(string)

	resp, err := registrarAPI.RegisterExternalDomain(&domain.RegistrarAPIRegisterExternalDomainRequest{
		Domain:    domainName,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := identity.SetGlobalIdentity(d, resp.Domain); err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("validation_token", resp.ValidationToken)
	_ = d.Set("organization_id", resp.OrganizationID)

	return resourceExternalDomainRead(ctx, d, m)
}

func resourceExternalDomainRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)

	resp, diags := readExternalDomainIntoState(ctx, d, registrarAPI)
	if diags.HasError() {
		return diags
	}

	if resp == nil {
		return nil
	}

	if err := identity.SetGlobalIdentity(d, resp.Domain); err != nil {
		return diag.FromErr(err)
	}

	validationToken := ""
	if resp.ExternalDomainRegistrationStatus != nil && resp.ExternalDomainRegistrationStatus.ValidationToken != "" {
		validationToken = resp.ExternalDomainRegistrationStatus.ValidationToken
	}

	_ = d.Set("validation_token", validationToken)

	return nil
}

// readExternalDomainIntoState fetches external domain data and sets schema
// attributes without Identity (works for both the resource and the data source).
// It does not set attributes that are not part of the data source schema
// (validation_token), which the resource read handles itself.
func readExternalDomainIntoState(ctx context.Context, d *schema.ResourceData, registrarAPI *domain.RegistrarAPI) (*domain.Domain, diag.Diagnostics) {
	resp, err := registrarAPI.GetDomain(&domain.RegistrarAPIGetDomainRequest{Domain: d.Id()}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			log.Printf("[WARN] domain external domain (%s) not found, removing from state", d.Id())
			d.SetId("")

			return nil, nil
		}

		return nil, diag.FromErr(err)
	}

	d.SetId(resp.Domain)
	_ = d.Set("domain", resp.Domain)
	_ = d.Set("project_id", resp.ProjectID)
	_ = d.Set("organization_id", resp.OrganizationID)

	if len(resp.DNSZones) > 0 {
		if err := d.Set("ns_servers", resp.DNSZones[0].NsDefault); err != nil {
			return resp, diag.FromErr(err)
		}
	}

	_ = d.Set("status", resp.Status.String())

	return resp, nil
}

func resourceExternalDomainDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if d.Get("status") != "active" {
		return diag.Errorf("cannot delete domain with status %s", d.Get("status"))
	}

	registrarAPI := NewRegistrarDomainAPI(m)

	_, err := registrarAPI.DeleteExternalDomain(&domain.RegistrarAPIDeleteExternalDomainRequest{
		Domain: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			log.Printf("[WARN] domain external domain (%s) not found, removing from state", d.Id())
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
