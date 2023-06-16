package main

import (
	"io"
	"os"

	"github.com/alecthomas/kong"
)

var commandLine struct {
	Build cmdBuild `kong:"cmd"`
}

func Main() {
	cmd := kong.Parse(&commandLine, kong.Name("dinghy"))

	cmd.BindTo(os.Stdout, (*io.Writer)(nil))
	cmd.FatalIfErrorf(cmd.Run())
}

func main() { Main() }
