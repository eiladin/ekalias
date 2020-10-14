package main

import (
	"os"

	"github.com/eiladin/ekalias/cmd"
)

var version = "dev"

func main() {
	cmd.Execute(version, os.Args[1:])
}
