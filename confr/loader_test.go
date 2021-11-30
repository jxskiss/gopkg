package confr

import (
	"flag"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/easy"
)

type DBConfig struct {
	MySQL string `json:"mysql" toml:"mysql" yaml:"mysql" flag:"mysql"`
	Redis string `json:"redis" toml:"redis" yaml:"redis"`
}

type MQConfig struct {
	Cluster string `json:"cluster" toml:"cluster" yaml:"cluster" default:"some_default_cluster"`
	Topic   string `json:"topic" toml:"topic" yaml:"topic"`
	Group   string `json:"group" toml:"group" yaml:"group"`
}

type TestConfig struct {
	Some1    string  `json:"some_1" toml:"some_1" yaml:"some_1" flag:"some1"`
	Some1Ptr *string `json:"some_1_ptr" toml:"some_1_ptr" yaml:"some_1_ptr" flag:"some1_ptr"`
	Some2    int     `json:"some_2" toml:"some_2" yaml:"some_2" flag:"some2"`
	Some2Ptr *int    `json:"some_2_ptr" toml:"some_2_ptr" yaml:"some_2_ptr"`
	Some3    int64   `json:"some_3" toml:"some_3" yaml:"some_3"`
	Some3Ptr *int64  `json:"some_3_ptr" toml:"some_3_ptr" yaml:"some_3_ptr"`
	Some4    int32   `json:"some_4" toml:"some_4" yaml:"some_4"`
	Some4Ptr *int32  `json:"some_4_ptr" toml:"some_4_ptr" yaml:"some_4_ptr"`

	Some5 []int    `json:"some_5" toml:"some_5" yaml:"some_5"`
	Some6 []int64  `json:"some_6" toml:"some_6" yaml:"some_6"`
	Some7 []int32  `json:"some_7" toml:"some_7" yaml:"some_7"`
	Some8 []string `json:"some_8" toml:"some_8" yaml:"some_8"`
	Some9 []string `json:"some_9" toml:"some_9" yaml:"some_9"`

	SomeBool     bool  `json:"some_bool" toml:"some_bool" yaml:"some_bool" flag:"some-bool"`
	SomeBoolPtr1 *bool `json:"some_bool_ptr1" toml:"some_bool_ptr1" yaml:"some_bool_ptr1" flag:"some-bool-ptr1"`
	SomeBoolPtr2 *bool `json:"some_bool_ptr2" toml:"some_bool_ptr2" yaml:"some_bool_ptr2" flag:"some-bool-ptr2"`

	OverrideEnvVar  int64    `json:"override_env_var" toml:"override_env_var" yaml:"override_env_var" env:"OVERRIDE_ENV_VAR"`
	ExplicitEnvVar1 string   `env:"EXPLICIT_ENV_VAR1"`
	ExplicitEnvVar2 *float64 `env:"EXPLICIT_ENV_VAR2"`

	ImplicitEnvVar1        *string // test with Implicit_Env_Var1
	ImplicitEnvVar2        string  // test with IMPLICIT_ENV_VAR2
	ImplicitEnvVarOverride string  `json:"implicit_env_var_override" toml:"implicit_env_var_override" yaml:"implicit_env_var_override"`

	SomeRemoteValue1    string    `custom:"some_remote_value_1"`
	SomeRemoteValue1Ptr *string   `custom:"some_remote_value_1_ptr"`
	SomeRemoteValue2    []string  `custom:"some_remote_value_2"`
	SomeRemoteStruct    *MQConfig `custom:"some_remote_struct"`

	DB    DBConfig  `json:"db" toml:"db" yaml:"db"`
	DBPtr *DBConfig `json:"db_ptr" toml:"db_ptr" yaml:"db_ptr"`

	DefaultMQ    MQConfig  `json:"default_mq" toml:"default_mq" yaml:"default_mq"`
	DefaultMQPtr *MQConfig `json:"default_mq_ptr" toml:"default_mq_ptr" yaml:"default_mq_ptr"`

	MQList []*MQConfig `json:"mq_list" toml:"mq_list" yaml:"mq_list"`
}

func TestLoad_SingleFile_JSON(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.json",
	}
	testLoad_SingleFile(t, configFiles...)
}

func TestLoad_SingleFile_TOML(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.toml",
	}
	testLoad_SingleFile(t, configFiles...)
}

func TestLoad_SingleFile_YAML(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.yml",
	}
	testLoad_SingleFile(t, configFiles...)
}

func testLoad_SingleFile(t *testing.T, files ...string) {
	cfg := &TestConfig{}
	err := New(&Config{Verbose: true}).loadFiles(cfg, files...)
	assert.Nil(t, err)
	assertSingleFileConfig(t, cfg)

	assert.Empty(t, cfg.Some4)                   // not set in test config
	assert.Equal(t, []int{5, 6, 7}, cfg.Some5)   // from test config
	assert.Equal(t, []int64{6, 7, 8}, cfg.Some6) // from test config
	assert.Empty(t, cfg.Some7)                   // not set in test config
}

func assertSingleFileConfig(t *testing.T, cfg *TestConfig, exclude ...string) {
	testFields := []struct {
		Key      string
		Expected interface{}
		Actual   interface{}
	}{
		{"some_1", "some_1", cfg.Some1},
		{"some_1_ptr", "some_1_ptr", *cfg.Some1Ptr},
		{"some_2", 2345, cfg.Some2},
		{"some_2_ptr", 23456, *cfg.Some2Ptr},
		{"some_8", []string{"a", "b", "c"}, cfg.Some8},
		{"mysql_dsn", "mysql_dsn", cfg.DBPtr.MySQL},
		{"redis_dsn", "redis_dsn", cfg.DBPtr.Redis},
		{"override_env_var", int64(12345), cfg.OverrideEnvVar},
	}
	for _, tf := range testFields {
		if !easy.InStrings(exclude, tf.Key) {
			assert.Equal(t, tf.Expected, tf.Actual)
		}
	}
	assert.NotEmpty(t, cfg.MQList)
}

func TestLoad_MultipleFiles_JSON(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.json",
		"./testdata/config.common.json",
	}
	testLoad_MultipleFiles(t, configFiles...)
}

func TestLoad_MultipleFiles_TOML(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.toml",
		"./testdata/config.common.toml",
	}
	testLoad_MultipleFiles(t, configFiles...)
}

func TestLoad_MultipleFiles_YAML(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.yml",
		"./testdata/config.common.yml",
	}
	testLoad_MultipleFiles(t, configFiles...)
}

func testLoad_MultipleFiles(t *testing.T, files ...string) {
	cfg := &TestConfig{}
	loader := New(&Config{Verbose: true})
	err := loader.loadFiles(cfg, files...)
	assert.Nil(t, err)
	assertSingleFileConfig(t, cfg)
	assertCommonConfig(t, cfg)

	assert.Equal(t, 2345, cfg.Some2) // config.test override config.common
}

func assertCommonConfig(t *testing.T, cfg *TestConfig) {
	assert.Equal(t, int64(-3456), cfg.Some3)
	assert.Equal(t, int64(-34567), *cfg.Some3Ptr)
	assert.Equal(t, "common_mysql_dsn", cfg.DB.MySQL)
	assert.Equal(t, "common_redis_dsn", cfg.DB.Redis)
	assert.Equal(t, "common_default_mq_topic", cfg.DefaultMQ.Topic)
	assert.Empty(t, cfg.DefaultMQ.Cluster) // processDefaults not called
	assert.Empty(t, cfg.DefaultMQ.Group)   // neither set in config.test.yml or config.common.yml

	assert.Empty(t, cfg.Some4)                      // not set in either test config or common config
	assert.Equal(t, []int{5, 6, 7}, cfg.Some5)      // from test config
	assert.Equal(t, []int64{6, 7, 8}, cfg.Some6)    // from test config
	assert.Equal(t, []int32{-7, -8, -9}, cfg.Some7) // from common config
	assert.Equal(t, "env_var_override", cfg.ImplicitEnvVarOverride)
}

func TestLoad_DefaultValues_JSON(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.json",
		"./testdata/config.common.json",
	}
	testLoad_DefaultValues(t, configFiles...)
}

func TestLoad_DefaultValues_TOML(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.toml",
		"./testdata/config.common.toml",
	}
	testLoad_DefaultValues(t, configFiles...)
}

func TestLoad_DefaultValues_YAML(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.yml",
		"./testdata/config.common.yml",
	}
	testLoad_DefaultValues(t, configFiles...)
}

func testLoad_DefaultValues(t *testing.T, files ...string) {
	cfg := &TestConfig{}
	loader := New(&Config{Verbose: true})
	err := loader.loadFiles(cfg, files...)
	assert.Nil(t, err)
	err = loader.processDefaults(cfg)
	assert.Nil(t, err)
	assertSingleFileConfig(t, cfg)

	assert.Empty(t, cfg.DefaultMQPtr) // pointer value not set
	assert.Equal(t, "some_default_cluster", cfg.DefaultMQ.Cluster)
	assert.Equal(t, "some_default_cluster", cfg.MQList[0].Cluster)
	assert.Equal(t, "topic_0", cfg.MQList[0].Topic)
	assert.Equal(t, "group_0", cfg.MQList[0].Group)
	assert.Equal(t, "some_default_cluster", cfg.MQList[1].Cluster)
	assert.Equal(t, "topic_1", cfg.MQList[1].Topic)
	assert.Equal(t, "group_1", cfg.MQList[1].Group)
	assert.Equal(t, "cluster_2", cfg.MQList[2].Cluster)
	assert.Equal(t, "topic_2", cfg.MQList[2].Topic)
	assert.Equal(t, "group_2", cfg.MQList[2].Group)
}

func TestLoad_CommandLineFlag_JSON(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.json",
		"./testdata/config.common.json",
	}
	testLoad_CommandLineFlag(t, configFiles...)
}

func TestLoad_CommandLineFlag_TOML(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.toml",
		"./testdata/config.common.toml",
	}
	testLoad_CommandLineFlag(t, configFiles...)
}

func TestLoad_CommandLineFlag_YAML(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.yml",
		"./testdata/config.common.yml",
	}
	testLoad_CommandLineFlag(t, configFiles...)
}

var defineFlagsOnce sync.Once

func testLoad_CommandLineFlag(t *testing.T, files ...string) {
	defineFlagsOnce.Do(func() {
		flag.String("some1", "", "some string value 1")
		flag.String("some1_ptr", "flag value some1_ptr", "some string ptr value 1")
		flag.Int("some2", 0, "some int value 2")
		flag.Bool("some-bool", false, "some bool value")
		flag.Bool("some-bool-ptr1", false, "some bool ptr value 1")
		flag.Bool("some-bool-ptr2", true, "some bool ptr value 2")
	})

	var err error
	err = flag.Set("some1", "some1 flag.Set called")
	assert.Nil(t, err)
	err = flag.Set("some2", "5432")
	assert.Nil(t, err)
	err = flag.Set("some-bool", "true")
	assert.Nil(t, err)
	err = flag.Set("some-bool-ptr1", "false")
	assert.Nil(t, err)

	cfg := &TestConfig{}
	loader := New(&Config{
		Verbose: true,
		FlagSet: flag.CommandLine,
	})
	err = loader.loadFiles(cfg, files...)
	assert.Nil(t, err)
	err = loader.processFlags(cfg)
	assert.Nil(t, err)
	assertSingleFileConfig(t, cfg, "some_1", "some_2")
	assertCommonConfig(t, cfg)

	// flags

	assert.Equal(t, "some1 flag.Set called", cfg.Some1) // flag override config.test
	assert.Equal(t, 5432, cfg.Some2)                    // flag override config.test and config.common

	assert.True(t, cfg.SomeBool)       // set to true
	assert.False(t, *cfg.SomeBoolPtr1) // set to false
	assert.True(t, *cfg.SomeBoolPtr2)  // flag default value true
}

func TestLoad_CustomLoader_JSON(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.json",
		"./testdata/config.common.json",
	}
	testLoad_CustomLoader(t, configFiles...)
}

func TestLoad_CustomLoader_TOML(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.toml",
		"./testdata/config.common.toml",
	}
	testLoad_CustomLoader(t, configFiles...)
}

func TestLoad_CustomLoader_YAML(t *testing.T) {
	configFiles := []string{
		"./testdata/config.test.yml",
		"./testdata/config.common.yml",
	}
	testLoad_CustomLoader(t, configFiles...)
}

func testLoad_CustomLoader(t *testing.T, files ...string) {
	cfg := &TestConfig{}
	loader := New(&Config{
		Verbose:      true,
		CustomLoader: testCustomLoader,
	})
	err := loader.loadFiles(cfg, files...)
	assert.Nil(t, err)
	assertSingleFileConfig(t, cfg)
	err = loader.processCustom(cfg)
	assert.Nil(t, err)

	assert.Equal(t, "remote_value_1", cfg.SomeRemoteValue1)
	assert.Equal(t, "remote_value_1_ptr", *cfg.SomeRemoteValue1Ptr)
	assert.Equal(t, []string{"a", "b", "c"}, cfg.SomeRemoteValue2)
	assert.Equal(t, MQConfig{
		Cluster: "remote_struct_cluster",
		Topic:   "remote_struct_topic",
		Group:   "remote_struct_group",
	}, *cfg.SomeRemoteStruct)
}

func testCustomLoader(typ reflect.Type, tag string) (interface{}, error) {
	return testCustomLoaderData[tag], nil
}

var testCustomLoaderData = map[string]interface{}{
	"some_remote_value_1":     "remote_value_1",
	"some_remote_value_1_ptr": "remote_value_1_ptr",
	"some_remote_value_2":     []string{"a", "b", "c"},
	"some_remote_struct": &MQConfig{
		Cluster: "remote_struct_cluster",
		Topic:   "remote_struct_topic",
		Group:   "remote_struct_group",
	},
}

func Test_getEnvName(t *testing.T) {
	testcases := [][]string{
		{"ManualOverride1", "Manual_Override1"},
		{"DefaultVar", "Default_Var"},
		{"AutoSplitVar", "Auto_Split_Var"},
		{"SomeID", "Some_ID"},
		{"SomeHTMLWord", "Some_HTML_Word"},
		{"Parent2_SomeID", "Parent2_Some_ID"},
		{"Parent2_SomeHTMLWord", "Parent2_Some_HTML_Word"},
	}
	loader := New(&Config{})
	for _, tc := range testcases {
		name := tc[0]
		want1 := "Confr_" + tc[1]
		got1 := loader.getEnvName("", name)
		assert.Equal(t, want1, got1)

		want2 := "MyApp_" + tc[1]
		got2 := loader.getEnvName("MyApp", name)
		assert.Equal(t, want2, got2)
	}
}

func TestLoad_ExplicitEnv(t *testing.T) {
	testEnv := [][]string{
		{"OVERRIDE_ENV_VAR", "54321"},
		{"EXPLICIT_ENV_VAR1", "ExplicitEnvVar1"},
	}
	for _, te := range testEnv {
		_ = os.Setenv(te[0], te[1])
	}
	defer func() {
		for _, te := range testEnv {
			_ = os.Unsetenv(te[0])
		}
	}()

	configFiles := []string{
		"./testdata/config.test.yml",
		"./testdata/config.common.yml",
	}
	cfg := &TestConfig{}
	loader := New(&Config{
		Verbose:      true,
		CustomLoader: testCustomLoader,
	})
	err := loader.loadFiles(cfg, configFiles...)
	assert.Nil(t, err)
	assertSingleFileConfig(t, cfg)
	err = loader.processCustom(cfg)
	assert.Nil(t, err)
	err = loader.processEnv(cfg, "")
	assert.Nil(t, err)

	assert.Equal(t, int64(54321), cfg.OverrideEnvVar)
	assert.Equal(t, "ExplicitEnvVar1", cfg.ExplicitEnvVar1)
	assert.Nil(t, cfg.ExplicitEnvVar2)

	assert.Empty(t, cfg.ImplicitEnvVar1)
	assert.Empty(t, cfg.ImplicitEnvVar2)
	assert.Equal(t, "env_var_override", cfg.ImplicitEnvVarOverride) // from config.common
}

func TestLoad_ImplicitEnv(t *testing.T) {
	testEnv := [][]string{
		{"Confr_Implicit_Env_Var1", "implicit env var1"},
		{"CONFR_IMPLICIT_ENV_VAR2", "implicit env var2"},
		{"Confr_Implicit_Env_Var_Override", "implicit env var override"},
	}
	for _, te := range testEnv {
		_ = os.Setenv(te[0], te[1])
	}
	defer func() {
		for _, te := range testEnv {
			_ = os.Unsetenv(te[0])
		}
	}()

	configFiles := []string{
		"./testdata/config.test.yml",
		"./testdata/config.common.yml",
	}
	cfg := &TestConfig{}
	loader := New(&Config{
		Verbose:           true,
		EnableImplicitEnv: true,
		CustomLoader:      testCustomLoader,
	})
	err := loader.loadFiles(cfg, configFiles...)
	assert.Nil(t, err)
	assertSingleFileConfig(t, cfg)
	err = loader.processCustom(cfg)
	assert.Nil(t, err)
	err = loader.processEnv(cfg, "")
	assert.Nil(t, err)

	assert.Equal(t, "implicit env var1", *cfg.ImplicitEnvVar1)
	assert.Equal(t, "implicit env var2", cfg.ImplicitEnvVar2)
	assert.Equal(t, "implicit env var override", cfg.ImplicitEnvVarOverride)
}

func TestLoad_AllowUnknownFields_JSON(t *testing.T) {
	configFiles := []string{"./testdata/config.unknown_fields.json"}
	testLoad_AllowUnknownFields(t, configFiles...)
}

func TestLoad_AllowUnknownFields_TOML(t *testing.T) {
	configFiles := []string{"./testdata/config.unknown_fields.toml"}
	testLoad_AllowUnknownFields(t, configFiles...)
}

func TestLoad_AllowUnknownFields_YAML(t *testing.T) {
	configFiles := []string{"./testdata/config.unknown_fields.yml"}
	testLoad_AllowUnknownFields(t, configFiles...)
}

func testLoad_AllowUnknownFields(t *testing.T, files ...string) {
	cfg := &TestConfig{}
	loader := New(&Config{})
	err := loader.Load(cfg, files...)
	assert.Nil(t, err)
}

func TestLoad_DisallowUnknownFields_JSON(t *testing.T) {
	configFiles := []string{"./testdata/config.unknown_fields.json"}
	testLoad_DisallowUnknownFields(t, configFiles...)
}

func TestLoad_DisallowUnknownFields_TOML(t *testing.T) {
	configFiles := []string{"./testdata/config.unknown_fields.toml"}
	testLoad_DisallowUnknownFields(t, configFiles...)
}

func TestLoad_DisallowUnknownFields_YAML(t *testing.T) {
	configFiles := []string{"./testdata/config.unknown_fields.yml"}
	testLoad_DisallowUnknownFields(t, configFiles...)
}

func testLoad_DisallowUnknownFields(t *testing.T, files ...string) {
	cfg := &TestConfig{}
	loader := New(&Config{DisallowUnknownFields: true})
	err := loader.Load(cfg, files...)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unknown_field")
}
