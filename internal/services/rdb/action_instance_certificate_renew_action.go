package rdb

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	rdb "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ action.Action              = (*InstanceCertificateRenewAction)(nil)
	_ action.ActionWithConfigure = (*InstanceCertificateRenewAction)(nil)
)

// InstanceCertificateRenewAction renews the certificate of an RDB instance.
type InstanceCertificateRenewAction struct {
	rdbAPI *rdb.API
	meta   *meta.Meta
}

func (a *InstanceCertificateRenewAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	a.meta = m
	a.rdbAPI = newAPI(m)
}

func (a *InstanceCertificateRenewAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdb_instance_certificate_renew_action"
}

type InstanceCertificateRenewActionModel struct {
	InstanceID types.String `tfsdk:"instance_id"`
	Region     types.String `tfsdk:"region"`
	Wait       types.Bool   `tfsdk:"wait"`
}

// NewInstanceCertificateRenewAction returns a new RDB instance certificate renew action.
func NewInstanceCertificateRenewAction() action.Action {
	return &InstanceCertificateRenewAction{}
}

func (a *InstanceCertificateRenewAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"instance_id": schema.StringAttribute{
				Required:    true,
				Description: "RDB instance ID to renew certificate for. Can be a plain UUID or a regional ID.",
			},
			"region": regional.SchemaAttribute(),
			"wait": schema.BoolAttribute{
				Optional:    true,
				Description: "Wait for the certificate renewal to complete before returning.",
			},
		},
	}
}

func (a *InstanceCertificateRenewAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data InstanceCertificateRenewActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.rdbAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured rdbAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	if data.InstanceID.IsNull() || data.InstanceID.IsUnknown() || data.InstanceID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing instance_id",
			"The instance_id attribute is required to renew an RDB instance certificate.",
		)

		return
	}

	instanceID := locality.ExpandID(data.InstanceID.ValueString())

	var region scw.Region

	if !data.Region.IsNull() && !data.Region.IsUnknown() && data.Region.ValueString() != "" {
		parsedRegion, err := scw.ParseRegion(data.Region.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid region value",
				fmt.Sprintf("The region attribute must be a valid Scaleway region. Got %q: %s", data.Region.ValueString(), err),
			)

			return
		}

		region = parsedRegion
	} else {
		if derivedRegion, id, parseErr := regional.ParseID(data.InstanceID.ValueString()); parseErr == nil {
			region = derivedRegion
			instanceID = id
		} else if a.meta != nil {
			defaultRegion, exists := a.meta.ScwClient().GetDefaultRegion()
			if !exists {
				resp.Diagnostics.AddError(
					"Unable to determine region",
					"Failed to get default region from provider configuration. Please set the region attribute, use a regional instance_id, or configure a default region in the provider.",
				)

				return
			}

			region = defaultRegion
		}
	}

	if region == "" {
		resp.Diagnostics.AddError(
			"Missing region",
			"Could not determine region for RDB instance certificate renewal. Please set the region attribute, use a regional instance_id, or configure a default region in the provider.",
		)

		return
	}

	renewReq := &rdb.RenewInstanceCertificateRequest{
		Region:     region,
		InstanceID: instanceID,
	}

	err := a.rdbAPI.RenewInstanceCertificate(renewReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing RDB RenewInstanceCertificate action",
			fmt.Sprintf("Failed to renew certificate for instance %s: %s", instanceID, err),
		)

		return
	}

	if data.Wait.ValueBool() {
		_, err = waitForRDBInstance(ctx, a.rdbAPI, region, instanceID, defaultInstanceTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error waiting for RDB certificate renewal completion",
				fmt.Sprintf("Certificate renewal for instance %s did not complete: %s", instanceID, err),
			)

			return
		}
	}
}
