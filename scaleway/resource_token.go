package scaleway

import (
	"github.com/hashicorp/terraform/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
)

func resourceScalewayToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayTokenCreate,
		Read:   resourceScalewayTokenRead,
		Update: resourceScalewayTokenUpdate,
		Delete: resourceScalewayTokenDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The account email. Defaults to registered user.",
			},
			"user_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The userid of the associated user.",
			},
			"access_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The access_key.",
			},
			"secret_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The secret_key.",
				Sensitive:   true,
			},
			"creation_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ip used to create the key.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The token description.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User password, in case a login is require",
			},
			"expires": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Defines if the token is set to expire",
			},
			"expiration_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The tokens expiration date",
			},
		},
	}
}

func resourceScalewayTokenCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	mu.Lock()
	email := ""
	if mail, ok := d.GetOk("email"); ok {
		email = mail.(string)
	} else {
		user, err := scaleway.GetUser()
		if err != nil {
			return err
		}
		email = user.Email
	}

	token, err := scaleway.CreateToken(&api.CreateTokenRequest{
		Email:    email,
		Password: d.Get("password").(string),
		Expires:  d.Get("expires").(bool),
	})
	mu.Unlock()
	if err != nil {
		return err
	}

	d.SetId(token.ID)
	return resourceScalewayTokenUpdate(d, m)
}

func resourceScalewayTokenRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	token, err := scaleway.GetToken(d.Id())
	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}
	d.Set("description", token.Description)
	d.Set("expiration_date", token.Expires)
	d.Set("expires", token.Expires != "")
	d.Set("user_id", token.UserID)
	d.Set("creation_ip", token.CreationIP)
	d.Set("access_key", token.AccessKey)
	d.Set("secret_key", token.SecretKey)
	user, err := scaleway.GetUser()
	if err != nil {
		return err
	}
	if user.ID == token.UserID {
		d.Set("email", user.Email)
	}

	return nil
}

func resourceScalewayTokenUpdate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	if d.HasChange("description") || d.HasChange("expires") {
		_, err := scaleway.UpdateToken(&api.UpdateTokenRequest{
			ID:          d.Id(),
			Expires:     d.Get("expires").(bool),
			Description: d.Get("description").(string),
		})
		if err != nil {
			return err
		}
	}

	return resourceScalewayTokenRead(d, m)
}

func resourceScalewayTokenDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	err := scaleway.DeleteToken(d.Id())
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}
