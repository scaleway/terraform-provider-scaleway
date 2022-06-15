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
			// TODO: use all of these values at least once
			Create:  schema.DefaultTimeout(defaultInstanceImageTimeout),
			Read:    schema.DefaultTimeout(defaultInstanceImageTimeout),
			Update:  schema.DefaultTimeout(defaultInstanceImageTimeout),
			Delete:  schema.DefaultTimeout(defaultInstanceImageTimeout),
			Default: schema.DefaultTimeout(defaultInstanceImageTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the image",
			},
			"root_volume": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "UUID of the snapshot", // TODO : really ?? the description in the dev documentation seems weird
			},
			"arch": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Architecture of the image",
			},
			"default_bootscript": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default bootscript of the image",
			},
			"extra_volumes": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Additional volumes attached to the image",
				Elem: &schema.Schema{
					Type: schema.TypeMap, // TODO: not sure about that, maybe i must list all nested attributes here, or maybe i can call the instance volume schema
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
			// TODO: add "organization" although it's deprecated ? I would say no because it's OnlyOneOf with "project"
			"project": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Project ID of the image",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of tags [\"tag1\", \"tag2\", ...] attached to the image",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"public": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true, // TODO : maybe ? to set it to the default value
				Description: "If true, the image will be public",
			},
			"from_server": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "??", // TODO: find a proper description of this attribute
			},
			// Computed
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the image",
			},
			"modification_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last modification of the Redis cluster",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The state of the image [ available | creating | error ]",
			},
			// Common
			"zone":            zoneSchema(),
			"project_id":      projectIDSchema(),      // TODO: do we need that ?
			"organization_id": organizationIDSchema(), // TODO: do we need that ?
		},
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
