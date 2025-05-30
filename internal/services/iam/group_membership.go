package iam

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

type EntityKind string

const (
	EntityKindUser        EntityKind = "user"
	EntityKindApplication EntityKind = "application"
)

func ResourceGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIamGroupMembershipCreate,
		ReadContext:   resourceIamGroupMembershipRead,
		DeleteContext: resourceIamGroupMembershipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"user_ids": {
				Type:         schema.TypeList,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Optional:     true,
				ExactlyOneOf: []string{"application_ids"},
				Description:  "The IDs of the users",
				ForceNew:     true,
			},
			"application_ids": {
				Type:         schema.TypeList,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Optional:     true,
				ExactlyOneOf: []string{"user_ids"},
				Description:  "The IDs of the applications",
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

	userIDs := types.ExpandStrings(d.Get("user_ids"))
	applicationIDs := types.ExpandStrings(d.Get("application_ids"))

	group, err := MakeSetGroupMembershipRequest(ctx, api, &iam.SetGroupMembersRequest{
		GroupID:        d.Get("group_id").(string),
		UserIDs:        userIDs,
		ApplicationIDs: applicationIDs,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(SetGroupMembershipResourceID(group.ID, userIDs, applicationIDs))

	return resourceIamGroupMembershipRead(ctx, d, m)
}

func resourceIamGroupMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewAPI(m)

	groupID, entityKind, entityIDs, err := ExpandGroupMembershipResourceID(d.Id())
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

	foundEntityIDs := make([]bool, len(entityIDs))

	if entityKind == EntityKindUser {
		for i, groupUserID := range group.UserIDs {
			if slices.Contains(entityIDs, groupUserID) {
				foundEntityIDs[i] = true
			}
		}
	} else if entityKind == EntityKindApplication {
		for i, groupApplicationID := range group.ApplicationIDs {
			if slices.Contains(entityIDs, groupApplicationID) {
				foundEntityIDs[i] = true
			}
		}
	}

	if slices.Contains(foundEntityIDs, false) {
		d.SetId("")

		return nil
	}

	_ = d.Set("group_id", groupID)
	_ = d.Set(fmt.Sprintf("%s_ids", entityKind), entityIDs)

	return nil
}

func resourceIamGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewAPI(m)

	groupID, _, _, err := ExpandGroupMembershipResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = MakeSetGroupMembershipRequest(ctx, api, &iam.SetGroupMembersRequest{
		GroupID: groupID,
	})

	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}

func SetGroupMembershipResourceID(groupID string, userIDs []string, applicationIDs []string) (resourceID string) {
	sort.Strings(userIDs)
	sort.Strings(applicationIDs)

	if len(userIDs) > 0 {
		resourceID = fmt.Sprintf("%s/%s/%s", groupID, EntityKindUser, strings.Join(userIDs, ","))
	} else if len(applicationIDs) > 0 {
		resourceID = fmt.Sprintf("%s/%s/%s", groupID, EntityKindApplication, strings.Join(applicationIDs, ","))
	}

	return
}

func ExpandGroupMembershipResourceID(id string) (groupID string, kind EntityKind, entityIDs []string, err error) {
	elems := strings.Split(id, "/")
	if len(elems) != 3 {
		return "", "", []string{}, fmt.Errorf("invalid group membership id format, expected {groupID}/{entityKind}/{entityIDs}, got: %s", id)
	}

	groupID = elems[0]
	kind = EntityKind(elems[1])
	if kind != EntityKindUser && kind != EntityKindApplication {
		return "", "", []string{}, fmt.Errorf("invalid entity kind, expected %s or %s, got: %s", EntityKindUser, EntityKindApplication, kind)
	}
	entityIDs = strings.Split(elems[2], ",")

	return
}

func MakeSetGroupMembershipRequest(ctx context.Context, api *iam.API, request *iam.SetGroupMembersRequest) (*iam.Group, error) {
	retryInterval := 250 * time.Millisecond
	maxRetries := 10

	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var response *iam.Group
	var err error

	// exponential backoff
	for i := 0; i < maxRetries; i++ {
		response, err = api.SetGroupMembers(request, scw.WithContext(ctx))
		if err != nil {
			if httperrors.Is409(err) && strings.Contains(err.Error(), fmt.Sprintf("resource group with ID %s is in a transient state: updating", request.GroupID)) {
				time.Sleep(retryInterval * time.Duration(i))
				continue
			}

			return nil, err
		}

		return response, nil
	}

	return nil, fmt.Errorf("failed to set group membership after %d retries", maxRetries)
}
