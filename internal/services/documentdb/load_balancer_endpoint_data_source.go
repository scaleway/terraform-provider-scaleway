package documentdb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	documentdb "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceEndpointLoadBalancer() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceDocumentDBLoadBalancerRead,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Instance on which the endpoint is attached",
				ConflictsWith:    []string{"instance_name"},
				ValidateFunc:     verify.IsUUIDorUUIDWithLocality(),
				DiffSuppressFunc: dsf.Locality,
			},
			"instance_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Instance Name on which the endpoint is attached",
				ConflictsWith: []string{"instance_id"},
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
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),
		},
	}
}

func DataSourceDocumentDBLoadBalancerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := documentDBAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	rawInstanceID, instanceIDExists := d.GetOk("instance_id")
	if !instanceIDExists {
		rawInstanceName := d.Get("instance_name").(string)
		res, err := api.ListInstances(&documentdb.ListInstancesRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(rawInstanceName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		foundRawInstance, err := datasource.FindExact(
			res.Instances,
			func(s *documentdb.Instance) bool { return s.Name == rawInstanceName },
			rawInstanceName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		rawInstanceID = foundRawInstance.ID
	}

	instanceID := locality.ExpandID(rawInstanceID)
	instance, err := waitForDocumentDBInstance(ctx, api, region, instanceID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	lb := getEndPointDocumentDBLoadBalancer(instance.Endpoints)
	_ = d.Set("instance_id", instanceID)
	_ = d.Set("instance_name", instance.Name)
	_ = d.Set("hostname", types.FlattenStringPtr(lb.Hostname))
	_ = d.Set("port", int(lb.Port))
	_ = d.Set("ip", types.FlattenIPPtr(lb.IP))
	_ = d.Set("name", lb.Name)

	d.SetId(datasource.NewRegionalID(lb.ID, region))

	return nil
}

func getEndPointDocumentDBLoadBalancer(endpoints []*documentdb.Endpoint) *documentdb.Endpoint {
	for _, endpoint := range endpoints {
		if endpoint.LoadBalancer != nil {
			return endpoint
		}
	}

	return nil
}
