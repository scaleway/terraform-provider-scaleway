package scaleway

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
)

func resourceScalewayUserData() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: `This resource is deprecated and will be removed in the next major version.
 Please use scaleway_instance_server instead.`,

		Create: resourceScalewayUserDataCreate,
		Read:   resourceScalewayUserDataRead,
		Update: resourceScalewayUserDataUpdate,
		Delete: resourceScalewayUserDataDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"server": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The server the meta data is associated with",
			},
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The key of the user data to manage",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The value of the user",
			},
		},
	}
}

func resourceScalewayUserDataCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	if err := scaleway.PatchUserdata(
		d.Get("server").(string),
		d.Get("key").(string),
		[]byte(d.Get("value").(string)),
		false); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("userdata-%s-%s", d.Get("server").(string), d.Get("key").(string)))
	return resourceScalewayUserDataRead(d, m)
}

func resourceScalewayUserDataRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	if d.Get("server").(string) == "" {
		// import case
		parts := strings.Split(d.Id(), "-")
		_ = d.Set("key", parts[len(parts)-1])
		_ = d.Set("server", strings.Join(parts[1:len(parts)-1], "-"))
	}
	userdata, err := scaleway.GetUserdata(
		d.Get("server").(string),
		d.Get("key").(string),
		false,
	)

	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	_ = d.Set("value", userdata.String())
	return nil
}

func resourceScalewayUserDataUpdate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	if err := scaleway.PatchUserdata(
		d.Get("server").(string),
		d.Get("key").(string),
		[]byte(d.Get("value").(string)),
		false); err != nil {
		return err
	}

	return resourceScalewayUserDataRead(d, m)
}

func resourceScalewayUserDataDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	err := scaleway.DeleteUserdata(
		d.Get("server").(string),
		d.Get("key").(string),
		false)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}
