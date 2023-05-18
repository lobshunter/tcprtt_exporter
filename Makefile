BIN ?= $(PWD)/bin

GOOS    := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
GOARCH  := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
GOENV   := GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO      := $(GOENV) go
DOCKER := DOCKER_BUILDKIT=1 docker

REPO := github.com/lobshunter/tcprtt_exporter

COMMIT    := $(shell git describe --no-match --always --dirty)
BRANCH    := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_DATE := $(shell date -Iseconds)

LDFLAGS := -w -s
LDFLAGS += -X "$(REPO)/pkg/version.gitHash=$(COMMIT)"
LDFLAGS += -X "$(REPO)/pkg/version.gitBranch=$(BRANCH)"
LDFLAGS += -X "$(REPO)/pkg/version.buildDate=$(BUILD_DATE)"

IMAGE ?= lobshunter/tcprtt_exporter
IMAGE_TAG ?= $(COMMIT)

default: build

build:
	$(GO) build -ldflags '$(LDFLAGS)' -o $(BIN)/$(GOOS)/$(GOARCH)/ .

build-image: build
	$(DOCKER) build --build-arg arch=$(GOARCH) --platform linux/$(GOARCH) -t $(IMAGE):$(IMAGE_TAG)-$(GOARCH) .

image-arm64:
	GOOS=linux GOARCH=arm64 make -j2 build-image

image-amd64:
	GOOS=linux GOARCH=amd64 make -j2 build-image

image: image-amd64 image-arm64

image-push-amd64: image-amd64
	$(DOCKER) push $(IMAGE):$(IMAGE_TAG)-amd64

image-push-arm64: image-arm64
	$(DOCKER) push $(IMAGE):$(IMAGE_TAG)-arm64

image-push: image-push-arm64 image-push-amd64
	$(DOCKER) manifest create $(IMAGE):$(IMAGE_TAG) \
		--amend $(IMAGE):$(IMAGE_TAG)-amd64 \
		--amend $(IMAGE):$(IMAGE_TAG)-arm64
	$(DOCKER) manifest push --purge $(IMAGE):$(IMAGE_TAG)
