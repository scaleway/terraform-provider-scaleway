package scaleway

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	account "github.com/scaleway/scaleway-sdk-go/api/account/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/assert"
)

func init() {
	resource.AddTestSweepers("scaleway_account_ssh_key", &resource.Sweeper{
		Name: "scaleway_account_ssh_key",
		F:    testSweepAccountSSHKey,
	})
}

func testSweepAccountSSHKey(_ string) error {
	return sweepZones([]scw.Zone{scw.ZoneFrPar1}, func(scwClient *scw.Client, zone scw.Zone) error {
		accountAPI := account.NewAPI(scwClient)

		l.Debugf("sweeper: destroying the SSH keys")

		listSSHKeys, err := accountAPI.ListSSHKeys(&account.ListSSHKeysRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing SSH keys in sweeper: %s", err)
		}

		for _, sshKey := range listSSHKeys.SSHKeys {
			err := accountAPI.DeleteSSHKey(&account.DeleteSSHKeyRequest{
				SSHKeyID: sshKey.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting SSH key in sweeper: %s", err)
			}
		}

		return nil
	})
}
func newTestToolsAndNameAndSSHKey(t *testing.T) (tt *TestTools, name string) {
	tt = NewTestTools(t)
	expectedName := "ssh-key-name"
	tt.Recorder.AddSaveFilter(newSSHKeyNameFilter(expectedName))
	if *UpdateCassettes {
		name = newRandomName("ssh-key")
	} else {
		name = expectedName
	}
	return
}

func TestAccScalewayAccountSSHKey_basic(t *testing.T) {
	tt, name := newTestToolsAndNameAndSSHKey(t)
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEEYrzDOZmhItdKaDAEqJQ4ORS2GyBMtBozYsK5kiXXX opensource@scaleway.com"
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayAccountSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + name + `"
						public_key = "` + SSHKey + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountSSHKeyExists(tt, "scaleway_account_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "public_key", SSHKey),
				),
			},
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + name + `-updated"
						public_key = "` + SSHKey + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountSSHKeyExists(tt, "scaleway_account_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "name", name+"-updated"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "public_key", SSHKey),
				),
			},
		},
	})
}

func TestAccScalewayAccountSSHKey_WithNewLine(t *testing.T) {
	tt, name := newTestToolsAndNameAndSSHKey(t)
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDjfkdWCwkYlVQMDUfiZlVrmjaGOfBYnmkucssae8Iup opensource@scaleway.com"
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayAccountSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + name + `"
						public_key = "\n\n` + SSHKey + `\n\n"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountSSHKeyExists(tt, "scaleway_account_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "public_key", SSHKey),
				),
			},
		},
	})
}

func testAccCheckScalewayAccountSSHKeyDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_account_ssh_key" {
				continue
			}

			accountAPI := accountAPI(tt.Meta)

			_, err := accountAPI.GetSSHKey(&account.GetSSHKeyRequest{
				SSHKeyID: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("SSH key (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}

func testAccCheckScalewayAccountSSHKeyExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		accountAPI := accountAPI(tt.Meta)

		_, err := accountAPI.GetSSHKey(&account.GetSSHKeyRequest{
			SSHKeyID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func newSSHKeyNameFilter(expectedSSHKeyName string) func(i *cassette.Interaction) error {
	return func(i *cassette.Interaction) error {
		if i.Request.Body != "" {
			var body account.CreateSSHKeyRequest
			err := json.Unmarshal([]byte(i.Request.Body), &body)
			if err != nil {
				return err
			}
			keyname := expectedSSHKeyName
			if strings.Contains(body.Name, "-updated") {
				keyname = fmt.Sprintf("%s-updated", keyname)
			}
			body.Name = keyname
			json, err := json.Marshal(body)
			if err != nil {
				return err
			}
			i.Request.Body = string(json)
		}
		if i.Response.Body != "" {
			var body account.SSHKey
			err := json.Unmarshal([]byte(i.Response.Body), &body)
			if err != nil {
				return err
			}
			keyname := expectedSSHKeyName
			if strings.Contains(body.Name, "-updated") {
				keyname = fmt.Sprintf("%s-updated", keyname)
			}
			body.Name = keyname
			json, err := json.Marshal(body)
			if err != nil {
				return err
			}
			i.Response.Body = string(json)
		}
		return nil
	}
}

func TestNewSSHKeyNameFilterWithRequestWithBodyContainingSSHKeyNameModifyTheBody(t *testing.T) {
	assert := assert.New(t)
	request := cassette.Request{
		Body: fmt.Sprintf("{\"name\":\"%s\",\"public_key\":\"ssh-ed25519\"}", newRandomName("test")),
	}
	interaction := cassette.Interaction{Request: request}

	err := newSSHKeyNameFilter("ssh-key-name")(&interaction)
	assert.NoError(err)

	assert.Equal("{\"name\":\"ssh-key-name\",\"public_key\":\"ssh-ed25519\"}", interaction.Request.Body)
}

func TestNewSSHKeyNameFilterWithResponseContainingSSHKeyNameModifyTheBody(t *testing.T) {
	assert := assert.New(t)
	response := cassette.Response{
		Body: fmt.Sprintf("{\"id\":\"\",\"name\":\"%s\",\"public_key\":\"ssh-ed25519\",\"fingerprint\":\"\",\"created_at\":null,\"updated_at\":null,\"creation_info\":null,\"organization_id\":\"\",\"project_id\":\"\"}", newRandomName("test")),
	}
	interaction := cassette.Interaction{Response: response}

	err := newSSHKeyNameFilter("ssh-key-name")(&interaction)
	assert.NoError(err)

	assert.Equal("{\"id\":\"\",\"name\":\"ssh-key-name\",\"public_key\":\"ssh-ed25519\",\"fingerprint\":\"\",\"created_at\":null,\"updated_at\":null,\"creation_info\":null,\"organization_id\":\"\",\"project_id\":\"\"}", interaction.Response.Body)
}

func TestNewSSHKeyNameFilterWithRequestWithBodyContainingUpdatedSSHKeyNameModifyTheBody(t *testing.T) {
	assert := assert.New(t)
	request := cassette.Request{
		Body: fmt.Sprintf("{\"name\":\"%s-updated\",\"public_key\":\"ssh-ed25519\"}", newRandomName("test")),
	}
	interaction := cassette.Interaction{Request: request}

	err := newSSHKeyNameFilter("ssh-key-name")(&interaction)
	assert.NoError(err)

	assert.Equal("{\"name\":\"ssh-key-name-updated\",\"public_key\":\"ssh-ed25519\"}", interaction.Request.Body)
}

func TestNewSSHKeyNameFilterWithResponseContainingUpdatedSSHKeyNameModifyTheBody(t *testing.T) {
	assert := assert.New(t)
	response := cassette.Response{
		Body: fmt.Sprintf("{\"id\":\"\",\"name\":\"%s-updated\",\"public_key\":\"ssh-ed25519\",\"fingerprint\":\"\",\"created_at\":null,\"updated_at\":null,\"creation_info\":null,\"organization_id\":\"\",\"project_id\":\"\"}", newRandomName("test")),
	}
	interaction := cassette.Interaction{Response: response}

	err := newSSHKeyNameFilter("ssh-key-name")(&interaction)
	assert.NoError(err)

	assert.Equal("{\"id\":\"\",\"name\":\"ssh-key-name-updated\",\"public_key\":\"ssh-ed25519\",\"fingerprint\":\"\",\"created_at\":null,\"updated_at\":null,\"creation_info\":null,\"organization_id\":\"\",\"project_id\":\"\"}", interaction.Response.Body)
}
