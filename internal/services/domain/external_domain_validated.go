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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceExternalDomainValidated() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExternalDomainValidatedCreate,
		ReadContext:   resourceExternalDomainValidatedRead,
		DeleteContext: resourceExternalDomainValidatedDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultDomainRegistrationTimeout),
			Delete:  schema.DefaultTimeout(defaultDomainRegistrationTimeout),
			Default: schema.DefaultTimeout(defaultDomainRegistrationTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    externalDomainValidatedSchema,
		Identity:      identity.DefaultGlobal(),
	}
}

func externalDomainValidatedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"domain": {
			Type:        schema.TypeString,
			Description: "The domain name to be validated.",
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
		"validated": {
			Type:        schema.TypeBool,
			Description: "Indicates if the domain has been validated by Scaleway.",
			Computed:    true,
		},
	}
}

func resourceExternalDomainValidatedCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)
	domainName := d.Get("domain").(string)

	_, err := waitForExternalDomainValidation(ctx, registrarAPI, domainName, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := identity.SetGlobalIdentity(d, domainName); err != nil {
		return diag.FromErr(err)
	}

	return resourceExternalDomainValidatedRead(ctx, d, m)
}

func resourceExternalDomainValidatedRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)

	resp, err := registrarAPI.GetDomain(&domain.RegistrarAPIGetDomainRequest{Domain: d.Id()}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			log.Printf("[WARN] domain external domain (%s) not found, removing from state", d.Id())
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if err := identity.SetGlobalIdentity(d, resp.Domain); err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("domain", resp.Domain)
	_ = d.Set("project_id", resp.ProjectID)
	_ = d.Set("organization_id", resp.OrganizationID)

	if len(resp.DNSZones) > 0 {
		if err := d.Set("ns_servers", resp.DNSZones[0].NsDefault); err != nil {
			return diag.FromErr(err)
		}
	}

	_ = d.Set("validated", resp.Status == domain.DomainStatusActive)

	return nil
}

func resourceExternalDomainValidatedDelete(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")

	return nil
}
