package main

import (
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/johnhoman/dinghy/internal/path"
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

		f2, err := os.Create("timing.txt")
		cmd.FatalIfErrorf(err, "failed to open file")
		defer func() {
			cmd.FatalIfErrorf(f2.Close())
		}()
		defer func() {
			for _, value := range path.ReqTiming {
				_, err = io.WriteString(f2, fmt.Sprintf("%s: %s\n", value.Duration, strings.TrimPrefix(value.URL, "https://github.com/")))
				cmd.FatalIfErrorf(err)
			}
		}()
	}

	cmd.BindTo(os.Stdout, (*io.Writer)(nil))
	cmd.FatalIfErrorf(cmd.Run())
}

func main() { Main() }
