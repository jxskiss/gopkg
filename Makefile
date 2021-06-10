.PHONY: test_linkname gen_set

list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'

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

gen_set:
	cd set && go run ./template.go
