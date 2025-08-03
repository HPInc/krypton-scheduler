DOCKER_IMAGE_NAME=krypton-sched-db

all: docker-image

docker-image:
	DOCKER_BUILDKIT=1 docker build -t $(DOCKER_IMAGE_NAME) --rm \
	-f Dockerfile .
	docker image prune -f

clean:
	docker rmi -f $(DOCKER_IMAGE_NAME)
	docker image prune -f