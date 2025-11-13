package main

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type QueryResponse struct {
	Threat bool   `json:"threat"`
	Feed   string `json:"feed"`
}

func queryHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		ip := c.Query("ip")

		resp := QueryResponse{
			Threat: true,
			Feed:   "malicious_ips",
		}

		slog.Info("query ip", "ip address", ip)

		c.IndentedJSON(http.StatusOK, resp)
	}
}
