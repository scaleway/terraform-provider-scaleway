package domaintestfuncs

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	domainSDK "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_domain_zone", &resource.Sweeper{
		Name: "scaleway_domain_zone",
		F:    testSweepDNSZone,
		// Revoke TEM domains first when that sweeper is registered (full make sweep).
		Dependencies: []string{"scaleway_tem_domain"},
	})
}

func testSweepDNSZone(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		domainAPI := domainSDK.NewAPI(scwClient)

		req := &domainSDK.ListDNSZonesRequest{}
		if acctest.TestDomain != "" {
			req.Domain = acctest.TestDomain
		}

		logging.L.Debugf("sweeper: deleting leftover test DNS zones")

		zones, err := domainAPI.ListDNSZones(req, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("error listing DNS zones in sweeper: %s", err)

			return nil
		}

		for _, zone := range zones.DNSZones {
			if !acctest.IsTestResource(zone.Subdomain) {
				continue
			}

			dnsZone := zone.Subdomain + "." + zone.Domain

			logging.L.Debugf("sweeper: deleting DNS zone %s", dnsZone)

			_, err := domainAPI.DeleteDNSZone(&domainSDK.DeleteDNSZoneRequest{
				DNSZone:   dnsZone,
				ProjectID: zone.ProjectID,
			})
			if err != nil {
				logging.L.Warningf("error deleting DNS zone %s in sweeper: %s", dnsZone, err)
			}
		}

		return nil
	})
}
