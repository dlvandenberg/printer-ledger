.PHONY: list
list:
		@LC_ALL=C $(MAKE) -pRrq -f $(firstword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/(^|\n)# Files(\n|$$)/,/(^|\n)# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | grep -E -v -e '^[^[:alnum:]]' -e '^$@$$'

DEV_DB := ./data/ledger_dev.db

run:
		go run . -db $(DEV_DB)

test:
		go test ./internal/...

test-verbose:
		go test ./internal/app -v

lint:
		golangci-lint run ./...
