###############################################################################
##                              Build stage
###############################################################################
FROM ghcr.io/hpinc/krypton/krypton-go-builder AS build-env

ADD . /go/src/protos
WORKDIR /go/src/protos

# Generate go code from the scheduler protos.
RUN make schedprotos

###############################################################################
##                              Packaging stage
###############################################################################
FROM scratch

WORKDIR /protos
COPY --from=build-env /go/src/protos/*.pb.go /protos/
COPY --from=build-env /go/src/protos/*.proto /protos/