package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	futil "github.com/contribsys/faktory/util"
	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/core"
	"github.com/contribsys/sparq/util"
)

func logPreamble() {
	log.SetFlags(0)
	log.Println(sparq.Name, sparq.Version)
	log.Printf("Copyright © %d Contributed Systems LLC", time.Now().Year())
	log.Println("Licensed under the GNU Affero Public License 3.0")
}

func main() {
	logPreamble()

	opts := ParseArguments()
	util.InitLogger(opts.LogLevel)
	futil.InitLogger(opts.LogLevel)
	util.Debugf("Options: %v", opts)

	s, err := core.NewService(opts)
	if err != nil {
		util.Error("Unable to start Sparq", err)
		return
	}

	go HandleSignals(s)

	// does not return until shutdown
	err = s.Run()
	if err != nil {
		util.Error("Error running Sparq", err)
		return
	}
}

func ParseArguments() core.Options {
	host := os.Getenv("SPARQ_HOSTNAME")
	if host == "" {
		host = "localhost.dev"
	}

	defaults := core.Options{
		Binding:          "localhost:9494",
		Hostname:         host,
		LogLevel:         "info",
		ConfigDirectory:  ".",
		StorageDirectory: ".",
	}

	flag.Usage = help
	flag.StringVar(&defaults.Binding, "b", "localhost:9494", "Network binding")
	flag.StringVar(&defaults.LogLevel, "l", "info", "Logging level (error, warn, info, debug)")

	// undocumented on purpose, we don't want people changing these if possible
	flag.StringVar(&defaults.StorageDirectory, "d", ".", "Storage directory")
	flag.StringVar(&defaults.ConfigDirectory, "c", ".", "Config directory")
	versionPtr := flag.Bool("v", false, "Show version")
	flag.Parse()

	if *versionPtr {
		os.Exit(0)
	}

	return defaults
}

func help() {
	log.Println("-h [hostname]\tInstance hostname, default: localhost.dev")
	log.Println("-b [binding]\tNetwork binding (use :9494 to listen on all interfaces), default: localhost:9494")
	log.Println("-l [level]\tSet logging level (error, warn, info, debug), default: info")
	log.Println("-v\t\tShow version and license information")
	log.Println("-h\t\tThis help screen")
}

var (
	Term os.Signal = syscall.SIGTERM
	Hup  os.Signal = syscall.SIGHUP
	Info os.Signal = syscall.SIGTTIN

	SignalHandlers = map[os.Signal]func(*core.Service){
		Term:         exit,
		os.Interrupt: exit,
		// Hup:          reload,
		Info: threadDump,
	}
)

func HandleSignals(s *core.Service) {
	signals := make(chan os.Signal, 1)
	for k := range SignalHandlers {
		signal.Notify(signals, k)
	}

	for {
		sig := <-signals
		util.Debugf("Received signal: %v", sig)
		funk := SignalHandlers[sig]
		funk(s)
	}
}

func exit(s *core.Service) {
	util.Infof("%s shutting down", sparq.Name)
	s.Close()
}

func threadDump(s *core.Service) {
	util.DumpProcessTrace()
}
