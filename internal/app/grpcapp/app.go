package grpcapp

import (
	"fmt"
	"log"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	authgrpc "github.com/mo0Oonnn/sso/internal/grpc/auth"
)

type App struct {
	logger      *slog.Logger
	gRPCServer  *grpc.Server
	authService authgrpc.Auth
	port        int
}

func New(logger *slog.Logger, authService authgrpc.Auth, port int) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer, authService)

	return &App{
		logger:     logger,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func (a *App) Run() error {
	const operation = "grpcapp.Run"

	logger := a.logger.With(slog.String("operation", operation))

	logger.Info("starting gRPC server", slog.Int("port", a.port))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	logger.Info("gRPC server is running", slog.Int("port", a.port))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return nil
}

func (a *App) Stop() {
	const operation = "grpcapp.Stop"

	a.logger.With(slog.String("operation", operation)).
		Info("stopping gRPC server")

	a.gRPCServer.GracefulStop()
}
