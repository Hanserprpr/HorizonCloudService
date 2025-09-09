package handlers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type HealthHandlerTestSuite struct {
	TestSuite
}

func (s *HealthHandlerTestSuite) TestHealthCheck() {
	// 健康检查端点通常不需要认证
	response := s.MakeRequest("GET", "/health", nil, nil)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "status")
	s.Equal("healthy", result["status"])
	s.Contains(result, "timestamp")
	s.Contains(result, "service")
	s.Equal("file-service", result["service"])
}

func (s *HealthHandlerTestSuite) TestHealthCheckWithAuth() {
	// 测试带认证的健康检查
	response := s.MakeRequest("GET", "/health", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	s.Equal("healthy", result["status"])
}

func (s *HealthHandlerTestSuite) TestLivenessProbe() {
	// Kubernetes liveness probe
	response := s.MakeRequest("GET", "/health/live", nil, nil)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	s.Equal("alive", result["status"])
}

func (s *HealthHandlerTestSuite) TestReadinessProbe() {
	// Kubernetes readiness probe
	response := s.MakeRequest("GET", "/health/ready", nil, nil)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "status")
	s.Contains(result, "checks")
	
	checks := result["checks"].(map[string]interface{})
	s.Contains(checks, "database")
	s.Contains(checks, "storage")
}

func (s *HealthHandlerTestSuite) TestDetailedHealthCheck() {
	// 详细健康检查
	response := s.MakeRequest("GET", "/health/detailed", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "status")
	s.Contains(result, "version")
	s.Contains(result, "uptime")
	s.Contains(result, "components")
	
	components := result["components"].(map[string]interface{})
	s.Contains(components, "database")
	s.Contains(components, "storage")
	s.Contains(components, "cache")
	s.Contains(components, "user_service")
}

func (s *HealthHandlerTestSuite) TestMetrics() {
	// 系统指标端点
	response := s.MakeRequest("GET", "/metrics", nil, nil)
	
	// 验证响应
	s.Equal(http.StatusOK, response.Code)
	
	// 检查是否返回Prometheus格式的指标
	body := response.Body.String()
	s.Contains(body, "# HELP")
	s.Contains(body, "# TYPE")
}

func (s *HealthHandlerTestSuite) TestServiceInfo() {
	// 服务信息端点
	response := s.MakeRequest("GET", "/info", nil, nil)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "name")
	s.Contains(result, "version")
	s.Contains(result, "description")
	s.Contains(result, "build_time")
	s.Contains(result, "git_commit")
	
	s.Equal("file-service", result["name"])
}

func (s *HealthHandlerTestSuite) TestVersionEndpoint() {
	// 版本信息端点
	response := s.MakeRequest("GET", "/version", nil, nil)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "version")
	s.Contains(result, "build_time")
	s.Contains(result, "git_commit")
}

func (s *HealthHandlerTestSuite) TestDatabaseHealthCheck() {
	// 数据库健康检查
	response := s.MakeRequest("GET", "/health/db", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "status")
	s.Contains(result, "response_time")
	s.Contains(result, "connection_count")
	
	if result["status"] == "healthy" {
		s.Contains(result, "version")
		s.Contains(result, "max_connections")
	}
}

func (s *HealthHandlerTestSuite) TestStorageHealthCheck() {
	// 存储健康检查
	response := s.MakeRequest("GET", "/health/storage", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "status")
	s.Contains(result, "response_time")
	
	if result["status"] == "healthy" {
		s.Contains(result, "available_space")
		s.Contains(result, "total_space")
		s.Contains(result, "usage_percentage")
	}
}

func (s *HealthHandlerTestSuite) TestCacheHealthCheck() {
	// 缓存健康检查
	response := s.MakeRequest("GET", "/health/cache", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "status")
	s.Contains(result, "response_time")
	
	if result["status"] == "healthy" {
		s.Contains(result, "connected_clients")
		s.Contains(result, "used_memory")
		s.Contains(result, "hit_rate")
	}
}

func (s *HealthHandlerTestSuite) TestExternalServiceHealthCheck() {
	// 外部服务健康检查
	response := s.MakeRequest("GET", "/health/external", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "services")
	
	services := result["services"].(map[string]interface{})
	s.Contains(services, "user_service")
	
	userService := services["user_service"].(map[string]interface{})
	s.Contains(userService, "status")
	s.Contains(userService, "response_time")
}

func (s *HealthHandlerTestSuite) TestSystemResourcesCheck() {
	// 系统资源检查
	response := s.MakeRequest("GET", "/health/resources", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "cpu")
	s.Contains(result, "memory")
	s.Contains(result, "disk")
	s.Contains(result, "goroutines")
	
	cpu := result["cpu"].(map[string]interface{})
	s.Contains(cpu, "usage_percent")
	
	memory := result["memory"].(map[string]interface{})
	s.Contains(memory, "used")
	s.Contains(memory, "total")
	s.Contains(memory, "usage_percent")
	
	disk := result["disk"].(map[string]interface{})
	s.Contains(disk, "used")
	s.Contains(disk, "total")
	s.Contains(disk, "usage_percent")
}

func (s *HealthHandlerTestSuite) TestServiceStatistics() {
	// 服务统计信息
	response := s.MakeRequest("GET", "/stats", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "requests")
	s.Contains(result, "files")
	s.Contains(result, "storage")
	s.Contains(result, "users")
	
	requests := result["requests"].(map[string]interface{})
	s.Contains(requests, "total")
	s.Contains(requests, "success_rate")
	s.Contains(requests, "average_response_time")
	
	files := result["files"].(map[string]interface{})
	s.Contains(files, "total_count")
	s.Contains(files, "total_size")
	
	storage := result["storage"].(map[string]interface{})
	s.Contains(storage, "used")
	s.Contains(storage, "available")
}

func (s *HealthHandlerTestSuite) TestPerformanceMetrics() {
	// 性能指标
	response := s.MakeRequest("GET", "/health/performance", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "response_times")
	s.Contains(result, "throughput")
	s.Contains(result, "error_rates")
	
	responseTimes := result["response_times"].(map[string]interface{})
	s.Contains(responseTimes, "average")
	s.Contains(responseTimes, "p95")
	s.Contains(responseTimes, "p99")
	
	throughput := result["throughput"].(map[string]interface{})
	s.Contains(throughput, "requests_per_second")
	s.Contains(throughput, "files_per_second")
}

func (s *HealthHandlerTestSuite) TestDeprecatedHealthEndpoint() {
	// 测试废弃的健康检查端点（如果有的话）
	response := s.MakeRequest("GET", "/api/v1/health", nil, nil)
	
	// 可能返回重定向或者直接返回健康状态
	if response.Code == http.StatusMovedPermanently || response.Code == http.StatusFound {
		s.NotEmpty(response.Header().Get("Location"))
	} else {
		result := s.AssertSuccessResponse(response, http.StatusOK)
		s.Equal("healthy", result["status"])
	}
}

func (s *HealthHandlerTestSuite) TestConfigurationCheck() {
	// 配置检查端点
	response := s.MakeRequest("GET", "/health/config", nil, s.AdminUser)
	
	// 验证响应（只有管理员可以访问）
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "environment")
	s.Contains(result, "debug_mode")
	s.Contains(result, "log_level")
	s.Contains(result, "features")
	
	// 验证敏感信息被隐藏
	s.NotContains(result, "password")
	s.NotContains(result, "secret")
	s.NotContains(result, "key")
}

func (s *HealthHandlerTestSuite) TestConfigurationCheckUnauthorized() {
	// 非管理员不能访问配置检查
	response := s.MakeRequest("GET", "/health/config", nil, s.TestUser)
	s.AssertErrorResponse(response, http.StatusForbidden)
}

func (s *HealthHandlerTestSuite) TestDeepHealthCheck() {
	// 深度健康检查（执行实际操作测试）
	response := s.MakeRequest("GET", "/health/deep", nil, s.AdminUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	
	s.Contains(result, "status")
	s.Contains(result, "tests")
	
	tests := result["tests"].(map[string]interface{})
	s.Contains(tests, "database_write_test")
	s.Contains(tests, "storage_write_test")
	s.Contains(tests, "cache_write_test")
	
	// 每个测试应该包含状态和执行时间
	dbTest := tests["database_write_test"].(map[string]interface{})
	s.Contains(dbTest, "status")
	s.Contains(dbTest, "duration")
}

func (s *HealthHandlerTestSuite) TestGracefulShutdownStatus() {
	// 优雅关闭状态检查
	response := s.MakeRequest("GET", "/health/shutdown", nil, nil)
	
	// 正常情况下应该返回运行状态
	result := s.AssertSuccessResponse(response, http.StatusOK)
	s.Equal("running", result["status"])
	s.Contains(result, "accepting_requests")
	s.True(result["accepting_requests"].(bool))
}

func TestHealthHandler(t *testing.T) {
	suite.Run(t, new(HealthHandlerTestSuite))
}