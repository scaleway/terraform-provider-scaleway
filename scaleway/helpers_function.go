package scaleway

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultFunctionNamespaceTimeout = 5 * time.Minute
	defaultFunctionTimeout          = 5 * time.Minute
	defaultFunctionRetryInterval    = 5 * time.Second
)

// functionAPIWithRegion returns a new container registry API and the region.
func functionAPIWithRegion(d *schema.ResourceData, m interface{}) (*function.API, scw.Region, error) {
	meta := m.(*Meta)
	api := function.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// functionAPIWithRegionAndID returns a new container registry API, region and ID.
func functionAPIWithRegionAndID(m interface{}, id string) (*function.API, scw.Region, string, error) {
	meta := m.(*Meta)
	api := function.NewAPI(meta.scwClient)

	region, id, err := parseRegionalID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}

func waitForFunctionNamespace(ctx context.Context, functionAPI *function.API, region scw.Region, id string, timeout time.Duration) (*function.Namespace, error) {
	retryInterval := defaultFunctionRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	ns, err := functionAPI.WaitForNamespace(&function.WaitForNamespaceRequest{
		Region:        region,
		NamespaceID:   id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return ns, err
}

func functionUpload(ctx context.Context, m interface{}, functionAPI *function.API, region scw.Region, functionID string, zipFile string) error {
	meta := m.(*Meta)
	zipStat, err := os.Stat(zipFile)
	if err != nil {
		return fmt.Errorf("failed to stat zip file: %w", err)
	}

	uploadURL, err := functionAPI.GetFunctionUploadURL(&function.GetFunctionUploadURLRequest{
		Region:        region,
		FunctionID:    functionID,
		ContentLength: uint64(zipStat.Size()),
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to fetch upload url: %w", err)
	}

	zip, err := os.Open(zipFile)
	if err != nil {
		return fmt.Errorf("failed to read zip file: %w", err)
	}
	defer zip.Close()

	req, err := http.NewRequest(http.MethodPut, uploadURL.URL, zip)
	if err != nil {
		return fmt.Errorf("failed to init request: %w", err)
	}

	req = req.WithContext(ctx)

	for headerName, headerList := range uploadURL.Headers {
		for _, header := range *headerList {
			req.Header.Add(headerName, header)
		}
	}

	if secretKey, secretKeyExists := meta.scwClient.GetSecretKey(); secretKeyExists {
		req.Header.Add("X-Auth-Token", secretKey)
	} else {
		return fmt.Errorf("missing secret key")
	}

	resp, err := meta.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return fmt.Errorf("failed to dump response: %w", err)
	}

	reqDump, err := httputil.DumpRequest(req, false)
	if err != nil {
		return fmt.Errorf("failed to dump request: %w", err)
	}

	tflog.Debug(ctx, "Request dump", map[string]interface{}{
		"url":      uploadURL.URL,
		"response": string(respDump),
		"request":  string(reqDump),
	})
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload function (Status: %d)", resp.StatusCode)
	}

	return nil
}

func functionDeploy(ctx context.Context, functionAPI *function.API, region scw.Region, functionID string) error {
	_, err := functionAPI.DeployFunction(&function.DeployFunctionRequest{
		Region:     region,
		FunctionID: functionID,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to deploy function")
	}
	return nil
}
