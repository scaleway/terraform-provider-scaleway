package temtestfuncs

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	domainSDK "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	temSDK "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

const temTestDNSDomain = "scaleway-terraform.com"

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_tem_domain", &resource.Sweeper{
		Name: "scaleway_tem_domain",
		F:    testSweepDomain,
	})
}

func testSweepDomain(_ string) error {
	err := acctest.SweepRegions([]scw.Region{scw.RegionFrPar, scw.RegionNlAms}, func(scwClient *scw.Client, region scw.Region) error {
		temAPI := temSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: revoking the tem domains in (%s)", region)

		listDomains, err := temAPI.ListDomains(&temSDK.ListDomainsRequest{Region: region}, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("error listing domains in (%s) in sweeper: %s", region, err)

			return nil
		}

		for _, ns := range listDomains.Domains {
			if !acctest.IsTestResource(ns.Name) {
				continue
			}

			_, err := temAPI.RevokeDomain(&temSDK.RevokeDomainRequest{
				DomainID: ns.ID,
				Region:   region,
			})
			if err != nil {
				logging.L.Warningf("error revoking domain in sweeper: %s", err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return sweepTEMTestDNSZones()
}

func sweepTEMTestDNSZones() error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		domainAPI := domainSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: deleting leftover TEM test DNS zones on %s", temTestDNSDomain)

		zones, err := domainAPI.ListDNSZones(&domainSDK.ListDNSZonesRequest{
			Domain: temTestDNSDomain,
		}, scw.WithAllPages())
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
