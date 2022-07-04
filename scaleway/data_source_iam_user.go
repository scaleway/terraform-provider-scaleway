package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayIamUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayIamUserRead,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The ID of the IAM user",
				ValidateFunc:  validationUUID(),
				ConflictsWith: []string{"email"},
			},
			"email": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The email address of the IAM user",
				ValidateFunc:  validationEmail(),
				ConflictsWith: []string{"user_id"},
			},

			// Default organization_id will be available on a major release. Please check #1337
			"organization_id": {
				Type:        schema.TypeString,
				Description: "The organization_id you want to attach the resource to",
				Required:    true,
			},
		},
	}
}

func dataSourceScalewayIamUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamAPI := iamAPI(meta)

	var email, organizationID string
	userID, ok := d.GetOk("user_id")
	if ok {
		userID = d.Get("user_id")
		res, err := iamAPI.GetUser(&iam.GetUserRequest{
			UserID: userID.(string),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		email = res.Email
		organizationID = res.OrganizationID
	} else {
		res, err := iamAPI.ListUsers(&iam.ListUsersRequest{
			OrganizationID: expandStringPtr(d.Get("organization_id")),
		}, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Users) == 0 {
			return diag.FromErr(fmt.Errorf("no user found with the email address %s", d.Get("email")))
		}
		for _, user := range res.Users {
			if user.Email == d.Get("email").(string) {
				if userID != "" {
					return diag.Errorf("more than 1 user found with the same email %s", d.Get("email"))
				}
				userID, email = user.ID, user.Email
			}
		}
		if userID == "" {
			return diag.Errorf("no user found with the email %s", d.Get("email"))
		}
	}

	d.SetId(userID.(string))
	err := d.Set("user_id", userID)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("user_id", userID)
	_ = d.Set("email", email)
	_ = d.Set("organization_id", organizationID)

	return nil
}
