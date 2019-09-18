package scaleway

import (
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	account "github.com/scaleway/scaleway-sdk-go/api/account/v2alpha1"
	baremetal "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	BaremetalServerWaitForTimeout   = 60 * time.Minute
	BaremetalServerRetryFuncTimeout = BaremetalServerWaitForTimeout + time.Minute // some RetryFunc are calling a WaitFor
)

var BaremetalServerResourceTimeout = BaremetalServerRetryFuncTimeout + time.Minute

// getInstanceAPIWithZone returns a new baremetal API and the zone for a Create request
func getBaremetalAPIWithZone(d *schema.ResourceData, m interface{}) (*baremetal.API, scw.Zone, error) {
	meta := m.(*Meta)
	baremetalAPI := baremetal.NewAPI(meta.scwClient)

	zone, err := getZone(d, meta)
	return baremetalAPI, zone, err
}

// getInstanceAPIWithZoneAndID returns an baremetal API with zone and ID extracted from the state
func getBaremetalAPIWithZoneAndID(m interface{}, id string) (*baremetal.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	baremetalAPI := baremetal.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	return baremetalAPI, zone, ID, err
}

// getAllUserSSHKeyIDs returns a slice containing all user SSH keys.
func getAllUserSSHKeyIDs(m interface{}) ([]string, error) {
	sshKeysResponse, err := getAccountAPI(m).ListSSHKeys(&account.ListSSHKeysRequest{})
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, sshKey := range sshKeysResponse.SSHKeys {
		ids = append(ids, sshKey.ID)
	}
	return ids, nil
}
