package webhosting

import (
	"context"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/webhosting/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultHostingTimeout = 15 * time.Minute
	hostingRetryInterval  = 5 * time.Second
)

// NewAPIWithRegionAndID returns a Hosting API with region and ID extracted from the state.
func NewAPIWithRegionAndID(m any, id string) (*webhosting.HostingAPI, scw.Region, string, error) {
	api := webhosting.NewHostingAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}

func waitForHosting(ctx context.Context, api *webhosting.HostingAPI, region scw.Region, hostingID string, timeout time.Duration) (*webhosting.Hosting, error) {
	retryInterval := hostingRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForHosting(&webhosting.WaitForHostingRequest{
		HostingID:     hostingID,
		Region:        region,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

func extractDNSRecordName(raw string) string {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return ""
	}

	return strings.TrimSuffix(fields[0], ".")
}
