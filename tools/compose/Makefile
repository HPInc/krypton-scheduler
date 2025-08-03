# we assume all test machines have either docker-compose or
# the compose plugin in docker command.
DOCKER_COMPOSE=docker-compose
HAS_DOCKER_COMPOSE := $(shell command -v docker-compose 2> /dev/null)
ifndef HAS_DOCKER_COMPOSE
  DOCKER_COMPOSE=docker compose
  DOCKER_COMPOSE_QUIET_PULL=--quiet-pull
endif

include .env

generate-keys:
	./create_scheduler_keys.sh

test: generate-keys start
	$(DOCKER_COMPOSE) -p$(PROJECT) \
		-f docker-compose-test.yml up $(DOCKER_COMPOSE_QUIET_PULL) \
		--exit-code-from $(TEST)

start:
	$(DOCKER_COMPOSE) -p$(PROJECT) \
		-f docker-compose-deps.yml -f docker-compose.yml $(DOCKER_COMPOSE_QUIET_PULL) up -d

stop:
	-$(DOCKER_COMPOSE) -p$(PROJECT) \
		-f docker-compose-deps.yml -f docker-compose.yml down
	-docker rm $(CLI).$(DOMAIN)

clean: stop
	-docker ps -aqf status=exited | xargs docker rm # remove exited
.PHONY: start stop test clean
