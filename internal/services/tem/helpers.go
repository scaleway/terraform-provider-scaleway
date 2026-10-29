package tem

import (
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	DefaultDomainTimeout           = 5 * time.Minute
	DefaultDomainCreateTimeout     = 60 * time.Minute
	defaultDomainValidationTimeout = 60 * time.Minute
	defaultDomainRetryInterval     = 15 * time.Second
)

// temAPIWithRegion returns a new Tem API and the region for a Create request
func temAPIWithRegion(d *schema.ResourceData, m any) (*tem.API, scw.Region, error) {
	api := tem.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// NewAPIWithRegionAndID returns a Tem API with zone and ID extracted from the state
func NewAPIWithRegionAndID(m any, id string) (*tem.API, scw.Region, string, error) {
	api := tem.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}

// FlattenMXRecordValue splits an API MX value ("PRIORITY EXCHANGE") into parts
// compatible with scaleway_domain_record data and priority attributes.
func FlattenMXRecordValue(value string) (priority int, exchange string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, ""
	}

	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 {
		return 0, value
	}

	prio, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, value
	}

	return prio, parts[1]
}
