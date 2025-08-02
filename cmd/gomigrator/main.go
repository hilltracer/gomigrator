package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hilltracer/gomigrator/internal/config"
	"github.com/hilltracer/gomigrator/internal/logger"
	"github.com/hilltracer/gomigrator/pkg/gomigrator"
)

var (
	configFile    string
	logLevel      string
	migrationsDir string
)

func init() {
	flag.StringVar(&configFile, "config", "configs/config.yaml", "Path to configuration file (YAML)")
	flag.StringVar(&logLevel, "log-level", "info", "Override log level from config (debug|info|error)")
	flag.StringVar(&migrationsDir, "dir", "migrations", "Directory for SQL migration files")
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage:\n")
		fmt.Fprintf(out, "  %s [flags] [dsn] <command>\n\n", os.Args[0])
		fmt.Fprintln(out, "Flags:")
		flag.PrintDefaults()

		fmt.Fprintln(out, "\nDsn:")
		fmt.Fprintln(out, "  Optional. PostgreSQL connection string in the form:")
		fmt.Fprintln(out, "  \"host=... port=... user=... password=... dbname=... sslmode=...\"")
		fmt.Fprintln(out, "  If omitted, dsn is loaded from the config file.")

		fmt.Fprintln(out, "\nCommand:")
		fmt.Fprintln(out, "  create <name>      Generate a new migration file (no DB connection needed)")
		fmt.Fprintln(out, "  up                 Apply all pending migrations")
		fmt.Fprintln(out, "  down               Rollback the last applied migration")
		fmt.Fprintln(out, "  redo               Rollback and re-apply the last migration")
		fmt.Fprintln(out, "  status             Print the status of all migrations")
		fmt.Fprintln(out, "  dbversion          Show the current DB version (or 0 if none)")
		fmt.Fprintln(out, "  version            Print gomigrator version")
		fmt.Fprintln(out, "  help               Print this help message")

		fmt.Fprintln(out, "\nEnvironment:")
		fmt.Fprintln(out, "  You can use environment variables in the config file using ${VAR} syntax.")
		fmt.Fprintln(out, "  Available: PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DB, PG_SSLMODE")
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

	switch cmd {
	case "help":
		flag.Usage()

	case "version":
		printVersion()

	case "create":
		// args offset depends on whether DSN was passed
		var nameIdx int
		if dsn == "" { // create <name>
			nameIdx = 1
		} else { // <DSN> create <name>
			nameIdx = 2
		}
		if len(args) <= nameIdx {
			logg.Error("usage: gomigrator [flags] [DSN] create <name>")
			return 1
		}

		filePath, err := gomigrator.Create(migrationsDir, args[nameIdx])
		if err != nil {
			logg.Error("create: " + err.Error())
			return 1
		}
		abs, _ := filepath.Abs(filePath)
		logg.Info("Created migration: " + abs)

	case "status", "up", "down", "redo", "dbversion":
		status := performDBOps(cmd, cfg.Storage.DSN, logg)
		if status != 0 {
			return status
		}
	default:
		logg.Error("unknown command: " + cmd)
		flag.Usage()
		return 1
	}
	return 0
}

func performDBOps(cmd string, dsn string, logg *logger.Logger) int {
	// mig, err := GoMigrator.NewFromDSN(context.Background(), dsn, migrationsDir)
	mig, err := gomigrator.New(context.Background(), gomigrator.Config{
		DSN: dsn,
		Dir: migrationsDir,
	})
	if err != nil {
		logg.Error("db connect: " + err.Error())
		return 1
	}
	defer mig.Close()

	switch cmd {
	case "status":
		statuses, err := mig.Status(context.Background())
		if err != nil {
			logg.Error(err.Error())
			return 1
		}
		if len(statuses) == 0 {
			logg.Info("no migrations found")
			return 0
		}

		for _, s := range statuses {
			fmt.Printf("%-14d %v\n", s.Version, s.IsApplied)
		}

	case "dbversion":
		v, err := mig.DBVersion(context.Background())
		if err != nil {
			logg.Error(err.Error())
			return 1
		}
		fmt.Println(v)

	case "up":
		if err := mig.Up(context.Background()); err != nil {
			logg.Error(err.Error())
			return 1
		}
		logg.Info("migrations applied")

	case "down":
		if err := mig.Down(context.Background()); err != nil {
			logg.Error(err.Error())
			return 1
		}
		logg.Info("migration rolled back")

	case "redo":
		if err := mig.Redo(context.Background()); err != nil {
			logg.Error(err.Error())
			return 1
		}
		logg.Info("migration redone")
	}
	return 0
}
