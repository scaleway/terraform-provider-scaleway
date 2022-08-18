package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayRdbReadReplica() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayRdbReadReplicaCreate,
		ReadContext:   resourceScalewayRdbReadReplicaRead,
		UpdateContext: resourceScalewayRdbReadReplicaUpdate,
		DeleteContext: resourceScalewayRdbReadReplicaDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultRdbInstanceTimeout),
			Read:    schema.DefaultTimeout(defaultRdbInstanceTimeout),
			Update:  schema.DefaultTimeout(defaultRdbInstanceTimeout),
			Delete:  schema.DefaultTimeout(defaultRdbInstanceTimeout),
			Default: schema.DefaultTimeout(defaultRdbInstanceTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Id of the rdb instance to replicate",
			},
			"endpoints": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"direct_access": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Direct access endpoint, it gives you an IP and a port to access your read-replica",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									// Endpoints common
									"endpoint_id": {
										Type:        schema.TypeString,
										Description: "UUID of the endpoint (UUID format).",
										Computed:    true,
									},
									"ip": {
										Type:        schema.TypeString,
										Description: "IPv4 address of the endpoint (IP address). Only one of ip and hostname may be set.",
										Computed:    true,
									},
									"port": {
										Type:        schema.TypeInt,
										Description: "TCP port of the endpoint.",
										Computed:    true,
									},
									"name": {
										Type:        schema.TypeString,
										Description: "Name of the endpoint.",
										Computed:    true,
									},
									"hostname": {
										Type:        schema.TypeString,
										Description: "Hostname of the endpoint. Only one of ip and hostname may be set.",
										Computed:    true,
									},
								},
							},
						},
						"private_network": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Private network endpoints",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									// Private network specific
									"private_network_id": {
										Type:         schema.TypeString,
										Description:  "UUID of the private network to be connected to the read replica (UUID format).",
										ValidateFunc: validationUUIDorUUIDWithLocality(),
										Required:     true,
									},
									"service_ip": {
										Type:         schema.TypeString,
										Description:  "Endpoint IPv4 address with a CIDR notation. Check documentation about IP and subnet limitations. (IP network).",
										Required:     true,
										ValidateFunc: validation.IsCIDR,
									},
									"zone": {
										Type:        schema.TypeString,
										Description: "Private network zone.",
										Computed:    true,
									},
									// Endpoints common
									"endpoint_id": {
										Type:        schema.TypeString,
										Description: "UUID of the endpoint (UUID format).",
										Computed:    true,
									},
									"ip": {
										Type:        schema.TypeString,
										Description: "IPv4 address of the endpoint (IP address). Only one of ip and hostname may be set.",
										Computed:    true,
									},
									"port": {
										Type:        schema.TypeInt,
										Description: "TCP port of the endpoint.",
										Computed:    true,
									},
									"name": {
										Type:        schema.TypeString,
										Description: "Name of the endpoint.",
										Computed:    true,
									},
									"hostname": {
										Type:        schema.TypeString,
										Description: "Hostname of the endpoint. Only one of ip and hostname may be set.",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			// Common
			"region": regionSchema(),
		},
	}
}

func resourceScalewayRdbReadReplicaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, err := rdbAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	endpointSpecs, err := expandReadReplicaEndpointsSpec(d.Get("endpoints.0"))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to expand endpoints: %w", err))
	}

	rr, err := rdbAPI.CreateReadReplica(&rdb.CreateReadReplicaRequest{
		Region:       region,
		InstanceID:   expandID(d.Get("instance_id")),
		EndpointSpec: endpointSpecs,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create read-replica: %w", err))
	}

	d.SetId(newRegionalIDString(region, rr.ID))

	_, err = waitForRDBReadReplica(ctx, rdbAPI, region, rr.ID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayRdbReadReplicaRead(ctx, d, meta)
}

func resourceScalewayRdbReadReplicaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, ID, err := rdbAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	rr, err := waitForRDBReadReplica(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("endpoints", flattenReadReplicaEndpoints(rr.Endpoints))

	_ = d.Set("region", string(region))

	return nil
}

//gocyclo:ignore
func resourceScalewayRdbReadReplicaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, ID, err := rdbAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// verify resource is ready
	_, err = waitForRDBReadReplica(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	newEndpoints := []*rdb.ReadReplicaEndpointSpec(nil)

	if d.HasChange("endpoints.0.direct_access") {
		_, directAccessExists := d.GetOk("endpoints.0.direct_access")
		tflog.Debug(ctx, "direct_access", map[string]interface{}{
			"exists": directAccessExists,
		})
		if !directAccessExists {
			err := rdbAPI.DeleteEndpoint(&rdb.DeleteEndpointRequest{
				Region:     region,
				EndpointID: expandID(d.Get("endpoints.0.direct_access.0.endpoint_id")),
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			newEndpoints = append(newEndpoints, expandReadReplicaEndpointsSpecDirectAccess(d.Get("endpoints.0.direct_access.0")))
		}
	}

	if d.HasChange("endpoints.0.private_network") {
		_, privateNetworkExists := d.GetOk("endpoints.0.private_network")
		if !privateNetworkExists {
			err := rdbAPI.DeleteEndpoint(&rdb.DeleteEndpointRequest{
				Region:     region,
				EndpointID: expandID(d.Get("endpoints.0.private_network.0.endpoint_id")),
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			pnEndpoint, err := expandReadReplicaEndpointsSpecPrivateNetwork(d.Get("endpoints.0.private_network.0"))
			if err != nil {
				return diag.FromErr(err)
			}
			newEndpoints = append(newEndpoints, pnEndpoint)
		}
	}

	if len(newEndpoints) > 0 {
		_, err := waitForRDBReadReplica(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutRead))
		if err != nil {
			return diag.FromErr(err)
		}
		_, err = rdbAPI.CreateReadReplicaEndpoint(&rdb.CreateReadReplicaEndpointRequest{
			Region:        region,
			ReadReplicaID: ID,
			EndpointSpec:  newEndpoints,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_, err = waitForRDBReadReplica(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayRdbReadReplicaRead(ctx, d, meta)
}

func resourceScalewayRdbReadReplicaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, ID, err := rdbAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// We first wait in case the instance is in a transient state
	_, err = waitForRDBReadReplica(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = rdbAPI.DeleteReadReplica(&rdb.DeleteReadReplicaRequest{
		Region:        region,
		ReadReplicaID: ID,
	}, scw.WithContext(ctx))

	if err != nil {
		return diag.FromErr(err)
	}

	// Lastly wait in case the instance is in a transient state
	_, err = waitForRDBReadReplica(ctx, rdbAPI, region, ID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
