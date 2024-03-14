package scaleway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultFunctionNamespaceTimeout = 5 * time.Minute
	defaultFunctionTimeout          = 15 * time.Minute
	defaultFunctionRetryInterval    = 5 * time.Second
	defaultFunctionAfterUpdateWait  = 1 * time.Second
	defaultFunctionCronTimeout      = 5 * time.Minute
)

// functionAPIWithRegion returns a new container registry API and the region.
func functionAPIWithRegion(d *schema.ResourceData, m interface{}) (*function.API, scw.Region, error) {
	api := function.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// FunctionAPIWithRegionAndID returns a new container registry API, region and ID.
func FunctionAPIWithRegionAndID(m interface{}, id string) (*function.API, scw.Region, string, error) {
	api := function.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}

func waitForFunctionNamespace(ctx context.Context, functionAPI *function.API, region scw.Region, id string, timeout time.Duration) (*function.Namespace, error) {
	retryInterval := defaultFunctionRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	ns, err := functionAPI.WaitForNamespace(&function.WaitForNamespaceRequest{
		Region:        region,
		NamespaceID:   id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return ns, err
}

func waitForFunction(ctx context.Context, functionAPI *function.API, region scw.Region, id string, timeout time.Duration) (*function.Function, error) {
	retryInterval := defaultFunctionRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	f, err := functionAPI.WaitForFunction(&function.WaitForFunctionRequest{
		Region:        region,
		FunctionID:    id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return f, err
}

func waitForFunctionCron(ctx context.Context, functionAPI *function.API, region scw.Region, cronID string, timeout time.Duration) (*function.Cron, error) {
	retryInterval := defaultFunctionRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return functionAPI.WaitForCron(&function.WaitForCronRequest{
		Region:        region,
		CronID:        cronID,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))
}

func waitForFunctionDomain(ctx context.Context, functionAPI *function.API, region scw.Region, id string, timeout time.Duration) (*function.Domain, error) {
	retryInterval := defaultFunctionRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	domain, err := functionAPI.WaitForDomain(&function.WaitForDomainRequest{
		Region:        region,
		DomainID:      id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return domain, err
}

func waitForFunctionTrigger(ctx context.Context, functionAPI *function.API, region scw.Region, id string, timeout time.Duration) (*function.Trigger, error) {
	retryInterval := defaultFunctionRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	trigger, err := functionAPI.WaitForTrigger(&function.WaitForTriggerRequest{
		Region:        region,
		TriggerID:     id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return trigger, err
}

func functionUpload(ctx context.Context, m interface{}, functionAPI *function.API, region scw.Region, functionID string, zipFile string) error {
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

	secretKey, _ := meta.ExtractScwClient(m).GetSecretKey()
	req.Header.Add("X-Auth-Token", secretKey)

	resp, err := meta.ExtractHTTPClient(m).Do(req)
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
		return errors.New("failed to deploy function")
	}
	return nil
}

func expandFunctionsSecrets(secretsRawMap interface{}) []*function.Secret {
	secretsMap := secretsRawMap.(map[string]interface{})
	secrets := make([]*function.Secret, 0, len(secretsMap))

	for k, v := range secretsMap {
		secrets = append(secrets, &function.Secret{
			Key:   k,
			Value: types.ExpandStringPtr(v),
		})
	}

	return secrets
}

func isFunctionDNSResolveError(err error) bool {
	responseError := &scw.ResponseError{}

	if !errors.As(err, &responseError) {
		return false
	}

	if strings.HasPrefix(responseError.Message, "could not validate domain") {
		return true
	}

	return false
}

func retryCreateFunctionDomain(ctx context.Context, functionAPI *function.API, req *function.CreateDomainRequest, timeout time.Duration) (*function.Domain, error) {
	timeoutChannel := time.After(timeout)

	for {
		select {
		case <-time.After(defaultFunctionRetryInterval):
			domain, err := functionAPI.CreateDomain(req, scw.WithContext(ctx))
			if err != nil && isFunctionDNSResolveError(err) {
				continue
			}
			return domain, err
		case <-timeoutChannel:
			return functionAPI.CreateDomain(req, scw.WithContext(ctx))
		}
	}
}
