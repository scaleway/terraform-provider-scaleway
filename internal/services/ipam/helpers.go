package ipam

import (
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ipam "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultIPRetryInterval     = 5 * time.Second
	defaultIPReverseDNSTimeout = 10 * time.Minute
)

// newAPIWithRegion returns a new ipam API and the region
func newAPIWithRegion(d *schema.ResourceData, m interface{}) (*ipam.API, scw.Region, error) {
	ipamAPI := ipam.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return ipamAPI, region, nil
}

// NewAPIWithRegionAndID returns a new ipam API with locality and ID extracted from the state
func NewAPIWithRegionAndID(m interface{}, id string) (*ipam.API, scw.Region, string, error) {
	ipamAPI := ipam.NewAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return ipamAPI, region, ID, err
}

func diffSuppressFuncStandaloneIPandCIDR(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	oldIP, oldNet, errOld := net.ParseCIDR(oldValue)
	if errOld != nil {
		oldIP = net.ParseIP(oldValue)
	}

	newIP, newNet, errNew := net.ParseCIDR(newValue)
	if errNew != nil {
		newIP = net.ParseIP(newValue)
	}

	if oldIP != nil && newIP != nil && oldIP.Equal(newIP) {
		return true
	}

	if oldNet != nil && newIP != nil && oldNet.Contains(newIP) {
		return true
	}

	if newNet != nil && oldIP != nil && newNet.Contains(oldIP) {
		return true
	}

	return false
}
