package acctest_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"testing"
)

func TestAccCassettes_Compressed(t *testing.T) {
	paths, err := getTestFiles()
	require.NoError(t, err)

	for path := range paths {
		c, err := cassette.Load(path)
		require.NoError(t, err)
		assert.NoError(t, checkCassetteCompressed(t, c))
	}
}

func checkCassetteCompressed(t *testing.T, c *cassette.Cassette) error {
	for i := 0; i < len(c.Interactions)-1; i++ {
		if c.Interactions[i].Response.Body == c.Interactions[i+1].Response.Body {
			t.Errorf("cassette %s contains two consecutive interactions with the same response body. Interaction %d and %d have the same body", c.Name, i, i+1)
		}
	}
	return nil
}
