// Command wisp is the Wisp database client launcher and CLI.
package main

import (
	"flag"
	"fmt"
	"os"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Bool("dev", false, "development mode (used by make dev)")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}
}
