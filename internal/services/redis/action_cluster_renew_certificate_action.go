package redis

import (
	"context"
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

type ClusterRenewCertificateAction struct {
	redisAPI *redis.API
}

func (a *ClusterRenewCertificateAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

	client := m.ScwClient()
	a.redisAPI = redis.NewAPI(client)
}

func (a *ClusterRenewCertificateAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redis_cluster_renew_certificate_action"
}

type ClusterRenewCertificateActionModel struct {
	ClusterID types.String `tfsdk:"cluster_id"`
	Zone      types.String `tfsdk:"zone"`
	Wait      types.Bool   `tfsdk:"wait"`
}

func NewClusterRenewCertificateAction() action.Action {
	return &ClusterRenewCertificateAction{}
}

func (a *ClusterRenewCertificateAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{
				Required:    true,
				Description: "Redis cluster ID to renew certificate for. Can be a plain UUID or a zonal ID.",
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Description: "Zone of the Redis cluster. If not set, the zone is derived from the cluster_id when possible or from the provider configuration.",
			},
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

	if data.ClusterID.IsNull() || data.ClusterID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing cluster_id",
			"The cluster_id attribute is required to renew the Redis cluster certificate.",
		)

		return
	}

	clusterID := locality.ExpandID(data.ClusterID.ValueString())

	var (
		zone scw.Zone
		err  error
	)

	if !data.Zone.IsNull() && data.Zone.ValueString() != "" {
		zone = scw.Zone(data.Zone.ValueString())
	} else {
		// Try to derive zone from the cluster_id if it is a zonal ID.
		if derivedZone, id, parseErr := zonal.ParseID(data.ClusterID.ValueString()); parseErr == nil {
			zone = derivedZone
			clusterID = id
		}
	}

	renewReq := &redis.RenewClusterCertificateRequest{
		ClusterID: clusterID,
	}

	if zone != "" {
		renewReq.Zone = zone
	}

	cluster, err := a.redisAPI.RenewClusterCertificate(renewReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing Redis RenewClusterCertificate action",
			fmt.Sprintf("Failed to renew certificate for cluster %s: %s", clusterID, err),
		)

		return
	}

	if data.Wait.ValueBool() {
		waitZone := cluster.Zone
		if waitZone == "" && zone != "" {
			waitZone = zone
		}

		if waitZone == "" {
			resp.Diagnostics.AddError(
				"Missing zone for wait operation",
				"Could not determine zone to wait for Redis cluster certificate renewal completion.",
			)

			return
		}

		_, err = waitForCluster(ctx, a.redisAPI, waitZone, clusterID, defaultRedisClusterTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error waiting for Redis cluster certificate renewal completion",
				fmt.Sprintf("Certificate renewal for cluster %s did not complete: %s", clusterID, err),
			)

			return
		}
	}
}
