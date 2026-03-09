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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
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
		Identity:      identity.DefaultRegional(),
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

func parseDomainIDForIdentity(domainIDStr string, d *schema.ResourceData, m any) (scw.Region, string, error) {
	region, id, err := regional.ParseID(domainIDStr)
	if err == nil {
		return region, id, nil
	}

	region, err = meta.ExtractRegion(d, m)
	if err != nil {
		return "", "", err
	}

	return region, domainIDStr, nil
}

func ResourceDomainValidationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := temAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	domainIDStr := d.Get("domain_id").(string)
	domainUUID := extractAfterSlash(domainIDStr)

	domain, err := api.GetDomain(&tem.GetDomainRequest{
		Region:   region,
		DomainID: domainUUID,
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

	identityRegion, _, err := parseDomainIDForIdentity(domainIDStr, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := identity.SetRegionalIdentity(d, identityRegion, domainUUID); err != nil {
		return diag.FromErr(err)
	}

	return ResourceDomainValidationRead(ctx, d, m)
}

func ResourceDomainValidationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	region, domainUUID, err := parseDomainIDForIdentity(d.Id(), d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	api, _, err := temAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	domain, err := api.GetDomain(&tem.GetDomainRequest{
		Region:   region,
		DomainID: domainUUID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if err := identity.SetRegionalIdentity(d, region, domainUUID); err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("domain_id", regional.NewIDString(region, domainUUID))
	_ = d.Set("region", string(region))
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
