package scaleway

import (
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestGetSSHKeyFingerprint(t *testing.T) {
	key := []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDYpDmIzRs5c+xs0jmljMbNYVcgV8fRruMCRDA4HKjGN2lqLTZhngGDXsdt/2kTNQQPAq2sR4N8mfX5wMRT/+jNb+8esPyY5WlElni0zmD7oLoPW4lYRES6f7EeAv6NttLfkDO42r15OtMnglcgWk1u4o3lOXuLbhzJT1qdicpDja22X3uR/xUy1AYhKBOoiSlQbkb7NhL0lA1xQNwerdaJJS8tFB+wViVDyP0f1HaIRxViFlTGuTbTuIJNR/7VJ9VBBuTnYXaRkPxz64sUXrtdVK8U0+4KsisyXwmgQKnvZBDj91wxz12OOzFSQ52iFprIj1JbkzuBmNWXUGKYzXJZ nicolai86@test")
	fingerprint, err := getSSHKeyFingerprint(key)

	if err != nil {
		t.Errorf("Expected no error, but got %v", err.Error())
	}
	if fingerprint != "d1:4c:45:59:a8:ee:e6:41:10:fb:3c:3e:54:98:5b:6f" {
		t.Errorf("Expected fingerprint of %q, but got %q", "d1:4c:45:59:a8:ee:e6:41:10:fb:3c:3e:54:98:5b:6f", fingerprint)
	}
}

func TestAccScalewaySSHKey_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewaySSHKeyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckScalewaySSHKeyConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scaleway_ssh_key.test", "id", "d1:4c:45:59:a8:ee:e6:41:10:fb:3c:3e:54:98:5b:6f"),
				),
			},
			resource.TestStep{
				Config: testAccCheckScalewaySSHKeysConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scaleway_ssh_key.test", "id", "d1:4c:45:59:a8:ee:e6:41:10:fb:3c:3e:54:98:5b:6f"),
					resource.TestCheckResourceAttr(
						"scaleway_ssh_key.test2", "id", "71:a9:e9:ec:5a:43:bc:49:0c:59:1d:74:0d:bb:a4:24"),
				),
			},
			resource.TestStep{
				Config: testAccCheckScalewaySSHKeyConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scaleway_ssh_key.test", "id", "d1:4c:45:59:a8:ee:e6:41:10:fb:3c:3e:54:98:5b:6f"),
				),
			},
		},
	})
}

func testAccCheckScalewaySSHKeyDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client).scaleway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway" {
			continue
		}

		user, err := client.GetUser()
		if err != nil {
			return err
		}
		for _, key := range user.SSHPublicKeys {
			if strings.Contains(key.Fingerprint, rs.Primary.ID) {
				return errors.New("key still exists.")
			}
		}
		return nil
	}

	return nil
}

var testAccCheckScalewaySSHKeyConfig = `
resource "scaleway_ssh_key" "test" {
	key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDYpDmIzRs5c+xs0jmljMbNYVcgV8fRruMCRDA4HKjGN2lqLTZhngGDXsdt/2kTNQQPAq2sR4N8mfX5wMRT/+jNb+8esPyY5WlElni0zmD7oLoPW4lYRES6f7EeAv6NttLfkDO42r15OtMnglcgWk1u4o3lOXuLbhzJT1qdicpDja22X3uR/xUy1AYhKBOoiSlQbkb7NhL0lA1xQNwerdaJJS8tFB+wViVDyP0f1HaIRxViFlTGuTbTuIJNR/7VJ9VBBuTnYXaRkPxz64sUXrtdVK8U0+4KsisyXwmgQKnvZBDj91wxz12OOzFSQ52iFprIj1JbkzuBmNWXUGKYzXJZ test"
}
`

var testAccCheckScalewaySSHKeysConfig = `
resource "scaleway_ssh_key" "test" {
	key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDYpDmIzRs5c+xs0jmljMbNYVcgV8fRruMCRDA4HKjGN2lqLTZhngGDXsdt/2kTNQQPAq2sR4N8mfX5wMRT/+jNb+8esPyY5WlElni0zmD7oLoPW4lYRES6f7EeAv6NttLfkDO42r15OtMnglcgWk1u4o3lOXuLbhzJT1qdicpDja22X3uR/xUy1AYhKBOoiSlQbkb7NhL0lA1xQNwerdaJJS8tFB+wViVDyP0f1HaIRxViFlTGuTbTuIJNR/7VJ9VBBuTnYXaRkPxz64sUXrtdVK8U0+4KsisyXwmgQKnvZBDj91wxz12OOzFSQ52iFprIj1JbkzuBmNWXUGKYzXJZ test"
}

resource "scaleway_ssh_key" "test2" {
	key = "ssh-dss AAAAB3NzaC1kc3MAAACBAOU48/Wjx5JYNdcxdbb0mfJ3vtRU5wzPXmJcaa5CbpWeG6x3wD2+1aasxgO54YdSfEJwVXcoqPwx5gQfpAEgZmYi3M7Yurv+iwAJaSk+CFHOdhxUNPdwWKxsuIA1vk+edhqTKPC5fMFPpMQU/QDr5XegLhCUq11oRjnpfhzmi96/AAAAFQCTRSG8CPxOGVfYbZF/NjRGRgDWMQAAAIArEya6WPd7Bz19rn6u0KC0LeBHmxjoe0M9hblrFHjL4sLpxW1qipUKN+zwXKR9lv4Y/voyzirc7a8DPrEIyMy6SjPu/CiTwx7zDv08nE4qx20V8X0FrusvPbm5jzQJBweUvUZFZcM7Ybvk7RwawaLCGBGZ6Mg/P2YWTfor88NnjwAAAIBUfM9wfQvn0bHso8bsxFFtdME0eZbyeIRlU8JPjOatei4/eyFMHYvfqeGxiGgox1/E7/qX+3rXuypQFTa8DlhyteZCproysFa8NRh8PJZ7uchWrgPZHuXISW+UwlJ/5cJpFxn3ijzdOzEj5EiDM+LBtFYtA5/obIq68eqK6tqM0w== test2"
}
`
