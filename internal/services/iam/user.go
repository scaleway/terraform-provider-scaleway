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
)

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
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The description of the iam user",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the user",
			},
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
				Description: "The status of user invitation.",
			},
			"mfa": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether or not the MFA is enabled",
			},
			"account_root_user_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the account root user associated with the iam user.",
			},
			"organization_id": account.OrganizationIDOptionalSchema(),
		},
	}
}

func resourceIamUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewAPI(m)
	email := d.Get("email").(string)
	user, err := api.CreateUser(&iam.CreateUserRequest{
		OrganizationID: d.Get("organization_id").(string),
		Email:          &email,
		Tags:           types.ExpandStrings(d.Get("tags")),
	}, scw.WithContext(ctx))
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

	_ = d.Set("email", user.Email)
	_ = d.Set("created_at", types.FlattenTime(user.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(user.UpdatedAt))
	_ = d.Set("organization_id", user.OrganizationID)
	_ = d.Set("deletable", user.Deletable)
	_ = d.Set("tags", types.FlattenSliceString(user.Tags))
	_ = d.Set("last_login_at", types.FlattenTime(user.LastLoginAt))
	_ = d.Set("type", user.Type)
	_ = d.Set("status", user.Status)
	_ = d.Set("mfa", user.Mfa)

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

	if d.HasChanges("tags") {
		_, err = api.UpdateUser(&iam.UpdateUserRequest{
			UserID: user.ID,
			Tags:   types.ExpandUpdatedStringsPtr(d.Get("tags")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
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
