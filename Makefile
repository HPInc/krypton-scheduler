include common.mk

# Create a docker image for the service.
docker-image:
	make -C protos docker-image
	make -C service docker-image

test:
	make -C tools/compose test

publish: docker-image
	make -C service publish
