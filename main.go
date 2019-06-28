package main

import (
	"github.com/ellcrys/mother/cmd"
)

var (
	version   = ""
	commit    = ""
	date      = ""
	goversion = "go1.10.4"
)

func main() {
	cmd.BuildVersion = version
	cmd.BuildCommit = commit
	cmd.BuildDate = date
	cmd.GoVersion = goversion
	cmd.Execute()
}
