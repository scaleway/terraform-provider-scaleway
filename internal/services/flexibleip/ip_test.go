package flexibleip_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	annotationsSDK "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	baremetalSDK "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	flexibleipSDK "github.com/scaleway/scaleway-sdk-go/api/flexibleip/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	annotationstestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/annotations/testfuncs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/baremetal"
	baremetalchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/baremetal/testfuncs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/flexibleip"
)

const (
	SSHKeyFlexibleIP = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIM7HUxRyQtB2rnlhQUcbDGCZcTJg7OvoznOiyC9W6IxH opensource@scaleway.com"
	BaremetalOffer   = "EM-B220E-NVME"
)

var DestroyWaitTimeout = 3 * time.Minute

func TestAccFlexibleIP_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckFlexibleIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_flexible_ip" "main" {}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.main"),
					resource.TestCheckResourceAttrWith("scaleway_flexible_ip.main", "srn", acctest.CheckSRN("flexible-ip", "flexible-ips")),
				),
			},
			{
				ResourceName:            "scaleway_flexible_ip.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"is_ipv6"},
			},
		},
	})
}

func TestAccFlexibleIP_WithZone(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckFlexibleIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_flexible_ip" "base" {}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "fr-par-1"),
				),
			},
			{
				Config: `
						resource "scaleway_flexible_ip" "base" {
							zone = "nl-ams-1"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "nl-ams-1"),
				),
			},
		},
	})
}

func TestAccFlexibleIP_IPv6(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckFlexibleIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_flexible_ip" "main" {
							is_ipv6 = true
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.main"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.main", "is_ipv6", "true"),
					testAccCheckFlexibleIPIsIPv6(tt, "scaleway_flexible_ip.main"),
				),
			},
		},
	})
}

func TestAccFlexibleIP_DefaultAnnotations(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	keyName := "tf_test_fip_default_key"
	valueName := "tf_test_fip_default_value"

	tt.Meta.SetDefaultAnnotations(map[string]string{keyName: valueName})

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckFlexibleIPDestroy(tt),
			annotationstestfuncs.IsAnnotationsKeyDestroyed(tt),
			annotationstestfuncs.IsAnnotationsValueDestroyed(tt),
			testAccCheckFlexibleIPDefaultAnnotationBindingDestroyed(tt, "scaleway_flexible_ip.main", keyName, valueName),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
						resource "scaleway_annotations_key" "main" {
							name            = %q
							organization_id = %q
						}

						resource "scaleway_annotations_value" "main" {
							key_id = scaleway_annotations_key.main.id
							name   = %q
						}

						resource "scaleway_flexible_ip" "main" {
							depends_on = [scaleway_annotations_value.main]
						}
					`, keyName, orgID, valueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.main"),
					resource.TestCheckResourceAttrWith("scaleway_flexible_ip.main", "srn", acctest.CheckSRN("flexible-ip", "flexible-ips")),
					testAccCheckFlexibleIPDefaultAnnotationBinding(tt, "scaleway_flexible_ip.main", keyName, valueName),
				),
			},
		},
	})
}

func TestAccFlexibleIP_DefaultAnnotationsUpdate(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	keyName := "tf_test_fip_default_update_key"
	valueAName := "tf_test_fip_default_update_value_a"
	valueBName := "tf_test_fip_default_update_value_b"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckFlexibleIPDestroy(tt),
			annotationstestfuncs.IsAnnotationsKeyDestroyed(tt),
			annotationstestfuncs.IsAnnotationsValueDestroyed(tt),
			testAccCheckFlexibleIPDefaultAnnotationBindingDestroyed(tt, "scaleway_flexible_ip.main", keyName, valueBName),
		),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					tt.Meta.SetDefaultAnnotations(map[string]string{keyName: valueAName})
				},
				Config: fmt.Sprintf(`
						resource "scaleway_annotations_key" "main" {
							name            = %q
							organization_id = %q
						}

						resource "scaleway_annotations_value" "a" {
							key_id = scaleway_annotations_key.main.id
							name   = %q
						}

						resource "scaleway_annotations_value" "b" {
							key_id = scaleway_annotations_key.main.id
							name   = %q
						}

						resource "scaleway_flexible_ip" "main" {
							description = "step 1"
							depends_on  = [scaleway_annotations_value.a, scaleway_annotations_value.b]
						}
					`, keyName, orgID, valueAName, valueBName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.main"),
					testAccCheckFlexibleIPDefaultAnnotationBinding(tt, "scaleway_flexible_ip.main", keyName, valueAName),
					// value B is not part of the wishlist yet, so no binding for it exists.
					testAccCheckFlexibleIPDefaultAnnotationBindingDestroyed(tt, "scaleway_flexible_ip.main", keyName, valueBName),
				),
			},
			{
				PreConfig: func() {
					tt.Meta.SetDefaultAnnotations(map[string]string{keyName: valueBName})
				},
				Config: fmt.Sprintf(`
						resource "scaleway_annotations_key" "main" {
							name            = %q
							organization_id = %q
						}

						resource "scaleway_annotations_value" "a" {
							key_id = scaleway_annotations_key.main.id
							name   = %q
						}

						resource "scaleway_annotations_value" "b" {
							key_id = scaleway_annotations_key.main.id
							name   = %q
						}

						resource "scaleway_flexible_ip" "main" {
							description = "step 2"
							depends_on  = [scaleway_annotations_value.a, scaleway_annotations_value.b]
						}
					`, keyName, orgID, valueAName, valueBName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.main"),
					// The update re-runs the reconcile, replacing the stale value A binding with value B.
					testAccCheckFlexibleIPDefaultAnnotationBinding(tt, "scaleway_flexible_ip.main", keyName, valueBName),
					testAccCheckFlexibleIPDefaultAnnotationBindingDestroyed(tt, "scaleway_flexible_ip.main", keyName, valueAName),
				),
			},
		},
	})
}

func TestAccFlexibleIP_DefaultAnnotationsUserBindingPreserved(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	defaultKeyName := "tf_test_fip_default_preserved_default_key"
	defaultValueName := "tf_test_fip_default_preserved_default_value"
	userKeyName := "tf_test_fip_default_preserved_user_key"
	userValueName := "tf_test_fip_default_preserved_user_value"

	tt.Meta.SetDefaultAnnotations(map[string]string{defaultKeyName: defaultValueName})

	userConfig := func(description string) string {
		return fmt.Sprintf(`
					resource "scaleway_annotations_key" "default" {
						name            = %q
						organization_id = %q
					}

					resource "scaleway_annotations_value" "default" {
						key_id = scaleway_annotations_key.default.id
						name   = %q
					}

					resource "scaleway_annotations_key" "user" {
						name            = %q
						organization_id = %q
					}

					resource "scaleway_annotations_value" "user" {
						key_id = scaleway_annotations_key.user.id
						name   = %q
					}

					resource "scaleway_flexible_ip" "main" {
						description = %q
						depends_on  = [scaleway_annotations_value.default]
					}

					resource "scaleway_annotations_binding" "user" {
						target_srn = scaleway_flexible_ip.main.srn
						value_id   = scaleway_annotations_value.user.id
					}
				`, defaultKeyName, orgID, defaultValueName, userKeyName, orgID, userValueName, description)
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckFlexibleIPDestroy(tt),
			annotationstestfuncs.IsAnnotationsKeyDestroyed(tt),
			annotationstestfuncs.IsAnnotationsValueDestroyed(tt),
			testAccCheckFlexibleIPDefaultAnnotationBindingDestroyed(tt, "scaleway_flexible_ip.main", defaultKeyName, defaultValueName),
		),
		Steps: []resource.TestStep{
			{
				Config: userConfig("step 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.main"),
					testAccCheckFlexibleIPDefaultAnnotationBinding(tt, "scaleway_flexible_ip.main", defaultKeyName, defaultValueName),
					testAccCheckFlexibleIPDefaultAnnotationBinding(tt, "scaleway_flexible_ip.main", userKeyName, userValueName),
				),
			},
			{
				// Forcing an update re-runs the reconcile; the user binding
				// (outside the wishlist) must be left untouched.
				Config: userConfig("step 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.main"),
					testAccCheckFlexibleIPDefaultAnnotationBinding(tt, "scaleway_flexible_ip.main", defaultKeyName, defaultValueName),
					testAccCheckFlexibleIPDefaultAnnotationBinding(tt, "scaleway_flexible_ip.main", userKeyName, userValueName),
				),
			},
		},
	})
}

func TestAccFlexibleIP_CreateAndAttachToBaremetalServer(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayFlexibleIP_CreateAndAttachToBaremetalServer"
	name := "TestAccScalewayFlexibleIP_CreateAndAttachToBaremetalServer"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckFlexibleIPDestroy(tt),
			baremetalchecks.CheckServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_flexible_ip" "base" {
							zone = "fr-par-1"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "fr-par-1"),
				),
			},
			{
				Config: fmt.Sprintf(`
						data "scaleway_baremetal_os" "by_id" {
							zone = "fr-par-1"
							name = "Ubuntu"
							version = "22.04 LTS (Jammy Jellyfish)"						
						}

						data "scaleway_baremetal_offer" "my_offer" {
							name = "%s"
					  		zone = "fr-par-1"
						}				

						resource "scaleway_iam_ssh_key" "main" {
							name 	   = "%s"
							public_key = "%s"
						}

						resource "scaleway_baremetal_server" "base" {
							name        = "%s"
							zone        = "fr-par-1"
							offer       = data.scaleway_baremetal_offer.my_offer.offer_id
							os          = data.scaleway_baremetal_os.by_id.os_id

							ssh_key_ids = [ scaleway_iam_ssh_key.main.id ]
						}
					`, BaremetalOffer, SSHKeyName, SSHKeyFlexibleIP, name),
			},
			{
				Config: fmt.Sprintf(`
						data "scaleway_baremetal_os" "by_id" {
							zone = "fr-par-1"
							name = "Ubuntu"
							version = "22.04 LTS (Jammy Jellyfish)"						
						}

						data "scaleway_baremetal_offer" "my_offer" {
							name = "%s"
					  		zone = "fr-par-1"
						}				

						resource "scaleway_iam_ssh_key" "main" {
							name 	   = "%s"
							public_key = "%s"
						}

						resource "scaleway_baremetal_server" "base" {
							name        = "%s"
							zone        = "fr-par-1"
							offer       = data.scaleway_baremetal_offer.my_offer.offer_id
							os          = data.scaleway_baremetal_os.by_id.os_id

							ssh_key_ids = [ scaleway_iam_ssh_key.main.id ]
						}

						resource "scaleway_flexible_ip" "base" {
							server_id = scaleway_baremetal_server.base.id
							zone = "fr-par-1"
						}
					`, BaremetalOffer, SSHKeyName, SSHKeyFlexibleIP, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					testAccCheckFlexibleIPAttachedToBaremetalServer(tt, "scaleway_flexible_ip.base", "scaleway_baremetal_server.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "fr-par-1"),
				),
			},
		},
	})
}

func TestAccFlexibleIP_AttachAndDetachFromBaremetalServer(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayFlexibleIP_AttachAndDetachFromBaremetalServer"
	name := "TestAccScalewayFlexibleIP_AttachAndDetachFromBaremetalServer"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckFlexibleIPDestroy(tt),
			baremetalchecks.CheckServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_flexible_ip" "base" {
							zone = "fr-par-1"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "fr-par-1"),
				),
			},
			{
				Config: fmt.Sprintf(`
						data "scaleway_baremetal_os" "by_id" {
							zone = "fr-par-1"
							name = "Ubuntu"
							version = "22.04 LTS (Jammy Jellyfish)"						
						}

						data "scaleway_baremetal_offer" "my_offer" {
							name = "%s"
					  		zone = "fr-par-1"
						}		

						resource "scaleway_iam_ssh_key" "main" {
							name 	   = "%s"
							public_key = "%s"
						}

						resource "scaleway_baremetal_server" "base" {
							name        = "%s"
							zone        = "fr-par-1"
							offer       = data.scaleway_baremetal_offer.my_offer.offer_id
							os          = data.scaleway_baremetal_os.by_id.os_id

							ssh_key_ids = [ scaleway_iam_ssh_key.main.id ]
						}
					`, BaremetalOffer, SSHKeyName, SSHKeyFlexibleIP, name),
			},
			{
				Config: fmt.Sprintf(`
						data "scaleway_baremetal_os" "by_id" {
							zone = "fr-par-1"
							name = "Ubuntu"
							version = "22.04 LTS (Jammy Jellyfish)"						
						}

						data "scaleway_baremetal_offer" "my_offer" {
							name = "%s"
					  		zone = "fr-par-1"
						}		

						resource "scaleway_iam_ssh_key" "main" {
							name 	   = "%s"
							public_key = "%s"
						}

						resource "scaleway_baremetal_server" "base" {
							name        = "%s"
							zone        = "fr-par-1"
							offer       = data.scaleway_baremetal_offer.my_offer.offer_id
							os          = data.scaleway_baremetal_os.by_id.os_id

							ssh_key_ids = [ scaleway_iam_ssh_key.main.id ]
						}

						resource "scaleway_flexible_ip" "base" {
							server_id = scaleway_baremetal_server.base.id
							zone = "fr-par-1"
						}
					`, BaremetalOffer, SSHKeyName, SSHKeyFlexibleIP, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					testAccCheckFlexibleIPAttachedToBaremetalServer(tt, "scaleway_flexible_ip.base", "scaleway_baremetal_server.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "fr-par-1"),
				),
			},
			{
				Config: `
						resource "scaleway_flexible_ip" "base" {
							zone = "fr-par-1"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "fr-par-1"),
				),
			},
		},
	})
}

func testAccCheckFlexibleIPExists(tt *acctest.TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		fipAPI, zone, ID, err := flexibleip.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = fipAPI.GetFlexibleIP(&flexibleipSDK.GetFlexibleIPRequest{
			FipID: ID,
			Zone:  zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckFlexibleIPDefaultAnnotationBinding(tt *acctest.TestTools, name, keyName, valueName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		srn := rs.Primary.Attributes["srn"]
		if srn == "" {
			return fmt.Errorf("resource %s has no srn", name)
		}

		annotationsAPI := annotationsSDK.NewAPI(tt.Meta.ScwClient())

		bindings, err := annotationsAPI.ListBindings(&annotationsSDK.ListBindingsRequest{
			TargetSrn: &srn,
		}, scw.WithAllPages())
		if err != nil {
			return err
		}

		for _, binding := range bindings.Bindings {
			if binding.Key != nil && binding.Key.Name == keyName &&
				binding.Value != nil && binding.Value.Name == valueName {
				return nil
			}
		}

		return fmt.Errorf("no default annotation binding found for key %q and value %q on %s", keyName, valueName, srn)
	}
}

func testAccCheckFlexibleIPDefaultAnnotationBindingDestroyed(tt *acctest.TestTools, name, keyName, valueName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		srn := rs.Primary.Attributes["srn"]
		if srn == "" {
			return fmt.Errorf("resource %s has no srn", name)
		}

		annotationsAPI := annotationsSDK.NewAPI(tt.Meta.ScwClient())

		bindings, err := annotationsAPI.ListBindings(&annotationsSDK.ListBindingsRequest{
			TargetSrn: &srn,
		}, scw.WithAllPages())
		if err != nil {
			if httperrors.Is404(err) {
				return nil
			}

			return err
		}

		for _, binding := range bindings.Bindings {
			if binding.Key != nil && binding.Key.Name == keyName &&
				binding.Value != nil && binding.Value.Name == valueName {
				return fmt.Errorf("default annotation binding for key %q and value %q still exists on %s", keyName, valueName, srn)
			}
		}

		return nil
	}
}

func testAccCheckFlexibleIPDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_flexible_ip" {
					continue
				}

				fipAPI, zone, id, err := flexibleip.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = fipAPI.GetFlexibleIP(&flexibleipSDK.GetFlexibleIPRequest{
					FipID: id,
					Zone:  zone,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("flexible IP (%s) still exists", rs.Primary.ID))
				// We check for 403 because instance API return 403 for deleted IP
				case httperrors.Is404(err) || httperrors.Is403(err):
					continue
				default:
					return retry.NonRetryableError(err)
				}
			}

			return nil
		})
	}
}

func testAccCheckFlexibleIPAttachedToBaremetalServer(tt *acctest.TestTools, ipResource, serverResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ipState, ok := s.RootModule().Resources[ipResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", ipResource)
		}

		serverState, ok := s.RootModule().Resources[serverResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", serverResource)
		}

		baremetalAPI, zoneID, err := baremetal.NewAPIWithZoneAndID(tt.Meta, ipState.Primary.ID)
		if err != nil {
			return err
		}

		server, err := baremetalAPI.GetServer(&baremetalSDK.GetServerRequest{
			Zone:     zoneID.Zone,
			ServerID: locality.ExpandID(serverState.Primary.ID),
		})
		if err != nil {
			return err
		}

		fipAPI, zone, ID, err := flexibleip.NewAPIWithZoneAndID(tt.Meta, ipState.Primary.ID)
		if err != nil {
			return err
		}

		ip, err := fipAPI.GetFlexibleIP(&flexibleipSDK.GetFlexibleIPRequest{
			FipID: ID,
			Zone:  zone,
		})
		if err != nil {
			return err
		}

		if ip.ServerID == nil || server.ID != *ip.ServerID {
			return fmt.Errorf("IDs should be the same in %s and %s: %v is different than %v", ipResource, serverResource, server.ID, ip.ServerID)
		}

		return nil
	}
}

func testAccCheckFlexibleIPIsIPv6(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		fipAPI, zone, ID, err := flexibleip.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		flexibleIP, err := fipAPI.GetFlexibleIP(&flexibleipSDK.GetFlexibleIPRequest{
			Zone:  zone,
			FipID: ID,
		})
		if err != nil {
			return err
		}

		if len(flexibleIP.IPAddress.IP.To16()) != net.IPv6len {
			return fmt.Errorf("expected an IPv6 address but got: %s", flexibleIP.IPAddress.String())
		}

		return nil
	}
}
