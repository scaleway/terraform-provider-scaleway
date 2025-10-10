package secret

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource                     = &DataSourceSecret{}
	_ datasource.DataSourceWithConfigure        = &DataSourceSecret{}
	_ datasource.DataSourceWithConfigValidators = &DataSourceSecret{}
)

type DataSourceSecret struct {
	secretAPI *secret.API
}

func NewDataSourceSecret() datasource.DataSource {
	return &DataSourceSecret{}
}

func (d DataSourceSecret) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_secret"
}

func (d DataSourceSecret) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			//	dsSchema["secret_id"] = &schema.Schema{
			//		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			//	}
			"secret_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the secret",
				Validators: []validator.String{
					verify.UUIDorUUIDWithLocalityValidator{},
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The secret name",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the secret",
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
				Description: "Location of the secret in the directory structure.",
			},
			"protected": schema.BoolAttribute{
				Optional:    true,
				Description: "True if secret protection is enabled on a given secret. A protected secret cannot be deleted.",
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of tags [\"tag1\", \"tag2\", ...] associated to secret",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: func() string {
					var t secret.SecretType

					secretTypes := t.Values()

					return fmt.Sprintf("Type of the secret could be any value among: %s", secretTypes)
				}(),
				Optional: true,
				Validators: []validator.String{
					verify.ValidatorFromEnum[secret.SecretType](secret.SecretType("")),
				},
			},
			"region":          regional.ResourceSchema("The region you want to attach the resource to"),
			"project_id":      account.ResourceProjectIDSchema("The project ID you want to attach the secret to"),
			"organization_id": account.DatasourceOrganizationIDSchema("The organization id the secret is attached to"),
			"ephemeral_policy": schema.ListNestedAttribute{
				Description: "Ephemeral policy of the secret. Policy that defines whether/when a secret's versions expire. By default, the policy is applied to all the secret's versions.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ttl": schema.StringAttribute{
							Optional:    true,
							Description: "Time frame, from one second and up to one year, during which the secret's versions are valid. Has to be specified in Go Duration format",
							Validators: []validator.String{
								verify.DurationValidator{},
							},
						},
						"expires_once_accessed": schema.BoolAttribute{
							Optional:    true,
							Description: "True if the secret version expires after a single user access.",
						},
						"action": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								verify.ValidatorFromEnum[secret.EphemeralPolicyAction](secret.EphemeralPolicyAction("")),
							},
							Description: "Action to perform when the version of a secret expires.",
						},
					},
				},
			},
			"versions": schema.ListNestedAttribute{
				Description: "List of the versions of the secret",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"revision": schema.StringAttribute{
							Computed:    true,
							Description: "The revision of secret version",
						},
						"secret_id": schema.StringAttribute{
							Computed:    true,
							Description: "The secret ID associated with this version",
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
							Description: "Date and time of secret version's creation (RFC 3339 format)",
						},
						"description": schema.StringAttribute{
							Optional:    true,
							Description: "Description of the secret version",
						},
						"latest": schema.BoolAttribute{
							Optional:    true,
							Description: "Returns true if the version is the latest.",
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx,
				timeouts.Opts{
					Create: true,
				},
			),
		},
	}
}

func (d DataSourceSecret) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	// Always perform a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*scw.Client)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *scw.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	d.secretAPI = secret.NewAPI(client)
}

func (d DataSourceSecret) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	// Prevent panic if the provider has not been configured.
	if d.secretAPI == nil {
		response.Diagnostics.AddError(
			"Unconfigured Scaleway Client",
			"Expected configured Scaleway Client. Please report this issue to the provider developers.",
		)

		return
	}

	//	secretID, ok := d.GetOk("secret_id")
	//	if !ok {
	//		secretName := d.Get("name").(string)
	//		request := &secret.ListSecretsRequest{
	//			Region:         region,
	//			Name:           types.ExpandStringPtr(secretName),
	//			ProjectID:      projectID,
	//			OrganizationID: types.ExpandStringPtr(d.Get("organization_id")),
	//			Path:           types.ExpandStringPtr(d.Get("path")),
	//		}
	//
	//		res, err := api.ListSecrets(request, scw.WithContext(ctx))
	//		if err != nil {
	//			return diag.FromErr(err)
	//		}
	//
	//		foundSecret, err := datasource.FindExact(
	//			res.Secrets,
	//			func(s *secret.Secret) bool { return s.Name == secretName },
	//			secretName,
	//		)
	//		if err != nil {
	//			return diag.FromErr(err)
	//		}
	//
	//		secretID = foundSecret.ID
	//	}
	//
	//	regionalID := datasource.NewRegionalID(secretID, region)
	//	d.SetId(regionalID)
	//
	//	err = d.Set("secret_id", regionalID)
	//	if err != nil {
	//		return diag.FromErr(err)
	//	}
	//
	//	diags := ResourceSecretRead(ctx, d, m)
	//	if diags != nil {
	//		return append(diags, diag.Errorf("failed to read secret")...)
	//	}
	//
	//	if d.Id() == "" {
	//		return diag.Errorf("secret (%s) not found", regionalID)
	//	}
	//
	//	return nil
}

func (d DataSourceSecret) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("name"),
			path.MatchRoot("secret_id"),
		),
		datasourcevalidator.Conflicting(
			path.MatchRoot("path"),
			path.MatchRoot("secret_id"),
		),
	}
}

//func DataSourceSecret() *schema.Resource {
//	// Set 'Optional' schema elements
//	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "path")
//

//
