package tem

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/tem_blocked_list_resource.md
var blockedListResourceDescription string

var (
	_ resource.Resource                = (*BlockedListResource)(nil)
	_ resource.ResourceWithConfigure   = (*BlockedListResource)(nil)
	_ resource.ResourceWithImportState = (*BlockedListResource)(nil)
	_ resource.ResourceWithIdentity    = (*BlockedListResource)(nil)
)

func NewBlockedListResource() resource.Resource {
	return &BlockedListResource{}
}

type BlockedListResource struct {
	api  *tem.API
	meta *meta.Meta
}

type blockedListResourceModel struct {
	ID        types.String `tfsdk:"id"`
	DomainID  types.String `tfsdk:"domain_id"`
	Email     types.String `tfsdk:"email"`
	Type      types.String `tfsdk:"type"`
	Reason    types.String `tfsdk:"reason"`
	Region    types.String `tfsdk:"region"`
	ProjectID types.String `tfsdk:"project_id"`
}

func (r *BlockedListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tem_blocked_list"
}

func (r *BlockedListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: blockedListResourceDescription,
		Description:         blockedListResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the domain affected by the blocklist.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Email address to block.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of the blocked list. (mailbox_full or mailbox_not_found)",
				Validators: []validator.String{
					verify.FrameworkValidateEnum[tem.BlocklistType](),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"reason": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("manual_block"),
				MarkdownDescription: "Reason for blocking the emails.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region":     regional.SchemaAttributeComputed(),
			"project_id": account.ProjectIDSchemaAttribute("The project_id you want to attach the resource to"),
		},
	}
}

func (r *BlockedListResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityfw.DefaultRegional()
}

func (r *BlockedListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.meta = m
	r.api = tem.NewAPI(r.meta.ScwClient())
}

func (r *BlockedListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan blockedListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, domainID, err := regional.ResolveID(plan.DomainID.ValueString(), plan.Region.ValueString(), r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve domain ID", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(plan.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	reason := plan.Reason.ValueString()

	createResp, err := r.api.BulkCreateBlocklists(&tem.BulkCreateBlocklistsRequest{
		Emails:   []string{plan.Email.ValueString()},
		Region:   region,
		DomainID: domainID,
		Type:     tem.BlocklistType(plan.Type.ValueString()),
		Reason:   &reason,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create TEM blocked list", err.Error())

		return
	}

	if len(createResp.Blocklists) == 0 {
		resp.Diagnostics.AddError("Failed to create TEM blocked list", "API returned no blocklists")

		return
	}

	blocklist := createResp.Blocklists[0]
	email := plan.Email.ValueString()

	// Mirror SDKv2 Create→Read (List) so VCR cassette interaction count stays aligned.
	blocklists, err := r.api.ListBlocklists(&tem.ListBlocklistsRequest{
		Region:   region,
		Email:    &email,
		DomainID: domainID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read TEM blocked list after create", err.Error())

		return
	}

	if len(blocklists.Blocklists) > 0 {
		blocklist = blocklists.Blocklists[0]
	}

	state := blockedListResourceModel{
		ID:        types.StringValue(regional.NewIDString(region, blocklist.ID)),
		DomainID:  types.StringValue(regional.NewIDString(region, blocklist.DomainID)),
		Email:     types.StringValue(blocklist.Email),
		Type:      types.StringValue(string(blocklist.Type)),
		Reason:    types.StringValue(blocklist.Reason),
		Region:    types.StringValue(region.String()),
		ProjectID: types.StringValue(projectID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(region, blocklist.ID))...)
}

func (r *BlockedListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state blockedListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, domainID, err := regional.ResolveID(state.DomainID.ValueString(), state.Region.ValueString(), r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve domain ID", err.Error())

		return
	}

	email := state.Email.ValueString()

	blocklists, err := r.api.ListBlocklists(&tem.ListBlocklistsRequest{
		Region:   region,
		Email:    &email,
		DomainID: domainID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read TEM blocked list", err.Error())

		return
	}

	if len(blocklists.Blocklists) == 0 {
		resp.State.RemoveResource(ctx)

		return
	}

	blocklist := blocklists.Blocklists[0]

	newState := blockedListResourceModel{
		ID:        types.StringValue(regional.NewIDString(region, blocklist.ID)),
		DomainID:  types.StringValue(regional.NewIDString(region, blocklist.DomainID)),
		Email:     types.StringValue(blocklist.Email),
		Type:      types.StringValue(string(blocklist.Type)),
		Reason:    types.StringValue(blocklist.Reason),
		Region:    types.StringValue(region.String()),
		ProjectID: state.ProjectID,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(region, blocklist.ID))...)
}

func (r *BlockedListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan blockedListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BlockedListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state blockedListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse blocked list ID", err.Error())

		return
	}

	err = r.api.DeleteBlocklist(&tem.DeleteBlocklistRequest{
		Region:      region,
		BlocklistID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed to delete TEM blocked list", err.Error())
	}
}

func (r *BlockedListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}
