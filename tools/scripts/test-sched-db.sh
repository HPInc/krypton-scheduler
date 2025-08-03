#!/bin/bash
NETWORK_BASE="krypton-net"
PROJECT_NAME="schedlocal"
SCHED_CONTAINER_NAME="scheduler"
DATABASE_CONTAINER_NAME="scheduler-db".$PROJECT_NAME
SCHED_IMAGE_NAME="krypton-scheduler"
DATABASE_IMAGE_NAME="krypton-sched-db"
NETWORK="$PROJECT_NAME"_"$NETWORK_BASE"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

if [[ -z "${CASSANDRA_PASSWORD}" ]]; then
    echo -n -e "${RED}Please specify the CASSANDRA_PASSWORD environment variable.${NC}"
    echo
    exit 1
fi

retval=$(docker inspect -f "{{.State.Running}}" $DATABASE_CONTAINER_NAME)
if [ "${retval[0]}" != true ]; then
    echo -e "${RED}Krypton scheduler database service is not started${NC}";
    exit 1
fi
docker ps --filter name=$DATABASE_CONTAINER_NAME

pwd="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
echo $pwd

# Determine the IP address of the scheduler container.
DB_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' \
$DATABASE_CONTAINER_NAME)

export SCHEMA_LOCATION="$pwd/../../service/db/schema"
export SCHEDULER_CONFIG_LOCATION="$pwd/../../service/config/config.yml"
export SCHEDULER_REGISTERED_SERVICE_CONFIG_FILE="$pwd/../../service/config/registered_services.yaml"
export SCHEDULER_DB_PASSWORD=${CASSANDRA_PASSWORD}
export SCHEDULER_DB_HOSTS=$DB_IP

go clean -testcache
CGO_ENABLED=0 go test -v ./...
