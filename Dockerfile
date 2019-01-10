FROM golang:1.11.2
COPY . /src/
WORKDIR /src/
RUN make clean \
  && make test \
  && make

FROM ubuntu:xenial
COPY --from=0 /src/event-forwarder-gelf /event-forwarder-gelf
ENTRYPOINT ["/event-forwarder-gelf"]
