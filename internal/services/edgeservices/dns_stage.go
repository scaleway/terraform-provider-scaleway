package edgeservices

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	edgeservices "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceDNSStage() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceDNSStageCreate,
		ReadContext:   ResourceDNSStageRead,
		UpdateContext: ResourceDNSStageUpdate,
		DeleteContext: ResourceDNSStageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    dnsStageSchema,
		Identity: identity.WrapSchemaMap(map[string]*schema.Schema{
			"dns_stage_id": {
				Type:              schema.TypeString,
				Description:       "DNS stage id (UUID format)",
				RequiredForImport: true,
			},
		}),
	}
}

func dnsStageSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"pipeline_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ID of the pipeline",
		},
		"backend_stage_id": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			Description:   "The backend stage ID the DNS stage will be linked to",
			ConflictsWith: []string{"cache_stage_id", "tls_stage_id"},
		},
		"tls_stage_id": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			Description:   "The TLS stage ID the DNS stage will be linked to",
			ConflictsWith: []string{"cache_stage_id", "backend_stage_id"},
		},
		"cache_stage_id": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			Description:   "The cache stage ID the DNS stage will be linked to",
			ConflictsWith: []string{"backend_stage_id", "tls_stage_id"},
		},
		"fqdns": {
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Description: "Fully Qualified Domain Name (in the format subdomain.example.com) to attach to the stage",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The type of the stage",
		},
		"default_fqdn": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Default Fully Qualified Domain Name attached to the stage",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the DNS stage",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the DNS stage",
		},
		"project_id": account.ProjectIDSchema(),
	}
}

func ResourceDNSStageCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	dnsStage, err := api.CreateDNSStage(&edgeservices.CreateDNSStageRequest{
		PipelineID:     d.Get("pipeline_id").(string),
		BackendStageID: types.ExpandStringPtr(d.Get("backend_stage_id").(string)),
		CacheStageID:   types.ExpandStringPtr(d.Get("cache_stage_id").(string)),
		TLSStageID:     types.ExpandStringPtr(d.Get("tls_stage_id").(string)),
		Fqdns:          types.ExpandStringsPtr(d.Get("fqdns")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetFlatIdentity(d, "dns_stage_id", dnsStage.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceDNSStageRead(ctx, d, m)
}

func ResourceDNSStageRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	dnsStage, err := api.GetDNSStage(&edgeservices.GetDNSStageRequest{
		DNSStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("backend_stage_id", types.FlattenStringPtr(dnsStage.BackendStageID))
	_ = d.Set("cache_stage_id", types.FlattenStringPtr(dnsStage.CacheStageID))
	_ = d.Set("pipeline_id", dnsStage.PipelineID)
	_ = d.Set("tls_stage_id", types.FlattenStringPtr(dnsStage.TLSStageID))
	_ = d.Set("created_at", types.FlattenTime(dnsStage.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(dnsStage.UpdatedAt))
	_ = d.Set("type", dnsStage.Type.String())
	_ = d.Set("default_fqdn", dnsStage.DefaultFqdn)

	oldFQDNs := d.Get("fqdns").([]any)
	oldFQDNsSet := make(map[string]bool)

	for _, fqdn := range oldFQDNs {
		oldFQDNsSet[fqdn.(string)] = true
	}

	newFQDNs := make([]string, 0)
	// add all FQDNs from the API response
	for _, fqdn := range dnsStage.Fqdns {
		if oldFQDNsSet[fqdn] || len(oldFQDNs) == 0 {
			// keep FQDNs that were in the old state or if there were no old FQDNs
			newFQDNs = append(newFQDNs, fqdn)
		}
	}
	// add any FQDNs from the old state that aren't in the API response
	for _, oldFQDN := range oldFQDNs {
		found := false

		for _, newFQDN := range newFQDNs {
			if oldFQDN.(string) == newFQDN {
				found = true

				break
			}
		}

		if !found {
			newFQDNs = append(newFQDNs, oldFQDN.(string))
		}
	}

	if err = d.Set("fqdns", newFQDNs); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourceDNSStageUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	hasChanged := false

	updateRequest := &edgeservices.UpdateDNSStageRequest{
		DNSStageID: d.Id(),
	}

	if d.HasChange("backend_stage_id") {
		updateRequest.BackendStageID = types.ExpandUpdatedStringPtr(d.Get("backend_stage_id"))
		hasChanged = true
	}

	if d.HasChange("cache_stage_id") {
		updateRequest.CacheStageID = types.ExpandUpdatedStringPtr(d.Get("cache_stage_id"))
		hasChanged = true
	}

	if d.HasChange("tls_stage_id") {
		updateRequest.TLSStageID = types.ExpandUpdatedStringPtr(d.Get("tls_stage_id"))
		hasChanged = true
	}

	if d.HasChange("fqdns") {
		updateRequest.Fqdns = types.ExpandUpdatedStringsPtr(d.Get("fqdns"))
		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateDNSStage(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceDNSStageRead(ctx, d, m)
}

func ResourceDNSStageDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	err := api.DeleteDNSStage(&edgeservices.DeleteDNSStageRequest{
		DNSStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
