module github.com/jxskiss/gopkg/v2

go 1.19

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/go-logr/logr v1.4.1
	github.com/gobwas/glob v0.2.3
	github.com/json-iterator/go v1.1.12
	github.com/spf13/cast v1.6.0
	github.com/stretchr/testify v1.8.4
	github.com/tidwall/gjson v1.17.1
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/sync v0.6.0
	golang.org/x/sys v0.18.0
	google.golang.org/protobuf v1.33.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
)

retract (
	v2.8.4 // Published accidentally.
	v2.3.5 // Has panic bug in easy.SplitMap.
)
