package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

type RateLimiter interface {
	Allow() bool
}

type IngestPayload struct {
	Feed    string   `json:"feed"`
	Entries []string `json:"entries"`
}

func ingestHandler() gin.HandlerFunc {

	return func(c *gin.Context) {
		c.ClientIP()
		var payload IngestPayload

		if err := c.BindJSON(&payload); err != nil {
			return
		}

		slog.Info("Ingested feed", "feed_name", payload.Feed, "feed_size", len(payload.Entries))

		c.Status(200)
	}
}
