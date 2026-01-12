package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/functions"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/applesilicon"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/baremetal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/cockpit"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/jobs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/keymanager"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mongodb"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/rdb"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpcgw"
)

var (
	_ provider.Provider            = &ScalewayProvider{}
	_ provider.ProviderWithActions = (*ScalewayProvider)(nil)
)

type ScalewayProvider struct {
	providerMeta *meta.Meta
}

func NewFrameworkProvider(m *meta.Meta) func() provider.Provider {
	return func() provider.Provider {
		return &ScalewayProvider{providerMeta: m}
	}
}

func (p *ScalewayProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "scaleway"
}

type ScalewayProviderModel struct {
	AccessKey      types.String `tfsdk:"access_key"`
	SecretKey      types.String `tfsdk:"secret_key"`
	Profile        types.String `tfsdk:"profile"`
	ProjectID      types.String `tfsdk:"project_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	APIURL         types.String `tfsdk:"api_url"`
	Region         types.String `tfsdk:"region"`
	Zone           types.String `tfsdk:"zone"`
}

func (p *ScalewayProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_key": schema.StringAttribute{
				Optional:    true,
				Description: "The Scaleway access key.",
			},
			"secret_key": schema.StringAttribute{
				Optional:    true,
				Description: "The Scaleway secret Key.",
			},
			"profile": schema.StringAttribute{
				Optional:    true,
				Description: "The Scaleway profile to use.",
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "The Scaleway project ID.",
			},
			"organization_id": schema.StringAttribute{
				Optional:    true,
				Description: "The Scaleway organization ID.",
			},
			"api_url": schema.StringAttribute{
				Optional:    true,
				Description: "The Scaleway API URL to use.",
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "The region you want to attach the resource to",
			},
			"zone": schema.StringAttribute{
				Description: "The zone you want to attach the resource to",
				Optional:    true,
			},
		},
	}
}

func (p *ScalewayProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ScalewayProviderModel

	// Read configuration data into model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var m *meta.Meta

	if p.providerMeta != nil {
		// Use pre-injected meta (from tests or config)
		resp.Diagnostics.Append(diag.NewWarningDiagnostic("using provider meta already initialized", "meta provider not empty"))

		m = p.providerMeta
	} else {
		config := &meta.Config{
			TerraformVersion: req.TerraformVersion,
		}

		var err error

		m, err = meta.NewMeta(ctx, config)
		if err != nil {
			resp.Diagnostics.AddError("error while configuring the provider", err.Error())

			return
		}
	}

	resp.ResourceData = m
	resp.DataSourceData = m
	resp.ActionData = m
	resp.EphemeralResourceData = m
}

func (p *ScalewayProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *ScalewayProvider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		keymanager.NewDecryptEphemeralResource,
		keymanager.NewEncryptEphemeralResource,
		keymanager.NewGenerateDataKeyEphemeralResource,
		keymanager.NewSignEphemeralResource,
	}
}

func (p *ScalewayProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *ScalewayProvider) Actions(_ context.Context) []func() action.Action {
	return []func() action.Action{
		applesilicon.NewRebootServerAction,
		baremetal.NewBaremetalServerAction,
		cockpit.NewTriggerTestAlertAction,
		instance.NewCreateSnapshot,
		instance.NewExportSnapshot,
		instance.NewServerAction,
		jobs.NewStartJobDefinitionAction,
		keymanager.NewRotateKeyAction,
		mongodb.NewInstanceSnapshotAction,
		rdb.NewDatabaseBackupExportAction,
		rdb.NewDatabaseBackupRestoreAction,
		rdb.NewInstanceCertificateRenewAction,
		rdb.NewInstanceLogPrepareAction,
		rdb.NewInstanceLogsPurgeAction,
		rdb.NewInstanceSnapshotAction,
		rdb.NewReadReplicaPromoteAction,
		rdb.NewReadReplicaResetAction,
		vpcgw.NewRefreshSSHKeysAction,
	}
}

func (p *ScalewayProvider) ListResources(_ context.Context) []func() list.ListResource {
	return []func() list.ListResource{}
}

func (p *ScalewayProvider) Functions(_ context.Context) []func() function.Function {
	var res []func() function.Function

	res = append(res, functions.NewRegionFromID)

	return res
}
