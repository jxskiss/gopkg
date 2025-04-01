package easy

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestJSONMarshalMapInterfaceInterface(t *testing.T) {
	m := make(map[any]any)
	m[1] = "1"
	m["2"] = 2
	got := JSON(m)
	want := `{"1":"1","2":2}`
	assert.Equal(t, want, got)
}

func TestJSONDisableEscapeHTML(t *testing.T) {
	m := map[string]string{
		"html": "<html></html>",
	}

	stdRet, err := json.Marshal(m)
	assert.Nil(t, err)
	assert.Equal(t, `{"html":"\u003chtml\u003e\u003c/html\u003e"}`, string(stdRet))

	got := JSON(m)
	assert.Equal(t, `{"html":"<html></html>"}`, got)
}

func TestLazyJSON(t *testing.T) {
	var x = &testObject{A: 123, B: "abc"}
	got1 := JSON(x)
	got2 := fmt.Sprintf("%v", LazyJSON(x))
	assert.Equal(t, got1, got2)
}

func TestLasyFunc(t *testing.T) {
	var x = &testObject{A: 123, B: "abc"}
	got1 := Pretty2(x)
	got2 := fmt.Sprint(LazyFunc(x, Pretty2))
	assert.Equal(t, got1, got2)
	got3 := fmt.Sprint(LazyFunc0(func() string { return Pretty2(x) }))
	assert.Equal(t, got1, got3)
}

var prettyTestWant = strings.TrimSpace(`
{
    "1": 123,
    "b": "<html>"
}`)

func TestPretty(t *testing.T) {
	test := map[string]any{
		"1": 123,
		"b": "<html>",
	}
	jsonString := JSON(test)
	assert.Equal(t, `{"1":123,"b":"<html>"}`, jsonString)

	got1 := Pretty(test)
	assert.Equal(t, prettyTestWant, got1)

	got2 := Pretty(jsonString)
	assert.Equal(t, prettyTestWant, got2)

	test3 := []byte("<fff> not a json object")
	got3 := Pretty(test3)
	assert.Equal(t, string(test3), got3)

	test4 := []byte{
		255, 253, 189, 240, 128, 200, 202, 204,
	}
	got4 := Pretty(test4)
	assert.Equal(t, "<pretty: non-printable bytes of length 8>", got4)

	got5 := Pretty2(map[string]any{"1": 123, "b": "<html>"})
	want5 := "{\n  \"1\": 123,\n  \"b\": \"<html>\"\n}"
	assert.Equal(t, want5, got5)
}

var parseJSONRecordsTestData = `
{
    "files": [
        {
            "displayName": "README.md",
            "repoName": "gopkg",
            "refName": "master",
            "path": "README.md",
            "preferredFileType": "readme",
            "tabName": "README",
            "loaded": true,
            "timedOut": false,
            "errorMessage": null,
            "headerInfo": {
                "toc": [
                    {
                        "level": 1,
                        "text": "gopkg",
                        "anchor": "gopkg",
                        "htmlText": "gopkg"
                    },
                    {
                        "level": 2,
                        "text": "Status",
                        "anchor": "status",
                        "htmlText": "Status"
                    },
                    {
                        "level": 2,
                        "text": "Code layout",
                        "anchor": "code-layout",
                        "htmlText": "Code layout"
                    },
                    {
                        "level": 2,
                        "text": "Packages",
                        "anchor": "packages",
                        "htmlText": "Packages"
                    }
                ],
            }
        },
        {
            "displayName": "LICENSE",
            "repoName": "gopkg",
            "refName": "master",
            "path": "LICENSE",
            "preferredFileType": "license",
            "tabName": "License",
            "loaded": true,
            "timedOut": false,
            "errorMessage": null,
            "headerInfo": {
                "toc": [],
            }
        }
    ],
    "processingTime": 31.543533999999998
}`

func TestParseJSONToMaps(t *testing.T) {
	mapping := JSONPathMapping{
		{"DisplayName", "displayName"},
		{"RepoName", "repoName"},
		{"Loaded", "loaded", "bool"},
		{"HeaderInfo", "headerInfo", "map"},
		{"HeaderInfoLevels", `headerInfo.toc.#(anchor="gopkg")#.level`, "array"},
	}

	j := gjson.Parse(parseJSONRecordsTestData).Get("files")
	got := ParseJSONToMaps(j.Array(), mapping)
	assert.Len(t, got, 2)
	assert.Equal(t, "README.md", got[0]["DisplayName"])
	assert.Equal(t, "LICENSE", got[1]["DisplayName"])
	assert.Equal(t, true, got[0]["Loaded"])
	assert.Equal(t, true, got[1]["Loaded"])
	assert.Equal(t, 4, len(got[0]["HeaderInfo"].(map[string]any)["toc"].([]any)))
	assert.Equal(t, 1, len(got[1]["HeaderInfo"].(map[string]any)))
	assert.Equal(t, []any{float64(1)}, got[0]["HeaderInfoLevels"])
	assert.Equal(t, 0, len(got[1]["HeaderInfoLevels"].([]any)))
}

func TestParseJSONRecords(t *testing.T) {
	type HeaderInfo struct {
		Level int    `mapping:"level"`
		Text  string `mapping:"text"`
	}
	type File struct {
		DisplayName        string                    `mapping:"displayName"`
		RepoName           string                    `mapping:"repoName"`
		Loaded             bool                      `mapping:"loaded"`
		HeaderInfo_1       map[string]*HeaderInfo    `mapping:"{\"toc\":headerInfo.toc.0}"`
		HeaderInfo_2       map[string]map[string]any `mapping:"{\"toc\":headerInfo.toc.1}"`
		HeaderInfoTOC_1    []*HeaderInfo             `mapping:"headerInfo.toc"`
		HeaderInfoTOC_2    []map[string]any          `mapping:"headerInfo.toc"`
		HeaderInfoLevels_1 []int                     `mapping:"headerInfo.toc.#(text=\"Code layout\")#.level"`
		HeaderInfoLevels_2 []any                     `mapping:"headerInfo.toc.#(anchor=\"code-layout\")#.level"`
	}

	j := gjson.Parse(parseJSONRecordsTestData).Get("files")

	var got []*File
	err := ParseJSONRecords(&got, j.Array())
	assert.Nil(t, err)
	assert.Len(t, got, 2)

	assert.Equal(t, "README.md", got[0].DisplayName)
	assert.Equal(t, "LICENSE", got[1].DisplayName)
	assert.Equal(t, true, got[0].Loaded)

	assert.Equal(t, 1, len(got[0].HeaderInfo_1))
	assert.Equal(t, HeaderInfo{1, "gopkg"}, *got[0].HeaderInfo_1["toc"])
	assert.Equal(t, 1, len(got[0].HeaderInfo_2))
	assert.Equal(t,
		map[string]any{"level": float64(2), "text": "Status", "anchor": "status", "htmlText": "Status"},
		got[0].HeaderInfo_2["toc"])

	assert.Equal(t, 4, len(got[0].HeaderInfoTOC_1))
	assert.Equal(t, HeaderInfo{1, "gopkg"}, *got[0].HeaderInfoTOC_1[0])
	assert.Equal(t, 4, len(got[0].HeaderInfoTOC_2))
	assert.Equal(t,
		map[string]any{"level": float64(2), "text": "Status", "anchor": "status", "htmlText": "Status"},
		got[0].HeaderInfoTOC_2[1],
	)

	assert.Equal(t, []int{2}, got[0].HeaderInfoLevels_1)
	assert.Equal(t, []any{float64(2)}, got[0].HeaderInfoLevels_2)
}

func TestTestParseJSONRecords_Options(t *testing.T) {
	type HeaderInfo struct {
		Level int    `mapping:"__HeaderInfo_Level"`
		Text  string `mapping:"text"`
	}
	type File struct {
		DisplayName        string `mapping:"displayName"`
		RepoName           string `mapping:"repoName"`
		Loaded             bool   `mapping:"loaded"`
		HeaderInfo_1       map[string]*HeaderInfo
		HeaderInfo_2       map[string]map[string]any `mapping:"__File_HeaderInfo_2"`
		HeaderInfoLevels_1 []int
		HeaderInfoLevels_2 []any `mapping:"__File_HeaderInfoLevels_2"`
	}

	j := gjson.Parse(parseJSONRecordsTestData).Get("files")

	var got []*File
	err := ParseJSONRecords(&got, j.Array(),
		WithDynamicJSONMapping(map[string]string{
			"__HeaderInfo_Level":        "level",
			"HeaderInfo_1":              `{"toc":headerInfo.toc.0}`,
			"__File_HeaderInfo_2":       `{"toc":headerInfo.toc.1}`,
			"HeaderInfoLevels_1":        `headerInfo.toc.#(text="Code layout")#.level`,
			"__File_HeaderInfoLevels_2": fmt.Sprintf("headerInfo.toc.#(anchor=%q)#.level", "code-layout"),
		}))
	assert.Nil(t, err)
	assert.Len(t, got, 2)

	assert.Equal(t, 1, len(got[0].HeaderInfo_1))
	assert.Equal(t, HeaderInfo{1, "gopkg"}, *got[0].HeaderInfo_1["toc"])
	assert.Equal(t, 1, len(got[0].HeaderInfo_2))
	assert.Equal(t,
		map[string]any{"level": float64(2), "text": "Status", "anchor": "status", "htmlText": "Status"},
		got[0].HeaderInfo_2["toc"])
	assert.Equal(t, []int{2}, got[0].HeaderInfoLevels_1)
	assert.Equal(t, []any{float64(2)}, got[0].HeaderInfoLevels_2)
}

func TestParseJSONRecords_Recursive(t *testing.T) {
	type Person struct {
		A          string
		Parent     *Person
		Children1  []*Person `mapping:"Children_1"`
		Children_2 map[string]*Person
	}
	testData := `[
	{
		"A": "test",
		"Parent": {
			"A": "parent"
		},
		"Children_1": [
			{
				"A": "child_1",
				"Parent": {"A": "test_child_1"}
			},
			{
				"A": "child_2",
				"Parent": {"A": "test_child_2"},
				"Children_1": [
					{"A": "child_2_1"},
					{"A": "child_2_2"}
				]
			}
		],
		"Children_2": {
			"child_1": {
				"A": "child_1",
				"Parent": {"A": "test_child_1"}
			},
			"child_2": {
				"A": "child_2",
				"Parent": {"A": "test_child_2"},
				"Children_2": {
					"child_2_1": {"A": "child_2_1"},
					"child_2_2": {"A": "child_2_2"}
				}
			}
		}
	}
]`
	var got []*Person
	err := ParseJSONRecords(&got, gjson.Parse(testData).Array())
	assert.Nil(t, err)
	assert.Len(t, got, 1)

	assert.Equal(t, "test", got[0].A)
	assert.Equal(t, "parent", got[0].Parent.A)
	assert.Equal(t, ([]*Person)(nil), got[0].Parent.Children1)
	assert.Equal(t, (map[string]*Person)(nil), got[0].Parent.Children_2)

	assert.Equal(t, "child_1", got[0].Children1[0].A)
	assert.Equal(t, "test_child_1", got[0].Children1[0].Parent.A)
	assert.Equal(t, ([]*Person)(nil), got[0].Children1[0].Children1)
	assert.Equal(t, (map[string]*Person)(nil), got[0].Children1[0].Children_2)

	assert.Equal(t, 2, len(got[0].Children1))
	assert.Equal(t, "child_2", got[0].Children1[1].A)
	assert.Equal(t, "test_child_2", got[0].Children1[1].Parent.A)
	assert.Equal(t, 2, len(got[0].Children1[1].Children1))
	assert.Equal(t, (map[string]*Person)(nil), got[0].Children1[0].Children_2)

	assert.Equal(t, 2, len(got[0].Children_2))
	assert.Equal(t, "child_1", got[0].Children_2["child_1"].A)
	assert.Equal(t, "test_child_1", got[0].Children_2["child_1"].Parent.A)
	assert.Equal(t, ([]*Person)(nil), got[0].Children_2["child_1"].Children1)
	assert.Equal(t, (map[string]*Person)(nil), got[0].Children_2["child_1"].Children_2)
	assert.Equal(t, "child_2", got[0].Children_2["child_2"].A)
	assert.Equal(t, "test_child_2", got[0].Children_2["child_2"].Parent.A)
	assert.Equal(t, "child_2_1", got[0].Children_2["child_2"].Children_2["child_2_1"].A)
	assert.Equal(t, "child_2_2", got[0].Children_2["child_2"].Children_2["child_2_2"].A)
}
