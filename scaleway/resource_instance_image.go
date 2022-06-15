package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceScalewayInstanceImage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayInstanceImageCreate,
		ReadContext:   resourceScalewayInstanceImageRead,
		UpdateContext: resourceScalewayInstanceImageUpdate,
		DeleteContext: resourceScalewayInstanceImageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			//TODO: use all of these values at least once
			Create:  schema.DefaultTimeout(defaultInstanceImageTimeout),
			Read:    schema.DefaultTimeout(defaultInstanceImageTimeout),
			Update:  schema.DefaultTimeout(defaultInstanceImageTimeout),
			Delete:  schema.DefaultTimeout(defaultInstanceImageTimeout),
			Default: schema.DefaultTimeout(defaultInstanceImageTimeout),
		},
		SchemaVersion: 0,
		Schema:        map[string]*schema.Schema{},
	}
}

func resourceScalewayInstanceImageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceScalewayInstanceImageRead(ctx, d, meta)
}

func resourceScalewayInstanceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceScalewayInstanceImageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceScalewayInstanceImageRead(ctx, d, meta)
}

func resourceScalewayInstanceImageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
