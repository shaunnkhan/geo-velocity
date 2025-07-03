package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shaunnkhan/geo-velocity/internal/geo"
)

func main() {
	// read environment variables
	// address port, default max travel speed, and unit of speed (mph & km/h)
	var addr = flag.String("addr", ":8080", "server port")
	var maxSpeed = flag.Float64("max-speed", 0.0, "a default max allowed speed")
	var unit = flag.String("unit", "km/h", "unit of speed as mph or km/h")

	// setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
	}))

	// new up handler & register http routes
	geoHandler := geo.NewGeoHandler(*maxSpeed, *unit, geo.NewMockRepository())

	mux := http.NewServeMux()
	geoHandler.RegisterRoutes(mux)

	// create server
	srv := &http.Server{
		Addr:         *addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// start server
	go func() {
		logger.Info("starting server", "addr", *addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// setup interrupt signal shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	stop()

	logger.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server exited")
}
