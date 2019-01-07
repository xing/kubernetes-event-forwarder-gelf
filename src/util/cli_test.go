package util

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	os.Args = []string{"bin/test", "--host", "kubernetes.io"}

	var options struct {
		Host string `long:"host" required:"true"`
	}

	ParseArgs(&options)
	flag.Set("logtostderr", "false")

	assert.Equal(t, "kubernetes.io", options.Host)
}

func TestParseArgsWithVerboseButNotSet(t *testing.T) {
	os.Args = []string{"bin/test"}

	var options struct {
		Verbose int `short:"v"`
	}

	ParseArgs(&options)
	flag.Set("logtostderr", "false")

	assert.Equal(t, 0, options.Verbose)
}

func TestParseArgsWithVerboseSet(t *testing.T) {
	os.Args = []string{"bin/test", "-v", "3"}

	var options struct {
		Verbose int `short:"v"`
	}

	ParseArgs(&options)
	flag.Set("logtostderr", "false")

	assert.Equal(t, 3, options.Verbose)
}
