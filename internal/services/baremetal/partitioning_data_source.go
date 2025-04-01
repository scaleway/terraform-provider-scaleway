package baremetal

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataPartitioning() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataPartitioningRead,
		Schema:      map[string]*schema.Schema{},
	}
}

func dataPartitioningRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
