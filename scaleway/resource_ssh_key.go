package scaleway

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
	"golang.org/x/crypto/ssh"
)

func resourceScalewaySSHKey() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "This resource is deprecated and will be removed in the next major version",

		Create: resourceScalewaySSHKeyCreate,
		Read:   resourceScalewaySSHKeyRead,
		Delete: resourceScalewaySSHKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ssh key",
			},
		},
	}
}

func sshKeyFingerprint(key []byte) (string, error) {
	pubkey, _, _, _, err := ssh.ParseAuthorizedKey(key)
	if err != nil {
		return "", err
	}
	return ssh.FingerprintLegacyMD5(pubkey), nil
}

func resourceScalewaySSHKeyCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	fingerprint, err := sshKeyFingerprint([]byte(d.Get("key").(string)))
	if err != nil {
		return err
	}

	user, err := scaleway.GetUser()
	if err != nil {
		return err
	}

	keys := []api.KeyDefinition{}
	exists := false
	for _, key := range user.SSHPublicKeys {
		exists = exists || key.Key == d.Get("key").(string)
		keys = append(keys, api.KeyDefinition{
			Key: key.Key,
		})
	}

	// remote already contains the key, nothing to do
	if exists {
		d.SetId(fingerprint)
		return nil
	}

	_, err = scaleway.PatchUserSSHKey(user.ID, api.UserPatchSSHKeyDefinition{
		SSHPublicKeys: append(keys, api.KeyDefinition{
			Key: strings.TrimSpace(d.Get("key").(string)),
		}),
	})

	if err != nil {
		return err
	}

	d.SetId(fingerprint)
	return nil
}

func resourceScalewaySSHKeyRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	user, err := scaleway.GetUser()
	if err != nil {
		return err
	}

	exists := false
	for _, key := range user.SSHPublicKeys {
		exists = exists || strings.Contains(key.Fingerprint, d.Id())
		if exists {
			_ = d.Set("key", key.Key)
			break
		}
	}
	if !exists {
		return fmt.Errorf("ssh key does not exist anymore")
	}

	return nil
}

func resourceScalewaySSHKeyDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	user, err := scaleway.GetUser()
	if err != nil {
		return err
	}

	keys := []api.KeyDefinition{}
	for _, key := range user.SSHPublicKeys {
		if !strings.Contains(key.Fingerprint, d.Id()) {
			keys = append(keys, api.KeyDefinition{
				Key: key.Key,
			})
		}
	}
	_, err = scaleway.PatchUserSSHKey(user.ID, api.UserPatchSSHKeyDefinition{
		SSHPublicKeys: keys,
	})

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}
