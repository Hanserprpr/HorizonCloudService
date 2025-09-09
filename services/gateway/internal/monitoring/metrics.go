package monitoring

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP请求总数计数器
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP请求持续时间直方图
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// 当前活跃连接数
	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being served",
		},
	)

	// 微服务健康状态
	serviceHealthStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_health_status",
			Help: "Health status of backend services (1=healthy, 0=unhealthy)",
		},
		[]string{"service_name"},
	)
)

// Init 初始化监控指标
func Init() {
	// 注册指标
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestsInFlight)
	prometheus.MustRegister(serviceHealthStatus)
}

// PrometheusMiddleware Prometheus指标收集中间件
func PrometheusMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 跳过metrics端点本身
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		
		// 增加活跃请求数
		httpRequestsInFlight.Inc()
		
		// 处理请求
		c.Next()
		
		// 减少活跃请求数
		httpRequestsInFlight.Dec()
		
		// 计算请求持续时间
		duration := time.Since(start).Seconds()
		
		// 记录指标
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		status := strconv.Itoa(c.Writer.Status())
		
		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
	})
}

// MetricsHandler 返回Prometheus指标处理器
func MetricsHandler() gin.HandlerFunc {
	handler := promhttp.Handler()
	return gin.WrapH(handler)
}

// UpdateServiceHealth 更新服务健康状态
func UpdateServiceHealth(serviceName string, isHealthy bool) {
	value := float64(0)
	if isHealthy {
		value = 1
	}
	serviceHealthStatus.WithLabelValues(serviceName).Set(value)
}