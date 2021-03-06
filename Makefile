.PHONY: test_linkname gen_set

list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'

gen_set:
	cd set && go run ./template.go

test_linkname:
	go clean -testcache ./internal/linkname && go test ./internal/linkname

test_forceexport:
	go test -gcflags=all=-l ./forceexport

test_monkey:
	go test -gcflags=all=-l ./monkey

test_json:
	go test ./json
	go test --tags gojson ./json
	go test --tags jsoniter ./json

test_coverage:
	mkdir -p _output/
	go test -count=1 -cover -coverprofile=_output/coverprofile.out ./... && go tool cover -html _output/coverprofile.out
