package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	account "github.com/scaleway/scaleway-sdk-go/api/account/v2alpha1"
)

func dataSourceScalewayAccountSSHKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalewayAccountSSHKeyRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the SSH key",
			},
			"ssh_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "The ID of the SSH key",
				ValidateFunc: validationUUIDorUUIDWithLocality(),
			},
			"public_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public SSH key",
			},
			"organization_id": organizationIDSchema(),
		},
	}
}

func dataSourceScalewayAccountSSHKeyRead(d *schema.ResourceData, m interface{}) error {
	accountAPI := accountAPI(m)

	var sshKey *account.SSHKey
	sshKeyID, ok := d.GetOk("ssh_key_id")
	if ok {
		res, err := accountAPI.GetSSHKey(&account.GetSSHKeyRequest{SSHKeyID: expandID(sshKeyID)})
		if err != nil {
			return err
		}
		sshKey = res
	} else {
		res, err := accountAPI.ListSSHKeys(&account.ListSSHKeysRequest{
			Name: String(d.Get("name").(string)),
		})
		if err != nil {
			return err
		}
		if len(res.SSHKeys) == 0 {
			return fmt.Errorf("no SSH Key found with the name %s", d.Get("name"))
		}
		if len(res.SSHKeys) > 1 {
			return fmt.Errorf("%d SSH Keys found with the same name %s", len(res.SSHKeys), d.Get("name"))
		}
		sshKey = res.SSHKeys[0]
	}

	d.SetId(sshKey.ID)
	_ = d.Set("name", sshKey.Name)
	_ = d.Set("ssh_key_id", sshKey.ID)
	_ = d.Set("public_key", sshKey.PublicKey)
	_ = d.Set("organization_id", sshKey.OrganizationID)

	return nil
}
