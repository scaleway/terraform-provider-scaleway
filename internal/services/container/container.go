package container

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

const (
	containerMaxConcurrencyLimit int = 80
)

func ResourceContainer() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceContainerCreate,
		ReadContext:   ResourceContainerRead,
		UpdateContext: ResourceContainerUpdate,
		DeleteContext: ResourceContainerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultContainerTimeout),
			Read:    schema.DefaultTimeout(defaultContainerTimeout),
			Update:  schema.DefaultTimeout(defaultContainerTimeout),
			Delete:  schema.DefaultTimeout(defaultContainerTimeout),
			Default: schema.DefaultTimeout(defaultContainerTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    containerSchema,
		Identity:      identity.DefaultRegional(),
	}
}

func containerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Computed:    true,
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
			ForceNew:    true,
			Description: "The container namespace associated",
		},
		"tags": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "List of tags [\"tag1\", \"tag2\", ...] attached to the container.",
		},
		"environment_variables": {
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Description: "The environment variables to be injected into your container at runtime.",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			ValidateDiagFunc: validation.MapKeyLenBetween(0, 100),
		},
		"secret_environment_variables": {
			Type:        schema.TypeMap,
			Optional:    true,
			Sensitive:   true,
			Description: "The secret environment variables to be injected into your container at runtime.",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			ValidateDiagFunc:      validation.MapKeyLenBetween(0, 100),
			DiffSuppressFunc:      dsf.CompareArgon2idPasswordAndHash,
			DiffSuppressOnRefresh: true,
		},
		"min_scale": {
			Type:        schema.TypeInt,
			Computed:    true,
			Optional:    true,
			Description: "The minimum of running container instances continuously.",
		},
		"max_scale": {
			Type:        schema.TypeInt,
			Computed:    true,
			Optional:    true,
			Description: "The maximum of number of instances this container can scale to.",
		},
		"memory_limit": {
			Type:        schema.TypeInt,
			Computed:    true,
			Optional:    true,
			Description: "The memory computing resources in MB to allocate to each container.",
		},
		"cpu_limit": {
			Type:        schema.TypeInt,
			Computed:    true,
			Optional:    true,
			Description: "The amount of vCPU computing resources to allocate to each container. Defaults to 70.",
		},
		"timeout": {
			Type:        schema.TypeInt,
			Computed:    true,
			Optional:    true,
			Description: "The maximum amount of time in seconds during which your container can process a request before we stop it. Defaults to 300s.",
		},
		"privacy": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The privacy type define the way to authenticate to your container",
			Default:          container.ContainerPrivacyPublic,
			ValidateDiagFunc: verify.ValidateEnum[container.ContainerPrivacy](),
		},
		"registry_image": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The scaleway registry image address",
		},
		"registry_sha256": {
			Type:         schema.TypeString,
			Optional:     true,
			RequiredWith: []string{"registry_image"},
			Description:  "The sha256 of your source registry image, changing it will re-apply the deployment. Can be any string",
		},
		"max_concurrency": {
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			Deprecated:   "Use scaling_option.concurrent_requests_threshold instead. This attribute will be removed.",
			Description:  "The maximum the number of simultaneous requests your container can handle at the same time.",
			ValidateFunc: validation.IntAtMost(containerMaxConcurrencyLimit),
		},
		"domain_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The native container domain name.",
		},
		"protocol": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The communication protocol http1 or h2c. Defaults to http1.",
			Default:          container.ContainerProtocolHTTP1.String(),
			ValidateDiagFunc: verify.ValidateEnum[container.ContainerProtocol](),
		},
		"port": {
			Type:        schema.TypeInt,
			Computed:    true,
			Optional:    true,
			Description: "The port to expose the container.",
		},
		"deploy": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "This allows you to control your production environment",
			Default:     false,
		},
		"http_option": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "HTTP traffic configuration",
			Default:          container.ContainerHTTPOptionEnabled.String(),
			ValidateDiagFunc: verify.ValidateEnum[container.ContainerHTTPOption](),
		},
		"sandbox": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Description:      "Execution environment of the container.",
			ValidateDiagFunc: verify.ValidateEnum[container.ContainerSandbox](),
		},
		"health_check": {
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Description: "Health check configuration of the container.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					// TCP has not been implemented yet in the API SDK, that's why the parameter is not in the schema.
					// See container.ContainerHealthCheckSpecTCPProbe.
					"http": {
						Type:        schema.TypeSet,
						Description: "HTTP health check configuration.",
						Required:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"path": {
									Type:        schema.TypeString,
									Description: "Path to use for the HTTP health check.",
									Required:    true,
								},
							},
						},
					},
					"failure_threshold": {
						Type:        schema.TypeInt,
						Description: "Number of consecutive health check failures before considering the container unhealthy.",
						Required:    true,
					},
					"interval": {
						Type:             schema.TypeString,
						Description:      "Period between health checks.",
						DiffSuppressFunc: dsf.Duration,
						ValidateDiagFunc: verify.IsDuration(),
						Required:         true,
					},
				},
			},
		},
		"scaling_option": {
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Description: "Configuration used to decide when to scale up or down.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"concurrent_requests_threshold": {
						Type:        schema.TypeInt,
						Description: "Scale depending on the number of concurrent requests being processed per container instance.",
						Optional:    true,
					},
					"cpu_usage_threshold": {
						Type:        schema.TypeInt,
						Description: "Scale depending on the CPU usage of a container instance.",
						Optional:    true,
					},
					"memory_usage_threshold": {
						Type:        schema.TypeInt,
						Description: "Scale depending on the memory usage of a container instance.",
						Optional:    true,
					},
				},
			},
		},
		"local_storage_limit": {
			Type:        schema.TypeInt,
			Description: "Local storage limit of the container (in MB)",
			Optional:    true,
			Computed:    true,
		},
		"command": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
			Description: "Command executed when the container starts. Overrides the command from the container image.",
		},
		"args": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
			Description: "Arguments passed to the command from the command \"field\". Overrides the arguments from the container image.",
		},
		"private_network_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "ID of the Private Network the container is connected to",
		},
		// computed
		"status": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The container status",
			Computed:    true,
		},
		"cron_status": {
			Type:        schema.TypeString,
			Description: "The cron status",
			Computed:    true,
		},
		"error_message": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The error description",
		},
		"region": regional.Schema(),
	}
}

func ResourceContainerCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceID := locality.ExpandID(d.Get("namespace_id").(string))
	// verify name space state
	_, err = waitForNamespace(ctx, api, region, namespaceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.Errorf("unexpected namespace error: %s", err)
	}

	req, err := setCreateContainerRequest(d, region)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.CreateContainer(req, scw.WithContext(ctx))
	if err != nil {
		return diag.Errorf("creation container error: %s", err)
	}

	// check if container should be deployed
	shouldDeploy := d.Get("deploy")
	if *types.ExpandBoolPtr(shouldDeploy) {
		_, err = waitForContainer(ctx, api, res.ID, region, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.Errorf("unexpected waiting container error: %s", err)
		}

		reqUpdate := &container.UpdateContainerRequest{
			Region:      res.Region,
			ContainerID: res.ID,
			Redeploy:    types.ExpandBoolPtr(shouldDeploy),
		}

		_, err = api.UpdateContainer(reqUpdate, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForContainer(ctx, api, res.ID, region, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.Errorf("unexpected waiting container error: %s", err)
		}
	}

	err = identity.SetRegionalIdentity(d, res.Region, res.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceContainerRead(ctx, d, m)
}

func ResourceContainerRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, containerID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	co, err := waitForContainer(ctx, api, containerID, region, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.Errorf("unexpected waiting container error: %s", err)
	}

	_ = d.Set("name", co.Name)
	_ = d.Set("namespace_id", regional.NewID(region, co.NamespaceID).String())
	_ = d.Set("status", co.Status.String())
	_ = d.Set("error_message", co.ErrorMessage)
	_ = d.Set("environment_variables", types.FlattenMap(co.EnvironmentVariables))
	_ = d.Set("min_scale", int(co.MinScale))
	_ = d.Set("max_scale", int(co.MaxScale))
	_ = d.Set("memory_limit", int(co.MemoryLimit))
	_ = d.Set("cpu_limit", int(co.CPULimit))
	_ = d.Set("timeout", co.Timeout.Seconds)
	_ = d.Set("privacy", co.Privacy.String())
	_ = d.Set("description", scw.StringPtr(*co.Description))
	_ = d.Set("registry_image", co.RegistryImage)
	_ = d.Set("max_concurrency", int(co.MaxConcurrency))
	_ = d.Set("domain_name", co.DomainName)
	_ = d.Set("protocol", co.Protocol.String())
	_ = d.Set("cron_status", co.Status.String())
	_ = d.Set("port", int(co.Port))
	_ = d.Set("deploy", scw.BoolPtr(*types.ExpandBoolPtr(d.Get("deploy"))))
	_ = d.Set("http_option", co.HTTPOption)
	_ = d.Set("sandbox", co.Sandbox)
	_ = d.Set("health_check", flattenHealthCheck(co.HealthCheck))
	_ = d.Set("scaling_option", flattenScalingOption(co.ScalingOption))
	_ = d.Set("region", co.Region.String())
	_ = d.Set("local_storage_limit", int(co.LocalStorageLimit))
	_ = d.Set("secret_environment_variables", flattenContainerSecrets(co.SecretEnvironmentVariables))
	_ = d.Set("tags", types.FlattenSliceString(co.Tags))
	_ = d.Set("command", types.FlattenSliceString(co.Command))
	_ = d.Set("args", types.FlattenSliceString(co.Args))

	if co.PrivateNetworkID != nil {
		_ = d.Set("private_network_id", regional.NewID(region, types.FlattenStringPtr(co.PrivateNetworkID).(string)).String())
	} else {
		_ = d.Set("private_network_id", nil)
	}

	return nil
}

func ResourceContainerUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, containerID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceID := d.Get("namespace_id")
	// verify name space state
	_, err = waitForNamespace(ctx, api, region, locality.ExpandID(namespaceID), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.Errorf("unexpected namespace error: %s", err)
	}

	// check for container state
	_, err = waitForContainer(ctx, api, containerID, region, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.Errorf("unexpected waiting container error: %s", err)
	}

	// update container
	req, err := setUpdateContainerRequest(d, region, containerID)
	if err != nil {
		return diag.FromErr(err)
	}

	con, err := api.UpdateContainer(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForContainer(ctx, api, con.ID, region, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceContainerRead(ctx, d, m)
}

func ResourceContainerDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, containerID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// check for container state
	_, err = waitForContainer(ctx, api, containerID, region, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	// delete container
	_, err = api.DeleteContainer(&container.DeleteContainerRequest{
		Region:      region,
		ContainerID: containerID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
