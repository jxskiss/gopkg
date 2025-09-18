package easy

import (
	"os"
	"path/filepath"
	"testing"
)

func makeGlobTestDir(t *testing.T) {
	t.Helper()
	path := "./testdata/a/b/c.d/e.f"
	err := os.MkdirAll(path, 0755)
	if err != nil {
		t.Fatalf("os.MkdirAll: %v", err)
	}
}

func removeGlobTestDir() {
	_ = os.RemoveAll("./testdata/a")
}

func TestGlob(t *testing.T) {
	makeGlobTestDir(t)
	defer removeGlobTestDir()

	// test pass-through to vanilla path/filepath
	{
		matches, err := Glob("./*/*/*/*.d")
		if err != nil {
			t.Fatalf("Glob: %s", err)
		}
		if len(matches) != 1 {
			t.Fatalf("got %d matches, expected 1", len(matches))
		}
		expected := filepath.Clean("testdata/a/b/c.d")
		if matches[0] != expected {
			t.Fatalf("matched [%s], expected [%s]", matches[0], expected)
		}
	}

	// test a single double-star
	{
		matches, err := Glob("./**/*.f")
		if err != nil {
			t.Fatalf("Glob: %s", err)
		}
		if len(matches) != 1 {
			t.Fatalf("got %d matches, expected 1", len(matches))
		}
		expected := filepath.Clean("testdata/a/b/c.d/e.f")
		if matches[0] != expected {
			t.Fatalf("matched [%s], expected [%s]", matches[0], expected)
		}
	}

	// test a single double-star
	{
		matches, err := Glob("./testdata/**/*.*")
		if err != nil {
			t.Fatalf("Glob: %s", err)
		}
		if len(matches) != 2 {
			t.Fatalf("got %d matches, expected 2", len(matches))
		}
		expected := []string{
			filepath.Clean("testdata/a/b/c.d"),
			filepath.Clean("testdata/a/b/c.d/e.f"),
		}
		for i, match := range matches {
			if match != expected[i] {
				t.Fatalf("matched [%s], expected [%s]", match, expected[i])
			}
		}
	}

	// test two double-stars
	{
		matches, err := Glob("./**/b/**/*.f")
		if err != nil {
			t.Fatalf("Glob: %s", err)
		}
		if len(matches) != 1 {
			t.Fatalf("got %d matches, expected 1", len(matches))
		}
		expected := filepath.Clean("testdata/a/b/c.d/e.f")
		if matches[0] != expected {
			t.Fatalf("matched [%s], expected [%s]", matches[0], expected)
		}
	}
}
