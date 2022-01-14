
SHELL := /bin/bash

export PROJECT_DIR            = $(shell 'pwd')
export BUILD_DIR              = $(PROJECT_DIR)/build

export COMPONENT_NAME ?= $(shell cat ./COMPONENT_NAME 2> /dev/null)
export COMPONENT_VERSION ?= $(shell cat ./COMPONENT_VERSION 2> /dev/null)

export ARCH       ?= $(shell uname -m)
export ARCH_TYPE   = $(if $(patsubst x86_64,,$(ARCH)),$(ARCH),amd64)
export BUILD_DATE  = $(shell date +%m/%d@%H:%M:%S)
export VCS_REF     = $(if $(shell git status --porcelain),$(GIT_COMMIT)-$(BUILD_DATE),$(GIT_COMMIT))

export CGO_ENABLED  = 0
export GO111MODULE := on
export GOOS         = $(shell go env GOOS)
export GOARCH       = $(ARCH_TYPE)
export GOPACKAGES   = $(shell go list ./... | grep -v /vendor | grep -v /internal | grep -v /build | grep -v /test)

export COMPONENT_SCRIPTS_PATH = $(BUILD_DIR)

## WARNING: OPERATOR-SDK - IMAGE_DESCRIPTION & DOCKER_BUILD_OPTS MUST NOT CONTAIN ANY SPACES
export IMAGE_DESCRIPTION ?= Cluster_Lifecycle_e2e
export DOCKER_FILE        = $(BUILD_DIR)/Dockerfile
export DOCKER_REGISTRY   ?= quay.io
export DOCKER_NAMESPACE  ?= stolostron
export DOCKER_IMAGE      ?= $(COMPONENT_NAME)
export DOCKER_IMAGE_COVERAGE_POSTFIX ?= -coverage
export DOCKER_IMAGE_COVERAGE      ?= $(DOCKER_IMAGE)$(DOCKER_IMAGE_COVERAGE_POSTFIX)
export DOCKER_BUILD_TAG  ?= latest
export DOCKER_TAG        ?= $(shell whoami)
export DOCKER_BUILDER    ?= docker

export DOCKER_BUILD_OPTS  = --build-arg VCS_REF=$(VCS_REF) \
	--build-arg VCS_URL=$(GIT_REMOTE_URL) \
	--build-arg IMAGE_NAME=$(DOCKER_IMAGE) \
	--build-arg IMAGE_DESCRIPTION=$(IMAGE_DESCRIPTION) \
	--build-arg ARCH_TYPE=$(ARCH_TYPE) \
	--build-arg REMOTE_SOURCE=. \
	--build-arg REMOTE_SOURCE_DIR=/remote-source \
	--build-arg BUILD_HARNESS_EXTENSIONS_PROJECT=${BUILD_HARNESS_EXTENSIONS_PROJECT}

# Only use git commands if it exists
ifdef GIT
GIT_COMMIT      = $(shell git rev-parse --short HEAD)
GIT_REMOTE_URL  = $(shell git config --get remote.origin.url)
VCS_REF     = $(if $(shell git status --porcelain),$(GIT_COMMIT)-$(BUILD_DATE),$(GIT_COMMIT))
endif

.PHONY: deps
## Download all project dependencies
deps: init component/init

.PHONY: push
push:: docker/tag docker/login
	make docker/push

.PHONY: build
build:
	go get -u github.com/onsi/ginkgo/ginkgo@v1.16.4
	ginkgo build pkg/tests/metrics
	ginkgo build pkg/tests/create_cluster
	ginkgo build pkg/tests/create_cluster_bm
	ginkgo build pkg/tests/import_cluster
	ginkgo build pkg/tests/detach_destroy
	ginkgo build pkg/tests/destroy_bm

.PHONY: build-image
build-image:
	@$(DOCKER_BUILDER) build -t $(DOCKER_IMAGE) -f $(DOCKER_FILE) .
	echo "${DOCKER_REGISTRY}/${DOCKER_IMAGE}:$(DOCKER_BUILD_TAG)"
	@$(DOCKER_BUILDER) tag $(DOCKER_IMAGE) ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:$(DOCKER_BUILD_TAG)
