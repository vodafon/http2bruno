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
	flagBaseDir    = flag.String("base", ".", "base collection folder for request")
	flagEnvFile    = flag.String("e", "environments/base.bru", "environment file")
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
		err := DoFolder(*flagFolder, *flagBaseDir)
		if err != nil {
			raiseError(err)
		}
	case "request":
		err := DoRequest(*flagBaseDir, *flagEnvFile)
		if err != nil {
			raiseError(err)
		}
	default:
		raiseError(fmt.Errorf("invalid -o flag: %q", *flagOp))
	}
}

func raiseError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
