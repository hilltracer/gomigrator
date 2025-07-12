package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hilltracer/gomigrator/internal/config"
	"github.com/hilltracer/gomigrator/internal/logger"
)

var (
	configFile string
	logLevel   string
)

func init() {
	flag.StringVar(&configFile, "config", "configs/config.yaml", "Path to configuration file")
	flag.StringVar(&logLevel, "log-level", "", "Override log level from config (debug|info|error)")
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage: %s [global flags] <command> [args]\n\n", os.Args[0])
		fmt.Fprintln(out, "Commands:")
		fmt.Fprintln(out, "  create <name>   Generate a new migration file")
		fmt.Fprintln(out, "  up              Apply pending migrations")
		fmt.Fprintln(out, "  down            Rollback last migration")
		fmt.Fprintln(out, "  redo            Rollback + re-apply last migration")
		fmt.Fprintln(out, "  status          Show migration status table")
		fmt.Fprintln(out, "  dbversion       Print current DB version")
		fmt.Fprintln(out, "  version         Print gomigrator version")
		fmt.Fprintln(out, "  help            Print this message")
		flag.PrintDefaults()
	}
}

func main() { os.Exit(run()) }

func run() int {
	/* ---------- parse only global flags ---------- */
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		return 1
	}
	cmd := flag.Arg(0)

	/* ---------- config ---------- */
	cfg, err := config.New(configFile) // ← with ExpandEnv
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		return 1
	}
	if logLevel != "" {
		cfg.Logger.Level = logLevel
	}

	logg := logger.New(cfg.Logger.Level)
	logg.Debug("configuration loaded from " + configFile)

	switch cmd {
	case "help":
		flag.Usage()
	case "version":
		printVersion()
	case "create", "up", "down", "redo", "status", "dbversion":
		logg.Info("command " + cmd + " is not implemented yet (stub)")
	default:
		logg.Error("unknown command: " + cmd)
		flag.Usage()
		return 1
	}

	return 0
}
