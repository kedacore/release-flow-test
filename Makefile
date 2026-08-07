VERSION ?= main
SUFFIX ?=

IMAGE_REGISTRY ?= ghcr.io
IMAGE_REPO     ?= kedacore

IMAGE_CONTROLLER = $(IMAGE_REGISTRY)/$(IMAGE_REPO)/release-flow-test$(SUFFIX):$(VERSION)

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
	@echo $(IMAGE_CONTROLLER)