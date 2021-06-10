package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/xing/event-forwarder-gelf/src"
	"github.com/xing/event-forwarder-gelf/src/util"
)

// VERSION represents the current version of the release.
const VERSION = "v1.1.0"

var opts struct {
	Version     func() `long:"version" description:"Print version information"`
	Verbose     int    `env:"VERBOSE" short:"v" long:"verbose" description:"Show verbose debug information"`
	GraylogHost string `env:"GRAYLOG_HOST" long:"host" required:"true" description:"Graylog TCP endpoint host"`
	GraylogPort string `env:"GRAYLOG_PORT" long:"port" required:"true" description:"Graylog TCP endpoint port"`
	Cluster     string `env:"CLUSTER" long:"cluster" required:"true" description:"Name of this cluster"`
}

func main() {
	opts.Version = printVersion
	util.ParseArgs(&opts)

	gelfWriter := util.GetGelfWriter(opts.GraylogHost, opts.GraylogPort)
	controller := src.NewController(gelfWriter, opts.Cluster)

	util.InstallSignalHandler(controller.Stop)

	controller.Run()
}

func printVersion() {
	fmt.Printf("event-forwarder-gelf %s %s/%s %s\n", VERSION, runtime.GOOS, runtime.GOARCH, runtime.Version())
	os.Exit(0)
}
