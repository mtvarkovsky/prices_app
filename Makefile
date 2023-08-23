FILES=files
PRICES=prices

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

.PHONY: build-files
build-files: ## builds the files executable and places it to ./build/
	go build -o ${BUILD_OUT_FILES} ${BUILD_SRC_FILES}

.PHONY: build-prices
build-prices: ## builds the prices executable and places it to ./build/
	go build -o ${BUILD_OUT_PRICES} ${BUILD_SRC_PRICES}

.PHONY: build-test-data
build-test-data: ## builds the testdata executable and places it to ./build/
	go build -o ${BUILD_OUT_TESTDATA} ${BUILD_SRC_TESTDATA}

.PHONY: clean
clean: ## clean up, removes the ./build directory
	@rm -rf ${BUILD_DIR}

.PHONY: run-files
run-files: ## runs the built executable
	./$(BUILD_OUT_FILES) start

.PHONY: run-files
run-prices: ## runs the built executable
	./$(BUILD_OUT_PRICES) start

.PHONY: migrate
migrate:
	@go install github.com/golang-migrate/migrate/v4

.PHONY: add-migration
add-migration: migrate ## add new .sql migration file
	@migrate create -ext sql -dir pkg/migrations/sql -seq $(MIGRATION_NAME)

.PHONY: docker-build-files
docker-build-files: ## build files docker image
	DOCKER_BUILDKIT=1 docker build --ssh default . -f ./deployments/Dockerfile_Files -t $(FILES):latest

.PHONY: docker-build-prices
docker-build-prices: ## build prices docker image
	DOCKER_BUILDKIT=1 docker build --ssh default . -f ./deployments/Dockerfile_Prices -t $(PRICES):latest

.PHONY: codegen
codegen:
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen

.PHONY: prices-api
prices-api: codegen
	@oapi-codegen --config $(PRICES_API_CONFIG) -o $(PRICES_API_OUT_SRC) $(PRICES_API_SPEC)
