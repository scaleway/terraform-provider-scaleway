package file

import (
	"context"
	"time"

	file "github.com/scaleway/scaleway-sdk-go/api/file/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForFileSystem(ctx context.Context, fileAPI *file.API, region scw.Region, id string, timeout time.Duration) (*file.FileSystem, error) {
	retryInterval := defaultFileSystemRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	fileSystem, err := fileAPI.WaitForFileSystem(&file.WaitForFileSystemRequest{
		FilesystemID:  id,
		Region:        region,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return fileSystem, err
}
