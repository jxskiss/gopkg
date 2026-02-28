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
