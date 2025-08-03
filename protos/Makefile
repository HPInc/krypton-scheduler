SCHEDULER_PROTOS_DOCKER_IMAGE=krypton-schedulerprotos
SCHEDULER_GHCR_IMAGE=ghcr.io/hpinc/krypton/$(SCHEDULER_PROTOS_DOCKER_IMAGE)

PROTOS_DIR=.
PROTOC_PATH=/usr/local/bin
PROTOC_CMD=protoc
PROTOC_BUILD=$(PROTOC_PATH)/$(PROTOC_CMD)

all: docker-image

build-protos: dstsprotos schedprotos

schedprotos:
	$(PROTOC_BUILD) -I . -I $(PROTOS_DIR) \
	--go_out=paths=source_relative:$(PROTOS_DIR) \
	--go-grpc_out=paths=source_relative:$(PROTOS_DIR) \
	$(PROTOS_DIR)/*.proto

dstsprotos:
	make -C dstsprotos fetch-protos

docker-image:
	docker build -t $(SCHEDULER_PROTOS_DOCKER_IMAGE) -f Dockerfile .

tag:
	docker tag $(SCHEDULER_PROTOS_DOCKER_IMAGE):latest $(SCHEDULER_GHCR_IMAGE):latest

publish: docker-image tag
	docker push $(SCHEDULER_GHCR_IMAGE):latest

clean:
	docker rmi -f $(SCHEDULER_PROTOS_DOCKER_IMAGE)

.PHONY: build-proto schedprotos dstsprotos docker-image tag publish clean