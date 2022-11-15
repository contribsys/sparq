package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	futil "github.com/contribsys/faktory/util"
	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/core"
	"github.com/contribsys/sparq/util"
)

func runExec(args []string) {
	logPreamble()
	opts := parseArguments(args)
	util.InitLogger(opts.LogLevel)
	futil.InitLogger(opts.LogLevel)
	util.Debugf("Options: %v", opts)

	s, err := core.NewService(opts)
	if err != nil {
		util.Error("Unable to start Sparq", err)
		return
	}

	go handleSignals(s)

	// does not return until shutdown
	err = s.Run()
	if err != nil {
		util.Error("Error running Sparq", err)
		return
	}
}

func parseArguments(args []string) core.Options {
	flags := flag.NewFlagSet("run", flag.ExitOnError)
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

	flags.Usage = runHelp
	flags.StringVar(&defaults.Hostname, "h", "localhost.dev", "Instance hostname")
	flags.StringVar(&defaults.Binding, "b", "localhost:9494", "Network binding")
	flags.StringVar(&defaults.LogLevel, "l", "info", "Logging level (error, warn, info, debug)")

	// undocumented on purpose, we don't want people changing these if possible
	flags.StringVar(&defaults.StorageDirectory, "d", ".", "Storage directory")
	flags.StringVar(&defaults.ConfigDirectory, "c", ".", "Config directory")
	err := flags.Parse(args)
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}

	return defaults
}

func runHelp() {
	log.Println(`
The "run" command starts the Sparq server.

Usage:

	sparq run -h localhost.dev -b localhost:9494
	sparq run -h example.social -b :8080 -l debug

Arguments:

	-h [hostname]\tInstance hostname, default: localhost.dev
	-b [binding]\tNetwork binding (use :9494 to listen on all interfaces), default: localhost:9494
	-l [level]\tSet logging level (error, warn, info, debug), default: info`)
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

func handleSignals(s *core.Service) {
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
