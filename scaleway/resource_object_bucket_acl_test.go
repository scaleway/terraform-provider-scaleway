package scaleway

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayObjectBucketACL_Basic(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	testBucketName := fmt.Sprintf("terraform-test-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "main" {
						name = "%s"
						region = "nl-ams"
					}
				
					resource "scaleway_object_bucket_acl" "main" {
						bucket = scaleway_object_bucket.main.name
						acl = "private"
						region = "nl-ams"
					}
					`, testBucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_object_bucket_acl.main", "bucket", testBucketName),
					resource.TestCheckResourceAttr("scaleway_object_bucket_acl.main", "acl", "private"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "main" {
						name = "%s"
						region = "nl-ams"
					}
				
					resource "scaleway_object_bucket_acl" "main" {
						bucket = scaleway_object_bucket.main.name
						acl = "public-read"
						region = "nl-ams"
					}
					`, testBucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_object_bucket_acl.main", "bucket", testBucketName),
					resource.TestCheckResourceAttr("scaleway_object_bucket_acl.main", "acl", "public-read"),
				),
			},
		},
	})
}

func TestAccScalewayObjectBucketACL_Grantee(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	testBucketName := fmt.Sprintf("terraform-test-%d", time.Now().Unix())
	ownerID := "105bdce1-64c0-48ab-899d-868455867ecf"
	ownerIDChild := "50ab77d5-56bd-4981-a118-4e0fa5309b59"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "main" {
						name = "%[1]s"
					}
				
					resource "scaleway_object_bucket_acl" "main" {
						bucket = scaleway_object_bucket.main.name
						access_control_policy {
						  grant {
							grantee {
								id   = "%[2]s"
								type = "CanonicalUser"
							}
							permission = "FULL_CONTROL"
						  }
						
						  grant {
							grantee {
							  id   = "%[2]s"
							  type = "CanonicalUser"
							}
							permission = "WRITE"
						  }
						
						  owner {
							id = "%[2]s"
						  }
						}
					}
					`, testBucketName, ownerID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_object_bucket_acl.main", "bucket", testBucketName),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "main" {
						name = "%[1]s"
					}
				
					resource "scaleway_object_bucket_acl" "main" {
						bucket = scaleway_object_bucket.main.name
						access_control_policy {
							grant {
								grantee {
								  id   = "%[3]s"
								  type = "CanonicalUser"
								}
								permission = "FULL_CONTROL"
							}
						
							owner {
								id = "%[2]s"
							}
					}
				}
					`, testBucketName, ownerID, ownerIDChild),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_object_bucket_acl.main", "bucket", testBucketName),
				),
			},
			{
				ResourceName:      "scaleway_object_bucket_acl.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccScalewayObjectBucketACL_GranteeWithOwner(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	testBucketName := fmt.Sprintf("terraform-test-%d", time.Now().Unix())
	ownerID := "105bdce1-64c0-48ab-899d-868455867ecf"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "main" {
						name = "%[1]s"
					}
				
					resource "scaleway_object_bucket_acl" "main" {
						bucket = scaleway_object_bucket.main.name
						expected_bucket_owner = "%[2]s"
						access_control_policy {
						  grant {
							grantee {
								id   = "%[2]s"
								type = "CanonicalUser"
							}
							permission = "FULL_CONTROL"
						  }
						
						  grant {
							grantee {
							  id   = "%[2]s"
							  type = "CanonicalUser"
							}
							permission = "WRITE"
						  }
						
						  owner {
							id = "%[2]s"
						  }
						}
					}
					`, testBucketName, ownerID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_object_bucket_acl.main", "bucket", testBucketName),
				),
			},
		},
	})
}
