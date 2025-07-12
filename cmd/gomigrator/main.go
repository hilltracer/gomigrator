package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hilltracer/gomigrator/internal/config"
	"github.com/hilltracer/gomigrator/internal/logger"
	"github.com/hilltracer/gomigrator/internal/sqlstorage"
)

var (
	configFile string
	logLevel   string
)

func init() {
	flag.StringVar(&configFile, "config", "configs/config.yaml", "Path to configuration file (YAML)")
	flag.StringVar(&logLevel, "log-level", "", "Override log level from config (debug|info|error)")
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage:\n")
		fmt.Fprintf(out, "  %s [flags] [DSN] <command>\n\n", os.Args[0])
		fmt.Fprintln(out, "\nFlags:")
		flag.PrintDefaults()

		fmt.Fprintln(out, "  DSN                Optional. PostgreSQL connection string in the form:")
		fmt.Fprintln(out, "                     \"host=... port=... user=... password=... dbname=... sslmode=...\"")
		fmt.Fprintln(out, "                     If omitted, DSN is loaded from the config file.")

		fmt.Fprintln(out, "\nCommands:")
		fmt.Fprintln(out, "  create <name>      Generate a new migration file (no DB connection needed)")
		fmt.Fprintln(out, "  up                 Apply all pending migrations")
		fmt.Fprintln(out, "  down               Rollback the last applied migration")
		fmt.Fprintln(out, "  redo               Rollback and re-apply the last migration")
		fmt.Fprintln(out, "  status             Print the status of all migrations")
		fmt.Fprintln(out, "  dbversion          Show the current DB version")
		fmt.Fprintln(out, "  version            Print gomigrator version")
		fmt.Fprintln(out, "  help               Print this help message")

		fmt.Fprintln(out, "\nEnvironment:")
		fmt.Fprintln(out, "  You can use environment variables in the config file using ${VAR} syntax.")
		fmt.Fprintln(out, "  Examples: LOG_LEVEL, PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, etc.")
	}
}

func main() { os.Exit(run()) }

func run() int {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		flag.Usage()
		return 1
	}

	var (
		dsn string
		cmd string
	)
	if strings.Contains(args[0], "host=") {
		dsn = args[0]
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "missing command after DSN")
			flag.Usage()
			return 1
		}
		cmd = args[1]
	} else {
		cmd = args[0]
	}

	cfg, err := config.New(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		return 1
	}
	if logLevel != "" {
		cfg.Logger.Level = logLevel
	}
	logg := logger.New(cfg.Logger.Level)

	if dsn != "" {
		cfg.Storage.DSN = dsn
		logg.Debug("using DSN from CLI: " + dsn)
	}

	// ---------- CMDs without db connection ----------
	switch cmd {
	case "help":
		flag.Usage()
		return 0
	case "version":
		printVersion()
		return 0
	case "create":
		logg.Info("create command is not implemented yet")
		return 0
	}

	// ---------- CMDs with db connection ----------
	store, err := sqlstorage.Connect(context.Background(), cfg.Storage.DSN)
	if err != nil {
		logg.Error("db connect: " + err.Error())
		return 1
	}
	defer store.Close()

	switch cmd {
	case "status":
		versions, err := store.AppliedVersions(context.Background())
		if err != nil {
			logg.Error("status: " + err.Error())
			return 1
		}
		if len(versions) == 0 {
			fmt.Println("no migrations found")
			return 0
		}
		for v, ok := range versions {
			fmt.Printf("%d\t%v\n", v, ok)
		}
	case "up", "down", "redo", "dbversion":
		logg.Info("command " + cmd + " is not implemented yet")
	default:
		logg.Error("unknown command: " + cmd)
		flag.Usage()
		return 1
	}

	return 0
}
