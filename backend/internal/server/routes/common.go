package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	basemiddleware "github.com/Wei-Shaw/sub2api/internal/middleware"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RegisterCommonRoutes 注册通用路由（健康检查、状态等）
func RegisterCommonRoutes(r *gin.Engine, affiliateLanding *handler.AffiliateLandingHandler, redisClient *redis.Client) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	registerAffiliateLandingRoute(r, affiliateLanding, redisClient)

	// Claude Code 遥测日志（忽略，直接返回200）
	r.POST("/api/event_logging/batch", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup status endpoint (always returns needs_setup: false in normal mode)
	// This is used by the frontend to detect when the service has restarted after setup
	r.GET("/setup/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"needs_setup": false,
				"step":        "completed",
			},
		})
	})
}

func registerAffiliateLandingRoute(r *gin.Engine, affiliateLanding *handler.AffiliateLandingHandler, redisClient *redis.Client) {
	handle := func(c *gin.Context) {
		if affiliateLanding == nil {
			c.Status(http.StatusNotFound)
			return
		}
		affiliateLanding.Redirect(c)
	}
	if redisClient == nil {
		r.GET("/r/:affCode", handle)
		return
	}

	limiter := basemiddleware.NewRateLimiter(redisClient)
	failClose := basemiddleware.RateLimitFailClose
	r.GET("/r/:affCode",
		limiter.LimitWithOptions("affiliate-landing-ip", 60, time.Minute, basemiddleware.RateLimitOptions{
			FailureMode: failClose,
			KeyFunc:     func(c *gin.Context) string { return ip.GetTrustedClientIP(c) },
		}),
		limiter.LimitWithOptions("affiliate-landing-fingerprint", 30, time.Minute, basemiddleware.RateLimitOptions{
			FailureMode: failClose,
			KeyFunc: func(c *gin.Context) string {
				agent := c.Request.UserAgent()
				if len(agent) > 512 {
					agent = agent[:512]
				}
				return ip.GetTrustedClientIP(c) + "\n" + agent
			},
		}),
		limiter.LimitWithOptions("affiliate-landing-code", 300, time.Minute, basemiddleware.RateLimitOptions{
			FailureMode: failClose,
			KeyFunc: func(c *gin.Context) string {
				return strings.ToUpper(strings.TrimSpace(c.Param("affCode")))
			},
		}),
		handle,
	)
}
