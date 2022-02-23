package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	containerMaxLimit uint64 = 80
)

func resourceScalewayContainer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayContainerCreate,
		ReadContext:   resourceScalewayContainerRead,
		UpdateContext: resourceScalewayContainerUpdate,
		DeleteContext: resourceScalewayContainerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultContainerTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The container name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The container description",
			},
			"namespace_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The namespace associated with the container",
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The container status",
			},
			"environment_variables": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The environment variables to be injected into your container at runtime ",
			},
			"min_scale": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The minimum of running container instances continuously (default: 0)",
				Default:     0,
			},
			"max_scale": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum of number of instances this container can scale to (default: 20)",
				Default:     20,
			},
			"memory_limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The memory computing resources in MB to allocate to each container",
			},
			"cpu_limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The CPU computing resources to allocate to each container",
			},
			"timeout": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The maximum amount of time in seconds during which your container can process a request before we stop it.",
				ValidateFunc: validateDuration(),
			},
			"error_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The error description",
			},
			"privacy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The privacy type",
				Default:     container.ContainerPrivacyPublic,
				ValidateFunc: validation.StringInSlice([]string{
					container.ContainerPrivacyUnknownPrivacy.String(),
					container.ContainerPrivacyPublic.String(),
					container.ContainerPrivacyPrivate.String(),
				}, false),
			},
			"registry_image": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The registry image address where your container is stored.",
			},
			"max_concurrency": {
				Type:             schema.TypeInt,
				Optional:         true,
				Description:      "The maximum the number of simultaneous requests your container can handle at the same time.",
				ValidateDiagFunc: validateMaxLimit(containerMaxLimit),
				Default:          80,
			},
			"domain_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The container domain name.",
			},
			"protocol": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The communication protocol. default (unknown_protocol)",
				Default:     container.ContainerProtocolUnknownProtocol.String(),
				ValidateFunc: validation.StringInSlice([]string{
					container.ContainerProtocolH2c.String(),
					container.ContainerProtocolHTTP1.String(),
					container.ContainerProtocolUnknownProtocol.String()},
					false),
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The port to expose the container",
			},
			"cron_status": {
				Type:        schema.TypeString,
				Description: "The cron status",
				Computed:    true,
			},
			"redeploy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Allow redeploy container",
				Default:     false,
			},
			"region": regionSchema(),
		},
	}
}

func resourceScalewayContainerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := containerAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceID := d.Get("namespace_id")
	// verify name space state
	_, err = api.WaitForNamespace(&container.WaitForNamespaceRequest{
		NamespaceID:   expandID(namespaceID),
		Region:        region,
		Timeout:       scw.TimeDurationPtr(defaultRegistryNamespaceTimeout),
		RetryInterval: DefaultWaitRetryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.Errorf("unexpected namespace error: %s", err)
	}

	req := setCreateContainerRequest(d, region)
	res, err := api.CreateContainer(req, scw.WithContext(ctx))
	if err != nil {
		return diag.Errorf("unexpected waiting container error: %s", err)
	}

	d.SetId(newRegionalIDString(region, res.ID))

	return resourceScalewayContainerRead(ctx, d, meta)
}

func resourceScalewayContainerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, containerID, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceID := d.Get("namespace_id")
	// verify name space state
	_, err = api.WaitForNamespace(&container.WaitForNamespaceRequest{
		NamespaceID:   expandID(namespaceID),
		Region:        region,
		Timeout:       scw.TimeDurationPtr(defaultRegistryNamespaceTimeout),
		RetryInterval: DefaultWaitRetryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.Errorf("unexpected namespace error: %s", err)
	}

	// check for container state
	co, err := api.WaitForContainer(&container.WaitForContainerRequest{ContainerID: containerID,
		Region:        region,
		Timeout:       scw.TimeDurationPtr(defaultContainerTimeout),
		RetryInterval: DefaultWaitRetryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.Errorf("unexpected waiting container error: %s", err)
	}

	_ = d.Set("name", co.Name)
	_ = d.Set("namespace_id", newRegionalID(region, co.NamespaceID))
	_ = d.Set("status", co.Status.String())
	_ = d.Set("error_message", co.ErrorMessage)
	_ = d.Set("environment_variables", flattenMap(co.EnvironmentVariables))
	_ = d.Set("min_scale", int(co.MinScale))
	_ = d.Set("max_scale", int(co.MaxScale))
	_ = d.Set("max_scale", int(co.MemoryLimit))
	_ = d.Set("cpu_limit", int(co.CPULimit))
	_ = d.Set("timeout", flattenDuration(co.Timeout.ToTimeDuration()))
	_ = d.Set("privacy", co.Privacy.String())
	_ = d.Set("description", *co.Description)
	_ = d.Set("registry_image", co.RegistryImage)
	_ = d.Set("max_concurrency", int(co.MaxConcurrency))
	_ = d.Set("domain_name", co.DomainName)
	_ = d.Set("protocol", co.Protocol.String())
	_ = d.Set("cron_status", co.Status.String())
	_ = d.Set("port", int(co.Port))
	_ = d.Set("redeploy", *expandBoolPtr(d.Get("redeploy")))
	_ = d.Set("region", co.Region.String())

	return nil
}

func resourceScalewayContainerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, containerID, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceID := d.Get("namespace_id")
	// verify name space state
	_, err = api.WaitForNamespace(&container.WaitForNamespaceRequest{
		NamespaceID:   expandID(namespaceID),
		Region:        region,
		Timeout:       scw.TimeDurationPtr(defaultRegistryNamespaceTimeout),
		RetryInterval: DefaultWaitRetryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.Errorf("unexpected namespace error: %s", err)
	}

	// check for container state
	_, err = api.WaitForContainer(&container.WaitForContainerRequest{ContainerID: containerID,
		Region:        region,
		Timeout:       scw.TimeDurationPtr(defaultContainerTimeout),
		RetryInterval: DefaultWaitRetryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.Errorf("unexpected waiting container error: %s", err)
	}

	// Warning or Errors can be collected as warnings
	var diags diag.Diagnostics

	// check triggers associated
	triggers, errList := api.ListCrons(&container.ListCronsRequest{
		Region:      region,
		ContainerID: containerID,
	}, scw.WithContext(ctx))
	if errList != nil {
		return diag.FromErr(errList)
	}

	// wait for triggers state
	for _, c := range triggers.Crons {
		_, err := api.WaitForCron(&container.WaitForCronRequest{
			CronID:        c.ID,
			Region:        region,
			Timeout:       scw.TimeDurationPtr(defaultContainerTimeout),
			RetryInterval: DefaultWaitRetryInterval,
		}, scw.WithContext(ctx))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Warning waiting cron job",
				Detail:   err.Error(),
			})
		}
	}

	// update container
	req := &container.UpdateContainerRequest{
		Region:      region,
		ContainerID: containerID,
	}

	if d.HasChanges("environment_variables") {
		envVariablesRaw := d.Get("environment_variables")
		req.EnvironmentVariables = expandMapStr(envVariablesRaw)
	}

	if d.HasChanges("min_scale") {
		req.MinScale = toUint32(d.Get("min_scale"))
	}

	if d.HasChanges("max_scale") {
		req.MaxScale = toUint32(d.Get("max_scale"))
	}

	if d.HasChanges("memory_limit") {
		req.MemoryLimit = toUint32(d.Get("memory_limit"))
	}

	if d.HasChanges("timeout") {
		req.Timeout = toDuration(d.Get("timeout"))
	}

	if d.HasChanges("privacy") {
		req.Privacy = container.ContainerPrivacy(*expandStringPtr(d.Get("privacy")))
	}

	if d.HasChanges("description") {
		req.Description = expandStringPtr(d.Get("description"))
	}

	if d.HasChanges("registry_image") {
		req.RegistryImage = expandStringPtr(d.Get("registry_image"))
	}

	if d.HasChanges("domain_name") {
		req.DomainName = expandStringPtr(d.Get("domain_name"))
	}

	if d.HasChanges("max_concurrency") {
		req.MaxConcurrency = toUint32(d.Get("max_concurrency"))
	}

	if d.HasChanges("protocol") {
		req.Protocol = container.ContainerProtocol(*expandStringPtr(d.Get("protocol")))
	}

	if d.HasChanges("port") {
		req.Port = toUint32(d.Get("port"))
	}

	if d.HasChanges("redeploy") {
		req.Redeploy = expandBoolPtr(d.Get("redeploy"))
	}

	_, err = api.UpdateContainer(
		req,
		scw.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return append(diags, resourceScalewayContainerRead(ctx, d, meta)...)
}

func resourceScalewayContainerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, containerID, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceID := d.Get("namespace_id")
	// verify name space state
	_, err = api.WaitForNamespace(&container.WaitForNamespaceRequest{
		NamespaceID:   expandID(namespaceID),
		Region:        region,
		Timeout:       scw.TimeDurationPtr(defaultRegistryNamespaceTimeout),
		RetryInterval: DefaultWaitRetryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.Errorf("unexpected namespace error: %s", err)
	}

	// check for container state
	_, err = api.WaitForContainer(&container.WaitForContainerRequest{ContainerID: containerID,
		Region:        region,
		Timeout:       scw.TimeDurationPtr(defaultContainerTimeout),
		RetryInterval: DefaultWaitRetryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.Errorf("unexpected waiting container error: %s", err)
	}

	// Warning or Errors can be collected as warnings
	var diags diag.Diagnostics

	// check triggers associated
	triggers, errList := api.ListCrons(&container.ListCronsRequest{
		Region:      region,
		ContainerID: containerID,
	}, scw.WithContext(ctx))
	if errList != nil {
		return diag.FromErr(errList)
	}

	// wait for triggers state
	for _, c := range triggers.Crons {
		_, err := api.WaitForCron(&container.WaitForCronRequest{
			CronID:        c.ID,
			Region:        region,
			Timeout:       scw.TimeDurationPtr(defaultContainerTimeout),
			RetryInterval: DefaultWaitRetryInterval,
		}, scw.WithContext(ctx))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Warning waiting cron job",
				Detail:   err.Error(),
			})
		}
	}

	// delete triggers

	// delete container
	_, err = api.DeleteContainer(&container.DeleteContainerRequest{
		Region:      region,
		ContainerID: containerID},
		scw.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
