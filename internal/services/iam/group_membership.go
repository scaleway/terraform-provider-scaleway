package iam

import (
	"context"
	"fmt"
	"slices"
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
		UpdateContext: resourceIamGroupMembershipUpdate,
		DeleteContext: resourceIamGroupMembershipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the group to add the users or applications to",
				ForceNew:    true,
			},
			"user_ids": {
				Type:         schema.TypeList,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Optional:     true,
				Description:  "The IDs of the users to add to the group",
				AtLeastOneOf: []string{"application_ids"},
			},
			"application_ids": {
				Type:         schema.TypeList,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Optional:     true,
				Description:  "The IDs of the applications to add to the group",
				AtLeastOneOf: []string{"user_ids"},
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

	groupID, entityIDs, err := ExpandGroupMembershipResourceID(d.Id())
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

	for _, userID := range entityIDs[EntityKindUser] {
		if !slices.Contains(group.UserIDs, userID) {
			return diag.FromErr(fmt.Errorf("user %s not found in group %s", userID, groupID))
		}
	}

	for _, applicationID := range entityIDs[EntityKindApplication] {
		if !slices.Contains(group.ApplicationIDs, applicationID) {
			return diag.FromErr(fmt.Errorf("application %s not found in group %s", applicationID, groupID))
		}
	}

	_ = d.Set("group_id", groupID)
	_ = d.Set("user_ids", entityIDs[EntityKindUser])
	_ = d.Set("application_ids", entityIDs[EntityKindApplication])

	return nil
}

func resourceIamGroupMembershipUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewAPI(m)

	groupID, _, err := ExpandGroupMembershipResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	userIDs := types.ExpandStrings(d.Get("user_ids"))
	applicationIDs := types.ExpandStrings(d.Get("application_ids"))

	request := &iam.SetGroupMembersRequest{
		GroupID:        groupID,
		UserIDs:        userIDs,
		ApplicationIDs: applicationIDs,
	}

	if d.HasChanges("user_ids", "application_ids") {
		group, err := MakeSetGroupMembershipRequest(ctx, api, request)
		if err != nil {
			return diag.FromErr(err)
		}

		if group.ID != groupID {
			return diag.FromErr(fmt.Errorf("group id changed from %s to %s", groupID, group.ID))
		}

		d.SetId(SetGroupMembershipResourceID(groupID, userIDs, applicationIDs))
	}

	return resourceIamGroupMembershipRead(ctx, d, m)
}

func resourceIamGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewAPI(m)

	groupID, _, err := ExpandGroupMembershipResourceID(d.Id())
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

// Build a parsable state with the following format:
// groupID/user:userID,application:applicationID
func SetGroupMembershipResourceID(groupID string, userIDs []string, applicationIDs []string) (resourceID string) {
	entityIDs := make([]string, 0)

	for _, userID := range userIDs {
		entityIDs = append(entityIDs, fmt.Sprintf("%s:%s", EntityKindUser, userID))
	}

	for _, applicationID := range applicationIDs {
		entityIDs = append(entityIDs, fmt.Sprintf("%s:%s", EntityKindApplication, applicationID))
	}

	resourceID = fmt.Sprintf("%s/%s", groupID, strings.Join(entityIDs, ","))

	return
}

// Parse the group membership resource id and return the group id and the map of entity ids by kind
func ExpandGroupMembershipResourceID(id string) (groupID string, entityIDs map[EntityKind][]string, err error) {
	elems := strings.Split(id, "/")
	if len(elems) != 2 {
		return "", nil, fmt.Errorf("invalid group membership id format, expected {groupID}/{entityKind}:{entityIDs}, got: %s", id)
	}

	groupID = elems[0]

	// entityKind:entityID,entityKind:entityID
	entityKindAndIDs := strings.Split(elems[1], ",")
	entityIDs = make(map[EntityKind][]string)

	for _, entityKindAndID := range entityKindAndIDs {
		splitted := strings.Split(entityKindAndID, ":")
		if len(splitted) != 2 {
			return "", nil, fmt.Errorf("invalid entity kind and id format, expected {entityKind}:{entityID}, got: %s", entityKindAndID)
		}

		entityKind, entityID := EntityKind(splitted[0]), splitted[1]
		if entityKind != EntityKindUser && entityKind != EntityKindApplication {
			return "", nil, fmt.Errorf("invalid entity kind, expected %s or %s, got: %s", EntityKindUser, EntityKindApplication, entityKind)
		}

		entityIDs[entityKind] = append(entityIDs[entityKind], entityID)
	}

	return
}

func MakeSetGroupMembershipRequest(ctx context.Context, api *iam.API, request *iam.SetGroupMembersRequest) (*iam.Group, error) {
	retryInterval := 250 * time.Millisecond
	maxRetries := 10

	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	// the IAM API often returns a 409 when the group is in a transient state
	// so we retry with an exponential backoff
	for i := range maxRetries {
		response, err := api.SetGroupMembers(request, scw.WithContext(ctx))
		if err != nil {
			if httperrors.Is409(err) && strings.Contains(err.Error(), fmt.Sprintf("resource group with ID %s is in a transient state: updating", request.GroupID)) {
				time.Sleep(retryInterval * time.Duration(i)) // lintignore: R018

				continue
			}

			return nil, err
		}

		return response, nil
	}

	return nil, fmt.Errorf("failed to set group membership after %d retries", maxRetries)
}
