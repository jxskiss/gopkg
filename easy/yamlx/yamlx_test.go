package yamlx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/jxskiss/gopkg/v2/perf/json"
)

func TestMarshal(t *testing.T) {
	data := map[string]any{
		"a": 1,
		"b": "2",
		"c": 3.14,
	}
	buf, err := Marshal(data)
	require.Nil(t, err)
	assert.True(t, len(buf) > 0)
}

func TestUnmarshal(t *testing.T) {
	t.Run("without extended feature", func(t *testing.T) {
		yamlData, err := os.ReadFile("./testdata/normal.yaml")
		require.Nil(t, err)

		var out map[string]any
		err = Unmarshal(yamlData, &out)
		require.Nil(t, err)
		assert.Len(t, out, 2)
		assert.Len(t, out["definitions"], 1)

		jsonData, err := json.MarshalToString(out)
		require.Nil(t, err)
		assert.Equal(t, "production", gjson.Get(jsonData, "pipelines.branches.main.1.step.deployment").String())
		assert.Equal(t, "manual", gjson.Get(jsonData, "pipelines.branches.main.1.step.trigger").String())
	})

	t.Run("env", func(t *testing.T) {
		os.Setenv("MY_ENV_1", "value1")
		os.Setenv("MY_ENV_3", "value3")

		yamlData, err := os.ReadFile("./testdata/env.yaml")
		require.Nil(t, err)

		var out any
		err = Unmarshal(yamlData, &out, EnableEnv())
		require.Nil(t, err)
		assert.Len(t, out, 5)

		jsonData, err := json.MarshalToString(out)
		require.Nil(t, err)
		assert.Equal(t, "value1", gjson.Get(jsonData, "0").String())
		assert.Equal(t, "value1", gjson.Get(jsonData, "1").String())
		assert.Equal(t, "value3", gjson.Get(jsonData, "2").String())
		assert.Equal(t, "value1", gjson.Get(jsonData, "3.key1").String())
		assert.Equal(t, "value1", gjson.Get(jsonData, "3.key2").String())
		assert.Equal(t, "value3", gjson.Get(jsonData, "3.key3").String())
		assert.True(t, gjson.Get(jsonData, "4.key4").Exists())
		assert.Equal(t, "", gjson.Get(jsonData, "4.key4").String())
	})

	t.Run("include / success", func(t *testing.T) {
		yamlData, err := os.ReadFile("./testdata/include1.yaml")
		require.Nil(t, err)

		var out map[string]any
		err = Unmarshal(yamlData, &out, EnableInclude())
		require.Nil(t, err)
		assert.Len(t, out, 3)

		jsonData, err := json.MarshalToString(out)
		require.Nil(t, err)
		assert.Equal(t, "value1", gjson.Get(jsonData, "key1.subkey1").String())
		assert.Equal(t, "production", gjson.Get(jsonData, "key1.subkey3.pipelines.branches.main.1.step.deployment").String())
		assert.Equal(t, "production", gjson.Get(jsonData, "key1.subkey4.pipelines.branches.main.1.step.deployment").String())
		assert.EqualValues(t, 12345, gjson.Get(jsonData, "key2.array_1.1").Int())
		assert.EqualValues(t, 12345, gjson.Get(jsonData, "key3.sub1.sub2.0.sub3_k1").Int())
		assert.Equal(t, "production", gjson.Get(jsonData, "key3.sub1.sub2.0.sub3_k2.pipelines.branches.main.1.step.deployment").String())
	})

	t.Run("include / circular", func(t *testing.T) {
		yamlData, err := os.ReadFile("./testdata/include_circular_1.yaml")
		require.Nil(t, err)

		var out map[string]any
		err = Unmarshal(yamlData, &out, EnableInclude())
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "circular include detected: ")
		assert.Contains(t, err.Error(), filepath.Join("testdata", "include_circular_3.yaml"))
	})

	t.Run("reference / success", func(t *testing.T) {
		yamlData, err := os.ReadFile("./testdata/ref.yaml")
		require.Nil(t, err)

		var out map[string]any
		err = Unmarshal(yamlData, &out)
		require.Nil(t, err)
		assert.Len(t, out, 9)

		jsonData, err := json.MarshalToString(out)
		require.Nil(t, err)
		assert.Equal(t, "bar", gjson.Get(jsonData, "obj1.foo").String())
		assert.Equal(t, "bar", gjson.Get(jsonData, "test_ref1").String())
		assert.EqualValues(t, 3, gjson.Get(jsonData, "test_ref2").Int())
		assert.EqualValues(t, 3, gjson.Get(jsonData, "test_ref3").Int())
		assert.Equal(t, "bar", gjson.Get(jsonData, "test_ref4.key1.key2").String())
		assert.EqualValues(t, 3, gjson.Get(jsonData, "test_ref4.key1.key3").Int())
		assert.Equal(t, "bar", gjson.Get(jsonData, "friends.0.last").String())
		assert.EqualValues(t, 3, gjson.Get(jsonData, "friends.1.age").Int())
		assert.Equal(t, []any{"Dale", "Roger", "Jane"}, gjson.Get(jsonData, "test_ref5").Value())
		assert.Equal(t,
			map[string]any{
				"key2": "bar",
				"key3": float64(3),
				"key4": []any{"Dale", "Roger", "Jane"},
			},
			gjson.Get(jsonData, "test_ref6.key1").Value())
	})

	t.Run("reference / not found", func(t *testing.T) {
		yamlData, err := os.ReadFile("./testdata/ref_not_found.yaml")
		require.Nil(t, err)

		var out map[string]any
		err = Unmarshal(yamlData, &out)
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "cannot find referenced data: friends.0.phone_number")
	})

	t.Run("reference / circular", func(t *testing.T) {
		yamlData, err := os.ReadFile("./testdata/ref_circular.yaml")
		require.Nil(t, err)

		var out map[string]any
		err = Unmarshal(yamlData, &out)
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "circular reference detected: friends.#.age")
	})

	t.Run("variable / success", func(t *testing.T) {
		yamlData, err := os.ReadFile("./testdata/variable.yaml")
		require.Nil(t, err)

		var out map[string]any
		err = Unmarshal(yamlData, &out)
		require.Nil(t, err)
		assert.Len(t, out, 3)

		jsonData, err := json.MarshalToString(out)
		require.Nil(t, err)
		assert.Equal(t, []any{"mvn package"}, gjson.Get(jsonData, "vars.k1").Value())
		assert.Equal(t,
			map[string]any{
				"key1": "value1",
			},
			gjson.Get(jsonData, "vars.k2.k3").Value())
		assert.Equal(t, "test", gjson.Get(jsonData, "vars.k4.0").Value())
		assert.Equal(t,
			map[string]any{
				"name":       "Deploy",
				"deployment": "production",
				"script": []any{
					"./deploy.sh target/my-app.jar",
				},
				"trigger": "manual",
			},
			gjson.Get(jsonData, "vars.k5.k6.k7").Value())
	})

	t.Run("variable / circular_1", func(t *testing.T) {
		yamlData, err := os.ReadFile("./testdata/variable_circular_1.yaml")
		require.Nil(t, err)

		var out any
		err = Unmarshal(yamlData, &out)
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "circular variable reference detected: test_var")
	})

	t.Run("variable / circular_2", func(t *testing.T) {
		yamlData, err := os.ReadFile("./testdata/variable_circular_2.yaml")
		require.Nil(t, err)

		var out any
		err = Unmarshal(yamlData, &out)
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "circular variable reference detected: test_var")
	})

	t.Run("function / success", func(t *testing.T) {
		yamlData := `
nowUnix: "@@fn nowUnix"
nowRFC3339: "@@fn   nowRFC3339"
nowFormat: '@@fn nowFormat("2006-01-02")'
uuid: '@@fn uuid'
key2:
  randStr: '@@fn randStr(5)'
  array:
    - '@@ref nowUnix'
    - '@@ref key2.randStr'
`
		var out map[string]any
		err := Unmarshal([]byte(yamlData), &out)
		require.Nil(t, err)

		jsonData, err := json.MarshalToString(out)
		require.Nil(t, err)

		assert.NotZero(t, gjson.Get(jsonData, "nowUnix").Int())
		assert.NotZero(t, gjson.Get(jsonData, "nowRFC3339").String())
		assert.NotZero(t, gjson.Get(jsonData, "nowFormat").String())
		assert.NotZero(t, gjson.Get(jsonData, "uuid").String())
		assert.NotZero(t, gjson.Get(jsonData, "key2.randStr").String())
		assert.Equal(t,
			gjson.Get(jsonData, "nowUnix").Int(),
			gjson.Get(jsonData, "key2.array.0").Int())
		assert.Equal(t,
			gjson.Get(jsonData, "key2.randStr").String(),
			gjson.Get(jsonData, "key2.array.1").String())
	})

	t.Run("function / custom", func(t *testing.T) {
		fn1 := func() int64 {
			return 123
		}
		fn2 := func() (string, error) {
			return "abc", nil
		}
		fn3 := func(i int, s string) (string, error) {
			slice := []string{"x", "y", "z"}
			return s + slice[i], nil
		}
		yamlData := `
k1: "@@fn fn1"
k2: "@@fn fn2()"
k3: '@@fn fn3(1, "123")'
k4: '@@fn fn3(2, "123")'
`
		var out map[string]any
		err := Unmarshal([]byte(yamlData), &out,
			WithFuncMap(map[string]any{
				"fn1": fn1, "fn2": fn2, "fn3": fn3,
			}))
		require.Nil(t, err)
		assert.Len(t, out, 4)

		assert.Equal(t, 123, out["k1"])
		assert.Equal(t, "abc", out["k2"])
		assert.Equal(t, "123y", out["k3"])
		assert.Equal(t, "123z", out["k4"])
	})

	t.Run("variable to function result", func(t *testing.T) {
		os.Setenv("ENTITY_ID", "12345")
		yamlData, err := os.ReadFile("./testdata/var_to_fn.yaml")
		require.Nil(t, err)

		var out map[string]any
		err = Unmarshal(yamlData, &out, EnableEnv())
		require.Nil(t, err)
		assert.Len(t, out, 8)
		assert.Len(t, out["key2"], 2)
		assert.Len(t, out["var1"], 2)
		assert.NotZero(t, out["cid"])
		assert.Equal(t, out["cid"], out["var2"])
		assert.Equal(t, "12345", out["key3"])
		assert.Equal(t, "12345", out["env1"])
	})

	t.Run("escape", func(t *testing.T) {
		yamlData := `
nowUnix: "\\@@fn nowUnix"
type: "\\@@type"
key2:
  - '\@@ref ref_1'
  - key1: "\\@@incl abc.yaml "
    key2: '\@@var   test_var'
    key3: '\\@@var test_var'
    key4: '\\@var test_var'
    key5: "\\@var test_var"
`
		var out map[string]any
		err := Unmarshal([]byte(yamlData), &out)
		require.Nil(t, err)

		assert.Equal(t, "@@fn nowUnix", out["nowUnix"])
		assert.Equal(t, "@@type", out["type"])
		assert.Equal(t,
			[]any{
				"@@ref ref_1",
				map[string]any{
					"key1": "@@incl abc.yaml ",
					"key2": `@@var   test_var`,
					"key3": `\@@var test_var`,
					"key4": `\\@var test_var`,
					"key5": `\@var test_var`,
				},
			},
			out["key2"])
	})
}

func TestLoad(t *testing.T) {
	var err error
	var out any

	err = Load("./testdata/normal.yaml", &out)
	require.Nil(t, err)

	err = Load("./testdata/include1.yaml", &out, EnableInclude())
	require.Nil(t, err)

	err = Load("./testdata/ref.yaml", &out)
	require.Nil(t, err)

	err = Load("./testdata/variable.yaml", &out)
	require.Nil(t, err)
}

func Test_unescapeStrValue(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{
			input: "\\@@var test_var",
			want:  "@@var test_var",
		},
		{
			input: `\@@incl file.yaml `,
			want:  "@@incl file.yaml ",
		},
		{
			input: `\@@@type`,
			want:  `@@@type`,
		},
		{
			input: `\\\\@@type`,
			want:  `\\\@@type`,
		},
		{
			input: `\\\@@@`,
			want:  `\\@@@`,
		},
	}
	for _, tc := range testCases {
		got := unescapeStrValue(tc.input)
		assert.Equalf(t, tc.want, got, "input= %q", tc.input)
	}
}
