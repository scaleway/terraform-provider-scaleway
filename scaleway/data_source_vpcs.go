package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayVPCs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayVPCsRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "VPCs with a name like it are listed.",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "VPCs with these exact tags are listed.",
			},
			"vpcs": {
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
						"tags": {
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
						"is_default": {
							Computed: true,
							Type:     schema.TypeBool,
						},
						"region":          regionSchema(),
						"organization_id": organizationIDSchema(),
						"project_id":      projectIDSchema(),
					},
				},
			},
			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
	}
}

func dataSourceScalewayVPCsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := vpcAPI.ListVPCs(&vpc.ListVPCsRequest{
		Region:    region,
		Tags:      expandStrings(d.Get("tags")),
		Name:      expandStringPtr(d.Get("name")),
		ProjectID: expandStringPtr(d.Get("project_id")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	vpcs := []interface{}(nil)
	for _, virtualPrivateCloud := range res.Vpcs {
		rawVpc := make(map[string]interface{})
		rawVpc["id"] = newRegionalIDString(region, virtualPrivateCloud.ID)
		rawVpc["name"] = virtualPrivateCloud.Name
		rawVpc["created_at"] = flattenTime(virtualPrivateCloud.CreatedAt)
		rawVpc["update_at"] = flattenTime(virtualPrivateCloud.UpdatedAt)
		rawVpc["is_default"] = virtualPrivateCloud.IsDefault
		if len(virtualPrivateCloud.Tags) > 0 {
			rawVpc["tags"] = virtualPrivateCloud.Tags
		}
		rawVpc["region"] = region.String()
		rawVpc["organization_id"] = virtualPrivateCloud.OrganizationID
		rawVpc["project_id"] = virtualPrivateCloud.ProjectID

		vpcs = append(vpcs, rawVpc)
	}

	d.SetId(region.String())
	_ = d.Set("vpcs", vpcs)

	return nil
}
