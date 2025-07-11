package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hilltracer/gomigrator/internal/app"
	"github.com/hilltracer/gomigrator/internal/config"
	"github.com/hilltracer/gomigrator/internal/logger"
	"github.com/hilltracer/gomigrator/internal/storage"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.yaml", "Path to configuration file")
}

func main() {
	os.Exit(run())
}

func run() int {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return 0
	}

	cfg, err := config.NewConfig(configFile)
	if err != nil {
		fmt.Printf("failed to read config: %v\n", err)
		return 1
	}

	logg := logger.New(cfg.Logger.Level)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	var storage storage.Repository
	switch cfg.Storage.Type {
	case "sql":
		pwd := os.Getenv(cfg.Storage.PG.PasswordEnv)
		dsn := fmt.Sprintf(
			"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.Storage.PG.User, pwd, cfg.Storage.PG.Host, cfg.Storage.PG.Port, cfg.Storage.PG.DBName, cfg.Storage.PG.SSLMode,
		)
		pgStore, err := sqlstorage.Connect(ctx, dsn)
		if err != nil {
			logg.Error("db connect: " + err.Error())
			return 1
		}
		storage = pgStore
	default:
		storage = memorystorage.New() // no needed
	}

	migrator := app.New(logg, storage) // not used

	logg.Info("gomigrator is running...")

	<-ctx.Done()

	return 0
}
