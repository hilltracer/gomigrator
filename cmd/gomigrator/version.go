package main

import "fmt"

var (
	release   = "dev" // ldflags: -X main.release=…
	buildDate = ""    // ldflags: -X main.buildDate=…
	gitHash   = ""    // ldflags: -X main.gitHash=…
)

func printVersion() {
	fmt.Printf("gomigrator %s (%s) %s\n", release, gitHash, buildDate)
}
