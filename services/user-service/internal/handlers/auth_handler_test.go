package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"user-service/internal/repository"
	"user-service/internal/services"
	
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// AuthHandlerTestSuite 认证处理器测试套件
type AuthHandlerTestSuite struct {
	suite.Suite
	router      *gin.Engine
	userService services.UserService
	authHandler *AuthHandler
}

// SetupSuite 设置测试套件
func (s *AuthHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	
	// 设置测试服务
	mockRepo := repository.NewMockUserRepository()
	s.userService = services.NewUserService(mockRepo, "test-jwt-secret")
	s.authHandler = NewAuthHandler(s.userService)
	
	// 设置路由
	s.router = gin.New()
	v1 := s.router.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/register", s.authHandler.Register)
		auth.POST("/login", s.authHandler.Login)
		auth.POST("/refresh", s.authHandler.RefreshToken)
		auth.POST("/logout", s.authHandler.Logout) // 这个需要认证中间件，测试时跳过
	}
}

// TestRegister 测试用户注册接口
func (s *AuthHandlerTestSuite) TestRegister() {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "成功注册",
			requestBody: services.RegisterRequest{
				StudentID: "test001",
				Password:  "password123",
				NickName:  "测试用户",
				Email:     "test@example.com",
				Phone:     "13800138000",
			},
			expectedStatus: http.StatusCreated,
			expectedMsg:    "注册成功",
		},
		{
			name: "学号为空",
			requestBody: services.RegisterRequest{
				StudentID: "",
				Password:  "password123",
				NickName:  "测试用户",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "密码过短",
			requestBody: services.RegisterRequest{
				StudentID: "test002",
				Password:  "123",
				NickName:  "测试用户",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "邮箱格式错误",
			requestBody: services.RegisterRequest{
				StudentID: "test003",
				Password:  "password123",
				NickName:  "测试用户",
				Email:     "invalid-email",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "重复学号",
			requestBody: services.RegisterRequest{
				StudentID: "test001", // 重复使用已注册的学号
				Password:  "password456",
				NickName:  "另一个用户",
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "请求体格式错误",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 准备请求体
			var reqBody []byte
			var err error
			
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(s.T(), err)
			}
			
			// 发送请求
			req, err := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
			assert.NoError(s.T(), err)
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, req)
			
			// 验证状态码
			assert.Equal(s.T(), tt.expectedStatus, w.Code)
			
			// 解析响应
			var response Response
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(s.T(), err)
			
			if tt.expectedMsg != "" {
				assert.Contains(s.T(), response.Message, tt.expectedMsg)
			}
			
			// 验证成功注册的响应结构
			if tt.expectedStatus == http.StatusCreated {
				assert.True(s.T(), response.Success)
				assert.NotNil(s.T(), response.Data)
				
				// 验证返回的认证数据
				authData, ok := response.Data.(map[string]interface{})
				assert.True(s.T(), ok)
				assert.NotEmpty(s.T(), authData["access_token"])
				assert.NotEmpty(s.T(), authData["refresh_token"])
				assert.NotNil(s.T(), authData["user"])
			}
		})
	}
}

// TestLogin 测试用户登录接口
func (s *AuthHandlerTestSuite) TestLogin() {
	// 先注册一个测试用户
	registerReq := services.RegisterRequest{
		StudentID: "login001",
		Password:  "password123",
		NickName:  "登录测试用户",
	}
	
	_, err := s.userService.Register(context.Background(), &registerReq)
	assert.NoError(s.T(), err)
	
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "成功登录",
			requestBody: services.LoginRequest{
				StudentID: "login001",
				Password:  "password123",
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "登录成功",
		},
		{
			name: "学号不存在",
			requestBody: services.LoginRequest{
				StudentID: "notexist",
				Password:  "password123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "密码错误",
			requestBody: services.LoginRequest{
				StudentID: "login001",
				Password:  "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "学号为空",
			requestBody: services.LoginRequest{
				StudentID: "",
				Password:  "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "请求体格式错误",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 准备请求体
			var reqBody []byte
			var err error
			
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(s.T(), err)
			}
			
			// 发送请求
			req, err := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
			assert.NoError(s.T(), err)
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, req)
			
			// 验证状态码
			assert.Equal(s.T(), tt.expectedStatus, w.Code)
			
			// 解析响应
			var response Response
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(s.T(), err)
			
			if tt.expectedMsg != "" {
				assert.Contains(s.T(), response.Message, tt.expectedMsg)
			}
			
			// 验证成功登录的响应结构
			if tt.expectedStatus == http.StatusOK {
				assert.True(s.T(), response.Success)
				assert.NotNil(s.T(), response.Data)
				
				// 验证返回的认证数据
				authData, ok := response.Data.(map[string]interface{})
				assert.True(s.T(), ok)
				assert.NotEmpty(s.T(), authData["access_token"])
				assert.NotEmpty(s.T(), authData["refresh_token"])
				assert.NotNil(s.T(), authData["user"])
			}
		})
	}
}

// TestRefreshToken 测试刷新令牌接口
func (s *AuthHandlerTestSuite) TestRefreshToken() {
	// 先注册并登录获取刷新令牌
	registerReq := services.RegisterRequest{
		StudentID: "refresh001",
		Password:  "password123",
		NickName:  "刷新令牌测试用户",
	}
	
	authResult, err := s.userService.Register(context.Background(), &registerReq)
	assert.NoError(s.T(), err)
	
	tests := []struct {
		name           string
		refreshToken   string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "成功刷新令牌",
			refreshToken:   authResult.RefreshToken,
			expectedStatus: http.StatusOK,
			expectedMsg:    "令牌刷新成功",
		},
		{
			name:           "无效的刷新令牌",
			refreshToken:   "invalid-refresh-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "刷新令牌为空",
			refreshToken:   "",
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		s.Run(tt.name, func() {
			requestBody := map[string]string{
				"refresh_token": tt.refreshToken,
			}
			
			reqBody, err := json.Marshal(requestBody)
			assert.NoError(s.T(), err)
			
			// 发送请求
			req, err := http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(reqBody))
			assert.NoError(s.T(), err)
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, req)
			
			// 验证状态码
			assert.Equal(s.T(), tt.expectedStatus, w.Code)
			
			// 解析响应
			var response Response
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(s.T(), err)
			
			if tt.expectedMsg != "" {
				assert.Contains(s.T(), response.Message, tt.expectedMsg)
			}
			
			// 验证成功刷新的响应结构
			if tt.expectedStatus == http.StatusOK {
				assert.True(s.T(), response.Success)
				assert.NotNil(s.T(), response.Data)
				
				// 验证返回的认证数据
				authData, ok := response.Data.(map[string]interface{})
				assert.True(s.T(), ok)
				assert.NotEmpty(s.T(), authData["access_token"])
				assert.NotEmpty(s.T(), authData["refresh_token"])
			}
		})
	}
}

// TestResponseFormat 测试响应格式一致性
func (s *AuthHandlerTestSuite) TestResponseFormat() {
	// 测试成功响应格式
	registerReq := services.RegisterRequest{
		StudentID: "format001",
		Password:  "password123",
		NickName:  "格式测试用户",
	}
	
	reqBody, err := json.Marshal(registerReq)
	assert.NoError(s.T(), err)
	
	req, err := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	assert.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	
	// 解析响应
	var response Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	
	// 验证响应结构
	assert.Equal(s.T(), http.StatusCreated, response.Code)
	assert.NotEmpty(s.T(), response.Message)
	assert.True(s.T(), response.Success)
	assert.NotNil(s.T(), response.Data)
	assert.NotZero(s.T(), response.Timestamp)
}

// TestConcurrentRequests 测试并发请求处理
func (s *AuthHandlerTestSuite) TestConcurrentRequests() {
	const numRequests = 10
	done := make(chan bool, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func(index int) {
			defer func() { done <- true }()
			
			registerReq := services.RegisterRequest{
				StudentID: fmt.Sprintf("concurrent%03d", index),
				Password:  "password123",
				NickName:  fmt.Sprintf("并发测试用户%d", index),
			}
			
			reqBody, err := json.Marshal(registerReq)
			assert.NoError(s.T(), err)
			
			req, err := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
			assert.NoError(s.T(), err)
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, req)
			
			// 所有请求都应该成功
			assert.Equal(s.T(), http.StatusCreated, w.Code)
		}(i)
	}
	
	// 等待所有请求完成
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

// 运行测试套件
func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}