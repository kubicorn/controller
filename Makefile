TARGET = controller
GOTARGET = github.com/kubicorn/$(TARGET)
REGISTRY ?= kubicorn
IMAGE = $(REGISTRY)/$(TARGET)
DIR := ${CURDIR}
DOCKER ?= docker

GIT_VERSION ?= $(shell git describe --always --dirty)
IMAGE_VERSION ?= $(shell git describe --always --dirty)
IMAGE_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD | sed 's/\///g')
GIT_REF = $(shell git rev-parse --short=8 --verify HEAD)

push:
	$(DOCKER) push $(REGISTRY)/$(TARGET):$(GIT_REF)

clean:
	rm -f $(TARGET)
	$(DOCKER) rmi $(REGISTRY)/$(TARGET) || true

container:
	$(DOCKER) build \
		-t $(REGISTRY)/$(TARGET):$(IMAGE_VERSION) \
		-t $(REGISTRY)/$(TARGET):$(IMAGE_BRANCH) \
		-t $(REGISTRY)/$(TARGET):$(GIT_REF) \
		.

run:
	$(DOCKER) run $(REGISTRY)/$(TARGET):$(IMAGE_VERSION)


build:
	go build -o bin/kubicorn-controller main.go
