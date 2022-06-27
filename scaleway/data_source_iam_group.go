package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayIamGroup() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayIamGroup().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "description", "user_ids", "application_ids")

	dsSchema["name"].ConflictsWith = []string{"group_id"}
	dsSchema["group_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the IAM group",
		ConflictsWith: []string{"name"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}
	// Requiring organization_id is temporary until we are able to get it from the sdk
	dsSchema["organization_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "The organization_id you want to attach the resource to",
		Required:    true,
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayIamGroupRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayIamGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)

	groupID, groupIDExists := d.GetOk("group_id")
	if !groupIDExists {
		req := &iam.ListGroupsRequest{
			OrganizationID: expandStringPtr(d.Get("organization_id")),
			Name:           expandStringPtr(d.Get("name")),
		}
		if appIDs := d.Get("application_ids"); appIDs != nil {
			req.ApplicationIDs = expandStrings(appIDs)
		}
		if userIDs := d.Get("user_ids"); userIDs != nil {
			req.ApplicationIDs = expandStrings(userIDs)
		}
		if groupIDs := d.Get("group_ids"); groupIDs != nil {
			req.ApplicationIDs = expandStrings(groupIDs)
		}

		res, err := api.ListGroups(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		for _, group := range res.Groups {
			if group.Name == d.Get("name").(string) {
				if groupID != "" {
					return diag.Errorf("more than 1 group found with the same name %s", d.Get("name"))
				}
				groupID = group.ID
			}
		}
		if groupID == "" {
			return diag.Errorf("no group found with the name %s", d.Get("name"))
		}
	}

	d.SetId(groupID.(string))
	err := d.Set("group_id", groupID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewayIamGroupRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read iam group state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("iam group (%s) not found", groupID)
	}

	return nil
}
