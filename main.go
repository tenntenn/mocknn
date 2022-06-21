package main

import (
	_ "embed"
	"os"

	"github.com/tenntenn/mocknn/internal"
)

//go:embed version.txt
var version string

func main() {
	os.Exit(internal.Main(version, os.Args[1:]))
}
