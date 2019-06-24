package scaleway

import (
	"time"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/utils"
)

const (
	ServerStateStopped = "stopped"
	ServerStateStarted = "started"
	ServerStateStandby = "standby"

	ServerWaitForTimeout = 10 * time.Minute
)

// getInstanceAPIWithZone returns a new instance API and the zone for a Create request
func getInstanceAPIWithZone(d *schema.ResourceData, m interface{}) (*instance.API, utils.Zone, error) {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, err := getZone(d, meta)
	return instanceApi, zone, err
}

// getInstanceAPIWithZoneAndID returns an instance API with zone and ID extracted from the state
func getInstanceAPIWithZoneAndID(m interface{}, id string) (*instance.API, utils.Zone, string, error) {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	return instanceApi, zone, ID, err
}

// expandRootVolume expands the current root volume or returns the default root volume
func expandRootVolume(v interface{}) map[string]interface{} {
	flattenVolume := map[string]interface{}{}

	vs, ok := v.([]map[string]interface{})
	if ok && len(vs) > 0 {
		flattenVolume = vs[0]
	}

	if _, exist := flattenVolume["delete_on_termination"]; !exist {
		flattenVolume["delete_on_termination"] = true // default value does not work on list
	}

	return flattenVolume
}

func schemaSetUserData(v interface{}) int {
	userData := v.(map[string]interface{})
	return hashcode.String(userData["key"].(string) + userData["value"].(string))
}
