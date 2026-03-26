package domain

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceExternalDomainValidated() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExternalDomainValidatedCreate,
		ReadContext:   resourceExternalDomainRead,
		DeleteContext: resourceExternalDomainValidatedDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultDomainRegistrationTimeout),
		},

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
			"organization_id": {
				Type:        schema.TypeString,
				Description: "The organization ID the domain is associated with.",
				Computed:    true,
			},
			"ns_servers": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of Ns servers for the domain.",
				Computed:    true,
			},
		},
	}
}

func resourceExternalDomainValidatedCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)
	domainName := d.Get("domain").(string)
	_, err := waitForExternalDomainValidation(ctx, registrarAPI, domainName, defaultDomainRegistrationTimeout)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(domainName)
	return resourceExternalDomainRead(ctx, d, m)
}

func resourceExternalDomainValidatedDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	d.SetId("")

	return nil
}
