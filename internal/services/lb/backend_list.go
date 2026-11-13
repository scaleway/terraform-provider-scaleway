package lb

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*BackendListResource)(nil)
	_ list.ListResourceWithConfigure    = (*BackendListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*BackendListResource)(nil)
)

type BackendListResource struct {
	meta  *meta.Meta
	lbAPI *lbSDK.ZonedAPI
}

func (r *BackendListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.lbAPI = lbSDK.NewZonedAPI(meta.ExtractScwClient(m))
}

func NewBackendListResource() list.ListResource {
	return &BackendListResource{}
}

func (r *BackendListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"lb_ids": schema.ListAttribute{
				Description: "Load Balancer IDs to list backends for.",
				Required:    true,
				ElementType: types.StringType,
			},
			"name":  listscw.NameAttribute("Name of the backend to filter for"),
			"zones": listscw.ZonesAttribute("Zones to filter for."),
		},
	}
}

func (r *BackendListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	backendResource := ResourceBackend()

	resp.ProtoV6Schema = translate.Schema(backendResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(backendResource.ProtoIdentitySchema(ctx)())
}

type BackendListResourceModel struct {
	LBIDs types.List   `tfsdk:"lb_ids"`
	Zones types.List   `tfsdk:"zones"`
	Name  types.String `tfsdk:"name"`
}

func (m *BackendListResourceModel) GetZones() types.List { return m.Zones }

func (r *BackendListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_backend"
}

type backendListTarget struct {
	zone scw.Zone
	lbID string
}

func (r *BackendListResource) FetchBackends(ctx context.Context, target backendListTarget, name *string) ([]*lbSDK.Backend, error) {
	response, err := r.lbAPI.ListBackends(&lbSDK.ZonedAPIListBackendsRequest{
		Zone: target.zone,
		LBID: target.lbID,
		Name: name,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		if httperrors.Is404(err) {
			return nil, nil
		}

		return nil, err
	}

	return response.Backends, nil
}

func (r *BackendListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data BackendListResourceModel

	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	zones, err := listscw.ExtractZones(ctx, &data, r.meta)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing zones", "An error was encountered when listing zones: "+err.Error()),
		})

		return
	}

	lbIDs, diags := locality.ExpandFrameworkIDs(ctx, data.LBIDs)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	targets := make([]backendListTarget, 0, len(zones)*len(lbIDs))
	for _, z := range zones {
		for _, id := range lbIDs {
			targets = append(targets, backendListTarget{zone: z, lbID: id})
		}
	}

	name := data.Name.ValueStringPointer()

	allBackends, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target backendListTarget) ([]*lbSDK.Backend, error) {
			return r.FetchBackends(ctx, target, name)
		},
		func(a, b *lbSDK.Backend) int {
			if a.LB.Zone != b.LB.Zone {
				return strings.Compare(string(a.LB.Zone), string(b.LB.Zone))
			}

			if a.LB.ID != b.LB.ID {
				return strings.Compare(a.LB.ID, b.LB.ID)
			}

			return strings.Compare(a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing LB Backends", "Failed to list LB Backends: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, backend := range allBackends {
			result := req.NewListResult(ctx)
			result.DisplayName = backend.Name

			backendResource := ResourceBackend()
			resourceData := backendResource.Data(&terraform.InstanceState{})

			err = identity.SetZonalIdentity(resourceData, backend.LB.Zone, backend.ID)
			if err != nil {
				result.Diagnostics.AddError("Retrieving identity data",
					"An error was encountered when retrieving the identity data: "+err.Error(),
				)

				if !push(result) {
					return
				}

				continue
			}

			tfTypeIdentity, errIdentityState := resourceData.TfTypeIdentityState()
			if errIdentityState != nil {
				result.Diagnostics.AddError(
					"Converting identity data",
					"An error was encountered when converting the identity data: "+errIdentityState.Error(),
				)
			}

			identitySetDiags := result.Identity.Set(ctx, *tfTypeIdentity)
			result.Diagnostics.Append(identitySetDiags...)

			sdkDiags := setBackendState(resourceData, backend, backend.LB.Zone)
			if sdkDiags.HasError() {
				tflog.Error(ctx, "error from setting backend state")

				for _, d := range sdkDiags {
					result.Diagnostics.AddError(d.Summary, d.Detail)
				}

				if !push(result) {
					return
				}

				continue
			}

			tfTypeResource, errTfTypeResourceState := resourceData.TfTypeResourceState()
			if errTfTypeResourceState != nil {
				result.Diagnostics.AddError(
					"Converting resource state",
					"An error was encountered when converting the resource state: "+errTfTypeResourceState.Error(),
				)
			}

			resourceSetDiags := result.Resource.Set(ctx, *tfTypeResource)
			result.Diagnostics.Append(resourceSetDiags...)

			if !push(result) {
				return
			}
		}
	}
}
