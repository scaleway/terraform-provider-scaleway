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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
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
		Identity:   identity.DefaultRegional(),
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
		"password": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Optional:    true,
			Description: "Password for the first user of the deployment",
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

	req := &datawarehouseapi.CreateDeploymentRequest{
		Region:       region,
		ProjectID:    d.Get("project_id").(string),
		Name:         d.Get("name").(string),
		Version:      d.Get("version").(string),
		ReplicaCount: uint32(d.Get("replica_count").(int)),
		CPUMin:       uint32(d.Get("cpu_min").(int)),
		CPUMax:       uint32(d.Get("cpu_max").(int)),
		RAMPerCPU:    uint32(d.Get("ram_per_cpu").(int)),
		Password:     d.Get("password").(string),
	}

	if v, ok := d.GetOk("tags"); ok {
		req.Tags = types.ExpandStrings(v)
	}

	req.Endpoints = []*datawarehouseapi.EndpointSpec{
		{
			Public: &datawarehouseapi.EndpointSpecPublicDetails{},
		},
	}

	deployment, err := api.CreateDeployment(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	deployment, err = waitForDatawarehouseDeployment(ctx, api, region, deployment.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetRegionalIdentity(d, deployment.Region, deployment.ID)
	if err != nil {
		return diag.FromErr(err)
	}

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
	_ = d.Set("status", string(deployment.Status))
	_ = d.Set("created_at", deployment.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", deployment.UpdatedAt.Format(time.RFC3339))

	publicBlock, hasPublic := flattenPublicNetwork(deployment.Endpoints)
	if hasPublic {
		_ = d.Set("public_network", publicBlock.([]map[string]any))
	} else {
		_ = d.Set("public_network", nil)
	}

	return nil
}

func resourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	_, err = waitForDatawarehouseDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &datawarehouseapi.UpdateDeploymentRequest{
		Region:       region,
		DeploymentID: id,
	}
	changed := false

	if d.HasChange("name") {
		req.Name = scw.StringPtr(d.Get("name").(string))
		changed = true
	}

	if d.HasChange("tags") {
		req.Tags = scw.StringsPtr(types.ExpandStrings(d.Get("tags")))
		changed = true
	}

	if d.HasChange("cpu_min") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "cpu_min cannot be updated in private beta",
			Detail:   "Modifying cpu_min has no effect until this feature is launched in general availability.",
		})
		req.CPUMin = scw.Uint32Ptr(uint32(d.Get("cpu_min").(int)))
		changed = true
	}

	if d.HasChange("cpu_max") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "cpu_max cannot be updated in private beta",
			Detail:   "Modifying cpu_max has no effect until this feature is launched in general availability.",
		})
		req.CPUMax = scw.Uint32Ptr(uint32(d.Get("cpu_max").(int)))
		changed = true
	}

	if d.HasChange("replica_count") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "replica_count cannot be updated in private beta",
			Detail:   "Modifying replica_count has no effect until this feature is launched in general availability.",
		})
		req.ReplicaCount = scw.Uint32Ptr(uint32(d.Get("replica_count").(int)))
		changed = true
	}

	if changed {
		resp, err := api.UpdateDeployment(req, scw.WithContext(ctx))
		_ = resp

		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForDatawarehouseDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	readDiags := resourceDeploymentRead(ctx, d, meta)

	return append(diags, readDiags...)
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
