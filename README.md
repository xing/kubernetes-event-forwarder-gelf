# event-forwarder-gelf

Forward Kubernetes Events using the Graylog Extended Log Format.

## Usage

    Usage:
      event-forwarder-gelf [OPTIONS]

    Application Options:
      -v, --verbose= Show verbose debug information [$VERBOSE]
          --host=    Graylog TCP endpoint host [$GRAYLOG_HOST]
          --port=    Graylog TCP endpoint port [$GRAYLOG_PORT]
          --cluster= Name of this cluster [$CLUSTER]

    Help Options:
      -h, --help     Show this help message

Run the pre-built image [`xingse/event-forwarder-gelf`] locally (with
local permission):

    echo CLUSTER=cluster-name >> .env
    echo GRAYLOG_HOST=graylog >> .env
    echo GRAYLOG_PORT=12222   >> .env
    docker run --env-file=.env xingse/event-forwarder-gelf

## Deployment

Run this controller on Kubernetes with the following commands:

    kubectl create serviceaccount event-forwarder-gelf \
      --namespace=kube-system

    kubectl create clusterrole xing:controller:event-forwarder-gelf \
      --verb=get,watch,list \
      --resource=events

    kubectl create clusterrolebinding xing:controller:event-forwarder-gelf \
      --clusterrole=xing:controller:event-forwarder-gelf \
      --serviceaccount=kube-system:event-forwarder-gelf

    kubectl run event-forwarder-gelf \
      --image=xingse/event-forwarder-gelf \
      --env=CLUSTER=cluster-name \
      --env=GRAYLOG_HOST=graylog \
      --env=GRAYLOG_PORT=12222 \
      --serviceaccount=event-forwarder-gelf

## Development

This project uses go modules introduced by [go 1.11][go-modules]. Please put the
project somewhere outside of your GOPATH to make go automatically recogninze
this.

All build and install steps are managed in the [Makefile](Makefile). `make test`
will fetch external dependencies, compile the code and run the tests. If all
goes well, hack along and submit a pull request. You might need to run the `go
mod tidy` after updating dependencies.


[`xingse/event-forwarder-gelf`]: https://hub.docker.com/r/xingse/event-forwarder-gelf
[go-modules]: https://github.com/golang/go/wiki/Modules