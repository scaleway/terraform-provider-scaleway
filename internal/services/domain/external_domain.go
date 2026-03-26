package domain

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
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

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Description: "The domain name to be managed.",
				Required:    true,
				ForceNew:    true,
			},
			"project_id": account.ProjectIDSchema(),
			"ns_servers": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of Ns servers for the domain.",
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
	d.SetId(domainName)

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
	_ = d.Set("domain", resp.Domain)
	_ = d.Set("project_id", resp.ProjectID)
	if len(resp.DNSZones) > 0 {
		_ = d.Set("ns_servers", resp.DNSZones[0].NsDefault)
	}
	_ = d.Set("status", resp.Status)
	if resp.ExternalDomainRegistrationStatus != nil && resp.ExternalDomainRegistrationStatus.ValidationToken != "" {
		_ = d.Set("validation_token", resp.ExternalDomainRegistrationStatus.ValidationToken)
	}

	return nil
}

func resourceExternalDomainDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if d.Get("status") == "pending" {
		return diag.Errorf("cannot delete domain with status %s", d.Get("status"))
	}
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
