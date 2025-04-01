package baremetal

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataEasyPartitioning() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataEasyPartitioningRead,
		Schema: map[string]*schema.Schema{
			"swap": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "set swap partition",
			},
			"ext_4": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "set ext_4 partition",
			},
			"ext_4_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "/data",
				Description: "ext_4 partition's name",
			},
		},
	}
}

func dataEasyPartitioningRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
