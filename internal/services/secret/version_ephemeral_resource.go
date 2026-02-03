package secret

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ ephemeral.EphemeralResource              = (*VersionEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*VersionEphemeralResource)(nil)
)

type VersionEphemeralResource struct {
	secretAPI *secret.API
	meta      *meta.Meta
}

func NewVersionEphemeralResource() ephemeral.EphemeralResource {
	return &VersionEphemeralResource{}
}

func (r *VersionEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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
	r.secretAPI = secret.NewAPI(client)
	r.meta = m
}

func (r *VersionEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_version"
}

type VersionEphemeralResourceModel struct {
	SecretID       types.String `tfsdk:"secret_id"`
	Region         types.String `tfsdk:"region"`
	Revision       types.String `tfsdk:"revision"`
	SecretName     types.String `tfsdk:"secret_name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ProjectID      types.String `tfsdk:"project_id"`
	// Output
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	Data        types.String `tfsdk:"data"`
}

//go:embed descriptions/version_ephemeral_resource.md
var versionEphemeralResourceDescription string

func (r *VersionEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         versionEphemeralResourceDescription,
		MarkdownDescription: versionEphemeralResourceDescription,
		Attributes: map[string]schema.Attribute{
			"secret_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The ID of the secret associated with the version. Either secret_id or secret_name must be specified.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithLocality(),
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("secret_id"),
						path.MatchRoot("secret_name"),
					),
				},
			},
			"secret_name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the secret.  Either secret_id or secret_name must be specified.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("secret_id"),
						path.MatchRoot("secret_name"),
					),
				},
			},
			"revision": schema.StringAttribute{
				Optional:    true,
				Description: "The revision of the secret version. Defaults to `latest`",
			},
			"region": regional.SchemaAttribute("The region of the secret version. If not set, the region is derived from the secret_id when possible or from the provider configuration."),
			"organization_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The organization ID to filter the secret version",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The project ID to filter the secret version",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the secret version",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Status of the secret version",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time of secret version's creation (RFC 3339 format)",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time of secret version's last update (RFC 3339 format)",
			},
			"data": schema.StringAttribute{
				Computed:    true,
				Description: "The payload of the secret version (base64 encoded)",
				Sensitive:   true,
			},
		},
	}
}

func (r *VersionEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data VersionEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if r.secretAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured secretAPI",
			"The ephemeral resource was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	var region scw.Region

	switch {
	case !data.Region.IsNull() && !data.Region.IsUnknown():
		region = scw.Region(data.Region.ValueString())
	case !data.SecretID.IsNull() && !data.SecretID.IsUnknown():
		if parsedRegion, _, err := regional.ParseID(data.SecretID.ValueString()); err == nil {
			region = parsedRegion
		} else {
			resp.Diagnostics.AddError(
				"Invalid secret_id",
				fmt.Sprintf("Failed to parse region from secret_id: %s", err),
			)

			return
		}
	default:
		defaultRegion, exists := r.meta.ScwClient().GetDefaultRegion()
		if !exists {
			resp.Diagnostics.AddError(
				"Missing region",
				"The region attribute is required to encrypt with a key. Please provide it explicitly or configure a default region in the provider.",
			)

			return
		}

		region = defaultRegion
	}

	revision := "latest"
	if !data.Revision.IsNull() && !data.Revision.IsUnknown() {
		revision = data.Revision.ValueString()
	}

	var secretID string

	switch {
	case !data.SecretName.IsNull() && !data.SecretName.IsUnknown():
		secretName := data.SecretName.ValueString()

		var organizationID *string

		if !data.OrganizationID.IsNull() && !data.OrganizationID.IsUnknown() {
			orgID := data.OrganizationID.ValueString()
			organizationID = &orgID
		}

		var projectID *string

		if !data.ProjectID.IsNull() && !data.ProjectID.IsUnknown() {
			projID := data.ProjectID.ValueString()
			projectID = &projID
		}

		secrets, err := r.secretAPI.ListSecrets(&secret.ListSecretsRequest{
			Region:         region,
			Name:           &secretName,
			ProjectID:      projectID,
			OrganizationID: organizationID,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error listing secrets",
				fmt.Sprintf("Failed to list secrets: %s", err),
			)

			return
		}

		foundSecret, err := datasource.FindExact(secrets.Secrets,
			func(s *secret.Secret) bool { return s.Name == secretName },
			secretName,
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error listing secrets",
				fmt.Sprintf("Failed to list secrets: %s", err),
			)

			return
		}

		secretID = foundSecret.ID
		data.SecretID = types.StringValue(secretID)
	case !data.SecretID.IsNull() && !data.SecretID.IsUnknown():
		secretID = locality.ExpandID(data.SecretID.ValueString())
	default:
		resp.Diagnostics.AddError(
			"Missing secret identifier",
			"Either secret_id or secret_name must be specified",
		)

		return
	}

	accessRequest := &secret.AccessSecretVersionRequest{
		Region:   region,
		SecretID: secretID,
		Revision: revision,
	}

	accessResponse, err := r.secretAPI.AccessSecretVersion(accessRequest, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error accessing secret version",
			fmt.Sprintf("Failed to access secret version: %s", err),
		)

		return
	}

	getRequest := &secret.GetSecretVersionRequest{
		Region:   region,
		SecretID: secretID,
		Revision: strconv.Itoa(int(accessResponse.Revision)),
	}

	getResponse, err := r.secretAPI.GetSecretVersion(getRequest, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting secret version details",
			fmt.Sprintf("Failed to get secret version details: %s", err),
		)

		return
	}

	if getResponse.Description != nil {
		data.Description = types.StringValue(*getResponse.Description)
	}

	data.Status = types.StringValue(getResponse.Status.String())
	if getResponse.CreatedAt != nil {
		data.CreatedAt = types.StringValue(getResponse.CreatedAt.Format(time.RFC3339))
	}

	if getResponse.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(getResponse.UpdatedAt.Format(time.RFC3339))
	}

	data.Data = types.StringValue(base64.StdEncoding.EncodeToString(accessResponse.Data))
	data.Revision = types.StringValue(revision)

	resp.Result.Set(ctx, &data)
}
