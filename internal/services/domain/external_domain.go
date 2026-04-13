package domain

import (
	"context"

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
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultDomainRegistrationTimeout),
			Read:    schema.DefaultTimeout(defaultDomainRegistrationTimeout),
			Delete:  schema.DefaultTimeout(defaultDomainRegistrationTimeout),
			Default: schema.DefaultTimeout(defaultDomainRegistrationTimeout),
		},
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
		"project_id": account.ProjectIDSchema(),
		"organization_id": {
			Type:        schema.TypeString,
			Description: "The organization ID the domain is associated with.",
			Computed:    true,
		},
		"ns_servers": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of default NS servers for the domain.",
			Computed:    true,
		},
		"status": {
			Type:        schema.TypeString,
			Description: "The status of the domain.",
			Computed:    true,
		},
		"validation_token": {
			Type:        schema.TypeString,
			Description: "The validation token to add as a TXT record to prove domain ownership.",
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

	_, err = registrarAPI.RegisterExternalDomain(&domain.RegistrarAPIRegisterExternalDomainRequest{
		Domain:    domainName,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := identity.SetGlobalIdentity(d, domainName); err != nil {
		return diag.FromErr(err)
	}

	return resourceExternalDomainRead(ctx, d, m)
}

func resourceExternalDomainRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)

	resp, err := registrarAPI.GetDomain(&domain.RegistrarAPIGetDomainRequest{Domain: d.Id()}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if err := identity.SetGlobalIdentity(d, resp.Domain); err != nil {
		return diag.FromErr(err)
	}

	persistExternalDomainFromRegistrarResponse(resp, d)

	return nil
}

func persistExternalDomainFromRegistrarResponse(resp *domain.Domain, d *schema.ResourceData) {
	_ = d.Set("domain", resp.Domain)
	_ = d.Set("project_id", resp.ProjectID)
	_ = d.Set("organization_id", resp.OrganizationID)
	_ = d.Set("status", resp.Status)

	if len(resp.DNSZones) > 0 {
		_ = d.Set("ns_servers", resp.DNSZones[0].NsDefault)
	}

	if resp.ExternalDomainRegistrationStatus != nil && resp.ExternalDomainRegistrationStatus.ValidationToken != "" {
		_ = d.Set("validation_token", resp.ExternalDomainRegistrationStatus.ValidationToken)
	}
}

func resourceExternalDomainDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)

	_, err := registrarAPI.DeleteExternalDomain(&domain.RegistrarAPIDeleteExternalDomainRequest{
		Domain: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
