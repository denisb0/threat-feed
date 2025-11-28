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

const (
	defaultRateLimit float64 = 1000
	defaultRateBurst int     = 10
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	err := godotenv.Load()
	if err != nil {
		slog.Warn(fmt.Sprintln("env file not loaded", err))
	}

	rateLimit := defaultRateLimit
	if rlEnv := os.Getenv("RATE_LIMIT"); rlEnv != "" {
		rateLimit, err = strconv.ParseFloat(rlEnv, 64)
		if err != nil {
			panic(fmt.Sprintln("invalid RATE_LIMIT value", err))
		}
	} else {
		slog.Info("rate limit value not set, using default", "rate_limit", rateLimit)
	}

	rateBurst := defaultRateBurst
	if rbEnv := os.Getenv("RATE_BURST"); rbEnv != "" {
		rateBurst, err = strconv.Atoi(rbEnv)
		if err != nil {
			panic(fmt.Sprintln("invalid RATE_BURST value", err))
		}
	} else {
		slog.Info("rate burst value not set, using default", "rate_burst", rateBurst)
	}

	storage := NewThreatMemStorage()

	router := gin.Default()
	router.GET("/query", queryHandler(storage))
	router.POST("/ingest", RLMiddleware(rate.Limit(rateLimit), rateBurst), ingestHandler(storage))

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Only enable pprof in development mode (when ENABLE_PPROF=true)
	enablePprof := os.Getenv("ENABLE_PPROF") == "true"
	var pprofServer *http.Server

	if enablePprof {
		pprofServer = &http.Server{
			Addr:    ":6060",
			Handler: http.DefaultServeMux,
		}
		slog.Warn("pprof profiling is ENABLED - this should only be used in development", "address", "localhost:6060")
	}

	go func() {
		slog.Info("starting server...", "address", "localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	if enablePprof {
		go func() {
			slog.Info("starting pprof server...", "address", "localhost:6060")
			if err := pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("pprof server error", "error", err)
			}
		}()
	}

	q := SignalHandler(func(sig os.Signal) {
		slog.Info("received signal, shutting down...", "signal", sig)
	}, func(s os.Signal) {
		if err := server.Shutdown(context.Background()); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
		if enablePprof && pprofServer != nil {
			if err := pprofServer.Shutdown(context.Background()); err != nil {
				slog.Error("pprof server shutdown error", "error", err)
			}
		}
	})

	<-q
	slog.Info("shutdown complete")
}
