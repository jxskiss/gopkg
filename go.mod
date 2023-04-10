module github.com/jxskiss/gopkg/v2

go 1.18

require (
	github.com/BurntSushi/toml v1.2.1
	github.com/bytedance/sonic v1.8.7
	github.com/davecgh/go-spew v1.1.1
	github.com/go-logr/logr v1.2.3
	github.com/json-iterator/go v1.1.12
	github.com/jxskiss/base62 v1.1.0
	github.com/spf13/cast v1.5.0
	github.com/stretchr/testify v1.8.1
	go.uber.org/atomic v1.10.0
	go.uber.org/zap v1.24.0
	golang.org/x/arch v0.3.0
	golang.org/x/sync v0.1.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
)

retract v2.3.5 // Has bug.
