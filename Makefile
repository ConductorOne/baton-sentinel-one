GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
BUILD_DIR = dist/${GOOS}_${GOARCH}
OUTPUT_PATH = ${BUILD_DIR}/baton-sentinel-one

ifeq ($(GOOS),windows)
OUTPUT_PATH = ${BUILD_DIR}/baton-sentinel-one.exe
else
OUTPUT_PATH = ${BUILD_DIR}/baton-sentinel-one
endif

.PHONY: build
build:
	go build -o ${OUTPUT_PATH} ./cmd/baton-sentinel-one

.PHONY: update-deps
update-deps:
	go get -d -u ./...
	go mod tidy -v
	go mod vendor

.PHONY: add-dep
add-dep:
	go mod tidy -v
	go mod vendor

.PHONY: lint
lint:
	golangci-lint run
