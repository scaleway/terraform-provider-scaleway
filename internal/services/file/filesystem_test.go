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

func TestFileSystemSizeInGBRangeValidation(t *testing.T) {
	sizeSchema := file.ResourceFileSystem().SchemaFunc()["size_in_gb"]

	tests := []struct {
		name      string
		sizeInGB  int
		wantError bool
	}{
		{
			name:     "minimum",
			sizeInGB: 100,
		},
		{
			name:     "five terabytes",
			sizeInGB: 5000,
		},
		{
			name:     "maximum",
			sizeInGB: 10000,
		},
		{
			name:      "below minimum",
			sizeInGB:  99,
			wantError: true,
		},
		{
			name:      "above maximum",
			sizeInGB:  10001,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := sizeSchema.ValidateFunc(tt.sizeInGB, "size_in_gb")
			if gotError := len(errors) > 0; gotError != tt.wantError {
				t.Fatalf("size_in_gb validation returned errors %v, wantError %t", errors, tt.wantError)
			}
		})
	}
}

func TestAccFileSystem_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	fileSystemName := "TestAccFileSystem_Basic"
	fileSystemNameUpdated := "TestAccFileSystem_BasicUpdate"
	sizeInGB := 100

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             filetestfuncs.CheckFileDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_file_filesystem" "fs" {
						name = "%s"
						size_in_gb = %d
					}
				`, fileSystemName, sizeInGB),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(tt, "scaleway_file_filesystem.fs"),
					resource.TestCheckResourceAttr("scaleway_file_filesystem.fs", "name", fileSystemName),
					resource.TestCheckResourceAttr("scaleway_file_filesystem.fs", "size_in_gb", strconv.Itoa(sizeInGB)),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_file_filesystem" "fs" {
						name = "%s"
						size_in_gb = %d
					}
				`, fileSystemNameUpdated, sizeInGB),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(tt, "scaleway_file_filesystem.fs"),
					resource.TestCheckResourceAttr("scaleway_file_filesystem.fs", "size_in_gb", strconv.Itoa(sizeInGB)),
				),
			},
			{
				ResourceName:      "scaleway_file_filesystem.fs",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFileSystem_SizeTooSmallFails(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	fileSystemName := "TestAccFileSystem_SizeTooSmallFails"
	sizeInGB := 10

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             filetestfuncs.CheckFileDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_file_filesystem" "fs" {
						name = "%s"
						size_in_gb = %d
					}
				`, fileSystemName, sizeInGB),
				ExpectError: regexp.MustCompile(`expected size_in_gb to be in the range \(100 - 10000\), got 10`),
			},
		},
	})
}

func TestAccFileSystem_InvalidSizeGranularityFails(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	fileSystemName := "TestAccFileSystem_InvalidSizeGranularityFails"
	sizeInGB := 250

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             filetestfuncs.CheckFileDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_file_filesystem" "fs" {
						name = "%s"
						size_in_gb = %d
					}
				`, fileSystemName, sizeInGB),
				ExpectError: regexp.MustCompile("size does not respect constraint, size must be a multiple of 100000000000"),
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
