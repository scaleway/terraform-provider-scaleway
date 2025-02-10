package acctest

import (
	"flag"
	"os"
	"regexp"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

var (
	TestDomain     = ""
	TestDomainZone = ""
	// prevent using production domain for testing
	reservedDomains = []*regexp.Regexp{
		regexp.MustCompile(`.*iliad.*`),
		regexp.MustCompile(`.*\.free\..*`),
		regexp.MustCompile(`.*\.online\..*`),
		regexp.MustCompile(`.*scaleway\..*`),
		regexp.MustCompile(`.*dedibox.*`),
	}
)

func init() {
	testDomainPtr := flag.String("test-domain", os.Getenv("TF_TEST_DOMAIN"), "Test domain")
	if testDomainPtr != nil && *testDomainPtr != "" {
		TestDomain = *testDomainPtr
	} else {
		logging.L.Infof("environment variable TF_TEST_DOMAIN is required")

		return
	}

	// check if the test domain is not a Scaleway reserved domain
	isReserved := false

	for _, reservedDomain := range reservedDomains {
		if reservedDomain.MatchString(TestDomain) {
			isReserved = true

			break
		}
	}

	if isReserved {
		logging.L.Warningf("TF_TEST_DOMAIN cannot be a Scaleway required domain. Please use another one.")

		return
	}

	logging.L.Infof("start domain record test with domain: %s", TestDomain)

	testDomainZonePtr := flag.String("test-domain-zone", os.Getenv("TF_TEST_DOMAIN_ZONE"), "Test domain zone")
	if testDomainZonePtr != nil && *testDomainZonePtr != "" {
		TestDomainZone = *testDomainZonePtr
	} else {
		logging.L.Infof("environment variable TF_TEST_DOMAIN_ZONE is required")

		return
	}

	logging.L.Infof("start domain record test with domain zone: %s", TestDomainZone)
}
