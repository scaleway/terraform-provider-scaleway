package iam

import (
	"context"
	_ "embed"
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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_                         ephemeral.EphemeralResource              = (*ApiKeyEphemeralResource)(nil)
	_                         ephemeral.EphemeralResourceWithConfigure = (*ApiKeyEphemeralResource)(nil)
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
	ReplaceResource       types.Bool   `tfsdk:"replace_resource"`
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
			"replace_resource": schema.BoolAttribute{
				Optional:    true,
				Description: "If true, always create a new resource and replace the existing one found by identifier (see `annotation_identifier` and `description_identifier`). If false, reuse existing resource if found.",
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

	var existingAPIKey *iam.APIKey
	var identifierDescription string
	var err error

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

	recreateResource := !data.ReplaceResource.IsNull() && !data.ReplaceResource.IsUnknown() && data.ReplaceResource.ValueBool()

	if recreateResource || existingAPIKey == nil {
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

		if recreateResource && existingAPIKey != nil {
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
}
