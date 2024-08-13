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
	$(LIBRARY_PATH_VAR)=/usr/local/lib CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go test ./...

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
