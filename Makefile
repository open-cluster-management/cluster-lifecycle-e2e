export GITHUB_USER    := $(shell echo $(GITHUB_USER) | sed 's/@/%40/g')
export GITHUB_TOKEN   ?=

export PROJECT_DIR            = $(shell 'pwd')
export BUILD_DIR              = $(PROJECT_DIR)/build

## WARNING: OPERATOR-SDK - IMAGE_DESCRIPTION & DOCKER_BUILD_OPTS MUST NOT CONTAIN ANY SPACES
export IMAGE_DESCRIPTION ?= Cluster_Lifecycle_e2e
export DOCKER_FILE        = $(BUILD_DIR)/Dockerfile
export DOCKER_REGISTRY   ?= quay.io
export DOCKER_NAMESPACE  ?= open-cluster-management
export DOCKER_IMAGE      ?= $(COMPONENT_NAME)
export DOCKER_IMAGE_COVERAGE_POSTFIX ?= -coverage
export DOCKER_IMAGE_COVERAGE      ?= $(DOCKER_IMAGE)$(DOCKER_IMAGE_COVERAGE_POSTFIX)
export DOCKER_BUILD_TAG  ?= latest
export DOCKER_TAG        ?= $(shell whoami)

USE_VENDORIZED_BUILD_HARNESS ?=

ifndef USE_VENDORIZED_BUILD_HARNESS
# -include $(shell curl -s -H 'Authorization: token ${GITHUB_TOKEN}' -H 'Accept: application/vnd.github.v4.raw' -L https://api.github.com/repos/itdove/build-harness-extensions/contents/templates/Makefile.build-harness-bootstrap?branch=code_coverage -o .build-harness-bootstrap; echo .build-harness-bootstrap)
-include $(shell curl -s -H 'Authorization: token ${GITHUB_TOKEN}' -H 'Accept: application/vnd.github.v4.raw' -L https://api.github.com/repos/open-cluster-management/build-harness-extensions/contents/templates/Makefile.build-harness-bootstrap -o .build-harness-bootstrap; echo .build-harness-bootstrap)
else
-include vbh/.build-harness-vendorized
endif

# Only use git commands if it exists
ifdef GIT
GIT_COMMIT      = $(shell git rev-parse --short HEAD)
GIT_REMOTE_URL  = $(shell git config --get remote.origin.url)
VCS_REF     = $(if $(shell git status --porcelain),$(GIT_COMMIT)-$(BUILD_DATE),$(GIT_COMMIT))
endif

.PHONY: build
build:
	make docker/info
	make docker/build

.PHONY: push
push:: docker/tag docker/login
	make docker/push

	