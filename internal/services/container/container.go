package container

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/container/v1"
	containerBeta "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
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
			Description: "The minimum of number of instances this container can scale to.",
		},
		"max_scale": {
			Type:        schema.TypeInt,
			Computed:    true,
			Optional:    true,
			Description: "The maximum of number of instances this container can scale to.",
		},
		"memory_limit": {
			Type:          schema.TypeInt,
			Computed:      true,
			Optional:      true,
			Description:   "The memory computing resources in MB to allocate to each container.",
			Deprecated:    "Please use memory_limit_bytes instead",
			ConflictsWith: []string{"memory_limit_bytes"},
		},
		"memory_limit_bytes": {
			Type:          schema.TypeInt,
			Computed:      true,
			Optional:      true,
			Description:   "The memory computing resources in bytes to allocate to each container.",
			ConflictsWith: []string{"memory_limit"},
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
			Description:      "The privacy type defines the way to authenticate to your container",
			Default:          container.ContainerPrivacyPublic,
			ValidateDiagFunc: verify.ValidateEnum[container.ContainerPrivacy](),
		},
		"registry_image": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			Description:  "The scaleway registry image address",
			Deprecated:   "Please use image instead",
			ExactlyOneOf: []string{"image"},
		},
		"image": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			Description:  "The image reference (e.g. \"rg.fr-par.scw.cloud/my-registry-namespace/image:tag\" or \"nginx:latest\").",
			ExactlyOneOf: []string{"registry_image"},
		},
		"registry_sha256": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The sha256 of your source registry image, changing it will re-apply the deployment. Can be any string",
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
			Deprecated:  "Containers are now automatically deployed or redeployed; setting this attribute will not have any effect.",
		},
		"https_connections_only": {
			Type:          schema.TypeBool,
			Optional:      true,
			Computed:      true,
			Description:   "If true, it will allow only HTTPS connections to access your container to prevent it from being triggered by insecure connections (HTTP).",
			ConflictsWith: []string{"http_option"},
		},
		"http_option": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Description:      "HTTP traffic configuration",
			Deprecated:       "Please use https_connections_only instead",
			ValidateDiagFunc: verify.ValidateEnum[containerBeta.ContainerHTTPOption](),
			ConflictsWith:    []string{"https_connections_only"},
		},
		"sandbox": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Description:      "Execution environment of the container.",
			ValidateDiagFunc: verify.ValidateEnum[container.ContainerSandbox](),
		},
		"health_check": {
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Description: "Health check configuration of the container.",
			Deprecated:  "Please use liveness_probe instead",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"tcp": {
						Type:        schema.TypeBool,
						Description: "Perform TCP check on the container",
						Optional:    true,
						Computed:    true,
					},
					"http": {
						Type:        schema.TypeList,
						Description: "HTTP health check configuration.",
						Computed:    true,
						Optional:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"path": {
									Type:        schema.TypeString,
									Description: "Path to use for the HTTP health check.",
									Optional:    true,
									Computed:    true,
								},
							},
						},
					},
					"failure_threshold": {
						Type:        schema.TypeInt,
						Description: "Number of consecutive health check failures before considering the container unhealthy.",
						Optional:    true,
						Computed:    true,
					},
					"interval": {
						Type:             schema.TypeString,
						Description:      "Period between health checks.",
						DiffSuppressFunc: dsf.Duration,
						ValidateDiagFunc: verify.IsDuration(),
						Optional:         true,
						Computed:         true,
					},
				},
			},
		},
		"liveness_probe": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Computed:    true,
			Description: "Defines how to check if the container is running.",
			Elem:        containerProbeSchema(),
		},
		"startup_probe": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "Defines how to check if the container has started successfully.",
			Elem:        containerProbeSchema(),
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
			Type:          schema.TypeInt,
			Description:   "Local storage limit of the container (in MB)",
			Deprecated:    "Please use local_storage_limit_bytes instead",
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"local_storage_limit_bytes"},
		},
		"local_storage_limit_bytes": {
			Type:          schema.TypeInt,
			Description:   "Local storage limit of the container (in bytes)",
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"local_storage_limit"},
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
		"domain_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The native container domain name.",
			Deprecated:  "This attribute will be removed in the future, please use public_endpoint instead",
		},
		"public_endpoint": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Public URL of the container. This is the default endpoint generated by Scaleway to access the container from the Internet.",
		},
		"status": {
			Type:        schema.TypeString,
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

func containerProbeSchema() any {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"failure_threshold": {
				Type:        schema.TypeInt,
				Description: "Number of consecutive failures before considering the container has to be restarted.",
				Required:    true,
			},
			"interval": {
				Type:             schema.TypeString,
				Description:      "Time interval between checks (in duration notation).",
				DiffSuppressFunc: dsf.Duration,
				ValidateDiagFunc: verify.IsDuration(),
				Required:         true,
			},
			"timeout": {
				Type:             schema.TypeString,
				Description:      "Duration before the check times out (in duration notation).",
				DiffSuppressFunc: dsf.Duration,
				ValidateDiagFunc: verify.IsDuration(),
				Required:         true,
			},
			"http": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Description: "Perform HTTP check on the container with the specified path.",
				Optional:    true,
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
			"tcp": {
				Type:        schema.TypeBool,
				Description: "Perform TCP check on the container",
				Optional:    true,
			},
		},
	}
}

func ResourceContainerCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceID := locality.ExpandID(d.Get("namespace_id").(string))
	// verify namespace state
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

	d.SetId(regional.NewIDString(region, res.ID))

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
	_ = d.Set("memory_limit_bytes", int(co.MemoryLimitBytes))
	_ = d.Set("memory_limit", int(co.MemoryLimitBytes/scw.MB))
	_ = d.Set("cpu_limit", int(co.MvcpuLimit))
	_ = d.Set("timeout", co.Timeout.Seconds)
	_ = d.Set("privacy", co.Privacy.String())
	_ = d.Set("description", co.Description)
	_ = d.Set("registry_image", co.Image)
	_ = d.Set("image", co.Image)
	_ = d.Set("public_endpoint", co.PublicEndpoint)
	_ = d.Set("domain_name", strings.TrimPrefix(co.PublicEndpoint, "https://"))
	_ = d.Set("protocol", co.Protocol.String())
	_ = d.Set("cron_status", co.Status.String())
	_ = d.Set("port", int(co.Port))
	_ = d.Set("https_connections_only", co.HTTPSConnectionsOnly)

	if co.HTTPSConnectionsOnly {
		_ = d.Set("http_option", containerBeta.ContainerHTTPOptionRedirected.String())
	} else {
		_ = d.Set("http_option", containerBeta.ContainerHTTPOptionEnabled.String())
	}

	_ = d.Set("sandbox", co.Sandbox)
	_ = d.Set("health_check", flattenLivenessProbeAsHealthCheck(co.LivenessProbe))
	_ = d.Set("liveness_probe", flattenContainerProbe(co.LivenessProbe))
	_ = d.Set("startup_probe", flattenContainerProbe(co.StartupProbe))
	_ = d.Set("scaling_option", flattenScalingOption(co.ScalingOption))
	_ = d.Set("region", co.Region.String())
	_ = d.Set("local_storage_limit_bytes", int(co.LocalStorageLimitBytes))
	_ = d.Set("local_storage_limit", int(co.LocalStorageLimitBytes/scw.MB))
	_ = d.Set("secret_environment_variables", co.SecretEnvironmentVariables)
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
	// verify namespace state
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
