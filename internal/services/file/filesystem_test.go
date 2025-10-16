package file_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	fileSDK "github.com/scaleway/scaleway-sdk-go/api/file/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/file"
	filetestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/file/testfuncs"
)

func TestAccFileSystem_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	fileSystemName := "TestAccFileSystem_Basic"
	fileSystemNameUpdated := "TestAccFileSystem_BasicUpdate"
	size := int64(100_000_000_000)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             filetestfuncs.CheckFileDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_file_filesystem" "fs" {
						name = "%s"
						size = %d
					}
				`, fileSystemName, size),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(tt, "scaleway_file_filesystem.fs"),
					resource.TestCheckResourceAttr("scaleway_file_filesystem.fs", "name", fileSystemName),
					resource.TestCheckResourceAttr("scaleway_file_filesystem.fs", "size", strconv.FormatInt(size, 10)),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_file_filesystem" "fs" {
						name = "%s"
						size = %d
					}
				`, fileSystemNameUpdated, size),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(tt, "scaleway_file_filesystem.fs"),
					resource.TestCheckResourceAttr("scaleway_file_filesystem.fs", "size", strconv.FormatInt(size, 10)),
				),
			},
		},
	})
}

func TestAccFileSystem_SizeTooSmallFails(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	fileSystemName := "TestAccFileSystem_SizeTooSmallFails"
	size := int64(10_000_000_000)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             filetestfuncs.CheckFileDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_file_filesystem" "fs" {
						name = "%s"
						size = %d
					}
				`, fileSystemName, size),
				ExpectError: regexp.MustCompile("size must be greater or equal to 100000000000"),
			},
		},
	})
}

func TestAccFileSystem_InvalidSizeGranularityFails(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	fileSystemName := "TestAccFileSystem_InvalidSizeGranularityFails"
	size := int64(25_000_000_000)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             filetestfuncs.CheckFileDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_file_filesystem" "fs" {
						name = "%s"
						size = %d
					}
				`, fileSystemName, size),
				ExpectError: regexp.MustCompile("size must be greater or equal to 100000000000"),
			},
		},
	})
}

func testAccCheckFileSystemExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		fileAPI, region, id, err := file.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = fileAPI.GetFileSystem(&fileSDK.GetFileSystemRequest{
			Region:       region,
			FilesystemID: id,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
