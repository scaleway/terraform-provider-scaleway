package tem

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func ResourceDomainValidation() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceDomainValidationCreate,
		ReadContext:   ResourceDomainValidationRead,
		DeleteContext: ResourceDomainValidationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultDomainValidationTimeout),
			Delete:  schema.DefaultTimeout(defaultDomainValidationTimeout),
			Default: schema.DefaultTimeout(defaultDomainValidationTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    domainValidationSchema,
		Identity: identity.WrapSchemaMap(map[string]*schema.Schema{
			"domain_id": {
				Type:              schema.TypeString,
				Description:       "The ID of the domain to validate.",
				RequiredForImport: true,
			},
		}),
	}
}

func domainValidationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"domain_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The id of domain name used when sending emails.",
		},
		"region": regional.Schema(),
		"timeout": {
			Type:        schema.TypeInt,
			Optional:    true,
			ForceNew:    true,
			Default:     300,
			Description: "Maximum wait time in second before returning an error.",
		},
		"validated": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Indicates if the domain is verified for email sending",
		},
	}
}

func ResourceDomainValidationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := temAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(d.Get("domain_id").(string))

	domain, err := api.GetDomain(&tem.GetDomainRequest{
		Region:   region,
		DomainID: extractAfterSlash(d.Get("domain_id").(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	duration := d.Get("timeout").(int)
	timeout := time.Duration(duration) * time.Second
	_ = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		domainCheck, _ := api.CheckDomain(&tem.CheckDomainRequest{
			Region:   region,
			DomainID: domain.ID,
		})
		if domainCheck == nil || domainCheck.Status == "pending" || domainCheck.Status == "unchecked" || domainCheck.Status == "autoconfiguring" {
			return retry.RetryableError(errors.New("retry"))
		}

		return nil
	})

	domainCheck, _ := api.CheckDomain(&tem.CheckDomainRequest{
		Region:   region,
		DomainID: domain.ID,
	})
	if domainCheck == nil || domainCheck.Status == "pending" || domainCheck.Status == "unchecked" || domainCheck.Status == "autoconfiguring" {
		d.SetId("")

		return diag.Errorf("domain validation did not complete in %d seconds", duration)
	}

	return ResourceDomainValidationRead(ctx, d, meta)
}

func ResourceDomainValidationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := temAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	domainID := d.Id()
	getDomainRequest := &tem.GetDomainRequest{
		Region:   region,
		DomainID: extractAfterSlash(domainID),
	}

	domain, err := api.GetDomain(getDomainRequest, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("validated", domain.Status == "checked")

	return nil
}

func ResourceDomainValidationDelete(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")

	return nil
}

func extractAfterSlash(s string) string {
	lastIndex := strings.LastIndex(s, "/")
	if lastIndex == -1 {
		return s
	}

	return s[lastIndex+1:]
}
