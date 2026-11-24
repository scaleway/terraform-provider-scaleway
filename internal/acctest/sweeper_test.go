package acctest_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/stretchr/testify/assert"
)

func TestSweepRegionsInvalidRegions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		wantErr string
		regions []scw.Region
	}{
		{name: "empty list", wantErr: "cannot sweep an empty region list"},
		{name: "region without zones", regions: []scw.Region{"all_regions"}, wantErr: `cannot sweep region "all_regions": no zones found`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := acctest.SweepRegions(tc.regions, func(*scw.Client, scw.Region) error {
				t.Fatal("unexpected sweep callback")

				return nil
			})
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestIsTestResource(t *testing.T) {
	assert.True(t, acctest.IsTestResource("tf_tests_mnq_sqs_queue_default_project"))
}
