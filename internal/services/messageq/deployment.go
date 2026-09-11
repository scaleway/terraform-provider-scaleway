package messageq

import (
	"context"
	_ "embed"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/deployment.md
var deploymentDescription string

func ResourceDeployment() *schema.Resource {
	return &schema.Resource{
		Description:   deploymentDescription,
		CreateContext: resourceDeploymentCreate,
		ReadContext:   resourceDeploymentRead,
		UpdateContext: resourceDeploymentUpdate,
		DeleteContext: resourceDeploymentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultDeploymentTimeout),
			Read:   schema.DefaultTimeout(defaultDeploymentReadTimeout),
			Update: schema.DefaultTimeout(defaultDeploymentTimeout),
			Delete: schema.DefaultTimeout(defaultDeploymentTimeout),
		},
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
			Optional:    true,
			Computed:    true,
			Description: "Name of the MessageQ deployment",
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
			Description: "MessageQ version to use",
		},
		"node_count": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Number of nodes. Can be updated via Upgrade without recreating the deployment",
		},
		"node_type": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Type of node",
		},
		"user_name": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Bootstrap username for the deployment. Prefer scaleway_messageq_user for additional users",
		},
		"password": {
			Type:          schema.TypeString,
			Sensitive:     true,
			Optional:      true,
			ForceNew:      true,
			Description:   "Bootstrap password for the deployment user. Only one of `password` or `password_wo` should be specified. Prefer scaleway_messageq_user for password rotation",
			ConflictsWith: []string{"password_wo"},
		},
		// WriteOnly is incompatible with ForceNew, so password_wo_version carries the
		// ForceNew that makes a rotation attempt visible in the plan.
		"password_wo": {
			Type:          schema.TypeString,
			Optional:      true,
			WriteOnly:     true,
			Description:   "Bootstrap password for the deployment user in [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) mode. Only one of `password` or `password_wo` should be specified. `password_wo` will not be set in the Terraform state. Bootstrap credentials are immutable, so changing `password_wo_version` recreates the deployment: use scaleway_messageq_user to rotate a password in place",
			ConflictsWith: []string{"password"},
			RequiredWith:  []string{"password_wo_version"},
		},
		"password_wo_version": {
			Type:         schema.TypeInt,
			Optional:     true,
			ForceNew:     true,
			Description:  "The version of the [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) password. To update the `password_wo`, you must also update the `password_wo_version`",
			RequiredWith: []string{"password_wo"},
		},
		"private_network": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Private network configuration",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"private_network_id": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
						Description:      "UUID of the Private Network",
					},
				},
			},
		},
		"volume": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Volume configuration",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:             schema.TypeString,
						Required:         true,
						ForceNew:         true,
						ValidateDiagFunc: verify.ValidateEnum[messageqapi.VolumeType](),
						Description:      "Volume type (sbs_5k, sbs_15k)",
					},
					"size_in_gb": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "Volume size in GB. Can be updated via Upgrade without recreating the deployment",
					},
				},
			},
		},
		"endpoints": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of endpoints",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Endpoint ID",
					},
					"services": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "List of services",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Service name",
								},
								"port": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "Service port",
								},
								"url": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Service URL",
								},
							},
						},
					},
					"public": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Whether the endpoint is public",
					},
					"private_network_id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Private network ID if applicable",
					},
				},
			},
		},
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
	api, region, err := newAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &messageqapi.CreateDeploymentRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
		Name:      types.ExpandOrGenerateString(d.Get("name"), "messageq"),
		Version:   d.Get("version").(string),
		NodeCount: uint32(d.Get("node_count").(int)),
		NodeType:  d.Get("node_type").(string),
	}

	if v, ok := d.GetOk("tags"); ok {
		req.Tags = types.ExpandStrings(v)
	}

	if v, ok := d.GetOk("user_name"); ok {
		req.UserName = types.ExpandStringPtr(v)
	}

	password := d.Get("password").(string)
	if _, ok := d.GetOk("password_wo_version"); ok {
		password = d.GetRawConfig().GetAttr("password_wo").AsString()
	}

	if password != "" {
		req.Password = &password
	}

	if v, ok := d.GetOk("volume"); ok {
		volumeList := v.([]any)
		if len(volumeList) > 0 {
			volumeMap := volumeList[0].(map[string]any)
			req.Volume = &messageqapi.Volume{
				Type:      messageqapi.VolumeType(volumeMap["type"].(string)),
				SizeBytes: expandVolumeSizeBytes(volumeMap["size_in_gb"].(int)),
			}
		}
	}

	pnID := ""

	if v, ok := d.GetOk("private_network"); ok {
		pnList := v.([]any)
		if len(pnList) > 0 {
			pnMap := pnList[0].(map[string]any)
			pnID = locality.ExpandID(pnMap["private_network_id"].(string))
		}
	}

	req.Endpoints = expandEndpointSpecsFromPrivateNetwork(pnID)

	deployment, err := api.CreateDeployment(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	deployment, err = waitForDeployment(ctx, api, region, deployment.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetRegionalIdentity(d, region, deployment.ID)
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

	deployment, err := waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	diags := setDeploymentState(d, deployment)

	err = identity.SetRegionalIdentity(d, deployment.Region, deployment.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func setDeploymentState(d *schema.ResourceData, deployment *messageqapi.Deployment) diag.Diagnostics {
	_ = d.Set("region", string(deployment.Region))
	_ = d.Set("project_id", deployment.ProjectID)
	_ = d.Set("name", deployment.Name)
	_ = d.Set("tags", types.FlattenSliceString(deployment.Tags))
	_ = d.Set("version", deployment.Version)
	_ = d.Set("node_count", int(deployment.NodeCount))
	_ = d.Set("node_type", deployment.NodeType)
	_ = d.Set("status", string(deployment.Status))

	if deployment.CreatedAt != nil {
		_ = d.Set("created_at", deployment.CreatedAt.Format(time.RFC3339))
	}

	if deployment.UpdatedAt != nil {
		_ = d.Set("updated_at", deployment.UpdatedAt.Format(time.RFC3339))
	}

	if deployment.Volume != nil {
		_ = d.Set("volume", []map[string]any{
			{
				"type":       string(deployment.Volume.Type),
				"size_in_gb": bytesToGB(deployment.Volume.SizeBytes),
			},
		})
	}

	// The API may briefly return both public and private endpoints while a switch is
	// in progress, and it keeps the public endpoint alive once a private one is added
	// (see updateDeploymentPrivateNetwork). In state, expose only the endpoint that
	// matches the Terraform config, falling back to every endpoint when nothing matches
	// so the attribute is never silently emptied.
	allEndpoints := deployment.Endpoints
	filteredEndpoints := allEndpoints

	if pnRaw, ok := d.GetOk("private_network"); ok {
		pnList := pnRaw.([]any)
		if len(pnList) > 0 {
			pnMap := pnList[0].(map[string]any)
			desiredPNID := locality.ExpandID(pnMap["private_network_id"].(string))

			filteredEndpoints = nil

			for _, ep := range allEndpoints {
				if ep == nil || ep.PrivateNetwork == nil {
					continue
				}

				if ep.PrivateNetwork.PrivateNetworkID == desiredPNID {
					filteredEndpoints = append(filteredEndpoints, ep)
				}
			}

			if len(filteredEndpoints) == 0 {
				filteredEndpoints = allEndpoints
			}
		}
	} else {
		filteredEndpoints = nil

		for _, ep := range allEndpoints {
			if ep == nil || ep.Public == nil || ep.PrivateNetwork != nil {
				continue
			}

			filteredEndpoints = append(filteredEndpoints, ep)
		}

		if len(filteredEndpoints) == 0 {
			filteredEndpoints = allEndpoints
		}
	}

	_ = d.Set("endpoints", flattenEndpoints(filteredEndpoints))

	return nil
}

func resourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("name", "tags") {
		req := &messageqapi.UpdateDeploymentRequest{
			Region:       region,
			DeploymentID: id,
		}

		if d.HasChange("name") {
			req.Name = types.ExpandStringPtr(d.Get("name"))
		}

		if d.HasChange("tags") {
			req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		}

		_, err := api.UpdateDeployment(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// UpgradeDeployment accepts precisely one of NodeCount or VolumeSizeBytes.
	if d.HasChange("node_count") {
		nodeCount := uint32(d.Get("node_count").(int))

		_, err := api.UpgradeDeployment(&messageqapi.UpgradeDeploymentRequest{
			Region:       region,
			DeploymentID: id,
			NodeCount:    &nodeCount,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("volume.0.size_in_gb") {
		sizeBytes := expandVolumeSizeBytes(d.Get("volume.0.size_in_gb").(int))

		_, err := api.UpgradeDeployment(&messageqapi.UpgradeDeploymentRequest{
			Region:          region,
			DeploymentID:    id,
			VolumeSizeBytes: &sizeBytes,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("private_network") {
		if diags := updateDeploymentPrivateNetwork(ctx, d, api, region, id); diags.HasError() {
			return diags
		}
	}

	return resourceDeploymentRead(ctx, d, meta)
}

func updateDeploymentPrivateNetwork(
	ctx context.Context,
	d *schema.ResourceData,
	api *messageqapi.API,
	region scw.Region,
	id string,
) diag.Diagnostics {
	deployment, err := waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	desiredPrivate := false

	var pnID string

	if v, ok := d.GetOk("private_network"); ok {
		pnList := v.([]any)
		if len(pnList) > 0 {
			desiredPrivate = true
			pnMap := pnList[0].(map[string]any)
			pnID = locality.ExpandID(pnMap["private_network_id"].(string))
		}
	}

	// Public endpoints are never deleted here: only private endpoints are reconciled
	// against the desired private_network_id. Adding a private network therefore leaves
	// the existing public endpoint in place, which setDeploymentState hides from the
	// `endpoints` attribute.
	var deletedEndpointIDs []string

	hasDesiredEndpoint := false

	for _, endpoint := range deployment.Endpoints {
		if endpoint == nil {
			continue
		}

		if endpoint.Public != nil && endpoint.PrivateNetwork == nil {
			if !desiredPrivate {
				hasDesiredEndpoint = true
			}

			continue
		}

		if endpoint.PrivateNetwork == nil {
			continue
		}

		if desiredPrivate && endpoint.PrivateNetwork.PrivateNetworkID == pnID {
			hasDesiredEndpoint = true

			continue
		}

		err := api.DeleteEndpoint(&messageqapi.DeleteEndpointRequest{
			Region:     region,
			EndpointID: endpoint.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		deletedEndpointIDs = append(deletedEndpointIDs, endpoint.ID)
	}

	// DeleteEndpoint returns immediately but the removal is asynchronous, so the deleted
	// endpoints must be gone before creating the new one.
	if len(deletedEndpointIDs) > 0 {
		err := waitForEndpointsDeleted(ctx, api, region, id, deletedEndpointIDs, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Create the desired endpoint only if it doesn't already exist.
	if !hasDesiredEndpoint {
		var spec *messageqapi.EndpointSpec
		if desiredPrivate {
			spec = &messageqapi.EndpointSpec{
				PrivateNetwork: &messageqapi.EndpointSpecPrivateNetworkDetails{
					PrivateNetworkID: pnID,
				},
			}
		} else {
			spec = &messageqapi.EndpointSpec{
				Public: &messageqapi.EndpointSpecPublicDetails{},
			}
		}

		_, err := api.CreateEndpoint(&messageqapi.CreateEndpointRequest{
			Region:       region,
			DeploymentID: id,
			EndpointSpec: spec,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceDeploymentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}

		return diag.FromErr(err)
	}

	_, err = api.DeleteDeployment(&messageqapi.DeleteDeploymentRequest{
		Region:       region,
		DeploymentID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}

		return diag.FromErr(err)
	}

	// Wait until the deployment is fully gone after Delete.
	_, err = waitForDeployment(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
