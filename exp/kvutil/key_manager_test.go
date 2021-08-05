package kvutil

import (
	"testing"
)

func Test_Key(t *testing.T) {
	var km KeyManager
	key := km.NewKey("abc:def:%d:%s")
	want := "abc:def:1234567:x0BtEadepz6L"

	got := key(1234567, "x0BtEadepz6L")
	if got != want {
		t.Errorf("failed Test_Key: got=%v want=%v", got, want)
	}
}

func Test_Key_Corner(t *testing.T) {
	cases := []struct {
		key  string
		args []interface{}
		want string
	}{
		{
			key:  "%v:blah:%v",
			args: []interface{}{1234567, "x0BtEadepz6L"},
			want: "1234567:blah:x0BtEadepz6L",
		},
		{
			key:  "%v%v:blah:",
			args: []interface{}{1234567, "x0BtEadepz6L"},
			want: "1234567x0BtEadepz6L:blah:",
		},
		{
			key:  "%v:%v:blah",
			args: []interface{}{1234567, "x0BtEadepz6L"},
			want: "1234567:x0BtEadepz6L:blah",
		},
		{
			key:  "blah:%v%v",
			args: []interface{}{1234567, "x0BtEadepz6L"},
			want: "blah:1234567x0BtEadepz6L",
		},
	}

	var km KeyManager
	for _, c := range cases {
		key := km.NewKey(c.key)
		got := key(c.args...)

		if got != c.want {
			t.Errorf("failed Test_Key_Corner: key=%v got=%v want=%v", c.key, got, c.want)
		}
	}
}

func Test_Key_CurlyBrace(t *testing.T) {
	cases := []struct {
		key  string
		args []interface{}
		want string
	}{
		{
			key:  "{}:blah:{}",
			args: []interface{}{1234567, "x0BtEadepz6L"},
			want: "1234567:blah:x0BtEadepz6L",
		},
		{
			key:  "{}{}:blah:",
			args: []interface{}{1234567, "x0BtEadepz6L"},
			want: "1234567x0BtEadepz6L:blah:",
		},
		{
			key:  "{}:{}:blah",
			args: []interface{}{1234567, "x0BtEadepz6L"},
			want: "1234567:x0BtEadepz6L:blah",
		},
		{
			key:  "blah:{}{}",
			args: []interface{}{1234567, "x0BtEadepz6L"},
			want: "blah:1234567x0BtEadepz6L",
		},
	}

	var km KeyManager
	for _, c := range cases {
		key := km.NewKey(c.key)
		got := key(c.args...)

		if got != c.want {
			t.Errorf("failed Test_Key_CurlyBrace: key=%v got=%v want=%v", c.key, got, c.want)
		}
	}
}

func Test_Key_NamedArgs(t *testing.T) {
	var km KeyManager
	key := km.NewKey("abc:{some_id}:{dummy}")
	want := "abc:1234567:x0BtEadepz6L"
	got := key(1234567, "x0BtEadepz6L")
	if got != want {
		t.Errorf("failed Test_Key_NamedArgs: got=%v, want=%v", got, want)
	}
}

func Test_Key_WithArgNames(t *testing.T) {
	var km KeyManager
	key := km.NewKey("{{some_id_1}_foo_bar_count}:{some_id_2}", "some_id_1", "some_id_2")
	want := "{111_foo_bar_count}:222"
	got := key(111, 222)
	if got != want {
		t.Errorf("failed Test_Key_WithArgNames: got= %v, want= %v", got, want)
	}
}

func Test_SetKeyPrefix(t *testing.T) {
	var km KeyManager
	km.SetPrefix("some_prefix:")

	key := km.NewKey("abc:def:%d:%s")
	want := "some_prefix:abc:def:1234567:x0BtEadepz6L"
	got := key(1234567, "x0BtEadepz6L")
	if got != want {
		t.Errorf("failed Test_Key: got=%v want=%v", got, want)
	}
}
