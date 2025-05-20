package iam

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

var descParameterIgnoredForGuest = " (this parameter is ignored in case of guest users)"

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIamUserCreate,
		ReadContext:   resourceIamUserRead,
		UpdateContext: resourceIamUserUpdate,
		DeleteContext: resourceIamUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"organization_id": account.OrganizationIDOptionalSchema(),
			// User input data
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The email of the user, which is not editable for guests",
			},
			"tags": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "The tags associated with the user",
			},
			"send_password_email": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether or not to send an email containing the member's password" + descParameterIgnoredForGuest,
			},
			"send_welcome_email": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether or not to send a welcome email that includes onboarding information" + descParameterIgnoredForGuest,
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The member's username" + descParameterIgnoredForGuest,
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The member's password for first access" + descParameterIgnoredForGuest,
			},
			"first_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The member's first name" + descParameterIgnoredForGuest,
			},
			"last_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The member's last name" + descParameterIgnoredForGuest,
			},
			"phone_number": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The member's phone number" + descParameterIgnoredForGuest,
			},
			"locale": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The member's locale" + descParameterIgnoredForGuest,
			},
			// Computed data
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the iam user",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the iam user",
			},
			"deletable": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether or not the iam user is editable",
			},
			"last_login_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of last login of the iam user",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of the iam user",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of user invitation",
			},
			"mfa": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether or not the MFA is enabled",
			},
			"account_root_user_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the account root user associated with the iam user",
			},
			"locked": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines whether the user is locked",
			},
		},
	}
}

func createUserRequestBody(d *schema.ResourceData, isMember bool) *iam.CreateUserRequest {
	if isMember {
		// Create and return a member.
		return &iam.CreateUserRequest{
			OrganizationID: d.Get("organization_id").(string),
			Tags:           types.ExpandStrings(d.Get("tags")),
			Member: &iam.CreateUserRequestMember{
				Email:             d.Get("email").(string),
				SendPasswordEmail: d.Get("send_password_email").(bool),
				SendWelcomeEmail:  d.Get("send_welcome_email").(bool),
				Username:          d.Get("username").(string),
				Password:          d.Get("password").(string),
				FirstName:         d.Get("first_name").(string),
				LastName:          d.Get("last_name").(string),
				PhoneNumber:       d.Get("phone_number").(string),
				Locale:            d.Get("locale").(string),
			},
		}
	} else {
		// Create and return a guest.
		return &iam.CreateUserRequest{
			OrganizationID: d.Get("organization_id").(string),
			Email:          scw.StringPtr(d.Get("email").(string)),
			Tags:           types.ExpandStrings(d.Get("tags")),
		}
	}
}

func resourceIamUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewAPI(m)

	var user *iam.User

	var err error

	if d.Get("username").(string) != "" {
		// Create a member.
		user, err = api.CreateUser(createUserRequestBody(d, true), scw.WithContext(ctx))
	} else {
		// Create a guest.
		user, err = api.CreateUser(createUserRequestBody(d, false), scw.WithContext(ctx))
	}

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.ID)

	return resourceIamUserRead(ctx, d, m)
}

func resourceIamUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewAPI(m)

	user, err := api.GetUser(&iam.GetUserRequest{
		UserID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("organization_id", user.OrganizationID)
	// User input data
	_ = d.Set("email", user.Email)
	_ = d.Set("tags", types.FlattenSliceString(user.Tags))
	_ = d.Set("username", user.Username)
	_ = d.Set("first_name", user.FirstName)
	_ = d.Set("last_name", user.LastName)
	_ = d.Set("phone_number", user.PhoneNumber)
	_ = d.Set("locale", user.Locale)
	// Computed data
	_ = d.Set("created_at", types.FlattenTime(user.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(user.UpdatedAt))
	_ = d.Set("deletable", user.Deletable)
	_ = d.Set("last_login_at", types.FlattenTime(user.LastLoginAt))
	_ = d.Set("type", user.Type)
	_ = d.Set("status", user.Status)
	_ = d.Set("mfa", user.Mfa)
	_ = d.Set("account_root_user_id", user.AccountRootUserID)
	_ = d.Set("locked", user.Locked)

	return nil
}

func resourceIamUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewAPI(m)

	user, err := api.GetUser(&iam.GetUserRequest{
		UserID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if user.Type == "guest" {
		// Users of type "guest" only support the update of tags. The update of the email is not supported.
		if d.HasChanges("tags") {
			_, err = api.UpdateUser(&iam.UpdateUserRequest{
				UserID: user.ID,
				Tags:   types.ExpandUpdatedStringsPtr(d.Get("tags")),
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if d.HasChange("email") {
			return diag.FromErr(errors.New("the email of a guest user cannot be updated, you need to create a new user"))
		}
	} else {
		req := &iam.UpdateUserRequest{UserID: user.ID}

		if d.HasChanges("tags", "email", "first_name") {
			req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
			req.Email = scw.StringPtr(d.Get("email").(string))
			req.FirstName = scw.StringPtr(d.Get("first_name").(string))
		}

		if d.HasChanges("last_name", "phone_number", "locale") {
			req.LastName = scw.StringPtr(d.Get("last_name").(string))
			req.PhoneNumber = scw.StringPtr(d.Get("phone_number").(string))
			req.Locale = scw.StringPtr(d.Get("locale").(string))
		}

		_, err = api.UpdateUser(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		// The update of the 'username' field is made through a different endpoint and payload.
		if d.HasChange("username") {
			_, err = api.UpdateUserUsername(&iam.UpdateUserUsernameRequest{
				UserID:   user.ID,
				Username: d.Get("username").(string),
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceIamUserRead(ctx, d, m)
}

func resourceIamUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewAPI(m)

	err := api.DeleteUser(&iam.DeleteUserRequest{
		UserID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
