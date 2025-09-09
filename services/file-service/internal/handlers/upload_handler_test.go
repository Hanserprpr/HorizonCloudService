package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"file-service/internal/models"
	"file-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type UploadHandlerTestSuite struct {
	TestSuite
}

func (s *UploadHandlerTestSuite) TestSimpleUpload() {
	// 创建测试文件
	fileContent := "This is a simple upload test file content"
	files := map[string][]byte{
		"test-simple.txt": []byte(fileContent),
	}
	fields := map[string]string{
		"folder_id":   "",
		"description": "Simple upload test",
	}

	// 发送简单上传请求
	response := s.MakeMultipartRequest("POST", "/api/v1/upload/simple", fields, files, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal("test-simple.txt", data["name"])
	s.Equal(float64(len(fileContent)), data["size"])
	s.Equal(s.TestUser.ID, uint(data["user_id"].(float64)))
}

func (s *UploadHandlerTestSuite) TestSimpleUploadToFolder() {
	// 创建测试文件夹
	folder := s.CreateTestFolder(s.TestUser.ID, "upload-folder", nil)

	// 创建测试文件
	fileContent := "Upload to specific folder"
	files := map[string][]byte{
		"folder-test.txt": []byte(fileContent),
	}
	fields := map[string]string{
		"folder_id": fmt.Sprintf("%d", folder.ID),
	}

	// 发送上传请求
	response := s.MakeMultipartRequest("POST", "/api/v1/upload/simple", fields, files, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal("folder-test.txt", data["name"])
	s.Equal(float64(folder.ID), data["folder_id"])
}

func (s *UploadHandlerTestSuite) TestInitiateChunkedUpload() {
	// 准备分片上传请求
	initData := map[string]interface{}{
		"file_name":    "large-file.txt",
		"file_size":    1024 * 1024 * 5, // 5MB
		"content_type": "text/plain",
		"chunk_size":   1024 * 1024, // 1MB chunks
		"folder_id":    nil,
	}

	// 初始化分片上传
	response := s.MakeRequest("POST", "/api/v1/upload/initiate", initData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.NotEmpty(data["session_id"])
	s.Equal("large-file.txt", data["file_name"])
	s.Equal(float64(5*1024*1024), data["file_size"])
	s.Greater(data["total_chunks"], float64(0))
	s.Equal("initiated", data["status"])
}

func (s *UploadHandlerTestSuite) TestUploadChunk() {
	// 首先初始化上传会话
	initData := map[string]interface{}{
		"file_name":    "chunk-test.txt",
		"file_size":    1024,
		"content_type": "text/plain",
		"chunk_size":   256,
	}

	initResponse := s.MakeRequest("POST", "/api/v1/upload/initiate", initData, s.TestUser)
	initResult := s.AssertSuccessResponse(initResponse)
	sessionData := initResult["data"].(map[string]interface{})
	sessionID := sessionData["session_id"].(string)

	// 上传第一个分片
	chunkContent := strings.Repeat("a", 256)
	chunkData := map[string]interface{}{
		"session_id":  sessionID,
		"chunk_index": 0,
		"chunk_size":  256,
		"chunk_data":  chunkContent,
	}

	response := s.MakeRequest("POST", "/api/v1/upload/chunk", chunkData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal(sessionID, data["session_id"])
	s.Equal(float64(0), data["chunk_index"])
	s.True(data["success"].(bool))
}

func (s *UploadHandlerTestSuite) TestCompleteChunkedUpload() {
	// 创建一个小文件的完整上传流程
	fileSize := int64(512) // 512 bytes
	chunkSize := int64(256) // 256 bytes per chunk
	
	// 初始化上传
	initData := map[string]interface{}{
		"file_name":    "complete-test.txt",
		"file_size":    fileSize,
		"content_type": "text/plain",
		"chunk_size":   chunkSize,
	}

	initResponse := s.MakeRequest("POST", "/api/v1/upload/initiate", initData, s.TestUser)
	initResult := s.AssertSuccessResponse(initResponse)
	sessionData := initResult["data"].(map[string]interface{})
	sessionID := sessionData["session_id"].(string)

	// 上传所有分片
	for i := 0; i < 2; i++ { // 2 chunks
		chunkContent := strings.Repeat(fmt.Sprintf("%d", i), int(chunkSize))
		chunkData := map[string]interface{}{
			"session_id":  sessionID,
			"chunk_index": i,
			"chunk_size":  chunkSize,
			"chunk_data":  chunkContent,
		}

		chunkResponse := s.MakeRequest("POST", "/api/v1/upload/chunk", chunkData, s.TestUser)
		s.AssertSuccessResponse(chunkResponse)
	}

	// 完成上传
	completeData := map[string]interface{}{
		"session_id": sessionID,
	}

	response := s.MakeRequest("POST", "/api/v1/upload/complete", completeData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.NotNil(data["file"])
	fileData := data["file"].(map[string]interface{})
	s.Equal("complete-test.txt", fileData["name"])
	s.Equal(float64(fileSize), fileData["size"])
}

func (s *UploadHandlerTestSuite) TestAbortUpload() {
	// 初始化上传会话
	initData := map[string]interface{}{
		"file_name":    "abort-test.txt",
		"file_size":    1024,
		"content_type": "text/plain",
		"chunk_size":   256,
	}

	initResponse := s.MakeRequest("POST", "/api/v1/upload/initiate", initData, s.TestUser)
	initResult := s.AssertSuccessResponse(initResponse)
	sessionData := initResult["data"].(map[string]interface{})
	sessionID := sessionData["session_id"].(string)

	// 中止上传
	abortData := map[string]interface{}{
		"session_id": sessionID,
	}

	response := s.MakeRequest("POST", "/api/v1/upload/abort", abortData, s.TestUser)
	
	// 验证响应
	s.AssertSuccessResponse(response, http.StatusOK)

	// 验证会话已被删除 - 尝试获取会话应该失败
	statusResponse := s.MakeRequest("GET", fmt.Sprintf("/api/v1/upload/status/%s", sessionID), nil, s.TestUser)
	s.AssertErrorResponse(statusResponse, http.StatusNotFound)
}

func (s *UploadHandlerTestSuite) TestGetUploadStatus() {
	// 初始化上传会话
	initData := map[string]interface{}{
		"file_name":    "status-test.txt",
		"file_size":    1024,
		"content_type": "text/plain",
		"chunk_size":   256,
	}

	initResponse := s.MakeRequest("POST", "/api/v1/upload/initiate", initData, s.TestUser)
	initResult := s.AssertSuccessResponse(initResponse)
	sessionData := initResult["data"].(map[string]interface{})
	sessionID := sessionData["session_id"].(string)

	// 获取上传状态
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/upload/status/%s", sessionID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal(sessionID, data["session_id"])
	s.Equal("status-test.txt", data["file_name"])
	s.Equal("initiated", data["status"])
	s.Equal(float64(0), data["uploaded_chunks"])
	s.Equal(float64(4), data["total_chunks"]) // 1024/256 = 4
}

func (s *UploadHandlerTestSuite) TestGetUploadProgress() {
	// 初始化上传会话
	initData := map[string]interface{}{
		"file_name":    "progress-test.txt",
		"file_size":    1024,
		"content_type": "text/plain",
		"chunk_size":   256,
	}

	initResponse := s.MakeRequest("POST", "/api/v1/upload/initiate", initData, s.TestUser)
	initResult := s.AssertSuccessResponse(initResponse)
	sessionData := initResult["data"].(map[string]interface{})
	sessionID := sessionData["session_id"].(string)

	// 上传一个分片
	chunkContent := strings.Repeat("a", 256)
	chunkData := map[string]interface{}{
		"session_id":  sessionID,
		"chunk_index": 0,
		"chunk_size":  256,
		"chunk_data":  chunkContent,
	}
	s.MakeRequest("POST", "/api/v1/upload/chunk", chunkData, s.TestUser)

	// 获取上传进度
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/upload/progress/%s", sessionID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal(sessionID, data["session_id"])
	s.Equal(float64(4), data["total_chunks"])
	s.Equal(float64(1), data["uploaded_chunks"])
	s.Equal(float64(25), data["percentage"]) // 1/4 * 100%
}

func (s *UploadHandlerTestSuite) TestListUploadSessions() {
	// 创建多个上传会话
	for i := 0; i < 3; i++ {
		initData := map[string]interface{}{
			"file_name":    fmt.Sprintf("session-%d.txt", i),
			"file_size":    1024,
			"content_type": "text/plain",
			"chunk_size":   256,
		}
		s.MakeRequest("POST", "/api/v1/upload/initiate", initData, s.TestUser)
	}

	// 获取上传会话列表
	response := s.MakeRequest("GET", "/api/v1/upload/sessions?page=1&page_size=10", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertPaginatedResponse(response)
	data := result["data"].(map[string]interface{})
	
	sessions := data["data"].([]interface{})
	s.GreaterOrEqual(len(sessions), 3) // 应该至少有3个会话
	s.GreaterOrEqual(data["total"], float64(3))
}

func (s *UploadHandlerTestSuite) TestBatchInitiateUpload() {
	// 准备批量上传请求
	batchData := map[string]interface{}{
		"files": []map[string]interface{}{
			{
				"file_name":    "batch1.txt",
				"file_size":    1024,
				"content_type": "text/plain",
				"chunk_size":   256,
			},
			{
				"file_name":    "batch2.txt",
				"file_size":    2048,
				"content_type": "text/plain",
				"chunk_size":   512,
			},
		},
	}

	// 批量初始化上传
	response := s.MakeRequest("POST", "/api/v1/upload/batch-initiate", batchData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal(float64(2), data["total"])
	s.Equal(float64(2), data["successful"])
	s.Equal(float64(0), data["failed"])
	
	sessions := data["sessions"].([]interface{})
	s.Len(sessions, 2)
}

func (s *UploadHandlerTestSuite) TestResumeUpload() {
	// 初始化上传会话
	initData := map[string]interface{}{
		"file_name":    "resume-test.txt",
		"file_size":    1024,
		"content_type": "text/plain",
		"chunk_size":   256,
	}

	initResponse := s.MakeRequest("POST", "/api/v1/upload/initiate", initData, s.TestUser)
	initResult := s.AssertSuccessResponse(initResponse)
	sessionData := initResult["data"].(map[string]interface{})
	sessionID := sessionData["session_id"].(string)

	// 暂停上传（通过其他API或直接操作数据库）
	pauseData := map[string]interface{}{
		"session_id": sessionID,
	}
	s.MakeRequest("POST", "/api/v1/upload/pause", pauseData, s.TestUser)

	// 恢复上传
	resumeData := map[string]interface{}{
		"session_id": sessionID,
	}

	response := s.MakeRequest("POST", "/api/v1/upload/resume", resumeData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal(sessionID, data["session_id"])
	s.Equal("uploading", data["status"])
}

func (s *UploadHandlerTestSuite) TestPauseUpload() {
	// 初始化上传会话
	initData := map[string]interface{}{
		"file_name":    "pause-test.txt",
		"file_size":    1024,
		"content_type": "text/plain",
		"chunk_size":   256,
	}

	initResponse := s.MakeRequest("POST", "/api/v1/upload/initiate", initData, s.TestUser)
	initResult := s.AssertSuccessResponse(initResponse)
	sessionData := initResult["data"].(map[string]interface{})
	sessionID := sessionData["session_id"].(string)

	// 暂停上传
	pauseData := map[string]interface{}{
		"session_id": sessionID,
	}

	response := s.MakeRequest("POST", "/api/v1/upload/pause", pauseData, s.TestUser)
	
	// 验证响应
	s.AssertSuccessResponse(response, http.StatusOK)

	// 验证状态已改变
	statusResponse := s.MakeRequest("GET", fmt.Sprintf("/api/v1/upload/status/%s", sessionID), nil, s.TestUser)
	statusResult := s.AssertSuccessResponse(statusResponse)
	statusData := statusResult["data"].(map[string]interface{})
	s.Equal("paused", statusData["status"])
}

func (s *UploadHandlerTestSuite) TestUploadStatistics() {
	// 创建一些上传会话来生成统计数据
	for i := 0; i < 5; i++ {
		initData := map[string]interface{}{
			"file_name":    fmt.Sprintf("stats-%d.txt", i),
			"file_size":    1024 * (i + 1),
			"content_type": "text/plain",
			"chunk_size":   256,
		}
		s.MakeRequest("POST", "/api/v1/upload/initiate", initData, s.TestUser)
	}

	// 获取上传统计
	response := s.MakeRequest("GET", "/api/v1/upload/statistics", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	stats := result["data"].(map[string]interface{})
	
	s.Contains(stats, "total_sessions")
	s.Contains(stats, "completed_sessions")
	s.Contains(stats, "total_bytes")
	s.GreaterOrEqual(stats["total_sessions"], float64(5))
}

func (s *UploadHandlerTestSuite) TestUploadWithInvalidChunkIndex() {
	// 初始化上传会话
	initData := map[string]interface{}{
		"file_name":    "invalid-chunk.txt",
		"file_size":    1024,
		"content_type": "text/plain",
		"chunk_size":   256,
	}

	initResponse := s.MakeRequest("POST", "/api/v1/upload/initiate", initData, s.TestUser)
	initResult := s.AssertSuccessResponse(initResponse)
	sessionData := initResult["data"].(map[string]interface{})
	sessionID := sessionData["session_id"].(string)

	// 尝试上传无效的分片索引
	chunkData := map[string]interface{}{
		"session_id":  sessionID,
		"chunk_index": 999, // 超出范围的索引
		"chunk_size":  256,
		"chunk_data":  "invalid chunk",
	}

	response := s.MakeRequest("POST", "/api/v1/upload/chunk", chunkData, s.TestUser)
	
	// 验证错误响应
	s.AssertErrorResponse(response, http.StatusBadRequest)
}

func (s *UploadHandlerTestSuite) TestCompleteIncompleteUpload() {
	// 初始化上传会话
	initData := map[string]interface{}{
		"file_name":    "incomplete.txt",
		"file_size":    1024,
		"content_type": "text/plain",
		"chunk_size":   256,
	}

	initResponse := s.MakeRequest("POST", "/api/v1/upload/initiate", initData, s.TestUser)
	initResult := s.AssertSuccessResponse(initResponse)
	sessionData := initResult["data"].(map[string]interface{})
	sessionID := sessionData["session_id"].(string)

	// 只上传部分分片（缺少一些分片）
	chunkData := map[string]interface{}{
		"session_id":  sessionID,
		"chunk_index": 0,
		"chunk_size":  256,
		"chunk_data":  strings.Repeat("a", 256),
	}
	s.MakeRequest("POST", "/api/v1/upload/chunk", chunkData, s.TestUser)

	// 尝试完成不完整的上传
	completeData := map[string]interface{}{
		"session_id": sessionID,
	}

	response := s.MakeRequest("POST", "/api/v1/upload/complete", completeData, s.TestUser)
	
	// 验证错误响应
	s.AssertErrorResponse(response, http.StatusBadRequest)
}

func (s *UploadHandlerTestSuite) TestUnauthorizedUpload() {
	// 测试无token上传
	files := map[string][]byte{
		"test.txt": []byte("unauthorized test"),
	}
	response := s.MakeMultipartRequest("POST", "/api/v1/upload/simple", map[string]string{}, files, nil)
	s.AssertErrorResponse(response, http.StatusUnauthorized)
}

func (s *UploadHandlerTestSuite) TestUploadToInvalidFolder() {
	// 尝试上传到不存在的文件夹
	files := map[string][]byte{
		"test.txt": []byte("test content"),
	}
	fields := map[string]string{
		"folder_id": "99999", // 不存在的文件夹ID
	}

	response := s.MakeMultipartRequest("POST", "/api/v1/upload/simple", fields, files, s.TestUser)
	s.AssertErrorResponse(response, http.StatusBadRequest)
}

func (s *UploadHandlerTestSuite) TestUploadEmptyFile() {
	// 尝试上传空文件
	files := map[string][]byte{
		"empty.txt": []byte(""),
	}

	response := s.MakeMultipartRequest("POST", "/api/v1/upload/simple", map[string]string{}, files, s.TestUser)
	s.AssertErrorResponse(response, http.StatusBadRequest)
}

func (s *UploadHandlerTestSuite) TestUploadWithoutFile() {
	// 尝试没有文件的上传请求
	response := s.MakeRequest("POST", "/api/v1/upload/simple", map[string]interface{}{}, s.TestUser)
	s.AssertErrorResponse(response, http.StatusBadRequest)
}

func TestUploadHandler(t *testing.T) {
	suite.Run(t, new(UploadHandlerTestSuite))
}