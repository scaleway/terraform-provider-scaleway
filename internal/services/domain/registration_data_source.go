package domain

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
)

func DataSourceRegistration() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(registrationSchema())
	datasource.FixDatasourceSchemaFlags(dsSchema, true, "domain_name")
	datasource.AddOptionalFieldsToSchema(dsSchema, "project_id")

	dsSchema["domain_name"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The domain name to look up (e.g. example.com).",
	}

	delete(dsSchema, "domain_names")
	dsSchema["domain_names"] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: "List of domain names in the registration.",
	}

	return &schema.Resource{
		ReadContext: dataSourceRegistrationRead,
		Schema:      dsSchema,
	}
}

func dataSourceRegistrationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	registrarAPI := NewRegistrarDomainAPI(m)
	domainName := d.Get("domain_name").(string)

	var projectID *string
	if v, ok := d.GetOk("project_id"); ok && v.(string) != "" {
		s := v.(string)
		projectID = &s
	}

	task, err := FindTaskByDomain(ctx, registrarAPI, domainName, projectID)
	if err != nil {
		return diag.FromErr(err)
	}

	id := task.ProjectID + "/" + task.ID
	d.SetId(id)

	_ = d.Set("task_id", task.ID)
	_ = d.Set("project_id", task.ProjectID)
	_ = d.Set("domain_names", SplitDomains(task.Domain))

	return resourceRegistrationsRead(ctx, d, m)
}
