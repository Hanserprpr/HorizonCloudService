package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"file-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type ThumbnailHandlerTestSuite struct {
	TestSuite
}

func (s *ThumbnailHandlerTestSuite) TestGetThumbnail() {
	// 创建测试图片文件
	imageFile := s.CreateTestFile(s.TestUser.ID, "test-image.jpg", "fake image content")
	imageFile.ContentType = "image/jpeg"
	imageFile.Category = "image"
	s.DB.Save(imageFile)

	// 创建缩略图记录
	thumbnail := &models.Thumbnail{
		FileID:      imageFile.ID,
		Size:        "medium",
		Width:       300,
		Height:      200,
		Path: "/thumbnails/test-image-medium.jpg",
		ContentType: "image/jpeg",
		Quality:     80,
		FileSize:    5120, // 5KB
		Status:      models.ThumbnailStatusReady,
	}
	s.DB.Create(thumbnail)

	// 获取缩略图
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/thumbnails/%d?size=medium", imageFile.ID), nil, s.TestUser)
	
	// 验证重定向响应（缩略图URL）
	s.Equal(http.StatusFound, response.Code)
	s.NotEmpty(response.Header().Get("Location"))
}

func (s *ThumbnailHandlerTestSuite) TestGetThumbnailDefaultSize() {
	// 创建测试图片文件
	imageFile := s.CreateTestFile(s.TestUser.ID, "default-size.jpg", "fake image content")
	imageFile.ContentType = "image/jpeg"
	imageFile.Category = "image"
	s.DB.Save(imageFile)

	// 创建默认尺寸缩略图
	thumbnail := &models.Thumbnail{
		FileID:      imageFile.ID,
		Size:        "small",
		Width:       150,
		Height:      100,
		Path: "/thumbnails/default-size-small.jpg",
		ContentType: "image/jpeg",
		Quality:     80,
		FileSize:    2048,
		Status:      models.ThumbnailStatusReady,
	}
	s.DB.Create(thumbnail)

	// 不指定size参数，应该返回默认尺寸
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/thumbnails/%d", imageFile.ID), nil, s.TestUser)
	
	// 验证响应
	s.Equal(http.StatusFound, response.Code)
	s.NotEmpty(response.Header().Get("Location"))
}

func (s *ThumbnailHandlerTestSuite) TestGetThumbnailNotFound() {
	// 尝试获取不存在文件的缩略图
	response := s.MakeRequest("GET", "/api/v1/thumbnails/99999", nil, s.TestUser)
	
	// 验证错误响应
	s.AssertErrorResponse(response, http.StatusNotFound)
}

func (s *ThumbnailHandlerTestSuite) TestGetThumbnailUnauthorized() {
	// 创建其他用户的图片文件
	otherImageFile := s.CreateTestFile(999, "other-image.jpg", "fake image content")
	otherImageFile.ContentType = "image/jpeg"
	s.DB.Save(otherImageFile)

	// 尝试访问其他用户的缩略图
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/thumbnails/%d", otherImageFile.ID), nil, s.TestUser)
	
	// 验证错误响应
	s.AssertErrorResponse(response, http.StatusNotFound) // 返回404避免信息泄露
}

func (s *ThumbnailHandlerTestSuite) TestGetThumbnailNonImageFile() {
	// 创建非图片文件
	textFile := s.CreateTestFile(s.TestUser.ID, "document.txt", "text content")
	textFile.ContentType = "text/plain"
	textFile.Category = "document"
	s.DB.Save(textFile)

	// 尝试获取非图片文件的缩略图
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/thumbnails/%d", textFile.ID), nil, s.TestUser)
	
	// 验证错误响应
	s.AssertErrorResponse(response, http.StatusBadRequest)
}

func (s *ThumbnailHandlerTestSuite) TestGenerateThumbnail() {
	// 创建测试图片文件
	imageFile := s.CreateTestFile(s.TestUser.ID, "generate-test.jpg", "fake image content")
	imageFile.ContentType = "image/jpeg"
	imageFile.Category = "image"
	s.DB.Save(imageFile)

	// 请求生成缩略图
	generateData := map[string]interface{}{
		"sizes": []string{"small", "medium", "large"},
	}

	response := s.MakeRequest("POST", fmt.Sprintf("/api/v1/thumbnails/%d/generate", imageFile.ID), generateData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal(float64(imageFile.ID), data["file_id"])
	s.Contains(data, "job_id") // 异步任务ID
	s.Equal("queued", data["status"])
}

func (s *ThumbnailHandlerTestSuite) TestGenerateAllSizeThumbnails() {
	// 创建测试图片文件
	imageFile := s.CreateTestFile(s.TestUser.ID, "generate-all.jpg", "fake image content")
	imageFile.ContentType = "image/jpeg"
	imageFile.Category = "image"
	s.DB.Save(imageFile)

	// 生成所有尺寸的缩略图
	response := s.MakeRequest("POST", fmt.Sprintf("/api/v1/thumbnails/%d/generate-all", imageFile.ID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal(float64(imageFile.ID), data["file_id"])
	s.Contains(data, "job_id")
	s.Equal("queued", data["status"])
}

func (s *ThumbnailHandlerTestSuite) TestListThumbnails() {
	// 创建测试图片文件
	imageFile := s.CreateTestFile(s.TestUser.ID, "list-test.jpg", "fake image content")
	imageFile.ContentType = "image/jpeg"
	imageFile.Category = "image"
	s.DB.Save(imageFile)

	// 创建多个缩略图记录
	sizes := []string{"small", "medium", "large"}
	for _, size := range sizes {
		thumbnail := &models.Thumbnail{
			FileID:      imageFile.ID,
			Size:        size,
			Width:       150,
			Height:      100,
			Path: fmt.Sprintf("/thumbnails/list-test-%s.jpg", size),
			ContentType: "image/jpeg",
			Quality:     80,
			FileSize:    2048,
			Status:      models.ThumbnailStatusReady,
		}
		s.DB.Create(thumbnail)
	}

	// 获取缩略图列表
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/thumbnails/%d/list", imageFile.ID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	thumbnails := result["data"].([]interface{})
	
	s.Len(thumbnails, 3) // 应该有3个缩略图
	
	// 验证缩略图信息
	for _, thumbInterface := range thumbnails {
		thumb := thumbInterface.(map[string]interface{})
		s.Equal(float64(imageFile.ID), thumb["file_id"])
		s.Contains([]string{"small", "medium", "large"}, thumb["size"])
		s.Equal(float64(models.ThumbnailStatusReady), thumb["status"])
	}
}

func (s *ThumbnailHandlerTestSuite) TestGetThumbnailInfo() {
	// 创建测试图片文件
	imageFile := s.CreateTestFile(s.TestUser.ID, "info-test.jpg", "fake image content")
	imageFile.ContentType = "image/jpeg"
	s.DB.Save(imageFile)

	// 创建缩略图记录
	thumbnail := &models.Thumbnail{
		FileID:      imageFile.ID,
		Size:        "medium",
		Width:       300,
		Height:      200,
		Path: "/thumbnails/info-test-medium.jpg",
		ContentType: "image/jpeg",
		Quality:     80,
		FileSize:    5120,
		Status:      models.ThumbnailStatusReady,
	}
	s.DB.Create(thumbnail)

	// 获取缩略图详细信息
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/thumbnails/%d/info?size=medium", imageFile.ID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal(float64(imageFile.ID), data["file_id"])
	s.Equal("medium", data["size"])
	s.Equal(float64(300), data["width"])
	s.Equal(float64(200), data["height"])
	s.Equal("jpeg", data["format"])
	s.Equal(float64(80), data["quality"])
	s.Equal(float64(models.ThumbnailStatusReady), data["status"])
}

func (s *ThumbnailHandlerTestSuite) TestDeleteThumbnail() {
	// 创建测试图片文件
	imageFile := s.CreateTestFile(s.TestUser.ID, "delete-test.jpg", "fake image content")
	imageFile.ContentType = "image/jpeg"
	s.DB.Save(imageFile)

	// 创建缩略图记录
	thumbnail := &models.Thumbnail{
		FileID:      imageFile.ID,
		Size:        "medium",
		Width:       300,
		Height:      200,
		Path: "/thumbnails/delete-test-medium.jpg",
		ContentType: "image/jpeg",
		Quality:     80,
		FileSize:    5120,
		Status:      models.ThumbnailStatusReady,
	}
	s.DB.Create(thumbnail)

	// 删除缩略图
	response := s.MakeRequest("DELETE", fmt.Sprintf("/api/v1/thumbnails/%d?size=medium", imageFile.ID), nil, s.TestUser)
	
	// 验证响应
	s.AssertSuccessResponse(response, http.StatusOK)

	// 验证缩略图已被删除
	var deletedThumbnail models.Thumbnail
	result := s.DB.Where("file_id = ? AND size = ?", imageFile.ID, "medium").First(&deletedThumbnail)
	s.Error(result.Error) // 应该找不到记录
}

func (s *ThumbnailHandlerTestSuite) TestDeleteAllThumbnails() {
	// 创建测试图片文件
	imageFile := s.CreateTestFile(s.TestUser.ID, "delete-all-test.jpg", "fake image content")
	imageFile.ContentType = "image/jpeg"
	s.DB.Save(imageFile)

	// 创建多个缩略图记录
	sizes := []string{"small", "medium", "large"}
	for _, size := range sizes {
		thumbnail := &models.Thumbnail{
			FileID:      imageFile.ID,
			Size:        size,
			Width:       150,
			Height:      100,
			Path: fmt.Sprintf("/thumbnails/delete-all-test-%s.jpg", size),
			ContentType: "image/jpeg",
			Quality:     80,
			FileSize:    2048,
			Status:      models.ThumbnailStatusReady,
		}
		s.DB.Create(thumbnail)
	}

	// 删除所有缩略图
	response := s.MakeRequest("DELETE", fmt.Sprintf("/api/v1/thumbnails/%d/all", imageFile.ID), nil, s.TestUser)
	
	// 验证响应
	s.AssertSuccessResponse(response, http.StatusOK)

	// 验证所有缩略图都被删除
	var count int64
	s.DB.Model(&models.Thumbnail{}).Where("file_id = ?", imageFile.ID).Count(&count)
	s.Equal(int64(0), count)
}

func (s *ThumbnailHandlerTestSuite) TestRegenerateThumbnail() {
	// 创建测试图片文件
	imageFile := s.CreateTestFile(s.TestUser.ID, "regenerate-test.jpg", "fake image content")
	imageFile.ContentType = "image/jpeg"
	s.DB.Save(imageFile)

	// 创建现有缩略图
	thumbnail := &models.Thumbnail{
		FileID:      imageFile.ID,
		Size:        "medium",
		Width:       300,
		Height:      200,
		Path: "/thumbnails/regenerate-test-medium.jpg",
		ContentType: "image/jpeg",
		Quality:     80,
		FileSize:    5120,
		Status:      models.ThumbnailStatusReady,
	}
	s.DB.Create(thumbnail)

	// 重新生成缩略图
	regenerateData := map[string]interface{}{
		"size":    "medium",
		"quality": 90,
		"force":   true,
	}

	response := s.MakeRequest("POST", fmt.Sprintf("/api/v1/thumbnails/%d/regenerate", imageFile.ID), regenerateData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal(float64(imageFile.ID), data["file_id"])
	s.Contains(data, "job_id")
	s.Equal("queued", data["status"])
}

func (s *ThumbnailHandlerTestSuite) TestGetThumbnailGenerationStatus() {
	// 创建测试图片文件
	imageFile := s.CreateTestFile(s.TestUser.ID, "status-test.jpg", "fake image content")
	imageFile.ContentType = "image/jpeg"
	s.DB.Save(imageFile)

	// 创建处理中的缩略图记录
	thumbnail := &models.Thumbnail{
		FileID:      imageFile.ID,
		Size:        "medium",
		Width:       0, // 还未完成
		Height:      0,
		Path: "/thumbnails/status-test-medium.jpg",
		ContentType: "image/jpeg",
		Quality:     80,
		FileSize:    0,
		Status:      models.ThumbnailStatusGenerating,
	}
	s.DB.Create(thumbnail)

	// 获取生成状态
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/thumbnails/%d/status", imageFile.ID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal(float64(imageFile.ID), data["file_id"])
	statuses := data["thumbnails"].([]interface{})
	s.GreaterOrEqual(len(statuses), 1)
	
	status := statuses[0].(map[string]interface{})
	s.Equal(float64(models.ThumbnailStatusGenerating), status["status"])
	s.Equal("medium", status["size"])
}

func (s *ThumbnailHandlerTestSuite) TestBatchGenerateThumbnails() {
	// 创建多个测试图片文件
	imageFiles := make([]*models.File, 3)
	for i := 0; i < 3; i++ {
		file := s.CreateTestFile(s.TestUser.ID, fmt.Sprintf("batch-%d.jpg", i), "fake image content")
		file.ContentType = "image/jpeg"
		file.Category = "image"
		s.DB.Save(file)
		imageFiles[i] = file
	}

	// 批量生成缩略图
	batchData := map[string]interface{}{
		"file_ids": []uint{imageFiles[0].ID, imageFiles[1].ID, imageFiles[2].ID},
		"sizes":    []string{"small", "medium"},
	}

	response := s.MakeRequest("POST", "/api/v1/thumbnails/batch-generate", batchData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal(float64(3), data["total"])
	s.Equal(float64(3), data["successful"])
	s.Equal(float64(0), data["failed"])
	
	jobs := data["jobs"].([]interface{})
	s.Len(jobs, 3)
}

func (s *ThumbnailHandlerTestSuite) TestGetThumbnailWithCustomSize() {
	// 创建测试图片文件
	imageFile := s.CreateTestFile(s.TestUser.ID, "custom-size.jpg", "fake image content")
	imageFile.ContentType = "image/jpeg"
	s.DB.Save(imageFile)

	// 请求自定义尺寸缩略图
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/thumbnails/%d?width=400&height=300", imageFile.ID), nil, s.TestUser)
	
	// 验证响应（可能需要先生成）
	// 如果缩略图不存在，可能返回202 Accepted表示正在生成
	if response.Code == http.StatusAccepted {
		result := s.AssertSuccessResponse(response, http.StatusAccepted)
		data := result["data"].(map[string]interface{})
		s.Contains(data, "job_id")
		s.Equal("generating", data["status"])
	} else {
		s.Equal(http.StatusFound, response.Code)
		s.NotEmpty(response.Header().Get("Location"))
	}
}

func (s *ThumbnailHandlerTestSuite) TestInvalidThumbnailParameters() {
	// 创建测试图片文件
	imageFile := s.CreateTestFile(s.TestUser.ID, "invalid-params.jpg", "fake image content")
	imageFile.ContentType = "image/jpeg"
	s.DB.Save(imageFile)

	// 测试无效的尺寸参数
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/thumbnails/%d?size=invalid", imageFile.ID), nil, s.TestUser)
	s.AssertErrorResponse(response, http.StatusBadRequest)

	// 测试无效的宽度参数
	response = s.MakeRequest("GET", fmt.Sprintf("/api/v1/thumbnails/%d?width=invalid", imageFile.ID), nil, s.TestUser)
	s.AssertErrorResponse(response, http.StatusBadRequest)

	// 测试过大的尺寸
	response = s.MakeRequest("GET", fmt.Sprintf("/api/v1/thumbnails/%d?width=10000&height=10000", imageFile.ID), nil, s.TestUser)
	s.AssertErrorResponse(response, http.StatusBadRequest)
}

func (s *ThumbnailHandlerTestSuite) TestThumbnailAdminAccess() {
	// 创建其他用户的图片文件
	otherImageFile := s.CreateTestFile(999, "admin-access.jpg", "fake image content")
	otherImageFile.ContentType = "image/jpeg"
	s.DB.Save(otherImageFile)

	// 创建缩略图记录
	thumbnail := &models.Thumbnail{
		FileID:      otherImageFile.ID,
		Size:        "medium",
		Width:       300,
		Height:      200,
		Path: "/thumbnails/admin-access-medium.jpg",
		ContentType: "image/jpeg",
		Status:      models.ThumbnailStatusReady,
	}
	s.DB.Create(thumbnail)

	// 管理员应该能够访问任何用户的缩略图
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/admin/thumbnails/%d", otherImageFile.ID), nil, s.AdminUser)
	
	// 验证管理员可以访问
	s.Equal(http.StatusFound, response.Code)
	s.NotEmpty(response.Header().Get("Location"))
}

func TestThumbnailHandler(t *testing.T) {
	suite.Run(t, new(ThumbnailHandlerTestSuite))
}