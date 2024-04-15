package documentdb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	documentdb "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourcePrivateNetworkEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDocumentDBInstanceEndpointCreate,
		ReadContext:   resourceDocumentDBInstanceEndpointRead,
		UpdateContext: resourceDocumentDBInstanceEndpointUpdate,
		DeleteContext: resourceDocumentDBInstanceEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultDocumentDBInstanceTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Instance on which the endpoint is attached",
			},
			"private_network_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     verify.IsUUIDorUUIDWithLocality(),
				DiffSuppressFunc: dsf.Locality,
				Description:      "The private network ID",
				ForceNew:         true,
			},
			// Computed
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
				Description: "The IP of your private network service",
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
			"zone":   zonal.Schema(),
			"region": regional.Schema(),
		},
	}
}

func resourceDocumentDBInstanceEndpointCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := documentDBAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID := locality.ExpandID(d.Get("instance_id"))
	endpointSpecPN := &documentdb.EndpointSpecPrivateNetwork{}
	createEndpointRequest := &documentdb.CreateEndpointRequest{
		Region:       region,
		InstanceID:   instanceID,
		EndpointSpec: &documentdb.EndpointSpec{},
	}

	endpointSpecPN.PrivateNetworkID = locality.ExpandID(d.Get("private_network_id").(string))
	ipNet := d.Get("ip_net").(string)
	if len(ipNet) > 0 {
		ip, err := types.ExpandIPNet(ipNet)
		if err != nil {
			return diag.FromErr(err)
		}
		endpointSpecPN.ServiceIP = &ip
	} else {
		endpointSpecPN.IpamConfig = &documentdb.EndpointSpecPrivateNetworkIpamConfig{}
	}

	createEndpointRequest.EndpointSpec.PrivateNetwork = endpointSpecPN
	_, err = waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		if httperrors.Is404(err) {
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
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, endpoint.ID))

	return resourceDocumentDBInstanceEndpointRead(ctx, d, m)
}

func resourceDocumentDBInstanceEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	endpoint, err := api.GetEndpoint(&documentdb.GetEndpointRequest{
		EndpointID: id,
		Region:     region,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	pnID := regional.NewIDString(region, endpoint.PrivateNetwork.PrivateNetworkID)
	serviceIP, err := types.FlattenIPNet(endpoint.PrivateNetwork.ServiceIP)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("private_network_id", pnID)
	_ = d.Set("ip_net", serviceIP)
	_ = d.Set("zone", endpoint.PrivateNetwork.Zone)
	_ = d.Set("port", int(endpoint.Port))
	_ = d.Set("name", endpoint.Name)
	_ = d.Set("hostname", endpoint.Hostname)
	_ = d.Set("ip", types.FlattenIPPtr(endpoint.IP))
	_ = d.Set("region", region.String())

	return nil
}

func resourceDocumentDBInstanceEndpointUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &documentdb.MigrateEndpointRequest{
		EndpointID: id,
		Region:     region,
	}

	if d.HasChange("instance_id") {
		req.InstanceID = locality.ExpandID(d.Get("instance_id"))

		if _, err := api.MigrateEndpoint(req, scw.WithContext(ctx)); err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForDocumentDBInstance(ctx, api, region, req.InstanceID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDocumentDBInstanceEndpointRead(ctx, d, m)
}

func resourceDocumentDBInstanceEndpointDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteEndpoint(&documentdb.DeleteEndpointRequest{
		Region:     region,
		EndpointID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
