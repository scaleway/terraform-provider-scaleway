package iam

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	annotations "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	scw_ephemeral "github.com/scaleway/terraform-provider-scaleway/v2/internal/ephemeral"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_                         ephemeral.EphemeralResource              = (*ApiKeyEphemeralResource)(nil)
	_                         ephemeral.EphemeralResourceWithConfigure = (*ApiKeyEphemeralResource)(nil)
	_                         ephemeral.EphemeralResourceWithClose     = (*ApiKeyEphemeralResource)(nil)
	iamTerraformAnnotationKey                                          = "iam_terraform_identifier"
)

type ApiKeyEphemeralResource struct {
	iamAPI           *iam.API
	identifierClient *scw_ephemeral.ResourceIdentifierManager[iam.APIKey]
	meta             *meta.Meta
}

func NewApiKeyEphemeralResource() ephemeral.EphemeralResource {
	return &ApiKeyEphemeralResource{}
}

func (r *ApiKeyEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Ephemeral Resource Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	client := m.ScwClient()
	r.iamAPI = iam.NewAPI(client)
	annotationsAPI := annotations.NewAPI(client)
	r.identifierClient = scw_ephemeral.NewResourceIdentifierManager(scw_ephemeral.ResourceIdentifierManagerConfig[iam.APIKey]{
		AnnotationsAPI:  annotationsAPI,
		ResourceHandler: NewIAMResourceHandler(r.iamAPI),
		AnnotationKey:   iamTerraformAnnotationKey,
	})
	r.meta = m
}

func (r *ApiKeyEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_api_key"
}

type ApiKeyEphemeralResourceModel struct {
	AccessKey             types.String `tfsdk:"access_key"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
	ExpiresAt             types.String `tfsdk:"expires_at"`
	ApplicationID         types.String `tfsdk:"application_id"`
	UserID                types.String `tfsdk:"user_id"`
	AnnotationIdentifier  types.String `tfsdk:"annotation_identifier"`
	DescriptionIdentifier types.String `tfsdk:"description_identifier"`
	Description           types.String `tfsdk:"description"`
	SecretKey             types.String `tfsdk:"secret_key"`
	CreationIP            types.String `tfsdk:"creation_ip"`
	DefaultProjectID      types.String `tfsdk:"default_project_id"`
	EphemeralLifecycle    types.String `tfsdk:"ephemeral_lifecycle"`
}

//go:embed descriptions/api_key_ephemeral_resource.md
var apiKeyEphemeralResourceDescription string

func (r *ApiKeyEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         apiKeyEphemeralResourceDescription,
		MarkdownDescription: apiKeyEphemeralResourceDescription,
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The description of the iam api key. Conflicts with `annotation_identifier` and `description_identifier`.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("annotation_identifier"), path.MatchRoot("description_identifier")),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time of the creation of the iam api key",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time of the last update of the iam api key",
			},
			"expires_at": schema.StringAttribute{
				Description: "The date and time (UTC) of the expiration of the iam api key. Cannot be changed afterwards",
				Optional:    true,
			},
			"access_key": schema.StringAttribute{
				Computed:    true,
				Description: "The access key of the iam api key",
			},
			"secret_key": schema.StringAttribute{
				Computed:    true,
				Description: "The secret Key of the iam api key",
				Sensitive:   true,
			},
			"application_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the application attached to the api key",
				Validators: []validator.String{
					verify.IsStringUUID(),
					stringvalidator.ConflictsWith(path.MatchRoot("user_id")),
				},
			},
			"user_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the user attached to the api key",
				Validators: []validator.String{
					verify.IsStringUUID(),
					stringvalidator.ConflictsWith(path.MatchRoot("application_id")),
				},
			},
			"creation_ip": schema.StringAttribute{
				Computed:    true,
				Description: "The IPv4 Address of the device which created the API key",
			},
			"default_project_id": schema.StringAttribute{
				Description: "Default Project ID to use with Object Storage.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"ephemeral_lifecycle": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Controls the lifecycle behavior of the ephemeral API key. `persist` (default): the API key and its annotations are not deleted when the ephemeral resource is closed. `delete`: the API key and its annotations are deleted when the ephemeral resource is closed. `replace`: any existing API key with the same identifier is deleted and a new one is created. Requires either `annotation_identifier` or `description_identifier` to be set.",
				Validators: []validator.String{
					stringvalidator.OneOf("persist", "delete", "replace"),
					verify.StringAlsoRequiresOneOf(path.MatchRoot("annotation_identifier"), path.MatchRoot("description_identifier")),
				},
			},
			"annotation_identifier": schema.StringAttribute{
				Optional:    true,
				Description: "String value used as identifier. Must be a unique string (e.g., UUID v7) to identify the resource. This value is stored as an annotation with key 'iam_terraform_identifier'. Conflicts with `description` and `description_identifier`.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("description"), path.MatchRoot("description_identifier")),
				},
			},
			"description_identifier": schema.StringAttribute{
				Optional:    true,
				Description: "Unique description used as identifier. Must be a unique string (e.g., UUID v7) to identify the resource. Conflicts with `description` and `annotation_identifier`.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("description"), path.MatchRoot("annotation_identifier")),
				},
			},
		},
	}
}

func (r *ApiKeyEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data ApiKeyEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if r.iamAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured iamAPI",
			"The ephemeral resource was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	var (
		existingAPIKey        *iam.APIKey
		identifierDescription string
		err                   error
	)

	orgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
	if !exists {
		resp.Diagnostics.AddAttributeError(
			path.Root("organization_id"),
			"Organization ID is required",
			"Configure a default organization",
		)

		return
	}

	hasAnnotationIdentifier := !data.AnnotationIdentifier.IsNull() && !data.AnnotationIdentifier.IsUnknown()
	hasDescriptionIdentifier := !data.DescriptionIdentifier.IsNull() && !data.DescriptionIdentifier.IsUnknown()

	if hasAnnotationIdentifier {
		annotationValue := data.AnnotationIdentifier.ValueString()

		if annotationValue == "" {
			resp.Diagnostics.AddError(
				"Invalid annotation_identifier",
				"annotation_identifier must be a non-empty string",
			)

			return
		}

		existingAPIKey, err = r.identifierClient.FindResourceByAnnotation(ctx, annotationValue, orgID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error finding existing API key by annotations",
				err.Error(),
			)

			return
		}

		// TODO: annotation should be compatible with description and not set here
		// Use the annotation value as description for the new key
		identifierDescription = "annotation_identifier:" + annotationValue
	}

	if hasDescriptionIdentifier {
		descriptionIdentifier := data.DescriptionIdentifier.ValueString()

		existingAPIKey, err = r.identifierClient.FindResourceByDescription(ctx, descriptionIdentifier, orgID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error finding existing API key by description",
				err.Error(),
			)

			return
		}

		identifierDescription = descriptionIdentifier
	}

	lifecycle := "persist"
	if !data.EphemeralLifecycle.IsNull() && !data.EphemeralLifecycle.IsUnknown() {
		lifecycle = data.EphemeralLifecycle.ValueString()
	}

	shouldReplace := lifecycle == "replace"
	shouldCreateNew := shouldReplace || existingAPIKey == nil

	if shouldCreateNew {
		createReq := iam.CreateAPIKeyRequest{
			ApplicationID:    data.ApplicationID.ValueStringPointer(),
			UserID:           data.UserID.ValueStringPointer(),
			DefaultProjectID: data.DefaultProjectID.ValueStringPointer(),
			Description:      identifierDescription,
		}

		if !data.ExpiresAt.IsNull() && !data.ExpiresAt.IsUnknown() && data.ExpiresAt.ValueString() != "" {
			parsedExpiresAt, err := time.Parse(time.RFC3339, data.ExpiresAt.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Invalid expires_at value",
					fmt.Sprintf("The expires_at attribute must be a valid RFC3339 timestamp. Got %q: %s", data.ExpiresAt.ValueString(), err),
				)

				return
			}

			createReq.ExpiresAt = &parsedExpiresAt
		}

		res, err := r.iamAPI.CreateAPIKey(&createReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error executing IAM Api Key Create",
				fmt.Sprintf("%s", err),
			)

			return
		}

		if hasAnnotationIdentifier {
			annotationValue := data.AnnotationIdentifier.ValueString()
			if err := r.identifierClient.GetOrCreateAnnotationIdentifier(ctx, res.Srn, annotationValue, orgID); err != nil {
				resp.Diagnostics.AddError(
					"Error setting up identifier annotations",
					err.Error(),
				)

				return
			}
		}

		if shouldReplace && existingAPIKey != nil {
			deleteReq := iam.DeleteAPIKeyRequest{
				AccessKey: existingAPIKey.AccessKey,
			}
			if err := r.iamAPI.DeleteAPIKey(&deleteReq, scw.WithContext(ctx)); err != nil {
				resp.Diagnostics.AddError(
					"Error deleting old API key",
					fmt.Sprintf("Failed to delete old API key %s: %s", existingAPIKey.AccessKey, err),
				)

				return
			}
		}

		setApiKeyData(&data, res)
	} else {
		setApiKeyData(&data, existingAPIKey)
	}

	resp.Result.Set(ctx, &data)

	// Store access key, lifecycle and annotation in private state for Close operation
	// Values must be JSON-encoded for private state storage
	if !data.AccessKey.IsNull() {
		accessKeyJSON, err := json.Marshal(data.AccessKey.ValueString())
		if err == nil {
			resp.Private.SetKey(ctx, "access_key", accessKeyJSON)
		}
	}

	lifecycleJSON, err := json.Marshal(lifecycle)
	if err == nil {
		resp.Private.SetKey(ctx, "ephemeral_lifecycle", lifecycleJSON)
	}

	if hasAnnotationIdentifier {
		annotationJSON, err := json.Marshal(data.AnnotationIdentifier.ValueString())
		if err == nil {
			resp.Private.SetKey(ctx, "annotation_identifier", annotationJSON)
		}
	}
}

func (r *ApiKeyEphemeralResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	lifecycleBytes, err := req.Private.GetKey(ctx, "ephemeral_lifecycle")
	if err != nil {
		return
	}

	var lifecycle string
	if err := json.Unmarshal(lifecycleBytes, &lifecycle); err != nil {
		return
	}

	if lifecycle != "delete" {
		return
	}

	accessKeyBytes, err := req.Private.GetKey(ctx, "access_key")
	if err != nil || len(accessKeyBytes) == 0 {
		return
	}

	var accessKey string
	if err := json.Unmarshal(accessKeyBytes, &accessKey); err != nil {
		return
	}

	annotationIdentifierBytes, err := req.Private.GetKey(ctx, "annotation_identifier")
	if err != nil {
		annotationIdentifierBytes = nil
	}

	var annotationIdentifier string
	if len(annotationIdentifierBytes) > 0 {
		if err := json.Unmarshal(annotationIdentifierBytes, &annotationIdentifier); err != nil {
			annotationIdentifier = ""
		}
	}

	orgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
	if !exists {
		resp.Diagnostics.AddError(
			"Organization ID is required",
			"Configure a default organization",
		)

		return
	}

	deleteReq := iam.DeleteAPIKeyRequest{
		AccessKey: accessKey,
	}
	if err := r.iamAPI.DeleteAPIKey(&deleteReq, scw.WithContext(ctx)); err != nil {
		if !httperrors.Is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting API key",
				fmt.Sprintf("Failed to delete API key %s: %s", accessKey, err),
			)
		}
	}

	if annotationIdentifier != "" {
		if err := r.identifierClient.DeleteAnnotationIdentifier(ctx, annotationIdentifier, orgID); err != nil {
			resp.Diagnostics.AddWarning(
				"Annotation cleanup failed",
				fmt.Sprintf("Failed to delete annotation binding for %q: %s", annotationIdentifier, err),
			)
		}
	}
}
