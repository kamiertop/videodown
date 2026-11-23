package icons

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestCatalogResolves parses every catalog entry.
func TestCatalogResolves(t *testing.T) {
	for name := range catalog {
		if icon, ok := Get(name); !ok || icon == nil {
			t.Errorf("catalog entry %q does not resolve", name)
		}
	}
}

// iconLiteralPatterns match the ways source files name icons: direct
// icons.Must/Get literals, the trailing literal of the shared empty and
// unavailable helpers, and the names arrays that feed platform rails.
var (
	mustPattern  = regexp.MustCompile(`icons\.(?:Must|Get)\("([a-z0-9-]+)"\)`)
	emptyPattern = regexp.MustCompile(`(?:empty|unavailable)\((?:c(?:\.Key\("[^"]*"\))?, )?[^,]+, "([a-z0-9-]+)"\)`)
	namesPattern = regexp.MustCompile(`names := \[\]string\{([^}]*)\}`)
	quoted       = regexp.MustCompile(`"([a-z0-9-]+)"`)
)

// catalogHas reports whether the catalog maps name.
func catalogHas(name string) bool {
	_, ok := catalog[name]
	return ok
}

// TestSourceNamesAreInCatalog walks the GUI sources so a renamed or new icon
// name fails here instead of panicking at runtime.
func TestSourceNamesAreInCatalog(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	var missing []string
	seen := map[string]bool{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		src := string(data)
		if !strings.Contains(src, "icons.Must") && !strings.Contains(src, "icons.Get") && !strings.Contains(src, "empty(c") && !strings.Contains(src, "unavailable(c") {
			return nil
		}
		for _, re := range []*regexp.Regexp{mustPattern, emptyPattern} {
			for _, match := range re.FindAllStringSubmatch(src, -1) {
				if name := match[1]; !catalogHas(name) {
					missing = append(missing, name+" ("+path+")")
					seen[name] = true
				}
			}
		}
		// names arrays feed nav rails looked up with icons.Must(names[i]) or
		// icons.Get(names[i]) in the same file.
		if strings.Contains(src, "names[") {
			for _, match := range namesPattern.FindAllStringSubmatch(src, -1) {
				for _, name := range quoted.FindAllStringSubmatch(match[1], -1) {
					if !catalogHas(name[1]) {
						missing = append(missing, name[1]+" ("+path+")")
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) > 0 {
		t.Fatalf("icon names missing from the catalog:\n%s", strings.Join(missing, "\n"))
	}
}
