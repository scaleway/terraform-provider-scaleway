package scaleway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam"
)

const (
	SSHKey               = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAICJEoOOgQBLJPs4g/XcPTKT82NywNPpxeuA20FlOPlpO opensource@scaleway.com"
	SSHKeyWithoutComment = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAICJEoOOgQBLJPs4g/XcPTKT82NywNPpxeuA20FlOPlpO"
)

func init() {
	resource.AddTestSweepers("scaleway_iam_ssh_key", &resource.Sweeper{
		Name: "scaleway_iam_ssh_key",
		F:    testSweepIamSSHKey,
	})
}

func testSweepIamSSHKey(_ string) error {
	return sweep(func(scwClient *scw.Client) error {
		iamAPI := iamSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the SSH keys")

		listSSHKeys, err := iamAPI.ListSSHKeys(&iamSDK.ListSSHKeysRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing SSH keys in sweeper: %s", err)
		}

		for _, sshKey := range listSSHKeys.SSHKeys {
			if !isTestResource(sshKey.Name) {
				continue
			}
			err := iamAPI.DeleteSSHKey(&iamSDK.DeleteSSHKeyRequest{
				SSHKeyID: sshKey.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting SSH key in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayIamSSHKey_basic(t *testing.T) {
	name := "tf-test-iam-ssh-key-basic"
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      iam.CheckSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%1s"
						public_key = "%2s"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					iam.CheckSSHKeyExists(tt, "scaleway_iam_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "public_key", SSHKeyWithoutComment),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%1s-updated"
						public_key = "%2s"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					iam.CheckSSHKeyExists(tt, "scaleway_iam_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "name", name+"-updated"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "public_key", SSHKeyWithoutComment),
				),
			},
		},
	})
}

func TestAccScalewayIamSSHKey_WithNewLine(t *testing.T) {
	name := "tf-test-iam-ssh-key-newline"
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      iam.CheckSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%1s"
						public_key = "\n\n%2s\n\n"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					iam.CheckSSHKeyExists(tt, "scaleway_iam_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "public_key", SSHKeyWithoutComment),
				),
			},
		},
	})
}

func TestAccScalewayIamSSHKey_ChangeResourceName(t *testing.T) {
	name := "tf-test-iam-ssh-key-change-resource-name"
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      iam.CheckSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "first" {
						name 	   = "%1s"
						public_key = "\n\n%2s\n\n"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					iam.CheckSSHKeyExists(tt, "scaleway_iam_ssh_key.first"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.first", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.first", "public_key", SSHKeyWithoutComment),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "second" {
						name 	   = "%1s"
						public_key = "\n\n%2s\n\n"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					iam.CheckSSHKeyExists(tt, "scaleway_iam_ssh_key.second"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.second", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.second", "public_key", SSHKeyWithoutComment),
				),
			},
		},
	})
}

func TestAccScalewayIamSSHKey_Disabled(t *testing.T) {
	name := "tf-test-iam-ssh-key-disabled"
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      iam.CheckSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%1s"
						public_key = "\n\n%2s\n\n"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					iam.CheckSSHKeyExists(tt, "scaleway_iam_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "public_key", SSHKeyWithoutComment),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "disabled", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%1s"
						public_key = "\n\n%2s\n\n"
						disabled = "true"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					iam.CheckSSHKeyExists(tt, "scaleway_iam_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "public_key", SSHKeyWithoutComment),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "disabled", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%1s"
						public_key = "\n\n%2s\n\n"
						disabled = "false"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					iam.CheckSSHKeyExists(tt, "scaleway_iam_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "public_key", SSHKeyWithoutComment),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "disabled", "false"),
				),
			},
		},
	})
}
