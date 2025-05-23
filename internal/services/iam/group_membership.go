package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceGroupMembership() *schema.Resource {
	return &schema.Resource{
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
		CreateContext:                     resourceIamGroupMembershipCreate,
		ReadContext:                       resourceIamGroupMembershipRead,
		DeleteContext:                     resourceIamGroupMembershipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The ID of the user",
				ExactlyOneOf: []string{"application_id"},
				ForceNew:     true,
			},
			"application_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The ID of the user",
				ExactlyOneOf: []string{"user_id"},
				ForceNew:     true,
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the group to add the user to",
				ForceNew:    true,
			},
		},
	}
}

func resourceIamGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewAPI(m)

	userID := types.ExpandStringPtr(d.Get("user_id"))
	applicationID := types.ExpandStringPtr(d.Get("application_id"))

	group, err := api.AddGroupMember(&iam.AddGroupMemberRequest{
		GroupID:       d.Get("group_id").(string),
		UserID:        userID,
		ApplicationID: applicationID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(GroupMembershipID(group.ID, userID, applicationID))

	return resourceIamGroupMembershipRead(ctx, d, m)
}

func resourceIamGroupMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewAPI(m)

	groupID, userID, applicationID, err := ExpandGroupMembershipID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	group, err := api.GetGroup(&iam.GetGroupRequest{
		GroupID: groupID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	foundInGroup := false

	if userID != "" {
		for _, groupUserID := range group.UserIDs {
			if groupUserID == userID {
				foundInGroup = true

				break
			}
		}
	} else if applicationID != "" {
		for _, groupApplicationID := range group.ApplicationIDs {
			if groupApplicationID == applicationID {
				foundInGroup = true

				break
			}
		}
	}

	if !foundInGroup {
		d.SetId("")

		return nil
	}

	_ = d.Set("group_id", groupID)
	_ = d.Set("user_id", userID)
	_ = d.Set("application_id", applicationID)

	return nil
}

func resourceIamGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewAPI(m)

	groupID, userID, applicationID, err := ExpandGroupMembershipID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &iam.RemoveGroupMemberRequest{
		GroupID: groupID,
	}

	if userID != "" {
		req.UserID = &userID
	} else if applicationID != "" {
		req.ApplicationID = &applicationID
	}

	_, err = api.RemoveGroupMember(req, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}

func GroupMembershipID(groupID string, userID *string, applicationID *string) string {
	if userID != nil {
		return fmt.Sprintf("%s/user/%s", groupID, *userID)
	}

	return fmt.Sprintf("%s/app/%s", groupID, *applicationID)
}

func ExpandGroupMembershipID(id string) (groupID string, userID string, applicationID string, err error) {
	elems := strings.Split(id, "/")
	if len(elems) != 3 {
		return "", "", "", fmt.Errorf("invalid group member id format, expected {groupID}/{type}/{memberID}, got: %s", id)
	}

	groupID = elems[0]

	switch elems[1] {
	case "user":
		userID = elems[2]
	case "app":
		applicationID = elems[2]
	}

	return
}
