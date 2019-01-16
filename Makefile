.PHONY: install clean test image push-image release perform-release

IMAGE := xingse/event-forwarder-gelf
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

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

release: image
ifneq ($(BRANCH),master)
	$(error release only works from master, currently on '$(BRANCH)')
endif
	$(MAKE) perform-release

TAG = $(shell docker run --rm $(IMAGE) --version | grep -oE "event-forwarder-gelf [^ ]+" | cut -d ' ' -f2)

perform-release:
	git tag $(TAG)
	git push origin $(TAG)
	git push origin master
