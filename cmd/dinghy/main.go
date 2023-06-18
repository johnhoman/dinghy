package main

import (
	"io"
	"os"
	"runtime/pprof"

	"github.com/alecthomas/kong"
)

var commandLine struct {
	Build   cmdBuild `kong:"cmd"`
	Profile bool     `kong:"name=pprof"`
}

func Main() {
	cmd := kong.Parse(&commandLine, kong.Name("dinghy"))

	if commandLine.Profile {
		f, err := os.Create("profile.pprof")
		cmd.FatalIfErrorf(err, "failed to open file")
		defer func() {
			cmd.FatalIfErrorf(f.Close())
		}()

		// Start profiling
		cmd.FatalIfErrorf(pprof.StartCPUProfile(f))
		defer pprof.StopCPUProfile()
	}

	cmd.BindTo(os.Stdout, (*io.Writer)(nil))
	cmd.FatalIfErrorf(cmd.Run())
}

func main() { Main() }
