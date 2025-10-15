package secret

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	diagFramework "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = &ResourceSecret{}
	_ resource.ResourceWithConfigure   = &ResourceSecret{}
	_ resource.ResourceWithIdentity    = &ResourceSecret{}
	_ resource.ResourceWithImportState = &ResourceSecret{}
)

type ResourceSecret struct {
	secretAPI *secret.API
}

func NewResourceSecret() resource.Resource {
	return &ResourceSecret{}
}

func (r *ResourceSecret) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ResourceSecret) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}
	m, ok := request.ProviderData.(*meta.Meta)
	if !ok {
		response.Diagnostics.AddError(
			"Cannot get meta from provider",
			"cannot get meta from provider",
		)

		return
	}

	client := m.ScwClient()
	r.secretAPI = secret.NewAPI(client)
}

type ResourceSecretIdentityModel struct {
	ID     types.String `tfsdk:"id"`
	Region types.String `tfsdk:"region"`
}

func (r *ResourceSecret) IdentitySchema(ctx context.Context, request resource.IdentitySchemaRequest, response *resource.IdentitySchemaResponse) {
	response.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "ID of the secret",
			},
			"region": identityschema.StringAttribute{
				OptionalForImport: true,
				Description:       "Region of the secret",
			},
		},
	}
}

func (r *ResourceSecret) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_secret"
}

// useStateForUnknownModifier implements the plan modifier.
type cleanUpFilePath struct{}

// Description returns a human-readable description of the plan modifier.
func (m cleanUpFilePath) Description(_ context.Context) string {
	return "Remove diff "
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m cleanUpFilePath) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m cleanUpFilePath) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is no state (resource is being created)
	// or if we're not in a refresh operation (since PlanOnly tests check for no changes)
	if req.State.Raw.IsNull() || !req.PlanValue.IsUnknown() {
		return
	}

	// For refresh operations, we want to ensure the state matches the API
	// The API likely normalizes paths, so we should clean them
	cleanedPath := filepath.Clean(req.StateValue.String())

	// Special case: if the cleaned path is empty or root, ensure it's "/"
	if cleanedPath == "." {
		cleanedPath = "/"
	}

	resp.PlanValue = types.StringValue(cleanedPath)
}

func (r *ResourceSecret) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description:         "Schema for secret",
		MarkdownDescription: "Schema for secret",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Secret ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The secret name",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the secret",
				Default:     stringdefault.StaticString(""),
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Status of the secret",
			},
			"version_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of versions for this Secret",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time of secret's creation (RFC 3339 format)",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time of secret's creation (RFC 3339 format)",
			},
			"path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Location of the secret in the directory structure.",
				Default:     stringdefault.StaticString("/"),
				PlanModifiers: []planmodifier.String{
					cleanUpFilePath{},
				},
			},
			"protected": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "True if secret protection is enabled on a given secret. A protected secret cannot be deleted.",
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of tags [\"tag1\", \"tag2\", ...] associated to secret",
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"type": schema.StringAttribute{
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: func() string {
					var t secret.SecretType

					secretTypes := t.Values()

					return fmt.Sprintf("Type of the secret could be any value among: %s", secretTypes)
				}(),
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(secret.SecretTypeOpaque.String()),
				Validators: []validator.String{
					verify.ValidatorFromEnum[secret.SecretType](secret.SecretType("")),
				},
			},
			"region":     regional.ResourceSchema("The region you want to attach the resource to"),
			"project_id": account.ResourceProjectIDSchema("The project ID you want to attach the secret to"),
			//"ephemeral_policy": schema.ListNestedAttribute{
			//	Description: "Ephemeral policy of the secret. Policy that defines whether/when a secret's versions expire. By default, the policy is applied to all the secret's versions.",
			//	Optional:    true,
			//	NestedObject: schema.NestedAttributeObject{
			//		Attributes: map[string]schema.Attribute{
			//			"ttl": schema.StringAttribute{
			//				Optional:    true,
			//				Description: "Time frame, from one second and up to one year, during which the secret's versions are valid. Has to be specified in Go Duration format",
			//				Validators: []validator.String{
			//					verify.DurationValidator{},
			//				},
			//				PlanModifiers: []planmodifier.String{
			//					planModifiers.Duration{},
			//				},
			//			},
			//			"expires_once_accessed": schema.BoolAttribute{
			//				Optional:    true,
			//				Description: "True if the secret version expires after a single user access.",
			//			},
			//			"action": schema.StringAttribute{
			//				Required: true,
			//				Validators: []validator.String{
			//					verify.ValidatorFromEnum[secret.EphemeralPolicyAction](secret.EphemeralPolicyAction("")),
			//				},
			//				Description: "Action to perform when the version of a secret expires.",
			//			},
			//		},
			//	},
			//},
			//"versions": schema.ListNestedAttribute{
			//	Description: "List of the versions of the secret",
			//	Computed:    true,
			//	NestedObject: schema.NestedAttributeObject{
			//		Attributes: map[string]schema.Attribute{
			//			"revision": schema.StringAttribute{
			//				Computed:    true,
			//				Description: "The revision of secret version",
			//			},
			//			"secret_id": schema.StringAttribute{
			//				Computed:    true,
			//				Description: "The secret ID associated with this version",
			//			},
			//			"status": schema.StringAttribute{
			//				Computed:    true,
			//				Description: "Status of the secret version",
			//			},
			//			"created_at": schema.StringAttribute{
			//				Computed:    true,
			//				Description: "Date and time of secret version's creation (RFC 3339 format)",
			//			},
			//			"updated_at": schema.StringAttribute{
			//				Computed:    true,
			//				Description: "Date and time of secret version's creation (RFC 3339 format)",
			//			},
			//			"description": schema.StringAttribute{
			//				Optional:    true,
			//				Description: "Description of the secret version",
			//			},
			//			"latest": schema.BoolAttribute{
			//				Optional:    true,
			//				Description: "Returns true if the version is the latest.",
			//			},
			//		},
			//	},
			//},
		},
		//Blocks: map[string]schema.Block{
		//	"timeouts": timeouts.Block(ctx,
		//		timeouts.Opts{
		//			Create:            true,
		//			CreateDescription: "Timeout to apply on Create",
		//			Read:              true,
		//			ReadDescription:   "Timeout to apply on Read",
		//			Update:            true,
		//			UpdateDescription: "Timeout to apply on Update",
		//			Delete:            true,
		//			DeleteDescription: "Timeout to apply on Delete",
		//		},
		//	),
		//},
	}
}

type ResourceSecretModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Protected   types.Bool   `tfsdk:"protected"`
	Type        types.String `tfsdk:"type"`
	Tags        types.List   `tfsdk:"tags"`
	Description types.String `tfsdk:"description"`
	Path        types.String `tfsdk:"path"`
	// EphemeralPolicy types.String      `tfsdk:"ephemeral_policy"`
	Region       types.String `tfsdk:"region"`
	ProjectID    types.String `tfsdk:"project_id"`
	VersionCount types.Int32  `tfsdk:"version_count"`
	Status       types.String `tfsdk:"status"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	// Versions     types.List     `tfsdk:"versions"`
	// Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func NewModelFromSecret(ctx context.Context, s secret.Secret) (*ResourceSecretModel, diagFramework.Diagnostics) {
	model := &ResourceSecretModel{
		Name:         types.StringValue(s.Name),
		Status:       types.StringValue(s.Status.String()),
		ProjectID:    types.StringValue(s.ProjectID),
		Protected:    types.BoolValue(s.Protected),
		Description:  types.StringPointerValue(s.Description),
		CreatedAt:    types.StringValue(s.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:    types.StringValue(s.UpdatedAt.Format(time.RFC3339)),
		Type:         types.StringValue(s.Type.String()),
		VersionCount: types.Int32Value(int32(s.VersionCount)),
		ID:           types.StringValue(regional.NewIDString(s.Region, s.ID)),
		Region:       types.StringValue(s.Region.String()),
		Path:         types.StringValue(s.Path),
	}

	tags, diags := types.ListValueFrom(ctx, types.StringType, s.Tags)
	model.Tags = tags

	return model, diags
}

func (r *ResourceSecret) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data ResourceSecretModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	secretCreateRequest := &secret.CreateSecretRequest{
		Name:      data.Name.ValueString(),
		Protected: data.Protected.ValueBool(),
	}

	if !data.Type.IsNull() {
		secretCreateRequest.Type = secret.SecretType(data.Type.ValueString())
	}

	if !data.Region.IsNull() {
		secretCreateRequest.Region = scw.Region(data.Region.ValueString())
	}

	if !data.ProjectID.IsNull() {
		secretCreateRequest.ProjectID = data.ProjectID.ValueString()
	}

	if !data.Tags.IsNull() {
		elements := make([]string, 0, len(data.Tags.Elements()))
		diags := data.Tags.ElementsAs(ctx, &elements, false)
		response.Diagnostics.Append(diags...)
		secretCreateRequest.Tags = elements
	}

	if !data.Description.IsNull() {
		secretCreateRequest.Description = data.Description.ValueStringPointer()
	}

	if !data.Path.IsNull() {
		secretCreateRequest.Path = data.Path.ValueStringPointer()
	}

	//if !data.EphemeralPolicy.IsNull() {
	//	secretCreateRequest.EphemeralPolicy = data.EphemeralPolicy.String()
	//}
	//rawEphemeralPolicy, policyExists := d.GetOk("ephemeral_policy")
	//if policyExists {
	//	secretCreateRequest.EphemeralPolicy, err = expandEphemeralPolicy(rawEphemeralPolicy)
	//	if err != nil {
	//		return diag.FromErr(err)
	//	}
	//}

	apiResponse, err := r.secretAPI.CreateSecret(secretCreateRequest, scw.WithContext(ctx))
	if err != nil {
		response.Diagnostics.AddError(
			"error while creating secret",
			err.Error(),
		)

		return
	}
	if apiResponse == nil {
		response.Diagnostics.AddError(
			"nil answer while creating secret",
			"nil answer while creating secret",
		)

		return
	}

	// Set data returned by API in identity
	identity := ResourceSecretIdentityModel{
		ID:     types.StringValue(apiResponse.ID),
		Region: types.StringValue(apiResponse.Region.String()),
	}
	response.Diagnostics.Append(response.Identity.Set(ctx, identity)...)

	// Save data into Terraform state
	dataToSave, _ := NewModelFromSecret(ctx, *apiResponse)

	response.Diagnostics.Append(response.State.Set(ctx, &dataToSave)...)
}

func (r *ResourceSecret) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data ResourceSecretModel

	// Read Terraform prior state data into the model
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	secretResponse, err := r.secretAPI.GetSecret(&secret.GetSecretRequest{
		Region:   scw.Region(data.Region.ValueString()),
		SecretID: locality.ExpandID(data.ID.ValueString()),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			response.State.RemoveResource(ctx)

			return
		}

		response.Diagnostics.AddError(
			"cannot get secret",
			err.Error())

		return
	}

	//if len(secretResponse.Tags) > 0 {
	//	_ = d.Set("tags", types.FlattenSliceString(secretResponse.Tags))
	//}

	//versions, err := r.secretAPI.ListSecretVersions(&secret.ListSecretVersionsRequest{
	//	Region:   r.region,
	//	SecretID: r.id,
	//}, scw.WithAllPages(), scw.WithContext(ctx))
	//if err != nil {
	//	if httperrors.Is404(err) {
	//		d.SetId("")
	//
	//		return nil
	//	}
	//
	//	return diag.FromErr(err)
	//}

	// data.VersionCount = types.Int32Value(int32(versions.TotalCount))

	//_ = d.Set("ephemeral_policy", flattenEphemeralPolicy(secretResponse.EphemeralPolicy))

	//versionsList := make([]map[string]any, 0, len(versions.Versions))
	//for _, version := range versions.Versions {
	//	versionsList = append(versionsList, map[string]any{
	//		"revision":    strconv.Itoa(int(version.Revision)),
	//		"secret_id":   version.SecretID,
	//		"status":      version.Status.String(),
	//		"created_at":  types.FlattenTime(version.CreatedAt),
	//		"updated_at":  types.FlattenTime(version.UpdatedAt),
	//		"description": types.FlattenStringPtr(version.Description),
	//		"latest":      types.FlattenBoolPtr(&version.Latest),
	//	})
	//}
	//
	//_ = d.Set("versions", versionsList)

	// Save updated data into Terraform state
	dataToSave, _ := NewModelFromSecret(ctx, *secretResponse)
	response.Diagnostics.Append(response.State.Set(ctx, &dataToSave)...)
}

func (r *ResourceSecret) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state ResourceSecretModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	updateRequest := &secret.UpdateSecretRequest{
		Region:   scw.Region(plan.Region.ValueString()),
		SecretID: locality.ExpandID(plan.ID.ValueString()),
	}

	hasChanged := false

	if !plan.Description.Equal(state.Description) {
		desc := plan.Description.ValueString()
		updateRequest.Description = &desc
		hasChanged = true
	}

	if !plan.Name.Equal(state.Name) {
		updateRequest.Name = plan.Name.ValueStringPointer()
		hasChanged = true
	}

	if !plan.Tags.Equal(state.Tags) {
		updateRequest.Tags = &[]string{}

		elements := make([]string, 0, len(plan.Tags.Elements()))
		diags := plan.Tags.ElementsAs(ctx, &elements, false)
		response.Diagnostics.Append(diags...)

		if elements != nil {
			updateRequest.Tags = &elements
		}

		hasChanged = true
	}

	if !plan.Path.Equal(state.Path) {
		updateRequest.Path = plan.Path.ValueStringPointer()
		hasChanged = true
	}

	//if d.HasChange("ephemeral_policy") {
	//	updateRequest.EphemeralPolicy, err = expandEphemeralPolicy(d.Get("ephemeral_policy"))
	//	if err != nil {
	//		return diag.FromErr(err)
	//	}
	//	hasChanged = true
	//}

	if hasChanged {
		secretResponse, err := r.secretAPI.UpdateSecret(updateRequest, scw.WithContext(ctx))
		if err != nil {
			response.Diagnostics.AddError(
				"unable to update secret",
				err.Error(),
			)

			return
		}

		dataToSave, _ := NewModelFromSecret(ctx, *secretResponse)

		// Save updated data into Terraform state
		response.Diagnostics.Append(response.State.Set(ctx, &dataToSave)...)
	}

	//if !plan.Protected.Equal(state.Protected) {
	//	s, err := r.secretAPI.GetSecret(&secret.GetSecretRequest{
	//		Region:   r.region,
	//		SecretID: r.ID,
	//	})
	//	if err != nil {
	//		response.Diagnostics.AddError(
	//			"error while trying to change protection of a secret",
	//			err.Error(),
	//		)
	//		return
	//	}
	//
	//	if s.Protected == protected {
	//		return nil
	//	}
	//
	//	if protected {
	//		_, err = r.secretAPI.ProtectSecret(&secret.ProtectSecretRequest{
	//			Region:   r.region,
	//			SecretID: r.ID,
	//		})
	//		if err != nil {
	//			return fmt.Errorf("failed to protect secret %s: %w", secretID, err)
	//		}
	//	} else {
	//		_, err = r.secretAPI.UnprotectSecret(&secret.UnprotectSecretRequest{
	//			Region:   r.region,
	//			SecretID: r.ID,
	//		})
	//		if err != nil {
	//			return fmt.Errorf("failed to unprotect secret %s: %w", secretID, err)
	//		}
	//	}
	//}
}

func (r *ResourceSecret) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data ResourceSecretModel

	// Read Terraform prior state data into the model
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := r.secretAPI.DeleteSecret(&secret.DeleteSecretRequest{
		Region:   scw.Region(data.Region.ValueString()),
		SecretID: data.ID.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		response.Diagnostics.AddError(
			"Unable to delete secret",
			err.Error())
	}
}
