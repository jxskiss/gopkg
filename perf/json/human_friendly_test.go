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
]`
	var m []interface{}
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
	assert.True(t, bytes.HasPrefix(got, []byte("[\n    {\n        \"author\": {\n            \"avatar_url\"")))
	assert.True(t, bytes.Contains(got, []byte(",\n        \"comments_url\": \"")))
	assert.True(t, bytes.Contains(got, []byte("        \"commit\": {\n            \"author\": {\n                \"date\": \"")))
	assert.True(t, bytes.Contains(got, []byte(",\n        \"parents\": [\n            {\n                \"sha\": \"")))
}
