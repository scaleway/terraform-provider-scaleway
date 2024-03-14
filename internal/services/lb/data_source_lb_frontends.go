package lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceFrontends() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceLbFrontendsRead,
		Schema: map[string]*schema.Schema{
			"lb_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "frontends with a lb id like it are listed.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "frontends with a name like it are listed.",
			},
			"frontends": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"name": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"inbound_port": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"backend_id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"lb_id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"timeout_client": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"certificate_ids": {
							Computed: true,
							Type:     schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"created_at": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"update_at": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"enable_http3": {
							Computed: true,
							Type:     schema.TypeBool,
						},
					},
				},
			},
			"zone":            zonal.Schema(),
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
		},
	}
}

func DataSourceLbFrontendsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	_, lbID, err := zonal.ParseID(d.Get("lb_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := lbAPI.ListFrontends(&lb.ZonedAPIListFrontendsRequest{
		Zone: zone,
		LBID: lbID,
		Name: types.ExpandStringPtr(d.Get("name")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	frontends := []interface{}(nil)
	for _, frontend := range res.Frontends {
		rawFrontend := make(map[string]interface{})
		rawFrontend["id"] = zonal.NewIDString(zone, frontend.ID)
		rawFrontend["name"] = frontend.Name
		rawFrontend["lb_id"] = zonal.NewIDString(zone, frontend.LB.ID)
		rawFrontend["created_at"] = types.FlattenTime(frontend.CreatedAt)
		rawFrontend["update_at"] = types.FlattenTime(frontend.UpdatedAt)
		rawFrontend["inbound_port"] = frontend.InboundPort
		rawFrontend["backend_id"] = frontend.Backend.ID
		rawFrontend["timeout_client"] = types.FlattenDuration(frontend.TimeoutClient)
		rawFrontend["enable_http3"] = frontend.EnableHTTP3
		if len(frontend.CertificateIDs) > 0 {
			rawFrontend["certificate_ids"] = frontend.CertificateIDs
		}

		frontends = append(frontends, rawFrontend)
	}

	d.SetId(zone.String())
	_ = d.Set("frontends", frontends)

	return nil
}
