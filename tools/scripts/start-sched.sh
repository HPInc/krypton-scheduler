#!/bin/bash
NETWORK_BASE="krypton_net"
SCHED_CONTAINER_NAME="scheduler"
DATABASE_CONTAINER_NAME="scheduler-db"
SCHED_IMAGE_NAME="krypton-scheduler"
DATABASE_IMAGE_NAME="krypton-sched-db"
PROJECT_NAME="sched"
NETWORK="$PROJECT_NAME"_"$NETWORK_BASE"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# First check if the required AWS config environment variables are set.
if [[ -z "${AWS_ACCESS_KEY_ID}" ]]; then
    echo -n -e "${RED}Please specify the AWS_ACCESS_KEY_ID environment variable.${NC}"
    echo
    exit 1
fi

if [[ -z "${AWS_SECRET_ACCESS_KEY}" ]]; then
    echo -n -e "${RED}Please specify the AWS_SECRET_ACCESS_KEY environment variable.${NC}"
    echo
    exit 1
fi

if [[ -z "${CASSANDRA_PASSWORD}" ]]; then
    echo -n -e "${RED}Please specify the CASSANDRA_PASSWORD environment variable.${NC}"
    echo
    exit 1
fi

echo -e "${GREEN}Shutting down existing containers and cleaning up network ...${NC}"
docker rm --force $SCHED_CONTAINER_NAME
docker rm --force $DATABASE_CONTAINER_NAME

pwd="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
echo $pwd

# Create a docker network for the Scheduler service.
echo "Setting up network for Scheduler service ..."
docker network inspect $NETWORK >/dev/null 2>&1 || \
    docker network create $NETWORK

# Start up the scheduler database container
docker-compose -f "$pwd/docker-compose-db.yml" -p $PROJECT_NAME up -d

echo "Waiting for scheduler database container to start up ..."
sleep 90
retval=$(docker inspect -f "{{.State.Running}}" $DATABASE_CONTAINER_NAME)
if [ "${retval[0]}" != true ]; then
    echo -e "${RED}Failed to start the Krypton scheduler database service${NC}";
    exit 1
fi
docker ps --filter name=$DATABASE_CONTAINER_NAME

# Deploy the Scheduler service docker container into the network.
echo -e "${GREEN}Starting the Krypton Scheduler service ...${NC}"
docker run -d -p 9990:9990 --net $NETWORK \
-e GRPC_GO_LOG_VERBOSITY_LEVEL=99 -e GRPC_TRACE="all" \
-e GO_DEBUG="http2debug=2" -e GRPC_GO_LOG_SEVERITY_LEVEL="info" \
-e AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}" \
-e AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY}" -e AWS_REGION="us-west-2" \
-e DB_PASSWORD="${CASSANDRA_PASSWORD}" -e TEST_MODE="enabled" \
--name $SCHED_CONTAINER_NAME $SCHED_IMAGE_NAME

echo "Waiting for container to start up ..."
sleep 5
retval=$(docker inspect -f "{{.State.Running}}" $SCHED_CONTAINER_NAME)
if [ "${retval[0]}" != true ]; then
    echo -e "${RED}Failed to start the Krypton Scheduler service${NC}";
    exit 1
fi

docker ps --filter name=$SCHED_CONTAINER_NAME

# Determine the IP address of the Scheduler container.
DB_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' \
$DATABASE_CONTAINER_NAME)
SCHED_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' \
$SCHED_CONTAINER_NAME)

echo -e "${GREEN}Krypton Scheduler has been deployed into the docker network $NETWORK ${NC}"
echo -e " - Krypton Scheduler IP address: $SCHED_IP"
echo -e " - Krypton Scheduler DB IP address: $DB_IP"
