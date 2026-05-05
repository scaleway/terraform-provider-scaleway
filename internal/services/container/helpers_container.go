package container

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	containerV1 "github.com/scaleway/scaleway-sdk-go/api/container/v1"
	containerV1Beta1 "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultContainerNamespaceTimeout = 5 * time.Minute
	defaultContainerCronTimeout      = 5 * time.Minute
	defaultContainerTimeout          = 12*time.Minute + 30*time.Second
	defaultContainerDomainTimeout    = 10 * time.Minute
	DefaultContainerRetryInterval    = 5 * time.Second
)

// newAPIWithRegion returns a new container v1 API and the region.
func newAPIWithRegion(d *schema.ResourceData, m any) (*containerV1.API, scw.Region, error) {
	api := containerV1.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// newAPIBetaWithRegion returns a new container v1beta1 API and the region.
func newAPIBetaWithRegion(d *schema.ResourceData, m any) (*containerV1Beta1.API, scw.Region, error) {
	api := containerV1Beta1.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// NewAPIWithRegionAndID returns a new container v1 API, region and ID.
func NewAPIWithRegionAndID(m any, id string) (*containerV1.API, scw.Region, string, error) {
	api := containerV1.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}

// NewAPIBetaWithRegionAndID returns a new container v1beta1 API, region and ID.
func NewAPIBetaWithRegionAndID(m any, id string) (*containerV1Beta1.API, scw.Region, string, error) {
	api := containerV1Beta1.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}

func setCreateContainerRequest(d *schema.ResourceData, region scw.Region) (*containerV1.CreateContainerRequest, error) {
	// required
	nameRaw := d.Get("name")
	namespaceID := d.Get("namespace_id")

	name := types.ExpandOrGenerateString(nameRaw.(string), "co")
	privacyType := d.Get("privacy")
	protocol := d.Get("protocol")

	reqImage := ""
	if registryImage, ok := d.GetOk("registry_image"); ok {
		reqImage = registryImage.(string)
	} else if image, ok := d.GetOk("image"); ok {
		reqImage = image.(string)
	}

	req := &containerV1.CreateContainerRequest{
		Region:      region,
		Image:       reqImage,
		NamespaceID: locality.ExpandID(namespaceID),
		Name:        name,
		Privacy:     containerV1.ContainerPrivacy(privacyType.(string)),
		Protocol:    containerV1.ContainerProtocol(*types.ExpandStringPtr(protocol)),
	}

	// optional
	if envVariablesRaw, ok := d.GetOk("environment_variables"); ok {
		req.EnvironmentVariables = types.ExpandMapStringString(envVariablesRaw)
	}

	if secretEnvVariablesRaw, ok := d.GetOk("secret_environment_variables"); ok {
		req.SecretEnvironmentVariables = types.ExpandMapStringString(secretEnvVariablesRaw)
	}

	if minScale, ok := d.GetOk("min_scale"); ok {
		req.MinScale = new(uint32(minScale.(int)))
	}

	if maxScale, ok := d.GetOk("max_scale"); ok {
		req.MaxScale = new(uint32(maxScale.(int)))
	}

	if memoryLimitBytes, ok := d.GetOk("memory_limit_bytes"); ok {
		req.MemoryLimitBytes = new(scw.Size(memoryLimitBytes.(int)))
	} else if memoryLimitMB, ok := d.GetOk("memory_limit"); ok {
		req.MemoryLimitBytes = new(scw.Size(memoryLimitMB.(int)) * scw.MB)
	}

	if cpuLimit, ok := d.GetOk("cpu_limit"); ok {
		req.MvcpuLimit = new(uint32(cpuLimit.(int)))
	}

	if timeout, ok := d.GetOk("timeout"); ok {
		timeInt := timeout.(int)
		req.Timeout = &scw.Duration{Seconds: int64(timeInt)}
	}

	if port, ok := d.GetOk("port"); ok {
		req.Port = new(uint32(port.(int)))
	}

	if description, ok := d.GetOk("description"); ok {
		req.Description = types.ExpandStringPtr(description)
	}

	if httpsConnectionsOnly, ok := d.GetOk("https_connections_only"); ok {
		req.HTTPSConnectionsOnly = new(httpsConnectionsOnly.(bool))
	} else if httpOption, ok := d.GetOk("http_option"); ok {
		switch httpOption.(string) {
		case containerV1Beta1.ContainerHTTPOptionEnabled.String():
			req.HTTPSConnectionsOnly = types.ExpandBoolPtr(false)
		case containerV1Beta1.ContainerHTTPOptionRedirected.String():
			req.HTTPSConnectionsOnly = types.ExpandBoolPtr(true)
		}
	}

	if sandbox, ok := d.GetOk("sandbox"); ok {
		req.Sandbox = containerV1.ContainerSandbox(sandbox.(string))
	}

	if scalingOption, ok := d.GetOk("scaling_option"); ok {
		scalingOptionReq, err := expandScalingOption(scalingOption)
		if err != nil {
			return nil, err
		}

		req.ScalingOption = scalingOptionReq
	}

	if localStorageLimitBytes, ok := d.GetOk("local_storage_limit_bytes"); ok {
		req.LocalStorageLimitBytes = new(scw.Size(localStorageLimitBytes.(int)))
	} else if localStorageLimit, ok := d.GetOk("local_storage_limit"); ok {
		req.LocalStorageLimitBytes = new(scw.Size(localStorageLimit.(int)) * scw.MB)
	}

	livenessProbe, livenessProbeSet := d.GetOk("liveness_probe")
	if livenessProbeSet {
		livenessProbeReq, err := expandContainerProbe(livenessProbe, "liveness_probe")
		if err != nil {
			return nil, err
		}

		req.LivenessProbe = livenessProbeReq
	}

	healthCheck, healthCheckSet := d.GetOk("health_check")
	if healthCheckSet {
		if livenessProbeSet {
			return nil, errors.New("only one of health_check and liveness_probe must be set, we recommend using liveness_probe")
		}

		healthCheckList := healthCheck.([]map[string]any)
		healthCheckElem := healthCheckList[0]

		healthCheckReq, err := expandHealthCheck(healthCheckElem)
		if err != nil {
			return nil, err
		}

		req.LivenessProbe = &containerV1.ContainerProbe{
			FailureThreshold: healthCheckReq.FailureThreshold,
			Interval:         healthCheckReq.Interval,
		}

		if healthCheckReq.HTTP != nil {
			req.LivenessProbe.HTTP = &containerV1.ContainerProbeHTTPProbe{Path: healthCheckReq.HTTP.Path}
		} else if healthCheckReq.TCP != nil {
			req.LivenessProbe.TCP = &containerV1.ContainerProbeTCPProbe{}
		}
	}

	if startupProbe, ok := d.GetOk("startup_probe"); ok {
		startupProbeReq, err := expandContainerProbe(startupProbe, "startup_probe")
		if err != nil {
			return nil, err
		}

		req.StartupProbe = startupProbeReq
	}

	if tags, ok := d.GetOk("tags"); ok {
		req.Tags = types.ExpandStrings(tags)
	}

	if command, ok := d.GetOk("command"); ok {
		req.Command = types.ExpandStrings(command)
	}

	if args, ok := d.GetOk("args"); ok {
		req.Args = types.ExpandStrings(args)
	}

	if pnID, ok := d.GetOk("private_network_id"); ok {
		req.PrivateNetworkID = types.ExpandStringPtr(locality.ExpandID(pnID.(string)))
	}

	return req, nil
}

func setUpdateContainerRequest(d *schema.ResourceData, region scw.Region, containerID string) (*containerV1.UpdateContainerRequest, error) {
	req := &containerV1.UpdateContainerRequest{
		Region:      region,
		ContainerID: containerID,
	}

	if d.HasChanges("environment_variables") {
		envVariablesRaw := d.Get("environment_variables")
		req.EnvironmentVariables = types.ExpandMapPtrStringString(envVariablesRaw)
	}

	if d.HasChanges("secret_environment_variables") {
		newEnv := d.Get("secret_environment_variables")
		req.SecretEnvironmentVariables = filterSecretEnvsToPatch(types.ExpandMapStringString(newEnv))
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChanges("min_scale") {
		req.MinScale = new(uint32(d.Get("min_scale").(int)))
	}

	if d.HasChanges("max_scale") {
		req.MaxScale = new(uint32(d.Get("max_scale").(int)))
	}

	if d.HasChanges("memory_limit_bytes", "memory_limit") {
		oldMemoryLimitBytes, newMemoryLimitBytes := d.GetChange("memory_limit_bytes")
		if oldMemoryLimitBytes != newMemoryLimitBytes {
			req.MemoryLimitBytes = new(scw.Size(newMemoryLimitBytes.(int)))
		} else {
			newMemoryLimitMB := d.Get("memory_limit")
			req.MemoryLimitBytes = new(scw.Size(newMemoryLimitMB.(int)) * scw.MB)
		}
	}

	if d.HasChanges("cpu_limit") {
		req.MvcpuLimit = new(uint32(d.Get("cpu_limit").(int)))
	}

	if d.HasChanges("timeout") {
		req.Timeout = &scw.Duration{Seconds: int64(d.Get("timeout").(int))}
	}

	if d.HasChanges("privacy") {
		req.Privacy = containerV1.ContainerPrivacy(d.Get("privacy").(string))
	}

	if d.HasChanges("description") {
		req.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
	}

	if d.HasChanges("image", "registry_image", "registry_sha256") {
		if oldImage, newImage := d.GetChange("image"); oldImage != newImage {
			req.Image = types.ExpandStringPtr(d.Get("image"))
		} else if oldRegistryImage, newRegistryImage := d.GetChange("registry_image"); oldRegistryImage != newRegistryImage {
			req.Image = types.ExpandStringPtr(d.Get("registry_image"))
		} else {
			if image := d.Get("image"); image != "" {
				req.Image = types.ExpandStringPtr(image)
			} else {
				req.Image = types.ExpandStringPtr(d.Get("registry_image"))
			}
		}
	}

	if d.HasChanges("protocol") {
		req.Protocol = containerV1.ContainerProtocol(d.Get("protocol").(string))
	}

	if d.HasChanges("port") {
		req.Port = new(uint32(d.Get("port").(int)))
	}

	if d.HasChanges("https_connections_only", "http_option") {
		oldHttpsConnectionOnly, newHttpsConnectionOnly := d.GetChange("https_connections_only")
		if oldHttpsConnectionOnly != newHttpsConnectionOnly {
			req.HTTPSConnectionOnly = new(newHttpsConnectionOnly.(bool))
		} else {
			newHttpOption := d.Get("http_option")
			switch newHttpOption {
			case containerV1Beta1.ContainerHTTPOptionEnabled.String():
				req.HTTPSConnectionOnly = types.ExpandBoolPtr(false)
			case containerV1Beta1.ContainerHTTPOptionRedirected.String():
				req.HTTPSConnectionOnly = types.ExpandBoolPtr(true)
			}
		}
	}

	if d.HasChanges("sandbox") {
		req.Sandbox = containerV1.ContainerSandbox(d.Get("sandbox").(string))
	}

	if d.HasChanges("scaling_option") {
		scalingOption := d.Get("scaling_option")

		scalingOptionReq, err := expandScalingOption(scalingOption)
		if err != nil {
			return nil, err
		}

		req.ScalingOption = scalingOptionReq
	}

	if d.HasChanges("liveness_probe", "health_check") {
		oldLivenessProbe, newLivenessProbe := d.GetChange("liveness_probe")
		if !reflect.DeepEqual(oldLivenessProbe, newLivenessProbe) {
			livenessProbeReq, err := expandContainerProbe(newLivenessProbe, "liveness_probe")
			if err != nil {
				return nil, err
			}

			req.LivenessProbe = livenessProbeReq
		} else {
			oldHealthCheck, newHealthCheck := d.GetChange("health_check")
			if !reflect.DeepEqual(oldHealthCheck, newHealthCheck) {
				livenessProbeReq, err := expandContainerProbeFromHealthCheck(newHealthCheck)
				if err != nil {
					return nil, err
				}

				req.LivenessProbe = livenessProbeReq
			}
		}
	}

	if d.HasChanges("startup_probe") {
		newStartupProbe := d.Get("startup_probe")

		startupProbe, err := expandContainerProbe(newStartupProbe, "startup_probe")
		if err != nil {
			return nil, err
		}

		if startupProbe == nil {
			req.StartupProbe = nil
		} else {
			startupProbeReq := &containerV1.UpdateContainerRequestProbe{
				FailureThreshold: &startupProbe.FailureThreshold,
				Interval:         startupProbe.Interval,
				Timeout:          startupProbe.Timeout,
			}

			if startupProbe.HTTP != nil {
				startupProbeReq.HTTP = &containerV1.UpdateContainerRequestProbeHTTPProbe{
					Path: new(startupProbe.HTTP.Path),
				}
			} else if startupProbe.TCP != nil {
				startupProbeReq.TCP = &containerV1.UpdateContainerRequestProbeTCPProbe{}
			}

			req.StartupProbe = startupProbeReq
		}
	}

	if d.HasChanges("local_storage_limit_bytes", "local_storage_limit") {
		oldLocalStorageLimitBytes, newLocalStorageLimitBytes := d.GetChange("local_storage_limit_bytes")
		if oldLocalStorageLimitBytes != newLocalStorageLimitBytes {
			req.LocalStorageLimitBytes = new(scw.Size(newLocalStorageLimitBytes.(int)))
		} else {
			newLocalStorageLimitMB := d.Get("local_storage_limit")
			req.LocalStorageLimitBytes = new(scw.Size(newLocalStorageLimitMB.(int)) * scw.MB)
		}
	}

	if d.HasChanges("command") {
		req.Command = types.ExpandUpdatedStringsPtr(d.Get("command"))
	}

	if d.HasChanges("args") {
		req.Args = types.ExpandUpdatedStringsPtr(d.Get("args"))
	}

	if d.HasChanges("private_network_id") {
		req.PrivateNetworkID = types.ExpandUpdatedStringPtr(locality.ExpandID(d.Get("private_network_id")))
	}

	return req, nil
}

func expandContainerProbeFromHealthCheck(healthCheck any) (*containerV1.ContainerProbe, error) {
	healthCheckList := healthCheck.([]any)
	healthCheckElem := healthCheckList[0].(map[string]any)

	healthCheckReq, err := expandHealthCheck(healthCheckElem)
	if err != nil {
		return nil, err
	}

	livenessProbeReq := &containerV1.ContainerProbe{
		FailureThreshold: healthCheckReq.FailureThreshold,
		Interval:         healthCheckReq.Interval,
		Timeout:          &scw.Duration{Seconds: 1}, // Timeout is required in liveness probe configuration but absent from health check, so we use the value set by the API on default liveness probes.
	}
	if healthCheckReq.TCP != nil {
		livenessProbeReq.TCP = &containerV1.ContainerProbeTCPProbe{}
	} else if healthCheckReq.HTTP != nil {
		livenessProbeReq.HTTP = &containerV1.ContainerProbeHTTPProbe{
			Path: healthCheckReq.HTTP.Path,
		}
	}

	return livenessProbeReq, nil
}

func flattenLivenessProbeAsHealthCheck(livenessProbe *containerV1.ContainerProbe) any {
	if livenessProbe == nil {
		return nil
	}

	var interval *time.Duration
	if livenessProbe.Interval != nil {
		interval = livenessProbe.Interval.ToTimeDuration()
	}

	flattenedHealthCheck := map[string]any{
		"http":              flattenContainerProbeHTTP(livenessProbe.HTTP),
		"failure_threshold": types.FlattenUint32Ptr(&livenessProbe.FailureThreshold),
		"interval":          types.FlattenDuration(interval),
	}

	if livenessProbe.TCP != nil {
		flattenedHealthCheck["tcp"] = true
	}

	return []map[string]any{flattenedHealthCheck}
}

func expandHealthCheck(healthCheck map[string]any) (*containerV1Beta1.ContainerHealthCheckSpec, error) {
	// All attributes are in fact required, but we had to mark them as Optional/Computed to ensure backward
	// compatibility, so we need to ensure that they are all present.
	healthCheckSpec := &containerV1Beta1.ContainerHealthCheckSpec{}

	if httpRaw, ok := healthCheck["http"]; ok {
		var err error

		healthCheckSpec.HTTP, err = expandHealthCheckHTTP(httpRaw.([]any))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("missing required attribute health_check.0.http")
	}

	if failureThreshold, ok := healthCheck["failure_threshold"]; ok {
		healthCheckSpec.FailureThreshold = uint32(failureThreshold.(int))
	} else {
		return nil, errors.New("missing required attribute health_check.0.failure_threshold")
	}

	if interval, ok := healthCheck["interval"]; ok {
		duration, err := types.ExpandDuration(interval)
		if err != nil {
			return nil, err
		}

		healthCheckSpec.Interval = scw.NewDurationFromTimeDuration(*duration)
	} else {
		return nil, errors.New("missing required attribute health_check.0.interval")
	}

	return healthCheckSpec, nil
}

func expandHealthCheckHTTP(healthCheckHTTPSchema []any) (*containerV1Beta1.ContainerHealthCheckSpecHTTPProbe, error) {
	healthCheckHTTP, ok := healthCheckHTTPSchema[0].(map[string]any)
	if !ok {
		return &containerV1Beta1.ContainerHealthCheckSpecHTTPProbe{}, nil
	}

	httpProbe := &containerV1Beta1.ContainerHealthCheckSpecHTTPProbe{}
	if path, ok := healthCheckHTTP["path"].(string); ok {
		httpProbe.Path = path
	} else {
		return nil, errors.New("missing required attribute health_check.0.http.0.path")
	}

	return httpProbe, nil
}

func expandContainerProbe(containerProbeSchema any, attributeName string) (*containerV1.ContainerProbe, error) {
	containerProbe, ok := containerProbeSchema.([]any)
	if !ok || len(containerProbe) != 1 {
		return nil, nil
	}

	rawProbe, isRawProbe := containerProbe[0].(map[string]any)
	if !isRawProbe {
		return nil, fmt.Errorf("expected container probe of type map[string]any, got %T", containerProbe[0])
	}

	containerProbeSpec := &containerV1.ContainerProbe{
		FailureThreshold: uint32(rawProbe["failure_threshold"].(int)),
	}

	if interval, ok := rawProbe["interval"]; ok {
		duration, err := types.ExpandDuration(interval)
		if err != nil {
			return nil, err
		}

		containerProbeSpec.Interval = scw.NewDurationFromTimeDuration(*duration)
	}

	if timeout, ok := rawProbe["timeout"]; ok {
		duration, err := types.ExpandDuration(timeout)
		if err != nil {
			return nil, err
		}

		containerProbeSpec.Timeout = scw.NewDurationFromTimeDuration(*duration)
	}

	tcpSetTrue := rawProbe["tcp"].(bool)
	http := rawProbe["http"].([]any)
	httpSetOK := len(http) == 1

	if (httpSetOK && tcpSetTrue) || (!httpSetOK && !tcpSetTrue) {
		return nil, fmt.Errorf("exactly one of \"%[1]s.http\" or \"%[1]s.tcp\" (set to true) must be defined", attributeName)
	}

	if httpSetOK {
		containerProbeSpec.HTTP = expandContainerProbeHTTP(http[0].(map[string]any))
	} else {
		containerProbeSpec.TCP = &containerV1.ContainerProbeTCPProbe{}
	}

	return containerProbeSpec, nil
}

func expandContainerProbeHTTP(rawHTTPProbe map[string]any) *containerV1.ContainerProbeHTTPProbe {
	httpProbe := &containerV1.ContainerProbeHTTPProbe{}
	if path, ok := rawHTTPProbe["path"].(string); ok {
		httpProbe.Path = path
	}

	return httpProbe
}

func flattenContainerProbe(probe *containerV1.ContainerProbe) any {
	if probe == nil {
		return nil
	}

	var interval *time.Duration
	if probe.Interval != nil {
		interval = probe.Interval.ToTimeDuration()
	}

	var timeout *time.Duration
	if probe.Timeout != nil {
		timeout = probe.Timeout.ToTimeDuration()
	}

	flattenedContainerProbe := map[string]any{
		"failure_threshold": types.FlattenUint32Ptr(&probe.FailureThreshold),
		"interval":          types.FlattenDuration(interval),
		"timeout":           types.FlattenDuration(timeout),
	}

	if probe.TCP != nil {
		flattenedContainerProbe["tcp"] = true
	} else if probe.HTTP != nil {
		flattenedContainerProbe["http"] = flattenContainerProbeHTTP(probe.HTTP)
	}

	return []map[string]any{flattenedContainerProbe}
}

func flattenContainerProbeHTTP(containerProbeHTTP *containerV1.ContainerProbeHTTPProbe) any {
	if containerProbeHTTP == nil {
		return nil
	}

	flattenedContainerProbeHTTP := make([]map[string]any, 0, 1)
	flattenedContainerProbeHTTP = append(flattenedContainerProbeHTTP, map[string]any{
		"path": types.FlattenStringPtr(&containerProbeHTTP.Path),
	})

	return flattenedContainerProbeHTTP
}

func expandScalingOption(scalingOptionSchema any) (*containerV1.ContainerScalingOption, error) {
	scalingOption, ok := scalingOptionSchema.(*schema.Set)
	if !ok {
		return &containerV1.ContainerScalingOption{}, nil
	}

	for _, option := range scalingOption.List() {
		rawOption, isRawOption := option.(map[string]any)
		if !isRawOption {
			continue
		}

		setFields := 0

		cso := &containerV1.ContainerScalingOption{}
		if concurrentRequestThresold, ok := rawOption["concurrent_requests_threshold"].(int); ok && concurrentRequestThresold != 0 {
			cso.ConcurrentRequestsThreshold = new(uint32(concurrentRequestThresold))
			setFields++
		}

		if cpuUsageThreshold, ok := rawOption["cpu_usage_threshold"].(int); ok && cpuUsageThreshold != 0 {
			cso.CPUUsageThreshold = new(uint32(cpuUsageThreshold))
			setFields++
		}

		if memoryUsageThreshold, ok := rawOption["memory_usage_threshold"].(int); ok && memoryUsageThreshold != 0 {
			cso.MemoryUsageThreshold = new(uint32(memoryUsageThreshold))
			setFields++
		}

		if setFields > 1 {
			return &containerV1.ContainerScalingOption{}, errors.New("a maximum of one scaling option can be set")
		}

		return cso, nil
	}

	return &containerV1.ContainerScalingOption{}, nil
}

func flattenScalingOption(scalingOption *containerV1.ContainerScalingOption) any {
	if scalingOption == nil {
		return nil
	}

	flattenedScalingOption := make([]map[string]any, 0, 1)
	flattenedScalingOption = append(flattenedScalingOption, map[string]any{
		"concurrent_requests_threshold": types.FlattenUint32Ptr(scalingOption.ConcurrentRequestsThreshold),
		"cpu_usage_threshold":           types.FlattenUint32Ptr(scalingOption.CPUUsageThreshold),
		"memory_usage_threshold":        types.FlattenUint32Ptr(scalingOption.MemoryUsageThreshold),
	})

	return flattenedScalingOption
}

func isContainerDNSResolveError(err error) bool {
	responseError := &scw.ResponseError{}

	if !errors.As(err, &responseError) {
		return false
	}

	if strings.HasPrefix(responseError.Message, "could not validate domain") {
		return true
	}

	return false
}

func retryCreateContainerDomain(ctx context.Context, containerAPI *containerV1.API, req *containerV1.CreateDomainRequest, timeout time.Duration) (*containerV1.Domain, error) {
	timeoutChannel := time.After(timeout)

	for {
		select {
		case <-time.After(DefaultContainerRetryInterval):
			domain, err := containerAPI.CreateDomain(req, scw.WithContext(ctx))
			if err != nil && isContainerDNSResolveError(err) {
				continue
			}

			return domain, err
		case <-timeoutChannel:
			return containerAPI.CreateDomain(req, scw.WithContext(ctx))
		}
	}
}

// filterSecretEnvsToPatch builds the list of secrets to be patched.
// - New secrets (which values are not hashed) should be added,
// - Unchanged secrets (which values are already hashed) should be passed as an empty string to indicate that no change is needed,
// - Old secrets which don't end up in the final list will be deleted.
func filterSecretEnvsToPatch(newEnv map[string]string) *map[string]string {
	toPatch := map[string]string{}

	for key, value := range newEnv {
		if !strings.HasPrefix(value, "$argon2id") {
			toPatch[key] = value
		} else {
			toPatch[key] = ""
		}
	}

	return &toPatch
}
