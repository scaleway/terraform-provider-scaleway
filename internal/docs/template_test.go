package docs_test

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var gotypeRE = regexp.MustCompile(`\{\{.*gotype:.*}}`)

func TestGoTypeDefined(t *testing.T) {
	err := filepath.WalkDir("../../templates/resources", func(path string, _ fs.DirEntry, _ error) error {
		if isTemplate := strings.Contains(path, "tmpl"); isTemplate {
			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("cannot open %s", path)
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			if !scanner.Scan() {
				t.Logf("❌ %s: file is empty", path)
				t.Fail()
			}
			firstLine := scanner.Text()
			if gotypeRE.MatchString(firstLine) {
				return nil
			}
			t.Logf("gotype missing at top of file: %s", path)
			t.Fail()
		}

		return nil
	})
	if err != nil {
		return
	}
}
