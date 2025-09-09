package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
	"gateway/config"
)

var redisClient *redis.Client
var limiter *rate.Limiter

// InitRedis 初始化Redis连接
func InitRedis(cfg struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}
}

// Middleware 限流中间件
func Middleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 获取客户端IP
		clientIP := c.ClientIP()
		
		// Redis key
		key := fmt.Sprintf("rate_limit:%s", clientIP)
		
		ctx := context.Background()
		
		// 使用Redis实现滑动窗口限流
		// 每分钟允许60个请求
		windowSize := time.Minute
		maxRequests := int64(60)
		
		now := time.Now().Unix()
		pipeline := redisClient.Pipeline()
		
		// 删除超过窗口期的记录
		pipeline.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(now-int64(windowSize.Seconds()), 10))
		
		// 添加当前请求时间戳
		pipeline.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
		
		// 统计窗口期内的请求数量
		pipeline.ZCard(ctx, key)
		
		// 设置key过期时间
		pipeline.Expire(ctx, key, windowSize)
		
		results, err := pipeline.Exec(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"message": "Rate limit check failed",
			})
			c.Abort()
			return
		}
		
		// 获取当前请求数量
		count := results[2].(*redis.IntCmd).Val()
		
		if count > maxRequests {
			c.Header("X-RateLimit-Limit", strconv.FormatInt(maxRequests, 10))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(now+int64(windowSize.Seconds()), 10))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code": 429,
				"message": "Rate limit exceeded",
			})
			c.Abort()
			return
		}
		
		// 设置响应头
		remaining := maxRequests - count
		c.Header("X-RateLimit-Limit", strconv.FormatInt(maxRequests, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(now+int64(windowSize.Seconds()), 10))
		
		c.Next()
	})
}