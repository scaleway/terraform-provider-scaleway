package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceIamUserRead,
		SchemaFunc:  dataSourceUserSchema,
	}
}

func dataSourceUserSchema() map[string]*schema.Schema {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceUser().SchemaFunc())

	dsSchema["email"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The email address of the IAM user",
		ValidateDiagFunc: verify.IsEmail(),
		ConflictsWith:    []string{"user_id"},
	}
	dsSchema["user_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the IAM user",
		ValidateDiagFunc: verify.IsUUID(),
		ConflictsWith:    []string{"email"},
	}
	dsSchema["organization_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Description:   "The organization_id you want to attach the resource to",
		Optional:      true,
		ConflictsWith: []string{"user_id"},
	}
	dsSchema["tags"] = &schema.Schema{
		Type: schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional:    true,
		Description: "The tags associated with the user",
	}

	return dsSchema
}

func DataSourceIamUserRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	iamAPI := NewAPI(m)

	var (
		userID any
		err    error
	)

	if id, ok := d.GetOk("user_id"); ok {
		userID = id
	} else {
		email := d.Get("email").(string)

		res, err := iamAPI.ListUsers(&iam.ListUsersRequest{
			OrganizationID: account.GetOrganizationID(m, d),
		}, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		if len(res.Users) == 0 {
			return diag.FromErr(fmt.Errorf("no user found with the email address %s", email))
		}

		for _, user := range res.Users {
			if user.Email == email {
				if userID != nil {
					return diag.Errorf("more than 1 user found with the same email %s", email)
				}

				userID = user.ID
			}
		}

		if userID == nil {
			return diag.Errorf("no user found with the email %s", email)
		}
	}

	d.SetId(userID.(string))

	err = d.Set("user_id", userID)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := iamAPI.GetUser(&iam.GetUserRequest{
		UserID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	setUserState(d, res)

	return nil
}
