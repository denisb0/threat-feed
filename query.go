package main

import (
	"context"
	"errors"

	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ThreatReader interface {
	Query(context.Context, string) (QueryResponse, error)
}

type QueryResponse struct {
	Threat bool   `json:"threat"`
	Feed   string `json:"feed"`
}

func queryHandler(tr ThreatReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Query("ip")

		response, err := tr.Query(c.Request.Context(), ip)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, errors.New("threat storage error"))
			return
		}
		resp := QueryResponse{
			Threat: true,
			Feed:   response.Feed,
		}

		slog.Info("query ip", "ip address", ip, "threat", response.Threat, "feed", response.Feed)

		c.JSON(http.StatusOK, resp)
	}
}
