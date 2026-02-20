package kafka_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	kafkatestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/kafka/testfuncs"
)

func init() {
	kafkatestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
