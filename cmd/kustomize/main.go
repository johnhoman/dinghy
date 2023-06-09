package main

import (
	"github.com/alecthomas/kong"
)

var commandLine struct {
	Build cmdBuild `kong:"cmd"`
}

func main() {
	cmd := kong.Parse(&commandLine, kong.Name("kustomize"))
	cmd.FatalIfErrorf(cmd.Run())
}
