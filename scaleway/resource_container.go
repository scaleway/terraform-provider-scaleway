package scaleway

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/robfig/cron"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	containerMaxConcurrencyLimit      int = 80
	defaultWaitContainerRetryInterval     = 30 * time.Second
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
			Default: schema.DefaultTimeout(defaultContainerNamespaceTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
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
				Description: "The container namespace associated",
			},
			"environment_variables": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The environment variables to be injected into your container at runtime.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(0, 1000),
				},
				ValidateDiagFunc: validation.MapKeyLenBetween(0, 100),
			},
			"min_scale": {
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
				Description: "The minimum of running container instances continuously. Defaults to 0.",
			},
			"max_scale": {
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
				Description: "The maximum of number of instances this container can scale to. Default to 20.",
			},
			"memory_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
				Description: "The memory computing resources in MB to allocate to each container. Defaults to 128.",
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
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The privacy type define the way to authenticate to your container",
				Default:     container.ContainerPrivacyPublic,
				ValidateFunc: validation.StringInSlice([]string{
					container.ContainerPrivacyPublic.String(),
					container.ContainerPrivacyPrivate.String(),
				}, false),
			},
			"registry_image": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The scaleway registry image address",
			},
			"max_concurrency": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "The maximum the number of simultaneous requests your container can handle at the same time. Defaults to 50.",
				ValidateFunc: validation.IntAtMost(containerMaxConcurrencyLimit),
			},
			"domain_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The container domain name.",
			},
			"protocol": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The communication protocol http1 or h2c. Defaults to http1.",
				Default:     container.ContainerProtocolHTTP1.String(),
				ValidateFunc: validation.StringInSlice([]string{
					container.ContainerProtocolH2c.String(),
					container.ContainerProtocolHTTP1.String(),
				}, false),
			},
			"port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
				Description: "The port to expose the container. Defaults to 8080",
			},
			"deploy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "This allows you to control your production environment",
				Default:     false,
			},
			"cron_job": {
				Type:        schema.TypeSet,
				Optional:    true,
				Set:         cronContainerHash,
				Description: "Specifies how to run a job periodically on a given schedule, written in Cron format.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cron_job_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Cron Job ID",
						},
						"schedule": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateCronExpression(),
							Description:  "Cron format string, e.g. 0 * * * * or @hourly, as schedule time of its jobs to be created and executed.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Cron job status.",
						},
					},
				},
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
			"region": regionComputedSchema(),
		},
	}
}

func resourceScalewayContainerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := containerAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	if region.String() == "" {
		region = scw.RegionFrPar
	}
	namespaceID := d.Get("namespace_id")
	_, err = waitForContainerNamespace(ctx, api, region, expandID(namespaceID), d.Timeout(schema.TimeoutCreate))
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

	_, err = waitForContainer(ctx, api, res.ID, region, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.Errorf("unexpected waiting container error: %s", err)
	}

	// check if container should be deployed
	shouldDeploy := d.Get("deploy")
	if *expandBoolPtr(shouldDeploy) {
		reqUpdate := &container.UpdateContainerRequest{
			Region:      res.Region,
			ContainerID: res.ID,
			Redeploy:    expandBoolPtr(shouldDeploy),
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

	if cronJobsRaw, ok := d.GetOk("cron_job"); ok {
		cronJobs := cronJobsRaw.(*schema.Set).List()
		if len(cronJobs) > 0 {
			cronJobRequests := createCronJobsRequest(res.ID, region, cronJobs)
			cronErrors := createCronJobs(ctx, api, cronJobRequests, d.Timeout(schema.TimeoutCreate))
			if len(cronErrors) > 0 {
				for _, cronErr := range cronErrors {
					tflog.Error(ctx, cronErr.Error())
				}
				return diag.Errorf("error creating cron jobs")
			}
		}
	}

	d.SetId(newRegionalIDString(region, res.ID))

	return resourceScalewayContainerRead(ctx, d, meta)
}

func createCronJobs(ctx context.Context, api *container.API, cronJobs []*container.CreateCronRequest, timeout time.Duration) []error {
	var errors []error
	for _, r := range cronJobs {
		c, err := api.CreateCron(r, scw.WithContext(ctx))
		if err != nil {
			errors = append(errors, err)
			continue
		}
		tflog.Info(ctx, fmt.Sprintf("[INFO] Submitted new cron job: %#v", r.Schedule))
		_, err = waitForContainerCron(ctx, api, c.ID, r.Region, timeout)
		if err != nil {
			errors = append(errors, err)
		}
		tflog.Info(ctx, "[INFO] cron job ready")
	}

	return errors
}

func resourceScalewayContainerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, containerID, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	co, err := waitForContainer(ctx, api, containerID, region, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.Errorf("unexpected waiting container error: %s", err)
	}

	_ = d.Set("name", co.Name)
	_ = d.Set("namespace_id", newRegionalID(region, co.NamespaceID).String())
	_ = d.Set("status", co.Status.String())
	_ = d.Set("error_message", co.ErrorMessage)
	_ = d.Set("environment_variables", flattenMap(co.EnvironmentVariables))
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
	_ = d.Set("deploy", scw.BoolPtr(*expandBoolPtr(d.Get("deploy"))))
	_ = d.Set("region", co.Region.String())

	// read Container Cron Jobs
	cronJobs, err := api.ListCrons(&container.ListCronsRequest{
		ContainerID: co.ID,
		Region:      co.Region,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("cron_job", schema.NewSet(cronContainerHash, flattenContainerCronJobs(cronJobs.Crons)))
	return nil
}

func resourceScalewayContainerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, containerID, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceID := d.Get("namespace_id")
	// verify name space state
	_, err = waitForContainerNamespace(ctx, api, region, expandID(namespaceID), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.Errorf("unexpected namespace error: %s", err)
	}

	// check for container state
	_, err = waitForContainer(ctx, api, containerID, region, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.Errorf("unexpected waiting container error: %s", err)
	}

	// Warning or Errors can be collected as warnings
	var diags diag.Diagnostics

	// update container
	req := &container.UpdateContainerRequest{
		Region:      region,
		ContainerID: containerID,
	}

	if d.HasChanges("environment_variables") {
		envVariablesRaw := d.Get("environment_variables")
		req.EnvironmentVariables = expandMapStringStringPtr(envVariablesRaw)
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
		timeout := d.Get("timeout")
		req.Timeout = &scw.Duration{Seconds: timeout.(int64)}
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

	if d.HasChanges("deploy") {
		req.Redeploy = expandBoolPtr(d.Get("deploy"))
	}

	currentContainer, err := api.UpdateContainer(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("cron_job") {
		first, after := d.GetChange("cron_job")
		isEqual := first.(*schema.Set).HashEqual(after)
		tflog.Info(ctx, fmt.Sprintf("cron job has change: %v", isEqual))
		cronJobs := d.Get("cron_job").(*schema.Set).List()
		if len(cronJobs) > 0 {
			for index, c := range cronJobs {
				cj := c.(map[string]interface{})
				keyJob := fmt.Sprintf("cron_job.%d", index)
				if d.HasChange(keyJob) {
					req := &container.UpdateCronRequest{
						Region:      region,
						CronID:      cj["cron_job_id"].(string),
						ContainerID: &currentContainer.ID,
						Schedule:    expandStringPtr(cj["schedule"]),
					}
					_, err := api.UpdateCron(req, scw.WithContext(ctx))
					if err != nil {
						return diag.FromErr(err)
					}
				}
			}
		}

		// check triggers associated
		triggers, err := api.ListCrons(&container.ListCronsRequest{
			Region:      region,
			ContainerID: containerID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		// wait for triggers state
		for _, c := range triggers.Crons {
			_, err = waitForContainerCron(ctx, api, c.ID, region, d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  fmt.Sprintf("Warning waiting cron job %s", c.ID),
					Detail:   err.Error(),
				})
			}
		}
	}

	return append(diags, resourceScalewayContainerRead(ctx, d, meta)...)
}

func resourceScalewayContainerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, containerID, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// check for container state
	_, err = waitForContainer(ctx, api, containerID, region, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.Errorf("unexpected waiting container error: %s", err)
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

func validateCronExpression() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of '%s' to be string", k))
			return
		}
		_, err := cron.ParseStandard(v)
		if err != nil {
			es = append(es, fmt.Errorf("'%s' should be an valid Cron expression", k))
		}
		return
	}
}

func createCronJobsRequest(containerID string, region scw.Region, cronJobs []interface{}) []*container.CreateCronRequest {
	var requests []*container.CreateCronRequest

	for _, cronJob := range cronJobs {
		config := cronJob.(map[string]interface{})
		requests = append(requests, &container.CreateCronRequest{
			ContainerID: expandID(containerID),
			Region:      region,
			Schedule:    config["schedule"].(string),
			Args:        []byte(`{"scw":{"provider":"Scaleway"}}`), // ARGS will be implemented in a dedicated PR
		})
	}

	return requests
}
