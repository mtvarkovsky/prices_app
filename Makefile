FILES=files
PRICES=prices
LOAD_BALANCER=load_balancer

BUILD_DIR=./build

BUILD_SRC_FILES=./cmd/files
BUILD_OUT_FILES=$(BUILD_DIR)/files

BUILD_SRC_PRICES=./cmd/prices
BUILD_OUT_PRICES=$(BUILD_DIR)/prices

BUILD_SRC_TESTDATA=./tools/testdata
BUILD_OUT_TESTDATA=$(BUILD_DIR)/testdata

NOW=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

MIGRATION_NAME=new_migration_${NOW}

OPENAPI_DIR=./api/openapi
PRICES_API_CONFIG=${OPENAPI_DIR}/prices/spec.yaml
PRICES_API_SPEC=${OPENAPI_DIR}/prices/prices.yaml
PRICES_API_OUT_SRC=pkg/api/prices.gen.go

COVER_DIR=$(BUILD_DIR)/coverage

TEST_DATA_LINES=100000
TEST_DATA_OUT=./test

.PHONY: build-files
build-files: ## builds the files executable and places it to ./build/
	go build -o ${BUILD_OUT_FILES} ${BUILD_SRC_FILES}

.PHONY: build-prices
build-prices: ## builds the prices executable and places it to ./build/
	go build -o ${BUILD_OUT_PRICES} ${BUILD_SRC_PRICES}

.PHONY: build-test-data
build-test-data: ## builds the testdata executable and places it to ./build/
	go build -o ${BUILD_OUT_TESTDATA} ${BUILD_SRC_TESTDATA}

.PHONY: build
build: build-files build-prices build-test-data ## build application

.PHONY: clean
clean: ## clean up, removes the ./build directory
	rm -rf ${BUILD_DIR}

.PHONY: run-files
run-files: ## runs the built executable
	./$(BUILD_OUT_FILES) start

.PHONY: run-files
run-prices: ## runs the built executable
	./$(BUILD_OUT_PRICES) start

.PHONY: migrate
migrate: ## install migrations dependency
	go install github.com/golang-migrate/migrate/v4

.PHONY: add-migration
add-migration: migrate ## add new .sql migration file
	migrate create -ext sql -dir pkg/migrations/sql -seq $(MIGRATION_NAME)

.PHONY: generate-test-data
generate-test-data: ## generate test data
	$(BUILD_OUT_TESTDATA) $(TEST_DATA_LINES) $(TEST_DATA_OUT)

.PHONY: docker-build-files
docker-build-files: ## build files docker image
	DOCKER_BUILDKIT=1 docker build --ssh default . -f ./deployments/Dockerfile_Files -t $(FILES):latest

.PHONY: docker-run-files
docker-run-files: ## runs files docker image
	docker-compose -f ./deployments/docker-compose.yaml run fileParser

.PHONY: docker-build-prices
docker-build-prices: ## build prices docker image
	DOCKER_BUILDKIT=1 docker build --ssh default . -f ./deployments/Dockerfile_Prices -t $(PRICES):latest

.PHONY: docker-build-load-balancer
docker-build-load-balancer: ## build load balancer docker image
	DOCKER_BUILDKIT=1 docker build --ssh default ./deployments -f ./deployments/Dockerfile_LoadBalancer -t $(LOAD_BALANCER):latest

.PHONY: docker-run-prices
docker-run-prices: ## runs prices docker image
	docker-compose -f ./deployments/docker-compose.yaml run --service-ports  apiServer


.PHONY: docker-build
docker-build: docker-build-files docker-build-prices docker-build-load-balancer ## build application

.PHONY: docker-run
docker-run: ## run application
	docker-compose -f ./deployments/docker-compose.yaml up -d

.PHONY: docker-logs
docker-logs: ## get application logs
	docker-compose -f ./deployments/docker-compose.yaml logs -f

.PHONY: docker-stop
docker-stop: ## stop application
	docker-compose -f ./deployments/docker-compose.yaml down

.PHONY: docker-stop-clean
docker-stop-clean: ## stop application and delete all data
	docker-compose -f ./deployments/docker-compose.yaml down -v

.PHONY: codegen
codegen: ## install oapicodegen dependency
	go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen

.PHONY: gomock
gomock: ## install mockgen dependency
	go get github.com/golang/mock/mockgen
	go install github.com/golang/mock/mockgen

.PHONY: mocks
mocks: gomock ## generate gomock files
	go generate ./...

.PHONY: clean-mocks
clean-mocks:
	find . -name '*_mock.go' -delete

.PHONY: prices-api
prices-api: codegen ## build code stubs from open api definition
	oapi-codegen --config $(PRICES_API_CONFIG) -o $(PRICES_API_OUT_SRC) $(PRICES_API_SPEC)

.PHONY: install-lint
install-lint: ## install golangci dependency
	go get github.com/golangci/golangci-lint/cmd/golangci-lint
	go install github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: lint
lint: install-lint ## run golang linter against the source code
	golangci-lint run ./... --fix

.PHONY: test
test: mocks ## run unit-tests
	mkdir -p ${COVER_DIR}; CGO_ENABLED=1; go test -coverprofile=${COVER_DIR}/coverage.out ./...
	go tool cover -html=${COVER_DIR}/coverage.out -o ${COVER_DIR}/coverage.html

# generate help info from comments: thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## help information about make commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

