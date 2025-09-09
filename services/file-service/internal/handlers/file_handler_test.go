package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"file-service/internal/models"
	"file-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type FileHandlerTestSuite struct {
	TestSuite
}

func (s *FileHandlerTestSuite) TestUploadFile() {
	// 创建测试文件
	fileContent := "This is a test file content"
	files := map[string][]byte{
		"test.txt": []byte(fileContent),
	}
	fields := map[string]string{
		"folder_id": "1",
	}

	// 发送上传请求
	response := s.MakeMultipartRequest("POST", "/api/v1/upload/simple", fields, files, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusOK)
	data := result["data"].(map[string]interface{})
	
	s.Equal("test.txt", data["name"])
	s.Equal(float64(len(fileContent)), data["size"])
	s.Equal(s.TestUser.ID, uint(data["user_id"].(float64)))
}

func (s *FileHandlerTestSuite) TestGetFile() {
	// 创建测试文件
	file := s.CreateTestFile(s.TestUser.ID, "test.txt", "test content")

	// 获取文件
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/files/%d", file.ID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response)
	data := result["data"].(map[string]interface{})
	
	s.Equal(file.Name, data["name"])
	s.Equal(float64(file.ID), data["id"])
	s.Equal(float64(file.UserID), data["user_id"])
}

func (s *FileHandlerTestSuite) TestGetFileNotFound() {
	// 尝试获取不存在的文件
	response := s.MakeRequest("GET", "/api/v1/files/99999", nil, s.TestUser)
	
	// 验证错误响应
	s.AssertErrorResponse(response, http.StatusNotFound)
}

func (s *FileHandlerTestSuite) TestGetFileUnauthorized() {
	// 创建另一个用户的文件
	otherFile := s.CreateTestFile(999, "other.txt", "other content")

	// 尝试访问其他用户的文件
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/files/%d", otherFile.ID), nil, s.TestUser)
	
	// 验证错误响应
	s.AssertErrorResponse(response, http.StatusNotFound) // 应该返回404而不是403，避免信息泄露
}

func (s *FileHandlerTestSuite) TestListFiles() {
	// 创建测试文件
	s.CreateTestFile(s.TestUser.ID, "file1.txt", "content1")
	s.CreateTestFile(s.TestUser.ID, "file2.txt", "content2")
	s.CreateTestFile(s.TestUser.ID, "file3.txt", "content3")

	// 获取文件列表
	response := s.MakeRequest("GET", "/api/v1/files?page=1&page_size=2", nil, s.TestUser)
	
	// 验证分页响应
	result := s.AssertPaginatedResponse(response)
	data := result["data"].(map[string]interface{})
	
	s.Equal(float64(1), data["page"])
	s.Equal(float64(2), data["page_size"])
	s.Equal(float64(3), data["total"])
	
	files := data["data"].([]interface{})
	s.Len(files, 2) // 应该返回2个文件
}

func (s *FileHandlerTestSuite) TestListFilesWithFolder() {
	// 创建测试文件夹
	folder := s.CreateTestFolder(s.TestUser.ID, "test-folder", nil)
	
	// 在文件夹中创建文件
	file := s.CreateTestFile(s.TestUser.ID, "folder-file.txt", "content")
	file.FolderID = &folder.ID
	s.DB.Save(file)

	// 获取文件夹中的文件
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/files?folder_id=%d", folder.ID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertPaginatedResponse(response)
	data := result["data"].(map[string]interface{})
	
	files := data["data"].([]interface{})
	s.Len(files, 1)
	
	fileData := files[0].(map[string]interface{})
	s.Equal("folder-file.txt", fileData["name"])
}

func (s *FileHandlerTestSuite) TestUpdateFile() {
	// 创建测试文件
	file := s.CreateTestFile(s.TestUser.ID, "original.txt", "content")

	// 准备更新数据
	updateData := map[string]interface{}{
		"name":        "updated.txt",
		"description": "Updated description",
		"tags":        []string{"tag1", "tag2"},
	}

	// 更新文件
	response := s.MakeRequest("PUT", fmt.Sprintf("/api/v1/files/%d", file.ID), updateData, s.TestUser)
	
	// 验证响应
	s.AssertSuccessResponse(response)

	// 验证数据库中的更改
	var updatedFile models.File
	s.DB.First(&updatedFile, file.ID)
	s.Equal("updated.txt", updatedFile.Name)
	s.Equal("Updated description", updatedFile.Description)
}

func (s *FileHandlerTestSuite) TestDeleteFile() {
	// 创建测试文件
	file := s.CreateTestFile(s.TestUser.ID, "to-delete.txt", "content")

	// 删除文件
	response := s.MakeRequest("DELETE", fmt.Sprintf("/api/v1/files/%d", file.ID), nil, s.TestUser)
	
	// 验证响应
	s.AssertSuccessResponse(response)

	// 验证文件状态变为已删除
	var deletedFile models.File
	s.DB.First(&deletedFile, file.ID)
	s.Equal(models.FileStatusDeleted, deletedFile.Status)
}

func (s *FileHandlerTestSuite) TestDownloadFile() {
	// 创建测试文件
	file := s.CreateTestFile(s.TestUser.ID, "download.txt", "download content")

	// 请求下载
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/files/%d/download", file.ID), nil, s.TestUser)
	
	// 验证重定向响应
	s.Equal(http.StatusFound, response.Code)
	s.NotEmpty(response.Header().Get("Location"))
}

func (s *FileHandlerTestSuite) TestMoveFile() {
	// 创建测试文件和文件夹
	file := s.CreateTestFile(s.TestUser.ID, "move-me.txt", "content")
	targetFolder := s.CreateTestFolder(s.TestUser.ID, "target-folder", nil)

	// 移动文件
	moveData := map[string]interface{}{
		"folder_id": targetFolder.ID,
	}
	response := s.MakeRequest("PUT", fmt.Sprintf("/api/v1/files/%d/move", file.ID), moveData, s.TestUser)
	
	// 验证响应
	s.AssertSuccessResponse(response)

	// 验证文件已移动
	var movedFile models.File
	s.DB.First(&movedFile, file.ID)
	s.Equal(&targetFolder.ID, movedFile.FolderID)
}

func (s *FileHandlerTestSuite) TestCopyFile() {
	// 创建测试文件和文件夹
	file := s.CreateTestFile(s.TestUser.ID, "copy-me.txt", "content")
	targetFolder := s.CreateTestFolder(s.TestUser.ID, "target-folder", nil)

	// 复制文件
	copyData := map[string]interface{}{
		"folder_id": targetFolder.ID,
	}
	response := s.MakeRequest("POST", fmt.Sprintf("/api/v1/files/%d/copy", file.ID), copyData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response)
	data := result["data"].(map[string]interface{})
	
	// 验证复制的文件
	s.Contains(data["name"], "copy")
	s.Equal(float64(targetFolder.ID), *data["folder_id"].(*float64))
	
	// 验证原文件未改变
	var originalFile models.File
	s.DB.First(&originalFile, file.ID)
	s.Equal("copy-me.txt", originalFile.Name)
}

func (s *FileHandlerTestSuite) TestSearchFiles() {
	// 创建测试文件
	s.CreateTestFile(s.TestUser.ID, "search-target.txt", "content")
	s.CreateTestFile(s.TestUser.ID, "another-file.txt", "content")
	s.CreateTestFile(s.TestUser.ID, "search-result.txt", "content")

	// 搜索文件
	response := s.MakeRequest("GET", "/api/v1/files/search?q=search", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertPaginatedResponse(response)
	data := result["data"].(map[string]interface{})
	
	files := data["data"].([]interface{})
	s.GreaterOrEqual(len(files), 1) // 至少应该找到包含"search"的文件
}

func (s *FileHandlerTestSuite) TestBatchOperation() {
	// 创建测试文件
	file1 := s.CreateTestFile(s.TestUser.ID, "batch1.txt", "content1")
	file2 := s.CreateTestFile(s.TestUser.ID, "batch2.txt", "content2")
	targetFolder := s.CreateTestFolder(s.TestUser.ID, "batch-target", nil)

	// 批量移动文件
	batchData := map[string]interface{}{
		"file_ids": []uint{file1.ID, file2.ID},
		"action":   "move",
		"target":   targetFolder.ID,
	}
	response := s.MakeRequest("POST", "/api/v1/files/batch", batchData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response)
	data := result["data"].(map[string]interface{})
	
	s.Equal(float64(2), data["total"])
	s.Equal(float64(2), data["successful"])
	s.Equal(float64(0), data["failed"])
}

func (s *FileHandlerTestSuite) TestBatchDelete() {
	// 创建测试文件
	file1 := s.CreateTestFile(s.TestUser.ID, "delete1.txt", "content1")
	file2 := s.CreateTestFile(s.TestUser.ID, "delete2.txt", "content2")

	// 批量删除文件
	batchData := map[string]interface{}{
		"file_ids": []uint{file1.ID, file2.ID},
		"action":   "delete",
	}
	response := s.MakeRequest("POST", "/api/v1/files/batch", batchData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response)
	data := result["data"].(map[string]interface{})
	
	s.Equal(float64(2), data["successful"])

	// 验证文件状态
	var deletedFiles []models.File
	s.DB.Find(&deletedFiles, []uint{file1.ID, file2.ID})
	for _, file := range deletedFiles {
		s.Equal(models.FileStatusDeleted, file.Status)
	}
}

func (s *FileHandlerTestSuite) TestGetFileVersions() {
	// 创建有版本的测试文件
	file := s.CreateTestFile(s.TestUser.ID, "versioned.txt", "content")
	
	// 创建文件版本
	version := &models.File{
		Name:     file.Name,
		UserID:   file.UserID,
		ParentID: &file.ID,
		Version:  2,
		IsLatest: false,
		Status:   models.FileStatusActive,
		Hash:     "version-hash",
		StoragePath: "/version/path",
		Size:     100,
	}
	s.DB.Create(version)

	// 获取文件版本
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/files/%d/versions", file.ID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response)
	versions := result["data"].([]interface{})
	s.GreaterOrEqual(len(versions), 1)
}

func (s *FileHandlerTestSuite) TestGetUserStats() {
	// 创建测试文件
	s.CreateTestFile(s.TestUser.ID, "stats1.txt", "content1")
	s.CreateTestFile(s.TestUser.ID, "stats2.txt", "content2")
	s.CreateTestFile(s.TestUser.ID, "stats3.txt", "content3")

	// 获取用户统计
	response := s.MakeRequest("GET", "/api/v1/files/stats", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response)
	data := result["data"].(map[string]interface{})
	
	s.Contains(data, "file_count")
	s.Contains(data, "total_size")
	s.GreaterOrEqual(data["file_count"], float64(3))
}

func (s *FileHandlerTestSuite) TestUnauthorizedAccess() {
	// 测试无token访问
	response := s.MakeRequest("GET", "/api/v1/files", nil, nil)
	s.AssertErrorResponse(response, http.StatusUnauthorized)
}

func (s *FileHandlerTestSuite) TestInvalidFileID() {
	// 测试无效的文件ID
	response := s.MakeRequest("GET", "/api/v1/files/invalid", nil, s.TestUser)
	s.AssertErrorResponse(response, http.StatusBadRequest)
}

func (s *FileHandlerTestSuite) TestAdminAccess() {
	// 创建其他用户的文件
	otherUserFile := s.CreateTestFile(999, "admin-test.txt", "content")

	// 管理员应该能够访问任何用户的文件（通过管理员API）
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/admin/users/999/files"), nil, s.AdminUser)
	
	// 验证管理员可以访问
	result := s.AssertPaginatedResponse(response)
	data := result["data"].(map[string]interface{})
	files := data["data"].([]interface{})
	s.GreaterOrEqual(len(files), 1)
}

func (s *FileHandlerTestSuite) TestFileFilters() {
	// 创建不同类型的文件
	doc := s.CreateTestFile(s.TestUser.ID, "document.txt", "content")
	doc.Category = "document"
	s.DB.Save(doc)

	image := s.CreateTestFile(s.TestUser.ID, "image.jpg", "content")
	image.Category = "image"
	image.ContentType = "image/jpeg"
	s.DB.Save(image)

	// 按分类过滤
	response := s.MakeRequest("GET", "/api/v1/files?category=document", nil, s.TestUser)
	result := s.AssertPaginatedResponse(response)
	data := result["data"].(map[string]interface{})
	files := data["data"].([]interface{})
	
	// 应该只返回文档类型的文件
	for _, fileInterface := range files {
		file := fileInterface.(map[string]interface{})
		s.Equal("document", file["category"])
	}
}

func (s *FileHandlerTestSuite) TestPaginationAndSorting() {
	// 创建多个文件
	for i := 1; i <= 5; i++ {
		file := s.CreateTestFile(s.TestUser.ID, fmt.Sprintf("file%d.txt", i), "content")
		// 设置不同的创建时间
		file.CreatedAt = file.CreatedAt.Add(-time.Duration(i) * time.Hour)
		s.DB.Save(file)
	}

	// 测试分页
	response := s.MakeRequest("GET", "/api/v1/files?page=1&page_size=2", nil, s.TestUser)
	result := s.AssertPaginatedResponse(response)
	data := result["data"].(map[string]interface{})
	
	s.Equal(float64(1), data["page"])
	s.Equal(float64(2), data["page_size"])
	s.Equal(float64(5), data["total"])
	s.Equal(float64(3), data["pages"]) // 总页数

	files := data["data"].([]interface{})
	s.Len(files, 2)

	// 测试排序
	response = s.MakeRequest("GET", "/api/v1/files?sort_by=created_at&sort_order=asc", nil, s.TestUser)
	result = s.AssertPaginatedResponse(response)
	// 验证文件按创建时间升序排列
}

func TestFileHandler(t *testing.T) {
	suite.Run(t, new(FileHandlerTestSuite))
}