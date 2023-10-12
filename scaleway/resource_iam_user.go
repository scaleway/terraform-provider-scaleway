package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayIamUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayIamUserCreate,
		ReadContext:   resourceScalewayIamUserRead,
		DeleteContext: resourceScalewayIamUserDelete,
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
			"organization_id": organizationIDOptionalSchema(),
		},
	}
}

func resourceScalewayIamUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamAPI := iamAPI(meta)
	user, err := iamAPI.CreateUser(&iam.CreateUserRequest{
		OrganizationID: d.Get("organization_id").(string),
		Email:          d.Get("email").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.ID)

	return resourceScalewayIamUserRead(ctx, d, meta)
}

func resourceScalewayIamUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)
	user, err := api.GetUser(&iam.GetUserRequest{
		UserID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("email", user.Email)
	_ = d.Set("created_at", flattenTime(user.CreatedAt))
	_ = d.Set("updated_at", flattenTime(user.UpdatedAt))
	_ = d.Set("organization_id", user.OrganizationID)
	_ = d.Set("deletable", user.Deletable)
	_ = d.Set("last_login_at", flattenTime(user.LastLoginAt))
	_ = d.Set("type", user.Type)
	_ = d.Set("status", user.Status)
	_ = d.Set("mfa", user.Mfa)

	return nil
}

func resourceScalewayIamUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)

	err := api.DeleteUser(&iam.DeleteUserRequest{
		UserID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
