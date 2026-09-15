package tem_test

import (
	"strings"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

// temAccSubdomain returns a stable cassette name in replay mode and a unique
// subdomain when recording so leftover DNS zones do not collide on nightly.
func temAccSubdomain(stable, prefix string) string {
	if *acctest.UpdateCassettes {
		return strings.ToLower(sdkacctest.RandomWithPrefix(prefix))
	}

	return stable
}
