SHELL := /bin/bash

CGO_CFLAGS := -I/usr/local/include
CGO_LDFLAGS := -L/usr/local/lib

# Detect the OS and set the appropriate library path variable
ifeq ($(shell uname), Linux)
    LIBRARY_PATH_VAR := LD_LIBRARY_PATH
else ifeq ($(shell uname), Darwin)
    LIBRARY_PATH_VAR := DYLD_LIBRARY_PATH
else
    $(error Unsupported OS)
endif

# regexp to filter some directories from testing
EXCLUDE_DIR_REGEXP := E2E

.PHONY: build
build:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go build cmd/main.go

.PHONY: fmt
fmt:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go fmt ./...

.PHONY: lint
lint:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) golangci-lint run --timeout=30m

.PHONY: test
test:
	SKIP_E2E_FRAMEWORK_INIT=1 $(LIBRARY_PATH_VAR)=/usr/local/lib CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go test -skip $(EXCLUDE_DIR_REGEXP) ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY:vendor
vendor:
	go mod vendor

.PHONY: image
image:
	docker build . -t registry.corp.furiosa.ai/furiosa/furiosa-metrics-exporter:devel --progress=plain --platform=linux/amd64

.PHONY: image-no-cache
image-no-cache:
	docker build . --no-cache -t registry.corp.furiosa.ai/furiosa/furiosa-metrics-exporter:devel --progress=plain --platform=linux/amd64

.PHONY: helm-lint
helm-lint:
	helm lint ./deployments/helm

.PHONY:e2e
e2e:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) E2E_TEST_IMAGE_REGISTRY=$(E2E_TEST_IMAGE_REGISTRY) E2E_TEST_IMAGE_NAME=$(E2E_TEST_IMAGE_NAME) E2E_TEST_IMAGE_TAG=$(E2E_TEST_IMAGE_TAG) ginkgo ./e2e
