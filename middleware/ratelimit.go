package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	visitors map[string]*rateInfo
	mu       sync.RWMutex
	rate     int
	window   time.Duration
}

type rateInfo struct {
	tokens    int
	lastRefill time.Time
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*rateInfo),
		rate:     rate,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for k, v := range rl.visitors {
			if time.Since(v.lastRefill) > rl.window*2 {
				delete(rl.visitors, k)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		rl.mu.Lock()
		v, exists := rl.visitors[key]
		if !exists {
			v = &rateInfo{tokens: rl.rate, lastRefill: time.Now()}
			rl.visitors[key] = v
		}
		if time.Since(v.lastRefill) >= rl.window {
			v.tokens = rl.rate
			v.lastRefill = time.Now()
		}
		if v.tokens <= 0 {
			rl.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		v.tokens--
		rl.mu.Unlock()
		c.Next()
	}
}

var GlobalLimiter = NewRateLimiter(100, time.Minute)
