package blocktestfuncs

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	blockSDK "github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block"
)

func IsSnapshotPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, id, err := block.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetSnapshot(&blockSDK.GetSnapshotRequest{
			SnapshotID: id,
			Zone:       zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func IsVolumePresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, id, err := block.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetVolume(&blockSDK.GetVolumeRequest{
			VolumeID: id,
			Zone:     zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func IsVolumeDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_block_volume" {
				continue
			}

			api, zone, id, err := block.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteVolume(&blockSDK.DeleteVolumeRequest{
				VolumeID: id,
				Zone:     zone,
			})
			if err == nil {
				return fmt.Errorf("block volume (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) && !httperrors.Is410(err) {
				return err
			}
		}

		return nil
	}
}

func IsSnapshotDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_block_snapshot" {
				continue
			}

			api, zone, id, err := block.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteSnapshot(&blockSDK.DeleteSnapshotRequest{
				SnapshotID: id,
				Zone:       zone,
			})
			if err == nil {
				return fmt.Errorf("block snapshot (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}

// MatchAttrPairIgnorePrefix is a custom check function which compares two Terraform configuration
// fields and considers them equals if the expected value contains a region/zone prefix but not the actual value.
// Can be modified if need be in the future.
func MatchAttrPairIgnorePrefix(nameFirst, keyFirst, nameSecond, keySecond string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rsFirst, ok := s.RootModule().Resources[nameFirst]
		if !ok {
			return fmt.Errorf("first resource not found in state: %s", nameFirst)
		}

		valFirst := rsFirst.Primary.Attributes[keyFirst]

		rsSecond, ok := s.RootModule().Resources[nameSecond]
		if !ok {
			return fmt.Errorf("second resource not found in state: %s", nameSecond)
		}

		valSecond := rsSecond.Primary.Attributes[keySecond]

		splitExpected := strings.Split(valFirst, "/")
		splitActual := strings.Split(valSecond, "/")

		if len(splitExpected) == 2 && len(splitActual) == 1 {
			if splitExpected[1] == splitActual[0] {
				return nil
			}
		}

		return fmt.Errorf("expected (prefix-insensitive) %q, but got %q", valFirst, valSecond)
	}
}
