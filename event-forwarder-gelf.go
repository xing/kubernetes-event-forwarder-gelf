package main

import (
	"github.com/xing/event-forwarder-gelf/src"
	"github.com/xing/event-forwarder-gelf/src/util"
)

var opts struct {
	Verbose     int    `env:"VERBOSE" short:"v" long:"verbose" description:"Show verbose debug information"`
	GraylogHost string `env:"GRAYLOG_HOST" long:"host" required:"true" description:"Graylog TCP endpoint host"`
	GraylogPort string `env:"GRAYLOG_PORT" long:"port" required:"true" description:"Graylog TCP endpoint port"`
	Cluster     string `env:"CLUSTER" long:"cluster" required:"true" description:"Name of this cluster"`
}

func main() {
	util.ParseArgs(&opts)

	gelfWriter := util.GetGelfWriter(opts.GraylogHost, opts.GraylogPort)
	controller := src.NewController(gelfWriter, opts.Cluster)

	util.InstallSignalHandler(controller.Stop)

	controller.Run()
}
