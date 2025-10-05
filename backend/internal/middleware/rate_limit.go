package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	lastSeen time.Time
	count    int
}

var (
	visitors = make(map[string]*visitor) // key: IP, value: visitor info
	mu       sync.RWMutex
)

// RateLimitMiddleware 限制請求頻率，防止 DDoS 攻擊
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 100
	}

	// clean up old visitors every minute
	go cleanupVisitors()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &visitor{
				lastSeen: time.Now(),
				count:    1,
			}
			mu.Unlock()
			c.Next()
			return
		}

		// Check if exceeds rate limit
		if time.Since(v.lastSeen) < time.Minute {
			if v.count >= requestsPerMinute {
				mu.Unlock()
				requestID, _ := c.Get("RequestID")
				c.JSON(http.StatusTooManyRequests, gin.H{
					"code":       http.StatusTooManyRequests,
					"message":    "Too many requests, please try again later",
					"request_id": requestID,
				})
				c.Abort()
				return
			}
			v.count++
		} else {
			v.count = 1
			v.lastSeen = time.Now()
		}
		mu.Unlock()

		c.Next()
	}
}

func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}
