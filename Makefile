.PHONY: install clean test image push-image

IMAGE := xingse/event-forwarder-gelf

all: event-forwarder-gelf

event-forwarder-gelf:
	go build

clean:
	go clean ./...

test:
	go test -v ./...

image:
	docker build -t $(IMAGE) .

push-image:
	docker push $(IMAGE)
