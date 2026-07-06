package main

import "github.com/martinbhatta/ctrl/internal/cli"

var version = "dev"

func main() {
	cli.Execute(version)
}
