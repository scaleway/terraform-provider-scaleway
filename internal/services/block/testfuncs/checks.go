package blocktestfuncs

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	blockSDK "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block"
)

var (
	ErrBlockResourceNotFound = errors.New("resource not found")
	ErrBlockVolumeStillExists = errors.New("block volume still exists")
	ErrBlockSnapshotStillExists = errors.New("block snapshot still exists")
)

func IsSnapshotPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("%w: %s", ErrBlockResourceNotFound, n)
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
			return fmt.Errorf("%w: %s", ErrBlockResourceNotFound, n)
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
				return fmt.Errorf("%w (%s)", ErrBlockVolumeStillExists, rs.Primary.ID)
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
				return fmt.Errorf("%w (%s)", ErrBlockSnapshotStillExists, rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
