package main

import (
	"github.com/alecthomas/kong"
	"github.com/johnhoman/dinghy/internal/path"
	"github.com/spf13/afero"
	"io"
	"os"
)

var commandLine struct {
	Build cmdBuild `kong:"cmd"`
}

func Main() {
	cmd := kong.Parse(&commandLine, kong.Name("dinghy"))
	wd, err := os.Getwd()
	cmd.FatalIfErrorf(err, "failed to get working directory")

	cmd.BindTo(os.Stdout, (*io.Writer)(nil))
	cmd.BindTo(afero.NewOsFs(), (*afero.Fs)(nil))
	cmd.BindTo(path.NewFSPath(afero.NewOsFs(), wd), (*path.Path)(nil))
	cmd.FatalIfErrorf(cmd.Run())
}

func main() { Main() }
