module github.com/jxskiss/gopkg

go 1.16

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/davecgh/go-spew v1.1.1
	github.com/goccy/go-json v0.7.9-0.20210927113039-9df46fc918f2
	github.com/spf13/cast v1.4.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/atomic v1.7.0
	go.uber.org/zap v1.19.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/pkg/errors => github.com/jxskiss/errors v0.14.0
