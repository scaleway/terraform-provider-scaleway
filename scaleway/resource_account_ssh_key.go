package scaleway

import (
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	account "github.com/scaleway/scaleway-sdk-go/api/account/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayAccountSSKKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayAccountSSHKeyCreate,
		Read:   resourceScalewayAccountSSHKeyRead,
		Update: resourceScalewayAccountSSHKeyUpdate,
		Delete: resourceScalewayAccountSSHKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the SSH key",
			},
			"public_key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The public SSH key",
				// We dont consider trailing \n as diff
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.Trim(old, "\n") == strings.Trim(new, "\n")
				},
			},
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayAccountSSHKeyCreate(d *schema.ResourceData, m interface{}) error {
	accountAPI := getAccountAPI(m)

	res, err := accountAPI.CreateSSHKey(&account.CreateSSHKeyRequest{
		Name:           d.Get("name").(string),
		PublicKey:      strings.Trim(d.Get("public_key").(string), "\n"),
		OrganizationID: d.Get("organization_id").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(res.ID)

	return resourceScalewayAccountSSHKeyRead(d, m)
}

func resourceScalewayAccountSSHKeyRead(d *schema.ResourceData, m interface{}) error {
	accountAPI := getAccountAPI(m)

	res, err := accountAPI.GetSSHKey(&account.GetSSHKeyRequest{
		SSHKeyID: d.Id(),
	})
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", res.Name)
	d.Set("public_key", res.PublicKey)
	d.Set("organization_id", res.OrganizationID)

	return nil
}

func resourceScalewayAccountSSHKeyUpdate(d *schema.ResourceData, m interface{}) error {
	accountAPI := getAccountAPI(m)

	if d.HasChange("name") {
		_, err := accountAPI.UpdateSSHKey(&account.UpdateSSHKeyRequest{
			SSHKeyID: d.Id(),
			Name:     scw.StringPtr(d.Get("name").(string)),
		})
		if err != nil {
			return err
		}
	}

	return resourceScalewayAccountSSHKeyRead(d, m)
}

func resourceScalewayAccountSSHKeyDelete(d *schema.ResourceData, m interface{}) error {
	accountAPI := getAccountAPI(m)

	err := accountAPI.DeleteSSHKey(&account.DeleteSSHKeyRequest{
		SSHKeyID: d.Id(),
	})
	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
