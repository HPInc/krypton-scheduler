include ../common.mk

BINARY_DIR=../bin/scheduler
BINARY_NAME=schedservice

SCHED_DOCKER_IMAGE=krypton-scheduler
BUILD_CACHE_DOCKER_IMAGE=krypton-scheduler-build-cache
SCHED_GHCR_IMAGE=ghcr.io/hpinc/krypton/$(SCHED_DOCKER_IMAGE)

all: build docker-image

build: build-binaries

# Fetch dependency protobuf files and generate go code.
fetch-protos:
	docker pull ghcr.io/hpinc/krypton/krypton-go-builder
	docker pull ghcr.io/hpinc/krypton/krypton-go-base
	make -C ../protos dstsprotos

gosec:
	gosec -exclude-dir=protos ./...

# Build the scheduler protobuf files and generate go code.
build-protos:
	make -C ../protos schedprotos

# Build the binaries for the service.
build-binaries: build-protos
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) \
	-ldflags "-X main.gitCommitHash=$(GIT_COMMIT) -X main.builtAt=$(BUILD_DATE) \
	-X main.builtBy=$(USER) -X main.builtOn=$(BUILT_ON)" \
	-o $(BINARY_DIR)/$(BINARY_NAME) -v
	mkdir -p $(BINARY_DIR)/config && cp config/*.yaml $(BINARY_DIR)/config/
	cp -r db/schema $(BINARY_DIR)/

# Create a docker image for the service.
docker-image: fetch-protos
	@mkdir -p $(BINARY_DIR)
	docker build -t $(SCHED_DOCKER_IMAGE) -f docker/Dockerfile ../

# Start the Scheduler and its dependencies within a local Docker network.
start:
	bash ../tools/scripts/start-sched.sh

test:
	go clean -testcache && go test ./...

# Publish the Scheduler docker image to Github.
publish:
	docker tag $(SCHED_DOCKER_IMAGE):latest $(SCHED_GHCR_IMAGE):latest
	docker push $(SCHED_GHCR_IMAGE):latest

clean:
	$(GOCLEAN)
	$(GOCLEAN) -testcache
	@docker rmi -f $(SCHED_DOCKER_IMAGE)
	@docker image prune -f 2>/dev/null 1>&2
	@rm -rf $(BINARY_DIR)

.PHONY: build-binaries docker-image clean build-cache