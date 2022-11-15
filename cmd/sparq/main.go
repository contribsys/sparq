package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/contribsys/sparq"
)

func logPreamble() {
	log.Println(sparq.Name, sparq.Version)
	log.Printf("Copyright © %d Contributed Systems LLC", time.Now().Year())
	log.Println("Licensed under the GNU Affero Public License 3.0")
}

func versionExec(args []string) {
	fs := flag.NewFlagSet("sparq-version", flag.ContinueOnError)
	fs.Usage = usage
	if err := fs.Parse(args); err != nil {
		log.Println(err)
		return
	}

	logPreamble()
}

func usage() {
	log.Println(`
	sparq is a ActivityPub daemon and tools
	
	Usage:
	
		sparq <command> <arguments>
	
	Valid commands are:
	
		sparq run [-h hostname]
		sparq version
		sparq help
	`)
}

func main() {
	log.SetFlags(0)

	cmd := ""
	args := os.Args
	if len(args) > 1 {
		cmd, args = args[1], args[2:]
	} else {
		usage()
		os.Exit(0)
		return
	}

	switch cmd {
	case "version":
		versionExec(args)
	case "run":
		runExec(args)
	default:
		usage()
	}
}
