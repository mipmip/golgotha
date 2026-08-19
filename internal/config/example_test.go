package config

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// exampleAllowlist holds config keys deliberately not documented in
// config.example.yaml. It is intentionally empty: every schema field should be
// shown (commented or not). Add a key here only with a clear reason.
var exampleAllowlist = map[string]struct{}{}

// findRepoRoot walks up from the test's working directory (the package dir under
// `go test`) until it finds go.mod, returning the repo root. This locates the
// root-level config.example.yaml without embedding or symlinks.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found walking up from test dir")
		}
		dir = parent
	}
}

func exampleConfigPath(t *testing.T) string {
	return filepath.Join(findRepoRoot(t), "config.example.yaml")
}

// TestExampleConfigValid is level (a): the documented example must load and
// validate through the standard loader (which also rejects unknown/typo'd keys
// via strict decoding).
func TestExampleConfigValid(t *testing.T) {
	if _, err := LoadFrom(exampleConfigPath(t)); err != nil {
		t.Fatalf("config.example.yaml must load+validate cleanly: %v", err)
	}
}

// collectYAMLKeys recursively gathers every yaml tag name reachable from t,
// descending into structs, pointers and slice/map element types.
func collectYAMLKeys(t reflect.Type, into map[string]struct{}, seen map[reflect.Type]bool) {
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Array || t.Kind() == reflect.Map {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}
	if seen[t] {
		return
	}
	seen[t] = true

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("yaml")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		if name != "" {
			into[name] = struct{}{}
		}
		collectYAMLKeys(f.Type, into, seen)
	}
}

// TestExampleConfigComplete is level (b): every schema field must appear as a
// key (commented `# key:` counts) in the example, so optional-field drift is
// caught even when the example stays valid.
func TestExampleConfigComplete(t *testing.T) {
	keys := map[string]struct{}{}
	collectYAMLKeys(reflect.TypeOf(Config{}), keys, map[reflect.Type]bool{})
	if len(keys) == 0 {
		t.Fatal("collected no yaml keys from Config; reflection broken")
	}

	data, err := os.ReadFile(exampleConfigPath(t))
	if err != nil {
		t.Fatalf("read example: %v", err)
	}
	text := string(data)

	var missing []string
	for key := range keys {
		if _, ok := exampleAllowlist[key]; ok {
			continue
		}
		// Line-anchored: optional leading whitespace, optional comment marker,
		// optional YAML list dash, then `key:`. Avoids substring/prose false
		// positives while allowing `- name:` list items and `# key:` comments.
		re := regexp.MustCompile(`(?m)^\s*#?\s*(-\s*)?` + regexp.QuoteMeta(key) + `:`)
		if !re.MatchString(text) {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("config.example.yaml is missing documentation for schema keys: %v\n"+
			"document each as `key:` (a commented `# key:` counts), or add to exampleAllowlist with a reason", missing)
	}
}
