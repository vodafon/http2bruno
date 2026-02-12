package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	flagOp         = flag.String("o", "request", "operaton. collection|folder|request")
	flagCollection = flag.String("c", "", "collection name")
	flagFolder     = flag.String("f", "", "folder name")
)

func main() {
	flag.Parse()

	switch *flagOp {
	case "collection":
		err := DoStructure(*flagCollection, *flagFolder)
		if err != nil {
			raiseError(err)
		}
	case "folder":
		err := DoFolder(*flagFolder, ".")
		if err != nil {
			raiseError(err)
		}
	default:
		raiseError(fmt.Errorf("Invalid -o flag: %q", *flagOp))
	}
}

func raiseError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
