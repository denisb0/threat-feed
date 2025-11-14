package main

import (
	"context"
	"errors"

	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ThreatWriter interface {
	Ingest(context.Context, IngestPayload) error
}

type RateLimiter interface {
	Allow() bool
}

type IngestPayload struct {
	Feed    string   `json:"feed"`
	Entries []string `json:"entries"`
}

func ingestHandler(tw ThreatWriter) gin.HandlerFunc {

	return func(c *gin.Context) {
		var payload IngestPayload

		if err := c.BindJSON(&payload); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			slog.Error("payload error", "error", err)
			return
		}

		if err := tw.Ingest(c.Request.Context(), payload); err != nil {
			c.AbortWithError(http.StatusInternalServerError, errors.New("threat storage error"))
			return
		}

		slog.Info("Ingested feed", "feed_name", payload.Feed, "feed_size", len(payload.Entries))

		c.Status(http.StatusOK)
	}
}
