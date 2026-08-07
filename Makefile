VERSION ?= $(shell git rev-parse --short HEAD)
SUFFIX ?=

IMAGE_REGISTRY ?= ghcr.io
IMAGE_REPO     ?= kedacore

IMAGE_CONTROLLER_BASE = $(IMAGE_REGISTRY)/$(IMAGE_REPO)/release-flow-test
IMAGE_CONTROLLER = $(IMAGE_CONTROLLER_BASE):$(VERSION)

BUILD_PLATFORMS ?= linux/amd64,linux/arm64,linux/s390x
OUTPUT_TYPE     ?= registry
METADATA_FILE   ?=

COSIGN_FLAGS ?= -y -a GIT_HASH=${GIT_COMMIT} -a GIT_VERSION=${VERSION} -a BUILD_DATE=${DATE}

docker-build:
	DOCKER_BUILDKIT=1 docker build . -t ${IMAGE_CONTROLLER}

docker-publish-multiarch:
	docker buildx build --output=type=${OUTPUT_TYPE} --platform=${BUILD_PLATFORMS} $(if $(METADATA_FILE),--metadata-file ${METADATA_FILE}) . -t ${IMAGE_CONTROLLER}

sign-images:
	COSIGN_EXPERIMENTAL=1 cosign sign ${COSIGN_FLAGS} $(IMAGE_CONTROLLER)

print-image-controller:
	@echo $(IMAGE_CONTROLLER_BASE)

## Changelog targets to be moved to the new file + binaries

CHANGELOG_DIR ?= tools/changelog
CHANGELOG_ENTRIES_DIR ?= .changelog
CHANGELOG_OUTPUT ?= CHANGELOG.md
CHANGELOG_VERSION ?=
CHANGELOG_TARGET_BRANCH ?= main
CHANGELOG_ALLOWED_TYPES ?=

CHANGELOG_ENTRY_NAME ?=

changelog-create:
	@if [ -z "$(CHANGELOG_ENTRY_NAME)" ]; then echo "CHANGELOG_ENTRY_NAME is required"; exit 1; fi
	cd $(CHANGELOG_DIR) && go run . create "$(CHANGELOG_ENTRY_NAME)" --repo-root "$(CURDIR)" $(if $(CHANGELOG_ENTRIES_DIR),--entries-dir "$(CHANGELOG_ENTRIES_DIR)")

changelog-test:
	cd $(CHANGELOG_DIR) && go test ./...

changelog-validate:
	cd $(CHANGELOG_DIR) && go run . validate --repo-root "$(CURDIR)" $(if $(CHANGELOG_ENTRIES_DIR),--entries-dir "$(CHANGELOG_ENTRIES_DIR)") $(if $(CHANGELOG_ALLOWED_TYPES),--allowed-types "$(CHANGELOG_ALLOWED_TYPES)")

changelog-generate-branch:
	cd $(CHANGELOG_DIR) && go run . generate-branch --repo-root "$(CURDIR)" $(if $(CHANGELOG_ENTRIES_DIR),--entries-dir "$(CHANGELOG_ENTRIES_DIR)") --output "$(CHANGELOG_OUTPUT)" $(if $(CHANGELOG_VERSION),--version "$(CHANGELOG_VERSION)") $(if $(CHANGELOG_ALLOWED_TYPES),--allowed-types "$(CHANGELOG_ALLOWED_TYPES)")

changelog-generate-main:
	cd $(CHANGELOG_DIR) && go run . generate-main --repo-root "$(CURDIR)" $(if $(CHANGELOG_ENTRIES_DIR),--entries-dir "$(CHANGELOG_ENTRIES_DIR)") --target-branch "$(CHANGELOG_TARGET_BRANCH)" --output "$(CHANGELOG_OUTPUT)" $(if $(CHANGELOG_VERSION),--version "$(CHANGELOG_VERSION)") $(if $(CHANGELOG_ALLOWED_TYPES),--allowed-types "$(CHANGELOG_ALLOWED_TYPES)")