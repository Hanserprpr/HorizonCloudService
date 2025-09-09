package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"file-service/internal/models"
	"file-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type FolderHandlerTestSuite struct {
	TestSuite
}

func (s *FolderHandlerTestSuite) TestCreateFolder() {
	// 创建文件夹请求
	folderData := map[string]interface{}{
		"name":        "test-folder",
		"description": "Test folder description",
	}

	// 发送创建请求
	response := s.MakeRequest("POST", "/api/v1/folders", folderData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusCreated)
	data := result["data"].(map[string]interface{})
	
	s.Equal("test-folder", data["name"])
	s.Equal("Test folder description", data["description"])
	s.Equal(s.TestUser.ID, uint(data["user_id"].(float64)))
}

func (s *FolderHandlerTestSuite) TestCreateNestedFolder() {
	// 创建父文件夹
	parent := s.CreateTestFolder(s.TestUser.ID, "parent-folder", nil)

	// 在父文件夹中创建子文件夹
	folderData := map[string]interface{}{
		"name":      "child-folder",
		"parent_id": parent.ID,
	}

	// 发送创建请求
	response := s.MakeRequest("POST", "/api/v1/folders", folderData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response, http.StatusCreated)
	data := result["data"].(map[string]interface{})
	
	s.Equal("child-folder", data["name"])
	s.Equal(float64(parent.ID), data["parent_id"])
	s.Equal(s.TestUser.ID, uint(data["user_id"].(float64)))
}

func (s *FolderHandlerTestSuite) TestCreateFolderDuplicateName() {
	// 创建第一个文件夹
	s.CreateTestFolder(s.TestUser.ID, "duplicate-name", nil)

	// 尝试创建同名文件夹
	folderData := map[string]interface{}{
		"name": "duplicate-name",
	}

	response := s.MakeRequest("POST", "/api/v1/folders", folderData, s.TestUser)
	
	// 应该返回错误
	s.AssertErrorResponse(response, http.StatusConflict)
}

func (s *FolderHandlerTestSuite) TestGetFolder() {
	// 创建测试文件夹
	folder := s.CreateTestFolder(s.TestUser.ID, "get-test-folder", nil)

	// 获取文件夹
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/folders/%d", folder.ID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response)
	data := result["data"].(map[string]interface{})
	
	s.Equal(folder.Name, data["name"])
	s.Equal(float64(folder.ID), data["id"])
	s.Equal(float64(folder.UserID), data["user_id"])
}

func (s *FolderHandlerTestSuite) TestGetFolderNotFound() {
	// 尝试获取不存在的文件夹
	response := s.MakeRequest("GET", "/api/v1/folders/99999", nil, s.TestUser)
	
	// 验证错误响应
	s.AssertErrorResponse(response, http.StatusNotFound)
}

func (s *FolderHandlerTestSuite) TestGetFolderUnauthorized() {
	// 创建其他用户的文件夹
	otherFolder := s.CreateTestFolder(999, "other-folder", nil)

	// 尝试访问其他用户的文件夹
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/folders/%d", otherFolder.ID), nil, s.TestUser)
	
	// 验证错误响应
	s.AssertErrorResponse(response, http.StatusNotFound) // 返回404而不是403，避免信息泄露
}

func (s *FolderHandlerTestSuite) TestListFolders() {
	// 创建多个测试文件夹
	s.CreateTestFolder(s.TestUser.ID, "folder1", nil)
	s.CreateTestFolder(s.TestUser.ID, "folder2", nil)
	s.CreateTestFolder(s.TestUser.ID, "folder3", nil)

	// 获取文件夹列表
	response := s.MakeRequest("GET", "/api/v1/folders?page=1&page_size=2", nil, s.TestUser)
	
	// 验证分页响应
	result := s.AssertPaginatedResponse(response)
	data := result["data"].(map[string]interface{})
	
	s.Equal(float64(1), data["page"])
	s.Equal(float64(2), data["page_size"])
	s.GreaterOrEqual(data["total"], float64(3))
	
	folders := data["data"].([]interface{})
	s.Len(folders, 2) // 应该返回2个文件夹
}

func (s *FolderHandlerTestSuite) TestListRootFolders() {
	// 创建根目录文件夹
	s.CreateTestFolder(s.TestUser.ID, "root-folder", nil)
	
	// 创建父文件夹和子文件夹
	parent := s.CreateTestFolder(s.TestUser.ID, "parent", nil)
	child := s.CreateTestFolder(s.TestUser.ID, "child", &parent.ID)

	// 只获取根目录文件夹
	response := s.MakeRequest("GET", "/api/v1/folders?parent_id=root", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertPaginatedResponse(response)
	data := result["data"].(map[string]interface{})
	
	folders := data["data"].([]interface{})
	// 应该至少有一个根目录文件夹
	s.GreaterOrEqual(len(folders), 1)
	
	// 验证返回的都是根目录文件夹
	for _, folderInterface := range folders {
		folder := folderInterface.(map[string]interface{})
		s.Nil(folder["parent_id"])
	}
}

func (s *FolderHandlerTestSuite) TestListSubfolders() {
	// 创建父文件夹
	parent := s.CreateTestFolder(s.TestUser.ID, "parent-folder", nil)
	
	// 在父文件夹中创建多个子文件夹
	s.CreateTestFolder(s.TestUser.ID, "sub1", &parent.ID)
	s.CreateTestFolder(s.TestUser.ID, "sub2", &parent.ID)

	// 获取子文件夹列表
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/folders?parent_id=%d", parent.ID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertPaginatedResponse(response)
	data := result["data"].(map[string]interface{})
	
	folders := data["data"].([]interface{})
	s.Len(folders, 2) // 应该有2个子文件夹
	
	// 验证返回的都是该父文件夹的子文件夹
	for _, folderInterface := range folders {
		folder := folderInterface.(map[string]interface{})
		s.Equal(float64(parent.ID), folder["parent_id"])
	}
}

func (s *FolderHandlerTestSuite) TestUpdateFolder() {
	// 创建测试文件夹
	folder := s.CreateTestFolder(s.TestUser.ID, "original-name", nil)

	// 准备更新数据
	updateData := map[string]interface{}{
		"name":        "updated-name",
		"description": "Updated description",
	}

	// 更新文件夹
	response := s.MakeRequest("PUT", fmt.Sprintf("/api/v1/folders/%d", folder.ID), updateData, s.TestUser)
	
	// 验证响应
	s.AssertSuccessResponse(response)

	// 验证数据库中的更改
	var updatedFolder models.Folder
	s.DB.First(&updatedFolder, folder.ID)
	s.Equal("updated-name", updatedFolder.Name)
	s.Equal("Updated description", updatedFolder.Description)
}

func (s *FolderHandlerTestSuite) TestUpdateFolderMove() {
	// 创建测试文件夹
	folder := s.CreateTestFolder(s.TestUser.ID, "move-me", nil)
	targetParent := s.CreateTestFolder(s.TestUser.ID, "target-parent", nil)

	// 移动文件夹
	updateData := map[string]interface{}{
		"parent_id": targetParent.ID,
	}

	response := s.MakeRequest("PUT", fmt.Sprintf("/api/v1/folders/%d", folder.ID), updateData, s.TestUser)
	
	// 验证响应
	s.AssertSuccessResponse(response)

	// 验证数据库中的更改
	var movedFolder models.Folder
	s.DB.First(&movedFolder, folder.ID)
	s.Equal(&targetParent.ID, movedFolder.ParentID)
}

func (s *FolderHandlerTestSuite) TestDeleteFolder() {
	// 创建测试文件夹
	folder := s.CreateTestFolder(s.TestUser.ID, "to-delete", nil)

	// 删除文件夹
	response := s.MakeRequest("DELETE", fmt.Sprintf("/api/v1/folders/%d", folder.ID), nil, s.TestUser)
	
	// 验证响应
	s.AssertSuccessResponse(response)

	// 验证文件夹已被软删除
	var deletedFolder models.Folder
	s.DB.Unscoped().First(&deletedFolder, folder.ID)
	s.NotNil(deletedFolder.DeletedAt)
}

func (s *FolderHandlerTestSuite) TestDeleteFolderWithContents() {
	// 创建父文件夹
	parent := s.CreateTestFolder(s.TestUser.ID, "parent-with-contents", nil)
	
	// 在父文件夹中创建子文件夹和文件
	child := s.CreateTestFolder(s.TestUser.ID, "child", &parent.ID)
	file := s.CreateTestFile(s.TestUser.ID, "file-in-parent.txt", "content")
	file.FolderID = &parent.ID
	s.DB.Save(file)

	// 尝试删除非空文件夹（应该递归删除）
	response := s.MakeRequest("DELETE", fmt.Sprintf("/api/v1/folders/%d", parent.ID), nil, s.TestUser)
	
	// 验证响应
	s.AssertSuccessResponse(response)

	// 验证父文件夹和子内容都被删除
	var deletedParent models.Folder
	s.DB.Unscoped().First(&deletedParent, parent.ID)
	s.NotNil(deletedParent.DeletedAt)
	
	var deletedChild models.Folder
	s.DB.Unscoped().First(&deletedChild, child.ID)
	s.NotNil(deletedChild.DeletedAt)
}

func (s *FolderHandlerTestSuite) TestGetFolderTree() {
	// 创建文件夹层次结构
	root := s.CreateTestFolder(s.TestUser.ID, "root", nil)
	level1 := s.CreateTestFolder(s.TestUser.ID, "level1", &root.ID)
	level2 := s.CreateTestFolder(s.TestUser.ID, "level2", &level1.ID)

	// 获取文件夹树
	response := s.MakeRequest("GET", "/api/v1/folders/tree", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response)
	tree := result["data"].([]interface{})
	
	s.GreaterOrEqual(len(tree), 1) // 应该至少有一个根文件夹
}

func (s *FolderHandlerTestSuite) TestGetFolderContents() {
	// 创建测试文件夹
	folder := s.CreateTestFolder(s.TestUser.ID, "contents-test", nil)
	
	// 在文件夹中添加内容
	subfolder := s.CreateTestFolder(s.TestUser.ID, "subfolder", &folder.ID)
	file := s.CreateTestFile(s.TestUser.ID, "file-in-folder.txt", "content")
	file.FolderID = &folder.ID
	s.DB.Save(file)

	// 获取文件夹内容
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/folders/%d/contents", folder.ID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response)
	data := result["data"].(map[string]interface{})
	
	folders := data["folders"].([]interface{})
	files := data["files"].([]interface{})
	
	s.Len(folders, 1)
	s.Len(files, 1)
	
	// 验证文件夹内容
	folderData := folders[0].(map[string]interface{})
	s.Equal("subfolder", folderData["name"])
	
	fileData := files[0].(map[string]interface{})
	s.Equal("file-in-folder.txt", fileData["name"])
}

func (s *FolderHandlerTestSuite) TestMoveFolderWithContents() {
	// 创建源文件夹和目标文件夹
	source := s.CreateTestFolder(s.TestUser.ID, "source", nil)
	target := s.CreateTestFolder(s.TestUser.ID, "target", nil)
	
	// 在源文件夹中添加内容
	subfolder := s.CreateTestFolder(s.TestUser.ID, "subfolder", &source.ID)
	file := s.CreateTestFile(s.TestUser.ID, "file.txt", "content")
	file.FolderID = &source.ID
	s.DB.Save(file)

	// 移动整个文件夹
	moveData := map[string]interface{}{
		"target_folder_id": target.ID,
	}

	response := s.MakeRequest("PUT", fmt.Sprintf("/api/v1/folders/%d/move", source.ID), moveData, s.TestUser)
	
	// 验证响应
	s.AssertSuccessResponse(response)

	// 验证文件夹被移动
	var movedFolder models.Folder
	s.DB.First(&movedFolder, source.ID)
	s.Equal(&target.ID, movedFolder.ParentID)
}

func (s *FolderHandlerTestSuite) TestGetFolderPath() {
	// 创建文件夹层次结构
	root := s.CreateTestFolder(s.TestUser.ID, "root", nil)
	level1 := s.CreateTestFolder(s.TestUser.ID, "level1", &root.ID)
	level2 := s.CreateTestFolder(s.TestUser.ID, "level2", &level1.ID)

	// 获取深层文件夹的路径
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/folders/%d/path", level2.ID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response)
	path := result["data"].(map[string]interface{})
	
	pathString := path["path"].(string)
	s.Contains(pathString, "root")
	s.Contains(pathString, "level1")
	s.Contains(pathString, "level2")
	
	breadcrumbs := path["breadcrumbs"].([]interface{})
	s.Len(breadcrumbs, 3) // root, level1, level2
}

func (s *FolderHandlerTestSuite) TestCopyFolder() {
	// 创建源文件夹
	source := s.CreateTestFolder(s.TestUser.ID, "copy-source", nil)
	target := s.CreateTestFolder(s.TestUser.ID, "copy-target", nil)
	
	// 在源文件夹中添加内容
	subfolder := s.CreateTestFolder(s.TestUser.ID, "subfolder", &source.ID)
	file := s.CreateTestFile(s.TestUser.ID, "file.txt", "content")
	file.FolderID = &source.ID
	s.DB.Save(file)

	// 复制文件夹
	copyData := map[string]interface{}{
		"target_folder_id": target.ID,
	}

	response := s.MakeRequest("POST", fmt.Sprintf("/api/v1/folders/%d/copy", source.ID), copyData, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response)
	data := result["data"].(map[string]interface{})
	
	// 验证复制的文件夹
	s.Contains(data["name"], "copy")
	s.Equal(float64(target.ID), data["parent_id"])
	
	// 验证原文件夹未改变
	var originalFolder models.Folder
	s.DB.First(&originalFolder, source.ID)
	s.Equal("copy-source", originalFolder.Name)
	s.Nil(originalFolder.ParentID)
}

func (s *FolderHandlerTestSuite) TestSearchFolders() {
	// 创建测试文件夹
	s.CreateTestFolder(s.TestUser.ID, "search-target", nil)
	s.CreateTestFolder(s.TestUser.ID, "another-folder", nil)
	s.CreateTestFolder(s.TestUser.ID, "search-result", nil)

	// 搜索文件夹
	response := s.MakeRequest("GET", "/api/v1/folders/search?q=search", nil, s.TestUser)
	
	// 验证响应
	result := s.AssertPaginatedResponse(response)
	data := result["data"].(map[string]interface{})
	
	folders := data["data"].([]interface{})
	s.GreaterOrEqual(len(folders), 1) // 至少应该找到包含"search"的文件夹
}

func (s *FolderHandlerTestSuite) TestGetFolderStats() {
	// 创建测试文件夹
	folder := s.CreateTestFolder(s.TestUser.ID, "stats-folder", nil)
	
	// 在文件夹中添加内容
	s.CreateTestFolder(s.TestUser.ID, "sub1", &folder.ID)
	s.CreateTestFolder(s.TestUser.ID, "sub2", &folder.ID)
	
	file1 := s.CreateTestFile(s.TestUser.ID, "file1.txt", strings.Repeat("x", 1024))
	file1.FolderID = &folder.ID
	s.DB.Save(file1)
	
	file2 := s.CreateTestFile(s.TestUser.ID, "file2.txt", strings.Repeat("y", 2048))
	file2.FolderID = &folder.ID
	s.DB.Save(file2)

	// 获取文件夹统计信息
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/folders/%d/stats", folder.ID), nil, s.TestUser)
	
	// 验证响应
	result := s.AssertSuccessResponse(response)
	stats := result["data"].(map[string]interface{})
	
	s.Equal(float64(2), stats["folder_count"])
	s.Equal(float64(2), stats["file_count"])
	s.GreaterOrEqual(stats["total_size"], float64(3072)) // 1024 + 2048
}

func (s *FolderHandlerTestSuite) TestUnauthorizedAccess() {
	// 测试无token访问
	response := s.MakeRequest("GET", "/api/v1/folders", nil, nil)
	s.AssertErrorResponse(response, http.StatusUnauthorized)
}

func (s *FolderHandlerTestSuite) TestInvalidFolderID() {
	// 测试无效的文件夹ID
	response := s.MakeRequest("GET", "/api/v1/folders/invalid", nil, s.TestUser)
	s.AssertErrorResponse(response, http.StatusBadRequest)
}

func (s *FolderHandlerTestSuite) TestAdminAccess() {
	// 创建其他用户的文件夹
	otherUserFolder := s.CreateTestFolder(999, "admin-test-folder", nil)

	// 管理员应该能够访问任何用户的文件夹（通过管理员API）
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/admin/users/999/folders"), nil, s.AdminUser)
	
	// 验证管理员可以访问
	result := s.AssertPaginatedResponse(response)
	data := result["data"].(map[string]interface{})
	folders := data["data"].([]interface{})
	s.GreaterOrEqual(len(folders), 1)
}

func (s *FolderHandlerTestSuite) TestFolderPermissions() {
	// 创建文件夹
	folder := s.CreateTestFolder(s.TestUser.ID, "permission-test", nil)

	// 测试所有者可以访问
	response := s.MakeRequest("GET", fmt.Sprintf("/api/v1/folders/%d", folder.ID), nil, s.TestUser)
	s.AssertSuccessResponse(response)

	// 创建另一个用户，测试无法访问其他用户的文件夹
	// (这个测试在上面的TestGetFolderUnauthorized中已经覆盖)
}

func (s *FolderHandlerTestSuite) TestCreateFolderValidation() {
	// 测试空名称
	folderData := map[string]interface{}{
		"name": "",
	}
	response := s.MakeRequest("POST", "/api/v1/folders", folderData, s.TestUser)
	s.AssertErrorResponse(response, http.StatusBadRequest)

	// 测试名称过长
	longName := strings.Repeat("a", 256)
	folderData = map[string]interface{}{
		"name": longName,
	}
	response = s.MakeRequest("POST", "/api/v1/folders", folderData, s.TestUser)
	s.AssertErrorResponse(response, http.StatusBadRequest)

	// 测试无效父文件夹ID
	folderData = map[string]interface{}{
		"name":      "valid-name",
		"parent_id": 99999, // 不存在的父文件夹
	}
	response = s.MakeRequest("POST", "/api/v1/folders", folderData, s.TestUser)
	s.AssertErrorResponse(response, http.StatusBadRequest)
}

func TestFolderHandler(t *testing.T) {
	suite.Run(t, new(FolderHandlerTestSuite))
}