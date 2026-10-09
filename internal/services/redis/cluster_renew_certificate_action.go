package redis

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ action.Action              = (*ClusterRenewCertificateAction)(nil)
	_ action.ActionWithConfigure = (*ClusterRenewCertificateAction)(nil)
)

// ClusterRenewCertificateAction renews the TLS certificate of a Redis cluster.
type ClusterRenewCertificateAction struct {
	redisAPI *redis.API
	meta     *meta.Meta
}

func (a *ClusterRenewCertificateAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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
	a.redisAPI = newAPI(m)
}

func (a *ClusterRenewCertificateAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redis_cluster_renew_certificate"
}

type ClusterRenewCertificateActionModel struct {
	ClusterID types.String `tfsdk:"cluster_id"`
	Zone      types.String `tfsdk:"zone"`
	Wait      types.Bool   `tfsdk:"wait"`
}

// NewClusterRenewCertificateAction returns a new Redis cluster renew certificate action.
func NewClusterRenewCertificateAction() action.Action {
	return &ClusterRenewCertificateAction{}
}

//go:embed descriptions/cluster_renew_certificate_action.md
var clusterRenewCertificateDescription string

func (a *ClusterRenewCertificateAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: clusterRenewCertificateDescription,
		Description:         clusterRenewCertificateDescription,
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{
				Required:    true,
				Description: "Redis cluster ID to renew the certificate for. Can be a plain UUID or a zonal ID.",
			},
			"zone": zonal.SchemaAttribute("Zone of the Redis cluster. If not set, derived from cluster_id when possible or from the provider configuration."),
			"wait": schema.BoolAttribute{
				Optional:    true,
				Description: "Wait for the certificate renewal to complete before returning.",
			},
		},
	}
}

func (a *ClusterRenewCertificateAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ClusterRenewCertificateActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.redisAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured redisAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	if data.ClusterID.IsNull() || data.ClusterID.IsUnknown() || data.ClusterID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing cluster_id",
			"The cluster_id attribute is required to renew a Redis cluster certificate.",
		)

		return
	}

	clusterID := locality.ExpandID(data.ClusterID.ValueString())

	var zone scw.Zone

	if !data.Zone.IsNull() && !data.Zone.IsUnknown() && data.Zone.ValueString() != "" {
		parsedZone, err := scw.ParseZone(data.Zone.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid zone value",
				fmt.Sprintf("The zone attribute must be a valid Scaleway zone. Got %q: %s", data.Zone.ValueString(), err),
			)

			return
		}

		zone = parsedZone
	} else {
		if derivedZone, id, parseErr := zonal.ParseID(data.ClusterID.ValueString()); parseErr == nil {
			zone = derivedZone
			clusterID = id
		} else if a.meta != nil {
			defaultZone, exists := a.meta.ScwClient().GetDefaultZone()
			if !exists {
				resp.Diagnostics.AddError(
					"Unable to determine zone",
					"Failed to get default zone from provider configuration. Please set the zone attribute, use a zonal cluster_id, or configure a default zone in the provider.",
				)

				return
			}

			zone = defaultZone
		}
	}

	if zone == "" {
		resp.Diagnostics.AddError(
			"Missing zone",
			"Could not determine zone for Redis cluster certificate renewal. Please set the zone attribute, use a zonal cluster_id, or configure a default zone in the provider.",
		)

		return
	}

	_, err := a.redisAPI.RenewClusterCertificate(&redis.RenewClusterCertificateRequest{
		Zone:      zone,
		ClusterID: clusterID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing Redis RenewClusterCertificate action",
			fmt.Sprintf("Failed to renew certificate for cluster %s: %s", clusterID, err),
		)

		return
	}

	if data.Wait.ValueBool() {
		_, err = waitForCluster(ctx, a.redisAPI, zone, clusterID, defaultRedisClusterTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error waiting for Redis certificate renewal completion",
				fmt.Sprintf("Certificate renewal for cluster %s did not complete: %s", clusterID, err),
			)

			return
		}
	}
}
