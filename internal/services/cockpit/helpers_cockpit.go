package cockpit

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/validation"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	DefaultCockpitTimeout       = 5 * time.Minute
	defaultCockpitRetryInterval = 15 * time.Second
	pathMetricsURL              = "/api/v1/push"
	pathLogsURL                 = "/loki/api/v1/push"
	pathTracesURL               = "/otlp/v1/traces"
)

// NewGlobalAPI returns a new global cockpit API.
func NewGlobalAPI(m any) (*cockpit.GlobalAPI, error) {
	api := cockpit.NewGlobalAPI(meta.ExtractScwClient(m))

	return api, nil
}

// dataSourceBase holds the provider meta shared by the cockpit Framework data sources.
type dataSourceBase struct {
	meta *meta.Meta
}

func (b *dataSourceBase) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	b.meta = m
}

// regionalAPI returns a regional API and the region resolved from the attribute or the provider default.
func (b *dataSourceBase) regionalAPI(regionAttr types.String) (*cockpit.RegionalAPI, scw.Region, error) {
	region, err := meta.ExtractFrameworkRegion(regionAttr, b.meta.ScwClient())
	if err != nil {
		return nil, "", err
	}

	return cockpit.NewRegionalAPI(b.meta.ScwClient()), region, nil
}

// getRegionalIdentity reads the resource identity; it must be called before reading the state.
// The returned bool is false when the identity is missing (e.g. state written before identity support).
func getRegionalIdentity(ctx context.Context, identity *tfsdk.ResourceIdentity) (identityfw.RegionalIdentity, bool) {
	var ident identityfw.RegionalIdentity

	if identity == nil || identity.Get(ctx, &ident).HasError() {
		return ident, false
	}

	return ident, !ident.ID.IsNull() && !ident.ID.IsUnknown()
}

// resolveRegionalID resolves the region and ID of a resource from its identity when available, else from the state ID.
func resolveRegionalID(ident identityfw.RegionalIdentity, identityAvailable bool, stateID types.String) (scw.Region, string, error) {
	if identityAvailable {
		region, id := identityfw.ResolveRegionalIdentity(ident)

		return region, id, nil
	}

	return regional.ParseID(stateID.ValueString())
}

func NewAPIWithRegionAndID(m any, id string) (*cockpit.RegionalAPI, scw.Region, string, error) {
	api := cockpit.NewRegionalAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}

func waitForExporter(
	ctx context.Context,
	api *cockpit.RegionalAPI,
	region scw.Region,
	exporterID string,
	timeout time.Duration,
) (*cockpit.Exporter, error) {
	retryInterval := defaultCockpitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.Exporter, error) {
		return api.WaitForExporter(&cockpit.WaitForExporterRequest{
			Region:        region,
			ExporterID:    exporterID,
			Timeout:       new(timeout),
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))
	})
}

func cockpitTokenV1UpgradeFunc(_ context.Context, rawState map[string]any, _ any) (map[string]any, error) {
	defaultRegion := scw.RegionFrPar // Default to the 'fr-par' region as all tokens created with the v1beta1 API were implicitly set to this region

	if _, ok := rawState["region"]; !ok {
		rawState["region"] = defaultRegion.String()
	}

	if id, ok := rawState["id"].(string); ok && validation.IsUUID(id) {
		locality, ID, _ := regional.ParseID(id)
		if locality == "" {
			rawState["id"] = regional.NewIDString(defaultRegion, ID)
		}
	} else {
		return nil, fmt.Errorf("upgrade: expected 'id' to be a string, got %T", rawState["id"])
	}

	return rawState, nil
}
