package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mo0Oonnn/sso/internal/app"
	"github.com/mo0Oonnn/sso/internal/config"
	"github.com/mo0Oonnn/sso/internal/lib/logger/slogpretty"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	logger := setupLogger(cfg.Env)

	logger.Info("starting app",
		slog.String("env", cfg.Env),
	)
	logger.Debug("debug messages are enabled")

	application := app.New(logger, cfg.Port, cfg.StoragePath, cfg.TokenTTL)

	go application.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sig := <-stop

	application.GRPCServer.Stop()
	logger.Info("application stopped", slog.String("signal", sig.String()))
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	if env == envLocal {
		logger = setupPrettySlog()
	} else if env == envDev || env == envProd {
		logFile, err := os.Create("logs.log")
		if err != nil {
			log.Fatal("error creating log file:", err)
		}
		logger = slog.New(
			slog.NewJSONHandler(logFile, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}),
		)
	}
	return logger
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
