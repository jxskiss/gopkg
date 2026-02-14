package json

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHumanFriendlyIndentation(t *testing.T) {
	data := `[
    {
        "author": {
            "avatar_url": "https://github.com/images/error/octocat_happy.gif",
            "events_url": "https://api.github.com/users/octocat/events{/privacy}",
            "followers_url": "https://api.github.com/users/octocat/followers",
            "following_url": "https://api.github.com/users/octocat/following{/other_user}",
            "gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
            "gravatar_id": "",
            "html_url": "https://github.com/octocat",
            "id": 1,
            "login": "octocat",
            "node_id": "MDQ6VXNlcjE=",
            "organizations_url": "https://api.github.com/users/octocat/orgs",
            "received_events_url": "https://api.github.com/users/octocat/received_events",
            "repos_url": "https://api.github.com/users/octocat/repos",
            "site_admin": false,
            "starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
            "subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
            "type": "User",
            "url": "https://api.github.com/users/octocat"
        },
        "comments_url": "https://api.github.com/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e/comments",
        "commit": {
            "author": {
                "date": "2011-04-14T16:00:49Z",
                "email": "support@github.com",
                "name": "Monalisa Octocat"
            },
            "comment_count": 0,
            "committer": {
                "date": "2011-04-14T16:00:49Z",
                "email": "support@github.com",
                "name": "Monalisa Octocat"
            },
            "message": "Fix all the bugs",
            "tree": {
                "sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
                "url": "https://api.github.com/repos/octocat/Hello-World/tree/6dcb09b5b57875f334f61aebed695e2e4193db5e"
            },
            "url": "https://api.github.com/repos/octocat/Hello-World/git/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e",
            "verification": {
                "payload": null,
                "reason": "unsigned",
                "signature": null,
                "verified": false
            }
        },
        "committer": {
            "avatar_url": "https://github.com/images/error/octocat_happy.gif",
            "events_url": "https://api.github.com/users/octocat/events{/privacy}",
            "followers_url": "https://api.github.com/users/octocat/followers",
            "following_url": "https://api.github.com/users/octocat/following{/other_user}",
            "gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
            "gravatar_id": "",
            "html_url": "https://github.com/octocat",
            "id": 1,
            "login": "octocat",
            "node_id": "MDQ6VXNlcjE=",
            "organizations_url": "https://api.github.com/users/octocat/orgs",
            "received_events_url": "https://api.github.com/users/octocat/received_events",
            "repos_url": "https://api.github.com/users/octocat/repos",
            "site_admin": false,
            "starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
            "subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
            "type": "User",
            "url": "https://api.github.com/users/octocat"
        },
        "html_url": "https://github.com/octocat/Hello-World/commit/6dcb09b5b57875f334f61aebed695e2e4193db5e",
        "node_id": "MDY6Q29tbWl0NmRjYjA5YjViNTc4NzVmMzM0ZjYxYWViZWQ2OTVlMmU0MTkzZGI1ZQ==",
        "parents": [
            {
                "sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
                "url": "https://api.github.com/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e"
            }
        ],
        "sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
        "url": "https://api.github.com/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e"
    }
]
`
	var m []any
	err := UnmarshalFromString(data, &m)
	assert.Nil(t, err)

	got, err := HumanFriendly.MarshalIndent(m, "", "    ")
	assert.Nil(t, err)
	assert.True(t, bytes.HasPrefix(got, []byte("[\n    {\n        \"author\": {\n            \"avatar_url\"")))
	assert.True(t, bytes.Contains(got, []byte(",\n        \"comments_url\": \"")))
	assert.True(t, bytes.Contains(got, []byte("        \"commit\": {\n            \"author\": {\n                \"date\": \"")))
	assert.True(t, bytes.Contains(got, []byte(",\n        \"parents\": [\n            {\n                \"sha\": \"")))

	var buf bytes.Buffer
	err = HumanFriendly.NewEncoder(&buf).SetIndent("", "    ").Encode(m)
	assert.Nil(t, err)
	assert.Equal(t, data, buf.String())
}

func TestFloat64With6Digits(t *testing.T) {
	testCases := []struct {
		name     string
		input    float64
		expected string
	}{
		{"MoreThan6Digits", 1.2345678, `1.234568`},
		{"LessThan6Digits", 1.23, `1.23`},
		{"TrailingZeros", 1.230000, `1.23`},
		{"Integer", 1.0, `1`},
		{"Zero", 0.0, `0`},
		{"ManyTrailingZeros", 1.234000, `1.234`},
		{"EndsWithDot", 12345.000000, `12345`},
		{"NoDecimal", 123, `123`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := float64With6Digits(tc.input)
			b, err := f.MarshalJSON()
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, string(b))
		})
	}
}

func TestConvertAnyKeyMap(t *testing.T) {
	ptrVal := "pointer"
	testCases := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name: "SimpleMap",
			input: map[any]any{
				"b": 2,
				"a": 1,
			},
			expected: map[string]any{
				"a": 1,
				"b": 2,
			},
		},
		{
			name: "NestedMap",
			input: map[any]any{
				"b": map[any]any{"y": 2, "x": 1},
				"a": 1,
			},
			expected: map[string]any{
				"a": 1,
				"b": map[string]any{"x": 1, "y": 2},
			},
		},
		{
			name: "WithSlice",
			input: []any{
				map[any]any{"a": 1},
				map[any]any{"b": 2},
			},
			expected: []any{
				map[string]any{"a": 1},
				map[string]any{"b": 2},
			},
		},
		{
			name: "WithArray",
			input: [2]any{
				map[any]any{"a": 1},
				map[any]any{"b": 2},
			},
			expected: []any{
				map[string]any{"a": 1},
				map[string]any{"b": 2},
			},
		},
		{
			name: "WithFloat",
			input: map[any]any{
				"a": 1.2345678,
				"b": 2.0,
			},
			expected: map[string]any{
				"a": float64With6Digits(1.2345678),
				"b": 2.0,
			},
		},
		{
			name: "WithNil",
			input: map[any]any{
				"a": nil,
			},
			expected: map[string]any{
				"a": nil,
			},
		},
		{
			name: "WithPointer",
			input: map[any]any{
				"a": &ptrVal,
			},
			expected: map[string]any{
				"a": "pointer",
			},
		},
		{
			name: "WithNilPointer",
			input: map[any]any{
				"a": (*string)(nil),
			},
			expected: map[string]any{
				"a": nil,
			},
		},
		{
			name: "WithInterface",
			input: map[any]any{
				"a": any("hello"),
			},
			expected: map[string]any{
				"a": "hello",
			},
		},
		{
			name: "WithNilInterface",
			input: map[any]any{
				"a": any(nil),
			},
			expected: map[string]any{
				"a": nil,
			},
		},
		{
			name:     "NilInput",
			input:    nil,
			expected: nil,
		},
		{
			name:     "OtherType",
			input:    123,
			expected: 123,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := convertAnyKeyMap(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestKeyString(t *testing.T) {
	// To test keyString branches, we need maps with concrete key types.
	// map[interface{}]... results in Kind() == Interface, falling to default case.

	tests := []struct {
		name     string
		input    any
		expected map[string]any
	}{
		{
			name: "StringKey",
			input: map[string]any{
				"foo": "bar",
			},
			expected: map[string]any{
				"foo": "bar",
			},
		},
		{
			name: "IntKey",
			input: map[int]any{
				123: "val",
			},
			expected: map[string]any{
				"123": "val",
			},
		},
		{
			name: "Int8Key",
			input: map[int8]any{
				8: "val",
			},
			expected: map[string]any{
				"8": "val",
			},
		},
		{
			name: "Int16Key",
			input: map[int16]any{
				16: "val",
			},
			expected: map[string]any{
				"16": "val",
			},
		},
		{
			name: "Int32Key",
			input: map[int32]any{
				32: "val",
			},
			expected: map[string]any{
				"32": "val",
			},
		},
		{
			name: "Int64Key",
			input: map[int64]any{
				64: "val",
			},
			expected: map[string]any{
				"64": "val",
			},
		},
		{
			name: "UintKey",
			input: map[uint]any{
				456: "val",
			},
			expected: map[string]any{
				"456": "val",
			},
		},
		{
			name: "Uint8Key",
			input: map[uint8]any{
				8: "val",
			},
			expected: map[string]any{
				"8": "val",
			},
		},
		{
			name: "Uint16Key",
			input: map[uint16]any{
				16: "val",
			},
			expected: map[string]any{
				"16": "val",
			},
		},
		{
			name: "Uint32Key",
			input: map[uint32]any{
				32: "val",
			},
			expected: map[string]any{
				"32": "val",
			},
		},
		{
			name: "Uint64Key",
			input: map[uint64]any{
				64: "val",
			},
			expected: map[string]any{
				"64": "val",
			},
		},
		{
			name: "UintptrKey",
			input: map[uintptr]any{
				1: "val",
			},
			expected: map[string]any{
				"1": "val",
			},
		},
		{
			name: "Float32Key",
			input: map[float32]any{
				1.5: "val",
			},
			expected: map[string]any{
				"1.5": "val",
			},
		},
		{
			name: "Float64Key",
			input: map[float64]any{
				1.23: "val",
			},
			expected: map[string]any{
				"1.23": "val",
			},
		},
		{
			name: "BoolKey",
			input: map[bool]any{
				true: "val",
			},
			expected: map[string]any{
				"true": "val",
			},
		},
		// map[struct{}]any is not valid syntax directly if we want to ensure keyString logic.
		// But let's try.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := convertAnyKeyMap(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestHumanFriendlyErrorCases(t *testing.T) {
	// Function cannot be marshaled
	invalidInput := map[string]any{
		"func": func() {},
	}

	// Test Marshal error
	_, err := HumanFriendly.Marshal(invalidInput)
	assert.Error(t, err)

	// Test MarshalToString error
	_, err = HumanFriendly.MarshalToString(invalidInput)
	assert.Error(t, err)

	// Test MarshalIndent error
	_, err = HumanFriendly.MarshalIndent(invalidInput, "", "  ")
	assert.Error(t, err)

	// Test MarshalIndentString error
	_, err = HumanFriendly.MarshalIndentString(invalidInput, "", "  ")
	assert.Error(t, err)
}

func TestHumanFriendlyMarshal(t *testing.T) {
	m := map[any]any{
		"b": 2,
		"a": 1,
	}
	expected := `{"a":1,"b":2}`

	b, err := HumanFriendly.Marshal(m)
	assert.Nil(t, err)
	assert.Equal(t, expected, string(b))

	s, err := HumanFriendly.MarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, expected, s)
}

func TestHumanFriendlyMarshalIndentString(t *testing.T) {
	m := map[any]any{
		"b": 2,
		"a": 1,
	}
	expected := `{
    "a": 1,
    "b": 2
}`

	s, err := HumanFriendly.MarshalIndentString(m, "", "    ")
	assert.Nil(t, err)
	assert.Equal(t, expected, s)
}

func TestHFriendlyEncoder(t *testing.T) {
	m := map[string]string{
		"hello": "world",
		"a<b":   "c>d",
	}
	expected := "{\n\t\"a<b\": \"c>d\",\n\t\"hello\": \"world\"\n}"
	var buf bytes.Buffer
	enc := HumanFriendly.NewEncoder(&buf)
	enc.SetIndent("\t", "") // This is unusual, just for testing
	enc.SetEscapeHTML(true)
	err := enc.Encode(m)
	assert.Nil(t, err)
	assert.NotEqual(t, expected, buf.String())

	buf.Reset()
	enc = HumanFriendly.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	err = enc.Encode(m)
	assert.Nil(t, err)
	expected = "{\n  \"a<b\": \"c>d\",\n  \"hello\": \"world\"\n}\n"
	assert.Equal(t, expected, buf.String())
}
