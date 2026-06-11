package extract_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Priority merchants must have scrubbed on-disk fixtures (M1.4 / TESTING.md).
func TestMerchantFixtures_OnDisk(t *testing.T) {
	root := filepath.Join("..", "..", "test", "testdata", "email", "merchants")
	want := []string{"amazon", "screwfix", "toolstation", "bandq", "shell", "bp", "esso", "octopus", "edf"}
	for _, merchant := range want {
		dir := filepath.Join(root, merchant)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("%s: %v", merchant, err)
		}
		if len(entries) < 2 {
			t.Fatalf("%s: want ≥2 fixture files, got %d", merchant, len(entries))
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := strings.ToLower(e.Name())
			if !strings.HasSuffix(name, ".txt") && !strings.HasSuffix(name, ".eml") {
				continue
			}
			data, readErr := os.ReadFile(filepath.Join(dir, e.Name()))
			if readErr != nil {
				t.Fatal(readErr)
			}
			if len(data) == 0 {
				t.Fatalf("%s/%s: empty fixture", merchant, e.Name())
			}
		}
	}
}
