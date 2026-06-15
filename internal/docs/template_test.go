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

var (
	gotypeRE      = regexp.MustCompile(`{{\s*\-?\s*/\*\s*\.?gotype:.*\*/\s*\-?\s*}}`)
	subcategoryRE = regexp.MustCompile(`subcategory:\s*"([^"]+)"`)
)

func TestGoTypeDefined(t *testing.T) {
	err := filepath.WalkDir("../../templates/resources", func(path string, _ fs.DirEntry, _ error) error {
		if isTemplate := strings.Contains(path, "tmpl"); isTemplate {
			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("cannot open %s", path)
			}
			defer func(f *os.File) {
				err := f.Close()
				if err != nil {
					t.Fatal(err.Error())
				}
			}(f)

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

func TestSubcategoryHomogeneity(t *testing.T) {
	subcategoryMap := make(map[string]map[string]bool)

	docRoots := []string{"../../docs", "../../templates"}

	for _, root := range docRoots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if strings.Contains(path, "/guides/") {
				if d.IsDir() {
					return filepath.SkipDir
				}

				return nil
			}

			if (strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".tmpl")) && !d.IsDir() {
				if strings.HasSuffix(path, "index.md") || strings.HasSuffix(path, "index.md.tmpl") {
					return nil
				}

				content, err := os.ReadFile(path)
				if err != nil {
					t.Logf("❌ Cannot read file %s: %v", path, err)

					return nil
				}

				matches := subcategoryRE.FindStringSubmatch(string(content))
				if len(matches) == 0 {
					t.Logf("❌ No subcategory defined in file: %s", path)
					t.Fail()

					return nil
				}

				subcategory := matches[1]

				singularForm := subcategory
				if strings.HasSuffix(subcategory, "s") && !strings.Contains(subcategory, " ") {
					singularForm = strings.TrimSuffix(subcategory, "s")
				}

				if _, exists := subcategoryMap[singularForm]; !exists {
					subcategoryMap[singularForm] = make(map[string]bool)
				}

				subcategoryMap[singularForm][subcategory] = true
			}

			return nil
		})
		if err != nil {
			t.Logf("❌ Error walking directory %s: %v", root, err)
		}
	}

	conflictsFound := false

	for singularForm, variants := range subcategoryMap {
		if len(variants) > 1 {
			conflictsFound = true

			var variantList []string
			for variant := range variants {
				variantList = append(variantList, "\""+variant+"\"")
			}

			t.Logf("❌ Category conflict detected: singular form '%s' has variants: %s", singularForm, strings.Join(variantList, ", "))
		}
	}

	if conflictsFound {
		t.Fail()
	}
}
