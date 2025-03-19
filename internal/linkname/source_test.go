package linkname

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SourceCodeTestCase struct {
	MinVer, MaxVer int
	FileName       string
	Lines          []string
}

func newVer(major, minor, patch int) int {
	return major*1_000_000 + minor*1_000 + patch
}

func parseGoVer(ver string) int {
	ver = strings.TrimPrefix(ver, "go")
	if x := strings.IndexFunc(ver, func(r rune) bool {
		return r != '.' && !unicode.IsDigit(r)
	}); x > 0 {
		ver = ver[:x]
	}
	parts := strings.Split(ver, ".")
	intVer := 0
	multiplier := [3]int{1_000_000, 1_000, 1}
	for i := 0; i < 3 && i < len(parts); i++ {
		x := parts[i]
		a, err := strconv.Atoi(x)
		if err != nil {
			panic(err)
		}
		intVer += a * multiplier[i]
	}
	return intVer
}

func TestParseGoVer(t *testing.T) {
	testCases := []struct {
		str  string
		want int
	}{
		{"go1.19", 1_019_000},
		{"go1.19.13", 1_019_013},
		{"go1.20", 1_020_000},
		{"go1.21.0", 1_021_000},
		{"go1.22rc2", 1_022_000},
	}
	for _, tc := range testCases {
		t.Run(tc.str, func(t *testing.T) {
			assert.Equal(t, tc.want, parseGoVer(tc.str))
		})
	}
}

func TestSourceCode(t *testing.T) {
	markEnv := "TEST_LINKNAME_SOURCE"
	if os.Getenv(markEnv) == "" {
		t.Skipf("env %s not set, skip", markEnv)
	}

	versions := []string{
		"go1.21.0",
		"go1.21.13",
		"go1.22.0",
		"go1.22.10",
		"go1.23.0",
		"go1.23.7",
		"go1.24.1",
		"master",
	}

	fileURLTmpl := "https://raw.githubusercontent.com/golang/go/%s/src/%s"

	for _, goVersion := range versions {
		var intVer int
		if goVersion == "master" {
			intVer = newVer(1, 999, 999)
		} else {
			intVer = parseGoVer(goVersion)
		}

		// reflect source code
		for _, code := range reflectSourceCode {
			if (code.MinVer > 0 && intVer < code.MinVer) ||
				(code.MaxVer > 0 && intVer > code.MaxVer) {
				continue
			}

			testName := fmt.Sprintf("%s / %s", goVersion, code.FileName)
			t.Run(testName, func(t *testing.T) {
				fileURL := fmt.Sprintf(fileURLTmpl, goVersion, code.FileName)
				content, err := getFileContent(fileURL)
				require.Nil(t, err)

				for _, line := range code.Lines {
					rePattern := `(?m)^` + regexp.QuoteMeta(line) + `($|\s*\{.*)`
					re := regexp.MustCompile(rePattern)
					match := re.MatchString(content)
					assert.Truef(t, match, "line= %q", line)
				}
			})
		}

		// runtime source code
		for _, code := range runtimeSourceCode {
			if (code.MinVer > 0 && intVer < code.MinVer) ||
				(code.MaxVer > 0 && intVer > code.MaxVer) {
				continue
			}

			testName := fmt.Sprintf("%s / %s", goVersion, code.FileName)
			t.Run(testName, func(t *testing.T) {
				fileURL := fmt.Sprintf(fileURLTmpl, goVersion, code.FileName)
				content, err := getFileContent(fileURL)
				require.Nil(t, err)

				for _, line := range code.Lines {
					rePattern := `(?m)^` + regexp.QuoteMeta(line) + `($|\s*\{.*)`
					re := regexp.MustCompile(rePattern)
					match := re.MatchString(content)
					assert.Truef(t, match, "line= %q", line)
				}
			})
		}
	}
}

func getFileContent(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed new HTTP request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed get url %v: %w", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed read HTTP response body: %v: %w", url, err)
	}
	return string(body), nil
}
