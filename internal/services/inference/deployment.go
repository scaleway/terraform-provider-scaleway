package inference

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/inference/v1"
	ipamAPI "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/ipam"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceDeployment() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceDeploymentCreate,
		ReadContext:   ResourceDeploymentRead,
		UpdateContext: ResourceDeploymentUpdate,
		DeleteContext: ResourceDeploymentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultInferenceDeploymentTimeout),
			Read:    schema.DefaultTimeout(defaultInferenceDeploymentTimeout),
			Update:  schema.DefaultTimeout(defaultInferenceDeploymentTimeout),
			Delete:  schema.DefaultTimeout(defaultInferenceDeploymentTimeout),
			Default: schema.DefaultTimeout(defaultInferenceDeploymentTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    deploymentSchema,
		Identity:      identity.DefaultRegional(),
	}
}

func deploymentSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The deployment name",
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
		"node_type": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The node type to use for the deployment",
		},
		"model_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The model name to use for the deployment",
		},
		"model_id": {
			Type:             schema.TypeString,
			Required:         true,
			Description:      "The model id used for the deployment",
			ForceNew:         true,
			DiffSuppressFunc: dsf.Locality,
		},
		"accept_eula": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether or not the deployment is accepting eula",
		},
		"tags": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Optional:    true,
			Description: "The tags associated with the deployment",
		},
		"min_size": {
			Type:         schema.TypeInt,
			Optional:     true,
			Description:  "The minimum size of the pool",
			ValidateFunc: validation.IntAtLeast(1),
			Default:      1,
		},
		"max_size": {
			Type:         schema.TypeInt,
			Optional:     true,
			Description:  "The maximum size of the pool",
			ValidateFunc: validation.IntAtLeast(1),
			Default:      1,
		},
		"quantization": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The number of bits each model parameter should be quantized to",
		},
		"size": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The size of the pool",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the deployment",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the deployment",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the deployment",
		},
		"private_endpoint": {
			Type:         schema.TypeList,
			Optional:     true,
			MaxItems:     1,
			AtLeastOneOf: []string{"public_endpoint"},
			Description:  "List of endpoints",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Description: "The id of the private endpoint",
						Computed:    true,
					},
					"private_network_id": {
						Type:        schema.TypeString,
						Description: "The id of the private network",
						Optional:    true,
					},
					"disable_auth": {
						Type:        schema.TypeBool,
						Description: "Disable the authentication on the endpoint.",
						Optional:    true,
						Default:     false,
					},
					"url": {
						Type:        schema.TypeString,
						Description: "The URL of the endpoint.",
						Computed:    true,
					},
				},
			},
		},
		"public_endpoint": {
			Type:         schema.TypeList,
			Optional:     true,
			AtLeastOneOf: []string{"private_endpoint"},
			Description:  "Public endpoints",
			MaxItems:     1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Description: "The id of the public endpoint",
						Computed:    true,
					},
					"is_enabled": {
						Type:        schema.TypeBool,
						Description: "Enable or disable public endpoint",
						Optional:    true,
					},
					"disable_auth": {
						Type:        schema.TypeBool,
						Description: "Disable the authentication on the endpoint.",
						Optional:    true,
						Default:     false,
					},
					"url": {
						Type:        schema.TypeString,
						Description: "The URL of the endpoint.",
						Computed:    true,
					},
				},
			},
		},
		"private_ip": {
			Type:        schema.TypeList,
			Computed:    true,
			Optional:    true,
			Description: "The private IPv4 address associated with the deployment",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The ID of the IPv4 address resource",
					},
					"address": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The private IPv4 address",
					},
				},
			},
		},
	}
}

func ResourceDeploymentCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &inference.CreateDeploymentRequest{
		Region:       region,
		ProjectID:    d.Get("project_id").(string),
		Name:         d.Get("name").(string),
		NodeTypeName: d.Get("node_type").(string),
		ModelID:      locality.ExpandID(d.Get("model_id").(string)),
		Tags:         types.ExpandStrings(d.Get("tags")),
		Endpoints:    buildEndpoints(d),
	}

	if isAcceptingEula, ok := d.GetOk("accept_eula"); ok {
		req.AcceptEula = scw.BoolPtr(isAcceptingEula.(bool))
	}

	if minSize, ok := d.GetOk("min_size"); ok {
		req.MinSize = scw.Uint32Ptr(uint32(minSize.(int)))
	}

	if maxSize, ok := d.GetOk("max_size"); ok {
		req.MaxSize = scw.Uint32Ptr(uint32(maxSize.(int)))
	}

	if quantization, ok := d.GetOk("quantization"); ok {
		req.Quantization = &inference.DeploymentQuantization{
			Bits: uint32(quantization.(int)),
		}
	}

	deployment, err := api.CreateDeployment(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetRegionalIdentity(d, deployment.Region, deployment.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDeployment(ctx, api, region, deployment.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceDeploymentRead(ctx, d, m)
}

func buildEndpoints(d *schema.ResourceData) []*inference.EndpointSpec {
	var endpoints []*inference.EndpointSpec

	if publicEndpoint, ok := d.GetOk("public_endpoint"); ok {
		publicEndpointMap := publicEndpoint.([]any)[0].(map[string]any)
		if publicEndpointMap["is_enabled"].(bool) {
			publicEp := inference.EndpointSpec{
				PublicNetwork: &inference.EndpointPublicNetworkDetails{},
				DisableAuth:   publicEndpointMap["disable_auth"].(bool),
			}
			endpoints = append(endpoints, &publicEp)
		}
	}

	if privateEndpoint, ok := d.GetOk("private_endpoint"); ok {
		privateEndpointMap := privateEndpoint.([]any)[0].(map[string]any)
		if privateID, exists := privateEndpointMap["private_network_id"]; exists {
			privateEp := inference.EndpointSpec{
				PrivateNetwork: &inference.EndpointPrivateNetworkDetails{
					PrivateNetworkID: regional.ExpandID(privateID.(string)).ID,
				},
				DisableAuth: privateEndpointMap["disable_auth"].(bool),
			}
			endpoints = append(endpoints, &privateEp)
		}
	}

	return endpoints
}

func ResourceDeploymentRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deployment, err := waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", deployment.Name)
	_ = d.Set("region", deployment.Region)
	_ = d.Set("project_id", deployment.ProjectID)
	_ = d.Set("node_type", deployment.NodeTypeName)
	_ = d.Set("model_name", deployment.ModelName)
	_ = d.Set("min_size", int(deployment.MinSize))
	_ = d.Set("max_size", int(deployment.MaxSize))
	_ = d.Set("size", int(deployment.Size))
	_ = d.Set("status", deployment.Status)
	_ = d.Set("model_id", deployment.ModelID)
	_ = d.Set("tags", types.ExpandUpdatedStringsPtr(deployment.Tags))
	_ = d.Set("created_at", types.FlattenTime(deployment.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(deployment.UpdatedAt))

	var privateEndpoints []map[string]any

	var publicEndpoints []map[string]any

	for _, endpoint := range deployment.Endpoints {
		if endpoint.PrivateNetwork != nil {
			privateEndpointSpec := map[string]any{
				"id":                 endpoint.ID,
				"private_network_id": regional.NewID(deployment.Region, endpoint.PrivateNetwork.PrivateNetworkID).String(),
				"disable_auth":       endpoint.DisableAuth,
				"url":                endpoint.URL,
			}
			privateEndpoints = append(privateEndpoints, privateEndpointSpec)
		}

		if endpoint.PublicNetwork != nil {
			publicEndpointSpec := map[string]any{
				"id":           endpoint.ID,
				"is_enabled":   true,
				"disable_auth": endpoint.DisableAuth,
				"url":          endpoint.URL,
			}
			publicEndpoints = append(publicEndpoints, publicEndpointSpec)
		}
	}

	diags := diag.Diagnostics{}
	privateIPs := []map[string]any(nil)
	authorized := true

	if privateEndpoints != nil {
		_ = d.Set("private_endpoint", privateEndpoints)

		for _, endpoint := range deployment.Endpoints {
			if endpoint.PrivateNetwork == nil {
				continue
			}

			resourceType := ipamAPI.ResourceTypeLlmDeployment
			opts := &ipam.GetResourcePrivateIPsOptions{
				ResourceID:       &deployment.ID,
				ResourceType:     &resourceType,
				PrivateNetworkID: &endpoint.PrivateNetwork.PrivateNetworkID,
				ProjectID:        &deployment.ProjectID,
			}

			endpointPrivateIPs, err := ipam.GetResourcePrivateIPs(ctx, m, region, opts)

			switch {
			case err == nil:
				privateIPs = append(privateIPs, endpointPrivateIPs...)
			case httperrors.Is403(err):
				authorized = false

				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Warning,
					Summary:       "Unauthorized to read deployment's private IP, please check your IAM permissions",
					Detail:        err.Error(),
					AttributePath: cty.GetAttrPath("private_ip"),
				})
			default:
				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Warning,
					Summary:       fmt.Sprintf("Unable to get private IP for deployment %q", deployment.Name),
					Detail:        err.Error(),
					AttributePath: cty.GetAttrPath("private_ip"),
				})
			}

			if !authorized {
				break
			}
		}
	}

	if authorized {
		_ = d.Set("private_ip", privateIPs)
	}

	if publicEndpoints != nil {
		_ = d.Set("public_endpoint", publicEndpoints)
	}

	err = identity.SetRegionalIdentity(d, deployment.Region, deployment.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func ResourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deployment, err := waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	req := &inference.UpdateDeploymentRequest{
		Region:       region,
		DeploymentID: deployment.ID,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChange("min_size") {
		req.MinSize = types.ExpandUint32Ptr(d.Get("min_size"))
	}

	if d.HasChange("max_size") {
		req.MaxSize = types.ExpandUint32Ptr(d.Get("max_size"))
	}

	if _, err := api.UpdateDeployment(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return ResourceDeploymentRead(ctx, d, m)
}

func ResourceDeploymentDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteDeployment(&inference.DeleteDeploymentRequest{
		Region:       region,
		DeploymentID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
