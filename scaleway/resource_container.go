package scaleway

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
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
			"region": regionSchema(),
		},
	}
}

func resourceScalewayContainerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiHandler, err := newContainerHandler(ctx, d, meta)
	if err != nil {
		return err
	}

	namespaceID := d.Get("namespace_id")
	// verify name space state
	_, err = apiHandler.waitForNameSpace(expandStringPtr(namespaceID))
	if err != nil {
		return err
	}

	c, err := apiHandler.waitForContainerCreation(d)
	if err != nil {
		return err
	}

	d.SetId(newRegionalIDString(apiHandler.region, c.ID))

	return resourceScalewayContainerRead(ctx, d, meta)
}

func resourceScalewayContainerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiHandler, err := newContainerHandler(ctx, d, meta)
	if err != nil {
		return err
	}

	_, id, errID := parseRegionalID(d.Id())
	if err != nil {
		return diag.FromErr(errID)
	}

	namespaceID := d.Get("namespace_id")
	// verify name space state
	ns, err := apiHandler.waitForNameSpace(expandStringPtr(namespaceID))
	if err != nil {
		return err
	}

	co, err := apiHandler.waitForContainer(id)
	if err != nil {
		return err
	}

	_ = d.Set("name", co.Name)
	_ = d.Set("namespace_id", newRegionalID(apiHandler.region, ns.ID))
	_ = d.Set("status", co.Status.String())
	_ = d.Set("error_message", co.ErrorMessage)
	_ = d.Set("environment_variables", flattenMap(co.EnvironmentVariables))
	_ = d.Set("min_scale", int(co.MinScale))
	_ = d.Set("max_scale", int(co.MaxScale))
	_ = d.Set("max_scale", int(co.MemoryLimit))
	_ = d.Set("cpu_limit", int(co.CPULimit))
	_ = d.Set("timeout", flattenDuration(co.Timeout.ToTimeDuration()))
	_ = d.Set("privacy", co.Privacy)
	_ = d.Set("description", co.Description)
	_ = d.Set("registry_image", co.RegistryImage)
	_ = d.Set("max_concurrency", co.MaxConcurrency)
	_ = d.Set("domain_name", co.DomainName)
	_ = d.Set("protocol", co.Protocol.String())
	_ = d.Set("cron_status", co.Status.String())
	_ = d.Set("port", co.Port)
	_ = d.Set("region", co.Region)

	return nil
}

func resourceScalewayContainerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiHandler, err := newContainerHandler(ctx, d, meta)
	if err != nil {
		return err
	}

	_, containerID, errID := parseRegionalID(d.Id())
	if err != nil {
		return diag.FromErr(errID)
	}

	namespaceID := d.Get("namespace_id")
	// verify name space state
	_, err = apiHandler.waitForNameSpace(expandStringPtr(namespaceID))
	if err != nil {
		return err
	}

	req := &container.UpdateContainerRequest{
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

	errUpdate := apiHandler.waitForUpdate(req)
	if err != nil {
		return errUpdate
	}

	return resourceScalewayContainerRead(ctx, d, meta)
}

func resourceScalewayContainerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiHandler, err := newContainerHandler(ctx, d, meta)
	if err != nil {
		return err
	}

	_, containerID, errID := parseRegionalID(d.Id())
	if err != nil {
		return diag.FromErr(errID)
	}

	namespaceID := d.Get("namespace_id")
	// verify name space state
	_, err = apiHandler.waitForNameSpace(expandStringPtr(namespaceID))
	if err != nil {
		return err
	}

	return apiHandler.waitForDelete(containerID)
}
