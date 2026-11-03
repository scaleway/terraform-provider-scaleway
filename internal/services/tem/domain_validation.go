package tem

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/tem_domain_validation_resource.md
var domainValidationResourceDescription string

var (
	_ resource.Resource                = (*DomainValidationResource)(nil)
	_ resource.ResourceWithConfigure   = (*DomainValidationResource)(nil)
	_ resource.ResourceWithImportState = (*DomainValidationResource)(nil)
	_ resource.ResourceWithIdentity    = (*DomainValidationResource)(nil)
)

func NewDomainValidationResource() resource.Resource {
	return &DomainValidationResource{}
}

type DomainValidationResource struct {
	api  *tem.API
	meta *meta.Meta
}

type domainValidationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	DomainID  types.String `tfsdk:"domain_id"`
	Region    types.String `tfsdk:"region"`
	Timeout   types.Int64  `tfsdk:"timeout"`
	Validated types.Bool   `tfsdk:"validated"`
}

type domainValidationResourceIdentityModel = identityfw.RegionalIdentity

func (r *DomainValidationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tem_domain_validation"
}

func (r *DomainValidationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: domainValidationResourceDescription,
		Description:         domainValidationResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The id of domain name used when sending emails.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": regional.SchemaAttributeComputed(),
			"timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(300),
				MarkdownDescription: "Maximum wait time in second before returning an error.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"validated": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Indicates if the domain is verified for email sending",
			},
		},
	}
}

func (r *DomainValidationResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityfw.DefaultRegional()
}

func (r *DomainValidationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DomainValidationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainValidationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, domainUUID, err := regional.ResolveID(plan.DomainID.ValueString(), plan.Region.ValueString(), r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve domain ID", err.Error())

		return
	}

	domain, err := r.api.GetDomain(&tem.GetDomainRequest{
		Region:   region,
		DomainID: domainUUID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get TEM domain", err.Error())

		return
	}

	duration := int(plan.Timeout.ValueInt64())
	timeout := time.Duration(duration) * time.Second

	_ = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		domainCheck, checkErr := r.api.CheckDomain(&tem.CheckDomainRequest{
			Region:   region,
			DomainID: domain.ID,
		})
		if checkErr != nil {
			return retry.RetryableError(checkErr)
		}

		if domainCheck == nil || domainCheck.Status == "pending" || domainCheck.Status == "unchecked" || domainCheck.Status == "autoconfiguring" {
			return retry.RetryableError(errors.New("retry"))
		}

		return nil
	})

	domainCheck, err := r.api.CheckDomain(&tem.CheckDomainRequest{
		Region:   region,
		DomainID: domain.ID,
	})
	if err != nil || domainCheck == nil || domainCheck.Status == "pending" || domainCheck.Status == "unchecked" || domainCheck.Status == "autoconfiguring" {
		resp.Diagnostics.AddError(
			"Domain validation timed out",
			fmt.Sprintf("domain validation did not complete in %d seconds", duration),
		)

		return
	}

	// Refresh validated from GetDomain like SDKv2 Read
	domain, err = r.api.GetDomain(&tem.GetDomainRequest{
		Region:   region,
		DomainID: domainUUID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read TEM domain after validation", err.Error())

		return
	}

	state := domainValidationResourceModel{
		ID:        types.StringValue(regional.NewIDString(region, domainUUID)),
		DomainID:  types.StringValue(regional.NewIDString(region, domainUUID)),
		Region:    types.StringValue(region.String()),
		Timeout:   plan.Timeout,
		Validated: types.BoolValue(domain.Status == "checked"),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(region, domainUUID))...)
}

func (r *DomainValidationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    domainValidationResourceModel
		identity domainValidationResourceIdentityModel
	)

	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	identityAvailable := !resp.Diagnostics.HasError() && !identity.ID.IsNull() && !identity.ID.IsUnknown()

	if !identityAvailable && resp.Diagnostics.HasError() {
		resp.Diagnostics = nil
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		domainUUID string
		region     scw.Region
		err        error
	)

	if identityAvailable {
		region, domainUUID = identityfw.ResolveRegionalIdentity(identity)
	} else {
		region, domainUUID, err = regional.ResolveID(state.ID.ValueString(), state.Region.ValueString(), r.meta.ScwClient())
		if err != nil {
			resp.Diagnostics.AddError("Failed to parse domain validation ID", err.Error())

			return
		}
	}

	domain, err := r.api.GetDomain(&tem.GetDomainRequest{
		Region:   region,
		DomainID: domainUUID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read TEM domain validation", err.Error())

		return
	}

	newState := domainValidationResourceModel{
		ID:        types.StringValue(regional.NewIDString(region, domainUUID)),
		DomainID:  types.StringValue(regional.NewIDString(region, domainUUID)),
		Region:    types.StringValue(region.String()),
		Timeout:   state.Timeout,
		Validated: types.BoolValue(domain.Status == "checked"),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(region, domainUUID))...)
}

func (r *DomainValidationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan domainValidationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DomainValidationResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *DomainValidationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}
