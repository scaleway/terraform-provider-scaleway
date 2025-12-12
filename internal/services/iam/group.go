package iam

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIamGroupCreate,
		ReadContext:   resourceIamGroupRead,
		UpdateContext: resourceIamGroupUpdate,
		DeleteContext: resourceIamGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    groupSchema,
	}
}

func groupSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The name of the iam group",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The description of the iam group",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the group",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the group",
		},
		"user_ids": {
			Type:        schema.TypeSet,
			Description: "List of IDs of the users attached to the group",
			Optional:    true,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateDiagFunc: verify.IsUUID(),
			},
		},
		"application_ids": {
			Type:        schema.TypeSet,
			Description: "List of IDs of the applications attached to the group",
			Optional:    true,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateDiagFunc: verify.IsUUID(),
			},
		},
		"external_membership": {
			Type:        schema.TypeBool,
			Description: "Handle user and application memberships externally",
			Optional:    true,
			Default:     false,
		},
		"tags": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "The tags associated with the application",
		},
		"organization_id": account.OrganizationIDOptionalSchema(),
	}
}

func resourceIamGroupCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)
	req := &iam.CreateGroupRequest{
		OrganizationID: d.Get("organization_id").(string),
		Name:           types.ExpandOrGenerateString(d.Get("name"), "group"),
		Description:    d.Get("description").(string),
		Tags:           types.ExpandStrings(d.Get("tags")),
	}

	group, err := api.CreateGroup(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(group.ID)

	appIDs := types.ExpandStrings(d.Get("application_ids").(*schema.Set).List())
	userIDs := types.ExpandStrings(d.Get("user_ids").(*schema.Set).List())

	if !d.Get("external_membership").(bool) && (len(appIDs) > 0 || len(userIDs) > 0) {
		_, err := api.SetGroupMembers(&iam.SetGroupMembersRequest{
			GroupID:        group.ID,
			ApplicationIDs: appIDs,
			UserIDs:        userIDs,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceIamGroupRead(ctx, d, m)
}

func resourceIamGroupRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)

	group, err := api.GetGroup(&iam.GetGroupRequest{
		GroupID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", group.Name)
	_ = d.Set("description", group.Description)
	_ = d.Set("created_at", types.FlattenTime(group.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(group.UpdatedAt))
	_ = d.Set("organization_id", group.OrganizationID)
	_ = d.Set("tags", types.FlattenSliceString(group.Tags))

	if !d.Get("external_membership").(bool) {
		_ = d.Set("user_ids", group.UserIDs)
		_ = d.Set("application_ids", group.ApplicationIDs)
	}

	return nil
}

func resourceIamGroupUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)

	group, err := api.GetGroup(&iam.GetGroupRequest{
		GroupID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("name", "description", "tags") {
		_, err = api.UpdateGroup(&iam.UpdateGroupRequest{
			GroupID:     group.ID,
			Name:        types.ExpandUpdatedStringPtr(d.Get("name")),
			Description: types.ExpandUpdatedStringPtr(d.Get("description")),
			Tags:        types.ExpandUpdatedStringsPtr(d.Get("tags")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if !d.Get("external_membership").(bool) && d.HasChanges("application_ids", "user_ids") {
		appIDs := types.ExpandStrings(d.Get("application_ids").(*schema.Set).List())
		userIDs := types.ExpandStrings(d.Get("user_ids").(*schema.Set).List())

		if len(appIDs) > 0 || len(userIDs) > 0 {
			_, err = api.SetGroupMembers(&iam.SetGroupMembersRequest{
				ApplicationIDs: appIDs,
				UserIDs:        userIDs,
				GroupID:        group.ID,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			for i := range group.ApplicationIDs {
				_, err = api.RemoveGroupMember(&iam.RemoveGroupMemberRequest{
					GroupID:       group.ID,
					ApplicationID: &group.ApplicationIDs[i],
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}

			for i := range group.UserIDs {
				_, err = api.RemoveGroupMember(&iam.RemoveGroupMemberRequest{
					GroupID: group.ID,
					UserID:  &group.UserIDs[i],
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return resourceIamGroupRead(ctx, d, m)
}

func resourceIamGroupDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewAPI(m)

	err := api.DeleteGroup(&iam.DeleteGroupRequest{
		GroupID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
