package app

import (
	"log/slog"
	"time"

	"github.com/mo0Oonnn/sso/internal/app/grpcapp"
	"github.com/mo0Oonnn/sso/internal/services/auth"
	"github.com/mo0Oonnn/sso/internal/storage/sqlite"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(logger *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(logger, storage, storage, storage, tokenTTL)

	gRPCApp := grpcapp.New(logger, authService, grpcPort)

	return &App{
		GRPCServer: gRPCApp,
	}
}
