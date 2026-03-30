package datawarehouse

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	datawarehouseapi "github.com/scaleway/scaleway-sdk-go/api/datawarehouse/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceDeployment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeploymentCreate,
		ReadContext:   resourceDeploymentRead,
		UpdateContext: resourceDeploymentUpdate,
		DeleteContext: resourceDeploymentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
				cpuMin := diff.Get("cpu_min").(int)
				cpuMax := diff.Get("cpu_max").(int)

				if cpuMin > cpuMax {
					return fmt.Errorf("cpu_min (%d) must be <= cpu_max (%d)", cpuMin, cpuMax)
				}

				return nil
			},
		),
		SchemaFunc: deploymentSchema,
	}
}

func deploymentSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of the Datawarehouse deployment",
		},
		"tags": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
			Description: "List of tags to apply",
		},
		"version": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "ClickHouse version to use",
		},
		"replica_count": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Number of replicas",
		},
		"cpu_min": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Minimum CPU count",
		},
		"cpu_max": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Maximum CPU count",
		},
		"ram_per_cpu": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "RAM per CPU (GB)",
		},
		"started": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether the deployment should be running (`true`) or stopped (`false`). Maps to the Start deployment and Stop deployment API actions.",
		},
		"password": {
			Type:          schema.TypeString,
			Sensitive:     true,
			Optional:      true,
			Description:   "Password for the first user of the deployment. Only one of `password` or `password_wo` should be specified.",
			ConflictsWith: []string{"password_wo"},
		},
		"password_wo": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "Password for the first user of the deployment in [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) mode. Only one of `password` or `password_wo` should be specified. `password_wo` will not be set in the Terraform state. To update the `password_wo`, you must also update the `password_wo_version`. When updating, the password is rotated via the Data Warehouse Users API (the initial user is selected as an admin user when present, otherwise the first user by name).",
			WriteOnly:     true,
			ConflictsWith: []string{"password"},
			RequiredWith:  []string{"password_wo_version"},
		},
		"password_wo_version": {
			Type:         schema.TypeInt,
			Optional:     true,
			Description:  "The version of the [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) password. To update the `password_wo`, you must also update the `password_wo_version`.",
			RequiredWith: []string{"password_wo"},
		},
		"public_network": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Public endpoint configuration. A public endpoint is created by default.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "ID of the public endpoint",
					},
					"dns_record": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "DNS record for the public endpoint",
					},
					"services": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "List of services exposed on the public endpoint",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"protocol": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Service protocol (e.g. \"tcp\", \"https\", \"mysql\")",
								},
								"port": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "TCP port number",
								},
							},
						},
					},
				},
			},
		},
		"private_network": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			ForceNew:    true,
			Description: "Private network to expose your datawarehouse deployment",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"pn_id": {
						Type:             schema.TypeString,
						Required:         true,
						ForceNew:         true,
						ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
						DiffSuppressFunc: dsf.Locality,
						Description:      "The private network ID",
					},
					// Computed
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The endpoint ID",
					},
					"dns_record": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "DNS record for the private endpoint",
					},
					"services": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "List of services exposed on the private endpoint",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"protocol": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Service protocol (e.g. \"tcp\", \"https\", \"mysql\")",
								},
								"port": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "TCP port number",
								},
							},
						},
					},
				},
			},
		},
		// Computed
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the deployment",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of deployment creation (RFC 3339 format)",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of deployment last update (RFC 3339 format)",
		},
	}
}

func resourceDeploymentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := datawarehouseAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	var password string
	if _, ok := d.GetOk("password_wo_version"); ok {
		password = d.GetRawConfig().GetAttr("password_wo").AsString()
	} else {
		password = d.Get("password").(string)
	}

	req := &datawarehouseapi.CreateDeploymentRequest{
		Region:       region,
		ProjectID:    d.Get("project_id").(string),
		Name:         d.Get("name").(string),
		Version:      d.Get("version").(string),
		ReplicaCount: uint32(d.Get("replica_count").(int)),
		CPUMin:       uint32(d.Get("cpu_min").(int)),
		CPUMax:       uint32(d.Get("cpu_max").(int)),
		RAMPerCPU:    uint32(d.Get("ram_per_cpu").(int)),
		Password:     password,
	}

	if v, ok := d.GetOk("tags"); ok {
		req.Tags = types.ExpandStrings(v)
	}

	// Always create a public endpoint by default
	req.Endpoints = []*datawarehouseapi.EndpointSpec{
		{
			Public: &datawarehouseapi.EndpointSpecPublicDetails{},
		},
	}

	// Add private network endpoint if configured
	if privateNetworkList, ok := d.GetOk("private_network"); ok {
		privateNetworks := privateNetworkList.([]any)
		if len(privateNetworks) > 0 {
			pn := privateNetworks[0].(map[string]any)
			privateNetworkID := locality.ExpandID(pn["pn_id"].(string))

			req.Endpoints = append(req.Endpoints, &datawarehouseapi.EndpointSpec{
				PrivateNetwork: &datawarehouseapi.EndpointSpecPrivateNetworkDetails{
					PrivateNetworkID: privateNetworkID,
				},
			})
		}
	}

	deployment, err := api.CreateDeployment(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	deployment, err = waitForDatawarehouseDeployment(ctx, api, region, deployment.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	if !d.Get("started").(bool) {
		if err := stopDeploymentIfReady(ctx, api, region, deployment.ID, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(regional.NewIDString(region, deployment.ID))

	return resourceDeploymentRead(ctx, d, meta)
}

func resourceDeploymentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDatawarehouseDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	deployment, err := api.GetDeployment(&datawarehouseapi.GetDeploymentRequest{
		Region:       region,
		DeploymentID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("region", string(deployment.Region))
	_ = d.Set("project_id", deployment.ProjectID)
	_ = d.Set("name", deployment.Name)
	_ = d.Set("tags", types.FlattenSliceString(deployment.Tags))
	_ = d.Set("version", deployment.Version)
	_ = d.Set("replica_count", int(deployment.ReplicaCount))
	_ = d.Set("cpu_min", int(deployment.CPUMin))
	_ = d.Set("cpu_max", int(deployment.CPUMax))
	_ = d.Set("ram_per_cpu", int(deployment.RAMPerCPU))
	_ = d.Set("started", deploymentStatusIsRunning(deployment.Status))
	_ = d.Set("status", string(deployment.Status))
	_ = d.Set("created_at", deployment.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", deployment.UpdatedAt.Format(time.RFC3339))

	publicBlock, hasPublic := flattenPublicNetwork(deployment.Endpoints, deployment.Region)
	if hasPublic {
		_ = d.Set("public_network", publicBlock.([]map[string]any))
	} else {
		_ = d.Set("public_network", nil)
	}

	privateBlock, hasPrivate := flattenPrivateNetwork(deployment.Endpoints, deployment.Region)
	if hasPrivate {
		_ = d.Set("private_network", privateBlock.([]map[string]any))
	}

	return nil
}

func resourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	timeout := d.Timeout(schema.TimeoutUpdate)

	_, err = waitForDatawarehouseDeployment(ctx, api, region, id, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	deployment, err := api.GetDeployment(&datawarehouseapi.GetDeploymentRequest{
		Region:       region,
		DeploymentID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	configNeedsRunning := d.HasChanges("cpu_min", "cpu_max", "replica_count")
	wantStarted := d.Get("started").(bool)

	startedForScaleUpdate := false

	if configNeedsRunning && deployment.Status == datawarehouseapi.DeploymentStatusStopped {
		if err := startDeploymentIfStopped(ctx, api, region, id, timeout); err != nil {
			return diag.FromErr(err)
		}

		startedForScaleUpdate = true
	}

	req := &datawarehouseapi.UpdateDeploymentRequest{
		Region:       region,
		DeploymentID: id,
	}
	changed := false

	if d.HasChange("name") {
		req.Name = new(d.Get("name").(string))
		changed = true
	}

	if d.HasChange("tags") {
		req.Tags = new(types.ExpandStrings(d.Get("tags")))
		changed = true
	}

	if d.HasChange("cpu_min") {
		req.CPUMin = new(uint32(d.Get("cpu_min").(int)))
		changed = true
	}

	if d.HasChange("cpu_max") {
		req.CPUMax = new(uint32(d.Get("cpu_max").(int)))
		changed = true
	}

	if d.HasChange("replica_count") {
		req.ReplicaCount = new(uint32(d.Get("replica_count").(int)))
		changed = true
	}

	if changed {
		if _, err := api.UpdateDeployment(req, scw.WithContext(ctx)); err != nil {
			return diag.FromErr(err)
		}

		if _, err := waitForDatawarehouseDeployment(ctx, api, region, id, timeout); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("started") {
		if wantStarted {
			if err := startDeploymentIfStopped(ctx, api, region, id, timeout); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := stopDeploymentIfReady(ctx, api, region, id, timeout); err != nil {
				return diag.FromErr(err)
			}
		}
	} else if startedForScaleUpdate && !wantStarted {
		if err := stopDeploymentIfReady(ctx, api, region, id, timeout); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := d.GetOk("password_wo_version"); ok {
		if d.HasChange("password_wo_version") {
			userName, err := findInitialDeploymentUserName(ctx, api, region, id)
			if err != nil {
				return diag.FromErr(err)
			}

			pwd := d.GetRawConfig().GetAttr("password_wo").AsString()
			_, err = api.UpdateUser(&datawarehouseapi.UpdateUserRequest{
				Region:       region,
				DeploymentID: id,
				Name:         userName,
				Password:     types.ExpandStringPtr(pwd),
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			_, err = waitForDatawarehouseDeployment(ctx, api, region, id, timeout)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceDeploymentRead(ctx, d, meta)
}

func resourceDeploymentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDatawarehouseDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteDeployment(&datawarehouseapi.DeleteDeploymentRequest{
		Region:       region,
		DeploymentID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDatawarehouseDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}

func deploymentStatusIsRunning(s datawarehouseapi.DeploymentStatus) bool {
	return s != datawarehouseapi.DeploymentStatusStopped && s != datawarehouseapi.DeploymentStatusStopping
}

func startDeploymentIfStopped(ctx context.Context, api *datawarehouseapi.API, region scw.Region, deploymentID string, timeout time.Duration) error {
	dep, err := api.GetDeployment(&datawarehouseapi.GetDeploymentRequest{
		Region:       region,
		DeploymentID: deploymentID,
	}, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	if dep.Status != datawarehouseapi.DeploymentStatusStopped {
		return nil
	}

	if _, err := api.StartDeployment(&datawarehouseapi.StartDeploymentRequest{
		Region:       region,
		DeploymentID: deploymentID,
	}, scw.WithContext(ctx)); err != nil {
		return err
	}

	_, err = waitForDatawarehouseDeployment(ctx, api, region, deploymentID, timeout)

	return err
}

func stopDeploymentIfReady(ctx context.Context, api *datawarehouseapi.API, region scw.Region, deploymentID string, timeout time.Duration) error {
	dep, err := api.GetDeployment(&datawarehouseapi.GetDeploymentRequest{
		Region:       region,
		DeploymentID: deploymentID,
	}, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	if dep.Status != datawarehouseapi.DeploymentStatusReady {
		return nil
	}

	if _, err := api.StopDeployment(&datawarehouseapi.StopDeploymentRequest{
		Region:       region,
		DeploymentID: deploymentID,
	}, scw.WithContext(ctx)); err != nil {
		return err
	}

	_, err = waitForDatawarehouseDeployment(ctx, api, region, deploymentID, timeout)

	return err
}

func findInitialDeploymentUserName(ctx context.Context, api *datawarehouseapi.API, region scw.Region, deploymentID string) (string, error) {
	pageSize := uint32(100)

	resp, err := api.ListUsers(&datawarehouseapi.ListUsersRequest{
		Region:       region,
		DeploymentID: deploymentID,
		OrderBy:      datawarehouseapi.ListUsersRequestOrderByNameAsc,
		PageSize:     &pageSize,
	}, scw.WithContext(ctx))
	if err != nil {
		return "", err
	}

	if len(resp.Users) == 0 {
		return "", fmt.Errorf("no users found on datawarehouse deployment; cannot apply password_wo change")
	}

	var adminName string

	for _, u := range resp.Users {
		if !u.IsAdmin {
			continue
		}

		if adminName == "" || u.Name < adminName {
			adminName = u.Name
		}
	}

	if adminName != "" {
		return adminName, nil
	}

	return resp.Users[0].Name, nil
}
