package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(fmt.Sprintln("env file not loaded", err))
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	rateLimit, err := strconv.ParseFloat(os.Getenv("RATE_LIMIT"), 64)
	if err != nil {
		panic(fmt.Sprintln("invalid RATE_LIMIT value", err))
	}

	rateBurst, err := strconv.Atoi(os.Getenv("RATE_BURST"))
	if err != nil {
		panic(fmt.Sprintln("invalid RATE_BURST value", err))
	}

	storage := NewThreatMemStorage()

	router := gin.Default()
	router.GET("/query", queryHandler(storage))
	router.POST("/ingest", RLMiddleware(rate.Limit(rateLimit), rateBurst), ingestHandler(storage))

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	pprofServer := &http.Server{
		Addr:    ":6060",
		Handler: http.DefaultServeMux,
	}

	go func() {
		slog.Info("starting server...", "address", "localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	go func() {
		slog.Info("starting pprof server...", "address", "localhost:6060")
		if err := pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("pprof server error", "error", err)
		}
	}()

	q := SignalHandler(context.Background(), func(sig os.Signal) {
		slog.Info("received signal, shutting down...", "signal", sig)
	}, func(s os.Signal) {
		if err := server.Shutdown(context.Background()); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
		if err := pprofServer.Shutdown(context.Background()); err != nil {
			slog.Error("pprof server shutdown error", "error", err)
		}
	})

	<-q
	slog.Info("shutdown complete")
}
