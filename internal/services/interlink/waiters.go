package interlink

import (
	"context"
	"time"

	interlink "github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultLinkTimeout       = 10 * time.Minute
	defaultLinkRetryInterval = 15 * time.Second
)

func waitForLink(ctx context.Context, api *interlink.API, region scw.Region, linkID string, timeout time.Duration) (*interlink.Link, error) {
	retryInterval := defaultLinkRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	link, err := api.WaitForLink(&interlink.WaitForLinkRequest{
		LinkID:        linkID,
		Region:        region,
		Timeout:       &timeout,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return link, err
}
