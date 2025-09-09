package repository

import (
	"context"
	"testing"
	"user-service/internal/models"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// UserRepositoryTestSuite 用户仓库测试套件
type UserRepositoryTestSuite struct {
	suite.Suite
	repo UserRepository
	ctx  context.Context
}

// SetupSuite 设置测试套件
func (s *UserRepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.repo = NewMockUserRepository()
}

// TestCreateUser 测试创建用户
func (s *UserRepositoryTestSuite) TestCreateUser() {
	user := &models.User{
		StudentID:    "test001",
		Password:     "hashedpassword",
		NickName:     "测试用户",
		Email:        "test@example.com",
		Phone:        "13800138000",
		Status:       1,
		StorageQuota: 5368709120,
		StorageUsed:  0,
	}
	
	err := s.repo.Create(s.ctx, user)
	assert.NoError(s.T(), err)
	assert.NotZero(s.T(), user.ID) // ID应该被分配
	
	// 测试重复学号
	duplicateUser := &models.User{
		StudentID: "test001", // 相同学号
		Password:  "anotherpassword",
		NickName:  "另一个用户",
	}
	
	err = s.repo.Create(s.ctx, duplicateUser)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "already exists")
}

// TestGetByID 测试根据ID获取用户
func (s *UserRepositoryTestSuite) TestGetByID() {
	// 创建用户
	user := &models.User{
		StudentID: "test002",
		Password:  "hashedpassword",
		NickName:  "测试用户2",
	}
	
	err := s.repo.Create(s.ctx, user)
	assert.NoError(s.T(), err)
	
	// 根据ID获取用户
	retrievedUser, err := s.repo.GetByID(s.ctx, user.ID)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), retrievedUser)
	assert.Equal(s.T(), user.StudentID, retrievedUser.StudentID)
	assert.Equal(s.T(), user.NickName, retrievedUser.NickName)
	
	// 获取不存在的用户
	_, err = s.repo.GetByID(s.ctx, 99999)
	assert.Error(s.T(), err)
}

// TestGetByStudentID 测试根据学号获取用户
func (s *UserRepositoryTestSuite) TestGetByStudentID() {
	// 创建用户
	user := &models.User{
		StudentID: "test003",
		Password:  "hashedpassword",
		NickName:  "测试用户3",
	}
	
	err := s.repo.Create(s.ctx, user)
	assert.NoError(s.T(), err)
	
	// 根据学号获取用户
	retrievedUser, err := s.repo.GetByStudentID(s.ctx, "test003")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), retrievedUser)
	assert.Equal(s.T(), user.ID, retrievedUser.ID)
	assert.Equal(s.T(), user.NickName, retrievedUser.NickName)
	
	// 获取不存在的学号
	_, err = s.repo.GetByStudentID(s.ctx, "notexist")
	assert.Error(s.T(), err)
}

// TestGetByEmail 测试根据邮箱获取用户
func (s *UserRepositoryTestSuite) TestGetByEmail() {
	// 创建用户
	user := &models.User{
		StudentID: "test004",
		Password:  "hashedpassword",
		NickName:  "测试用户4",
		Email:     "test004@example.com",
	}
	
	err := s.repo.Create(s.ctx, user)
	assert.NoError(s.T(), err)
	
	// 根据邮箱获取用户
	retrievedUser, err := s.repo.GetByEmail(s.ctx, "test004@example.com")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), retrievedUser)
	assert.Equal(s.T(), user.ID, retrievedUser.ID)
	assert.Equal(s.T(), user.StudentID, retrievedUser.StudentID)
	
	// 获取不存在的邮箱
	_, err = s.repo.GetByEmail(s.ctx, "notexist@example.com")
	assert.Error(s.T(), err)
}

// TestUpdateUser 测试更新用户
func (s *UserRepositoryTestSuite) TestUpdateUser() {
	// 创建用户
	user := &models.User{
		StudentID: "test005",
		Password:  "hashedpassword",
		NickName:  "测试用户5",
	}
	
	err := s.repo.Create(s.ctx, user)
	assert.NoError(s.T(), err)
	
	// 更新用户信息
	user.NickName = "更新后的昵称"
	user.Email = "updated@example.com"
	
	err = s.repo.Update(s.ctx, user)
	assert.NoError(s.T(), err)
	
	// 验证更新结果
	retrievedUser, err := s.repo.GetByID(s.ctx, user.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "更新后的昵称", retrievedUser.NickName)
	assert.Equal(s.T(), "updated@example.com", retrievedUser.Email)
}

// TestDeleteUser 测试删除用户
func (s *UserRepositoryTestSuite) TestDeleteUser() {
	// 创建用户
	user := &models.User{
		StudentID: "test006",
		Password:  "hashedpassword",
		NickName:  "测试用户6",
	}
	
	err := s.repo.Create(s.ctx, user)
	assert.NoError(s.T(), err)
	
	// 删除用户
	err = s.repo.Delete(s.ctx, user.ID)
	assert.NoError(s.T(), err)
	
	// 验证用户已被删除
	_, err = s.repo.GetByID(s.ctx, user.ID)
	assert.Error(s.T(), err)
}

// TestListUsers 测试获取用户列表
func (s *UserRepositoryTestSuite) TestListUsers() {
	// 创建多个用户
	users := []*models.User{
		{StudentID: "list001", Password: "pass1", NickName: "用户1", Status: 1},
		{StudentID: "list002", Password: "pass2", NickName: "用户2", Status: 1},
		{StudentID: "list003", Password: "pass3", NickName: "用户3", Status: 2},
	}
	
	for _, user := range users {
		err := s.repo.Create(s.ctx, user)
		assert.NoError(s.T(), err)
	}
	
	// 获取所有用户
	allUsers, total, err := s.repo.List(s.ctx, 0, 10, 0)
	assert.NoError(s.T(), err)
	assert.True(s.T(), len(allUsers) >= 3)
	assert.True(s.T(), total >= 3)
	
	// 按状态筛选
	activeUsers, activeTotal, err := s.repo.List(s.ctx, 0, 10, 1)
	assert.NoError(s.T(), err)
	assert.True(s.T(), len(activeUsers) >= 2)
	assert.True(s.T(), activeTotal >= 2)
	
	// 测试分页
	pageUsers, pageTotal, err := s.repo.List(s.ctx, 0, 1, 0)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(pageUsers))
	assert.True(s.T(), pageTotal >= 3)
}

// TestSearchUsers 测试搜索用户
func (s *UserRepositoryTestSuite) TestSearchUsers() {
	// 创建用于搜索的用户
	users := []*models.User{
		{StudentID: "search001", Password: "pass1", NickName: "张三", Email: "zhangsan@example.com"},
		{StudentID: "search002", Password: "pass2", NickName: "李四", Email: "lisi@example.com"},
		{StudentID: "search003", Password: "pass3", NickName: "王五", Email: "wangwu@example.com"},
	}
	
	for _, user := range users {
		err := s.repo.Create(s.ctx, user)
		assert.NoError(s.T(), err)
	}
	
	// 搜索学号
	results, total, err := s.repo.Search(s.ctx, "search001", 0, 10)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), total)
	assert.Equal(s.T(), "search001", results[0].StudentID)
	
	// 搜索昵称
	results, total, err = s.repo.Search(s.ctx, "张三", 0, 10)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), total)
	assert.Equal(s.T(), "张三", results[0].NickName)
	
	// 搜索邮箱
	results, total, err = s.repo.Search(s.ctx, "lisi@example.com", 0, 10)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), total)
	assert.Equal(s.T(), "lisi@example.com", results[0].Email)
	
	// 搜索不存在的关键词
	results, total, err = s.repo.Search(s.ctx, "notexist", 0, 10)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(0), total)
	assert.Equal(s.T(), 0, len(results))
}

// TestGetUsersByRole 测试根据角色获取用户
func (s *UserRepositoryTestSuite) TestGetUsersByRole() {
	// 创建不同角色的用户
	users := []*models.User{
		{StudentID: "role001", Password: "pass1", NickName: "管理员1", RoleID: 2},
		{StudentID: "role002", Password: "pass2", NickName: "管理员2", RoleID: 2},
		{StudentID: "role003", Password: "pass3", NickName: "普通用户1", RoleID: 1},
	}
	
	for _, user := range users {
		err := s.repo.Create(s.ctx, user)
		assert.NoError(s.T(), err)
	}
	
	// 获取管理员用户
	admins, adminTotal, err := s.repo.GetUsersByRole(s.ctx, 2, 0, 10)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(2), adminTotal)
	assert.Equal(s.T(), 2, len(admins))
	
	// 获取普通用户
	normalUsers, normalTotal, err := s.repo.GetUsersByRole(s.ctx, 1, 0, 10)
	assert.NoError(s.T(), err)
	assert.True(s.T(), normalTotal >= 1)
	assert.True(s.T(), len(normalUsers) >= 1)
}

// TestStorageQuota 测试存储配额相关操作
func (s *UserRepositoryTestSuite) TestStorageQuota() {
	// 创建用户
	user := &models.User{
		StudentID:    "quota001",
		Password:     "hashedpassword",
		NickName:     "配额测试用户",
		StorageQuota: 5368709120, // 5GB
		StorageUsed:  0,
	}
	
	err := s.repo.Create(s.ctx, user)
	assert.NoError(s.T(), err)
	
	// 更新存储使用量
	err = s.repo.UpdateStorageUsed(s.ctx, user.ID, 1073741824) // 1GB
	assert.NoError(s.T(), err)
	
	// 获取配额信息
	quota, err := s.repo.GetUserQuota(s.ctx, user.ID)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), quota)
	assert.Equal(s.T(), user.ID, quota.UserID)
	assert.Equal(s.T(), int64(5368709120), quota.StorageQuota)
	assert.Equal(s.T(), int64(1073741824), quota.StorageUsed)
	assert.Equal(s.T(), 20.0, quota.UsagePercent) // 1GB / 5GB * 100
	
	// 更新存储配额
	err = s.repo.UpdateStorageQuota(s.ctx, user.ID, 10737418240) // 10GB
	assert.NoError(s.T(), err)
	
	// 验证配额更新
	quota, err = s.repo.GetUserQuota(s.ctx, user.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(10737418240), quota.StorageQuota)
}

// TestUpdateUserStatus 测试更新用户状态
func (s *UserRepositoryTestSuite) TestUpdateUserStatus() {
	// 创建用户
	user := &models.User{
		StudentID: "status001",
		Password:  "hashedpassword",
		NickName:  "状态测试用户",
		Status:    1,
	}
	
	err := s.repo.Create(s.ctx, user)
	assert.NoError(s.T(), err)
	
	// 更新用户状态为禁用
	err = s.repo.UpdateStatus(s.ctx, user.ID, 2)
	assert.NoError(s.T(), err)
	
	// 验证状态更新
	retrievedUser, err := s.repo.GetByID(s.ctx, user.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, retrievedUser.Status)
}

// TestUpdateLoginInfo 测试更新登录信息
func (s *UserRepositoryTestSuite) TestUpdateLoginInfo() {
	// 创建用户
	user := &models.User{
		StudentID:  "login001",
		Password:   "hashedpassword",
		NickName:   "登录测试用户",
		LoginCount: 0,
	}
	
	err := s.repo.Create(s.ctx, user)
	assert.NoError(s.T(), err)
	
	// 更新登录信息
	err = s.repo.UpdateLoginInfo(s.ctx, user.ID)
	assert.NoError(s.T(), err)
	
	// 验证登录次数增加
	retrievedUser, err := s.repo.GetByID(s.ctx, user.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, retrievedUser.LoginCount)
}

// TestActivityLogs 测试活动日志
func (s *UserRepositoryTestSuite) TestActivityLogs() {
	// 创建用户
	user := &models.User{
		StudentID: "activity001",
		Password:  "hashedpassword",
		NickName:  "活动日志测试用户",
	}
	
	err := s.repo.Create(s.ctx, user)
	assert.NoError(s.T(), err)
	
	// 创建活动日志
	logs := []*models.ActivityLog{
		{
			UserID:    user.ID,
			Action:    "login",
			Resource:  "user",
			Detail:    "用户登录",
			IPAddress: "192.168.1.100",
			UserAgent: "Test-Agent/1.0",
		},
		{
			UserID:    user.ID,
			Action:    "upload_file",
			Resource:  "file",
			Detail:    "上传文件: test.jpg",
			IPAddress: "192.168.1.100",
			UserAgent: "Test-Agent/1.0",
		},
	}
	
	for _, log := range logs {
		err = s.repo.CreateActivityLog(s.ctx, log)
		assert.NoError(s.T(), err)
	}
	
	// 获取用户活动日志
	retrievedLogs, total, err := s.repo.GetUserActivityLogs(s.ctx, user.ID, 0, 10)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(2), total)
	assert.Equal(s.T(), 2, len(retrievedLogs))
	
	// 验证日志内容
	assert.Equal(s.T(), "login", retrievedLogs[0].Action)
	assert.Equal(s.T(), "upload_file", retrievedLogs[1].Action)
	
	// 测试分页
	pagedLogs, pagedTotal, err := s.repo.GetUserActivityLogs(s.ctx, user.ID, 0, 1)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(2), pagedTotal)
	assert.Equal(s.T(), 1, len(pagedLogs))
}

// 运行测试套件
func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}