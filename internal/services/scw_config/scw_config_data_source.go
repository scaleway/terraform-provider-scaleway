package scw_config

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func DataSourceConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConfigRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_id_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_key_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_key_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConfigRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId("0")

	projectID, isDefault, err := meta.ExtractProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("project_id", projectID)
	if err != nil {
		return diag.FromErr(err)
	}

	if isDefault {
		err = d.Set("project_id_source", "default project ID")
	} else {
		err = d.Set("project_id_source", "resource configuration")
	}

	var diags diag.Diagnostics
	client := meta.ExtractScwClient(m)
	if accessKey, ok := client.GetAccessKey(); ok {
		err = d.Set("access_key", accessKey)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Access key not found",
		})
	}

	if secretKey, ok := client.GetSecretKey(); ok {
		err = d.Set("secret_key", secretKey)
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Secret key not found",
		})
	}

	return diags
}
