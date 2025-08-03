###############################################################################
##                              Build stage
###############################################################################
FROM ghcr.io/hpinc/krypton/krypton-go-builder:latest AS builder

# Copy the source code over and build the scheduler binary.
ADD . /go/src/krypton-scheduler
WORKDIR /go/src/krypton-scheduler/service

RUN make gosec build

###############################################################################
##                              Packaging stage
###############################################################################
FROM ghcr.io/hpinc/krypton/krypton-go-base

# set working directory
WORKDIR /go/bin

COPY --from=builder /go/src/krypton-scheduler/bin/scheduler/schedservice .
COPY --from=builder /go/src/krypton-scheduler/service/config/config.yaml .
COPY --from=builder /go/src/krypton-scheduler/service/config/secrets.yaml .
COPY --from=builder /go/src/krypton-scheduler/service/config/registered_services.yaml .
COPY --from=builder /go/src/krypton-scheduler/service/db/schema /go/bin/schema/

# Add AWS Root CA required to connect to the MQTT broker & Cassandra using SSL.
COPY --from=builder /go/src/krypton-scheduler/service/docker/sf-class2-root.crt .
COPY --from=builder /go/src/krypton-scheduler/service/docker/AmazonRootCA1.pem .

USER 1001

# Expose the REST port over which the Scheduler service listens. HTTPS required.
EXPOSE 9990/tcp

# Start up the scheduler service.
CMD ["/go/bin/schedservice"]
