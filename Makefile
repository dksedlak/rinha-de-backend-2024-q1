## ---------------------------
## VARIABLES
## ---------------------------
COVERPROFILE=./coverage.out

## ---------------------------
## TARGETS
## ---------------------------
.PHONY: help
help: logo ## list all the targets availables
	@echo "---"
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%s:\033[0m %s\n", $$1,$$2 } ' $(MAKEFILE_LIST)
	@echo ""

.PHONY: update
update: ## updates all the packages (go.mod)
	@echo "-> updating the packages (go.mod)"
	@go get -u ./...
	@go mod tidy
	@echo "-> the packages has been updated"

.PHONY: test
test: lint ## runs all the unit tests
	@echo "-> start [unit tests]"
	@go clean -testcache && go test -count=1 -v -cover -race ./... -coverprofile=$(COVERPROFILE)
	@echo "-> done [unit tests]"

.PHONY: test-integration
test-integration: ## runs all the integrations tests
	@echo "-> start [integration tests]"
	@docker-compose up -d
	@sleep 5
	@export ENABLE_INTEGRATION_TEST=1 && go test -count=1 -v -cover -race ./...
	@docker-compose down --remove-orphans --volumes
	@echo "-> done [integration tests]"

.PHONY: lint
lint: ## runs the linter
	@go mod tidy
	@echo "-> running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.2
	@golangci-lint run -c .golangci.yml --color always
	@echo "-> linter executed"

.PHONY: generate
generate: ## creates all the mocks
	@go get github.com/vektra/mockery/v2@v2.30.0
	@for dir in $(shell find . -type d -name 'mocks' -o -name 'mocks') ; do \
		echo "Removing $$dir..." && rm -rf $$dir \
	; done
	@go generate ./...
	@go mod tidy
	@echo "-> all the mocks has been created"

.PHONY: logo
logo:
	@echo "Rinha de Backend 2041 - Q1"
	@echo "\033[0;36m ____  _       _                 _        ____             _                  _  ";
	@echo "|  _ \(_)_ __ | |__   __ _    __| | ___  | __ )  __ _  ___| | _____ _ __   __| | ";
	@echo "| |_) | | '_ \| '_ \ / _\` |  / _\` |/ _ \ |  _ \ / _\` |/ __| |/ / _ \ '_ \ / _\` |";
	@echo "|  _ <| | | | | | | | (_| | | (_| |  __/ | |_) | (_| | (__|   <  __/ | | | (_| | ";
	@echo "|_| \_\_|_| |_|_| |_|\__,_|  \__,_|\___| |____/ \__,_|\___|_|\_\___|_| |_|\__,_| \033[0m";
	@echo ""