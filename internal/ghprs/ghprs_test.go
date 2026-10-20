package ghprs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// fakeGh writes a `gh` stub that answers `pr list` with the given JSON.
func fakeGh(t *testing.T, jsonOut string) {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\ncat <<'JSON'\n" + jsonOut + "\nJSON\n"
	path := filepath.Join(dir, "gh")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestCollectParsesAndTightensWindow(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell stub")
	}
	now := time.Now()
	fakeGh(t, fmt.Sprintf(`[
	  {"number": 12, "title": "Dark mode", "mergedAt": %q, "url": "https://x/12", "author": {"login": "dev"}},
	  {"number": 9, "title": "Old thing", "mergedAt": %q, "url": "https://x/9", "author": {"login": "dev"}}
	]`, now.Add(-2*time.Hour).UTC().Format(time.RFC3339), now.Add(-72*time.Hour).UTC().Format(time.RFC3339)))

	items, err := Collect(context.Background(), t.TempDir(), 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != "12" || items[0].Title != "Dark mode" {
		t.Fatalf("day-granular gh result not tightened to window: %+v", items)
	}
}
