package main

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

func RLMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	rl := NewIPRateLimiter(r, b)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !rl.Allow(ip) {
			c.JSON(429, gin.H{"error": "Rate limit"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (i *IPRateLimiter) Allow(ip string) bool {

	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if exists {
		return limiter.Allow()
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	if limiter, exists := i.ips[ip]; exists {
		return limiter.Allow()
	}

	newLimiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = newLimiter
	return newLimiter.Allow()
}
