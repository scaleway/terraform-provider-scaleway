package kafka

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kafkaapi "github.com/scaleway/scaleway-sdk-go/api/kafka/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		SchemaFunc: clusterSchema,
		Identity:   identity.DefaultRegional(),
	}
}

func clusterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of the Kafka cluster",
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
			Description: "Kafka version to use",
		},
		"node_amount": {
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
			Description: "Number of nodes in the cluster",
		},
		"node_type": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Node type to use for the cluster",
		},
		"volume_type": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: verify.ValidateEnum[kafkaapi.VolumeType](),
			Description:      "Type of volume where data is stored",
		},
		"volume_size_in_gb": {
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
			Description: "Volume size in GB",
		},
		"user_name": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Username for the Kafka user",
		},
		"password": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Optional:    true,
			ForceNew:    true,
			Description: "Password for the Kafka user",
		},
		"public_network": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Public endpoint configuration",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "ID of the public endpoint",
					},
					"dns_records": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "DNS records for the public endpoint",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"port": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "TCP port number",
					},
				},
			},
		},
		"private_network": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			ForceNew:    true,
			Description: "Private network to expose your Kafka cluster",
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
					"dns_records": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "DNS records for the private endpoint",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"port": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "TCP port number",
					},
				},
			},
		},
		// Computed
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the cluster",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of cluster creation (RFC 3339 format)",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of cluster last update (RFC 3339 format)",
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &kafkaapi.CreateClusterRequest{
		Region:     region,
		ProjectID:  d.Get("project_id").(string),
		Name:       d.Get("name").(string),
		Version:    d.Get("version").(string),
		NodeAmount: uint32(d.Get("node_amount").(int)),
		NodeType:   d.Get("node_type").(string),
		Volume: &kafkaapi.CreateClusterRequestVolumeSpec{
			Type:      kafkaapi.VolumeType(d.Get("volume_type").(string)),
			SizeBytes: scw.Size(uint64(d.Get("volume_size_in_gb").(int))) * scw.GB,
		},
	}

	if v, ok := d.GetOk("tags"); ok {
		req.Tags = types.ExpandStrings(v)
	}

	if v, ok := d.GetOk("user_name"); ok {
		userName := v.(string)
		req.UserName = &userName
	}

	if v, ok := d.GetOk("password"); ok {
		password := v.(string)
		req.Password = &password
	}

	// Configure endpoints
	// Note: Public endpoints are not yet supported by the Kafka API (returns 501)
	// For now, only create private network endpoints if configured
	if privateNetworkList, ok := d.GetOk("private_network"); ok {
		privateNetworks := privateNetworkList.([]any)
		if len(privateNetworks) > 0 {
			pn := privateNetworks[0].(map[string]any)
			privateNetworkID := locality.ExpandID(pn["pn_id"].(string))

			req.Endpoints = []*kafkaapi.EndpointSpec{
				{
					PrivateNetwork: &kafkaapi.EndpointSpecPrivateNetworkDetails{
						PrivateNetworkID: privateNetworkID,
					},
				},
			}
		}
	} else {
		// If no private network is configured, try to create a public endpoint
		// This will fail with 501 until public endpoints are supported
		req.Endpoints = []*kafkaapi.EndpointSpec{
			{
				PublicNetwork: &kafkaapi.EndpointSpecPublicDetails{},
			},
		}
	}

	cluster, err := api.CreateCluster(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	cluster, err = waitForKafkaCluster(ctx, api, region, cluster.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	identity.SetRegionalIdentity(d, region, cluster.ID)

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForKafkaCluster(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	cluster, err := api.GetCluster(&kafkaapi.GetClusterRequest{
		Region:    region,
		ClusterID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	diags := setClusterState(d, cluster)
	identity.SetRegionalIdentity(d, cluster.Region, cluster.ID)

	return diags
}

func setClusterState(d *schema.ResourceData, cluster *kafkaapi.Cluster) diag.Diagnostics {
	_ = d.Set("region", string(cluster.Region))
	_ = d.Set("project_id", cluster.ProjectID)
	_ = d.Set("name", cluster.Name)
	_ = d.Set("tags", types.FlattenSliceString(cluster.Tags))
	_ = d.Set("version", cluster.Version)
	_ = d.Set("node_amount", int(cluster.NodeAmount))
	_ = d.Set("node_type", cluster.NodeType)
	_ = d.Set("status", string(cluster.Status))
	_ = d.Set("created_at", cluster.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", cluster.UpdatedAt.Format(time.RFC3339))

	if cluster.Volume != nil {
		_ = d.Set("volume_type", string(cluster.Volume.Type))
		_ = d.Set("volume_size_in_gb", int(cluster.Volume.SizeBytes/scw.GB))
	}

	publicBlock, hasPublic := flattenPublicNetwork(cluster.Endpoints, cluster.Region)
	if hasPublic {
		_ = d.Set("public_network", publicBlock.([]map[string]any))
	} else {
		_ = d.Set("public_network", nil)
	}

	privateBlock, hasPrivate := flattenPrivateNetwork(cluster.Endpoints, cluster.Region)
	if hasPrivate {
		_ = d.Set("private_network", privateBlock.([]map[string]any))
	}

	return nil
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	_, err = waitForKafkaCluster(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &kafkaapi.UpdateClusterRequest{
		Region:    region,
		ClusterID: id,
	}

	hasChanged := false

	if d.HasChange("name") {
		name := d.Get("name").(string)
		req.Name = &name
		hasChanged = true
	}

	if d.HasChange("tags") {
		tags := types.ExpandStrings(d.Get("tags"))
		req.Tags = &tags
		hasChanged = true
	}

	if hasChanged {
		_, err = api.UpdateCluster(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForKafkaCluster(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	readDiags := resourceClusterRead(ctx, d, meta)

	return append(diags, readDiags...)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForKafkaCluster(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteCluster(&kafkaapi.DeleteClusterRequest{
		Region:    region,
		ClusterID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForKafkaCluster(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
