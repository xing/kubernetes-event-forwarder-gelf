.PHONY: install

all: event-forwarder-gelf

event-forwarder-gelf:
	go build

clean:
	go clean ./...

test:
	go test ./...

image:
	docker build -t xingse/event-forwarder-gelf .
