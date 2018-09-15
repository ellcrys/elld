package main

import (
	"github.com/ellcrys/elld/cmd"
)

var (
	version = ""
	commit  = ""
	date    = ""
)

func main() {
	cmd.BuildVersion = version
	cmd.BuildCommit = commit
	cmd.BuildDate = date
	cmd.Execute()
}
