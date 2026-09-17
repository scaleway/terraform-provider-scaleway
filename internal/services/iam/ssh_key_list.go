package iam

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*SSHKeyListResource)(nil)
	_ list.ListResourceWithConfigure    = (*SSHKeyListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*SSHKeyListResource)(nil)
)

type SSHKeyListResource struct {
	meta   *meta.Meta
	iamAPI *iamSDK.API
}

func (r *SSHKeyListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.iamAPI = iamSDK.NewAPI(meta.ExtractScwClient(m))
}

func NewSSHKeyListResource() list.ListResource {
	return &SSHKeyListResource{}
}

func (r *SSHKeyListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_ids": listscw.ProjectIDsAttribute("Project IDs to filter for"),
			"name":        listscw.NameAttribute("Name of the SSH key to filter for"),
			"disabled": schema.BoolAttribute{
				Description: "Filter SSH keys by disabled status",
				Optional:    true,
			},
		},
	}
}

func (r *SSHKeyListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	sshKeyResource := ResourceSSKKey()

	resp.ProtoV6Schema = translate.Schema(sshKeyResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(sshKeyResource.ProtoIdentitySchema(ctx)())
}

type SSHKeyListResourceModel struct {
	ProjectIDs types.List   `tfsdk:"project_ids"`
	Name       types.String `tfsdk:"name"`
	Disabled   types.Bool   `tfsdk:"disabled"`
}

func (m *SSHKeyListResourceModel) GetProjects() types.List { return m.ProjectIDs }

func (r *SSHKeyListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_ssh_key"
}

func (r *SSHKeyListResource) FetchSSHKeys(ctx context.Context, projectID string, data SSHKeyListResourceModel) ([]*iamSDK.SSHKey, error) {
	request := &iamSDK.ListSSHKeysRequest{
		ProjectID: &projectID,
		Name:      data.Name.ValueStringPointer(),
	}

	if !data.Disabled.IsNull() && !data.Disabled.IsUnknown() {
		disabled := data.Disabled.ValueBool()
		request.Disabled = &disabled
	}

	response, err := r.iamAPI.ListSSHKeys(request, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.SSHKeys, nil
}

func (r *SSHKeyListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data SSHKeyListResourceModel

	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	projects, err := listscw.ExtractProjects(ctx, &data, r.meta)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing projects", "An error was encountered when listing projects: "+err.Error()),
		})

		return
	}

	allSSHKeys, err := listscw.FetchConcurrently(ctx, projects,
		func(ctx context.Context, projectID string) ([]*iamSDK.SSHKey, error) {
			return r.FetchSSHKeys(ctx, projectID, data)
		},
		func(a, b *iamSDK.SSHKey) int {
			return listscw.CompareGlobalProjectItems(a.ProjectID, b.ProjectID, a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing IAM SSH Keys", "Failed to list IAM SSH Keys: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, sshKey := range allSSHKeys {
			result := req.NewListResult(ctx)
			result.DisplayName = sshKey.Name

			sshKeyResource := ResourceSSKKey()
			resourceData := sshKeyResource.Data(&terraform.InstanceState{})

			err = identity.SetGlobalIdentity(resourceData, sshKey.ID)
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

			setSSHKeyState(resourceData, sshKey)

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
