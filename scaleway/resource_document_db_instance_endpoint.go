package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	document_db "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayDocumentDBInstanceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayDocumentDBInstanceEndpointCreate,
		ReadContext:   resourceScalewayDocumentDBInstanceEndpointRead,
		UpdateContext: resourceScalewayDocumentDBInstanceEndpointUpdate,
		DeleteContext: resourceScalewayDocumentDBInstanceEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultRdbInstanceTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Instance on which the endpoint is attached",
			},
			"private_network": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The Private network details",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateFunc:     validationUUIDorUUIDWithLocality(),
							DiffSuppressFunc: diffSuppressFuncLocality,
							Description:      "The private network ID",
							ForceNew:         true,
						},
						// Computed
						"endpoint_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The endpoint ID",
						},
						"ip_net": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsCIDR,
							Description:  "The IP with the given mask within the private subnet",
						},
						"ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IP of your Instance within the private service",
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsPortNumber,
							Description:  "The port of your private service",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of your private service",
						},
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The hostname of your endpoint",
						},
						"zone": zoneSchema(),
					},
				},
			},
			"load_balancer": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Load balancer of the database instance",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Computed
						"endpoint_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The endpoint ID",
						},
						"ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IP of your load balancer service",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The port of your load balancer service",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of your load balancer service",
						},
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The hostname of your endpoint",
						},
					},
				},
			},
			"region": regionSchema(),
		},
	}
}

func resourceScalewayDocumentDBInstanceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := documentDBAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID := expandID(d.Get("instance_id"))
	createEndpointRequest := &document_db.CreateEndpointRequest{
		Region:       region,
		InstanceID:   instanceID,
		EndpointSpec: &document_db.EndpointSpec{},
	}

	endpointsConfig := createEndpointRequest.EndpointSpec
	pn, exist := d.GetOk("private_network")
	if exist {
		endpointsConfig.PrivateNetwork, err = expandDocumentDBPrivateNetwork(pn, exist)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	endpoint, err := api.CreateEndpoint(createEndpointRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, endpoint.ID))

	return resourceScalewayDocumentDBInstanceEndpointRead(ctx, d, meta)
}

func resourceScalewayDocumentDBInstanceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := documentDBAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	endpoint, err := api.GetEndpoint(&document_db.GetEndpointRequest{
		EndpointID: id,
		Region:     region,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	privateNetwork, err := flattenDocumentDBPrivateNetwork(endpoint.PrivateNetwork)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID := expandID(d.Get("instance_id"))
	instance, err := waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("private_network", privateNetwork)
	_ = d.Set("load_balancer", flattenDocumentDBLoadBalancer(instance.Endpoints))
	return nil
}

func resourceScalewayDocumentDBInstanceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := documentDBAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &document_db.MigrateEndpointRequest{
		EndpointID: id,
		Region:     region,
	}

	if d.HasChange("instance_id") {
		req.InstanceID = expandID(d.Get("instance_id"))

		if _, err := api.MigrateEndpoint(req, scw.WithContext(ctx)); err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForDocumentDBInstance(ctx, api, region, req.InstanceID, d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayDocumentDBInstanceEndpointRead(ctx, d, meta)
}

func resourceScalewayDocumentDBInstanceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := documentDBAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteEndpoint(&document_db.DeleteEndpointRequest{
		Region:     region,
		EndpointID: id,
	}, scw.WithContext(ctx))

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
