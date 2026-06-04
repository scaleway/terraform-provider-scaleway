package interlink

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	interlink "github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func interlinkAPIWithRegion(d *schema.ResourceData, m any) (*interlink.API, scw.Region, error) {
	interlinkAPI := interlink.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return interlinkAPI, region, err
}

func NewAPIWithRegionAndID(m any, id string) (*interlink.API, scw.Region, string, error) {
	interlinkAPI := interlink.NewAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return interlinkAPI, region, ID, err
}
