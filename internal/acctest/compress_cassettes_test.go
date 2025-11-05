package acctest_test

import (
	"strings"
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestAccCassettes_Compressed(t *testing.T) {
	paths, err := getTestFiles(false)
	require.NoError(t, err)

	var g errgroup.Group

	for path := range paths {
		g.Go(func() error {
			var (
				report         acctest.CompressReport
				errCompression error
			)

			folder := strings.Split(path, "/")[2]
			if acctest.FolderUsesVCRv4(folder) {
				report, errCompression = acctest.CompressCassetteV4(path)
			} else {
				report, errCompression = acctest.CompressCassetteV3(path)
			}

			require.NoError(t, errCompression)
			require.Zero(t, report.SkippedInteraction, "Issue with cassette: %s", report.Path)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		t.Errorf("error: %s", err)
	}
}
