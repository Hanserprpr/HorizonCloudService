package user

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"horizon-cloud-admin/internal/global/database"
	"horizon-cloud-admin/internal/global/jwt"
	"horizon-cloud-admin/internal/global/response"
	"horizon-cloud-admin/internal/model"
	"horizon-cloud-admin/tools"
)

// User 定义登录和注册请求的结构体
type User struct {
	StudentID string `json:"student_id" binding:"required"` // 学号，唯一标识用户
	Password  string `json:"password" binding:"required"`   // 密码，登录时验证，注册时加密
	NickName  string `json:"nick_name" binding:"required"`  // 用户昵称
}

// ChangePasswordReq 定义修改密码请求的结构体
type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"` // 旧密码，用于验证
	NewPassword string `json:"new_password" binding:"required"` // 新密码，需加密后保存
}

// Login 处理用户登录请求
func Login(c *gin.Context) {
	// 定义请求结构体并绑定 JSON 数据
	var req User
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("绑定登录请求失败", "error", err, "student_id", req.StudentID)
		response.Fail(c, response.ErrInvalidRequest.WithOrigin(err))
		return
	}

	// 查询用户是否存在
	var user model.User
	err := database.DB.Where("student_id = ?", req.StudentID).First(&user).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		log.Warn("用户不存在", "student_id", req.StudentID)
		response.Fail(c, response.ErrNotFound)
		return
	case err != nil:
		log.Error("数据库查询失败", "error", err, "student_id", req.StudentID)
		response.Fail(c, response.ErrDatabase.WithOrigin(err))
		return
	}

	// 验证密码
	if !tools.PasswordCompare(req.Password, user.Password) {
		log.Warn("密码错误", "student_id", req.StudentID)
		response.Fail(c, response.ErrInvalidPassword)
		return
	}

	// 记录登录成功的日志
	log.Info("用户登录成功",
		"student_id", user.StudentID,
		"nick_name", user.NickName,
		"role_id", user.RoleID)

	// 生成 JWT 令牌并返回用户信息
	response.Success(c, map[string]interface{}{
		"token": jwt.CreateToken(jwt.Payload{
			StudentID: user.StudentID,
			RoleID:    user.RoleID,
		}),
		"student_id": user.StudentID,
		"nick_name":  user.NickName,
		"role_id":    user.RoleID,
	})
}

// Register 处理用户注册请求
func Register(c *gin.Context) {
	// 定义请求结构体并绑定 JSON 数据
	var req User
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("绑定注册请求失败", "error", err, "student_id", req.StudentID)
		response.Fail(c, response.ErrInvalidRequest.WithOrigin(err))
		return
	}

	// 检查学号是否已存在
	var existingUser model.User
	err := database.DB.Where("student_id = ?", req.StudentID).First(&existingUser).Error
	if err == nil {
		log.Warn("用户已存在", "student_id", req.StudentID)
		response.Fail(c, response.ErrAlreadyExists)
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("数据库查询失败", "error", err, "student_id", req.StudentID)
		response.Fail(c, response.ErrDatabase.WithOrigin(err))
		return
	}

	// 加密密码
	encryptedPassword := tools.PasswordEncrypt(req.Password)

	// 创建新的用户
	user := model.User{
		StudentID: req.StudentID,
		Password:  encryptedPassword,
		NickName:  req.NickName,
		RoleID:    1, // 默认角色 ID，可根据需求调整
	}

	// 保存用户到数据库
	if err := database.DB.Create(&user).Error; err != nil {
		log.Error("创建用户失败", "error", err, "student_id", req.StudentID)
		response.Fail(c, response.ErrDatabase.WithOrigin(err))
		return
	}

	// 记录注册成功的日志
	log.Info("用户注册成功",
		"student_id", user.StudentID,
		"nick_name", user.NickName,
		"role_id", user.RoleID)

	// 返回成功响应
	response.Success(c, map[string]interface{}{
		"student_id": user.StudentID,
		"nick_name":  user.NickName,
		"role_id":    user.RoleID,
	})
}

// ChangePassword 处理用户修改密码请求
// 验证旧密码正确性后更新新密码，要求用户已通过认证
// 参数:
//   - c: gin 上下文，用于接收请求和发送响应
func ChangePassword(c *gin.Context) {
	// 获取认证信息
	payload, exists := c.Get("payload")
	if !exists {
		response.Fail(c, response.ErrUnauthorized)
		return
	}
	userPayload, ok := payload.(jwt.Payload)
	if !ok {
		response.Fail(c, response.ErrUnauthorized)
		return
	}

	// 定义请求结构体并绑定 JSON 数据
	var req ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("绑定修改密码请求失败", "error", err, "student_id", userPayload.StudentID)
		response.Fail(c, response.ErrInvalidRequest.WithOrigin(err))
		return
	}

	// 查询用户
	var user model.User
	err := database.DB.Where("student_id = ?", userPayload.StudentID).First(&user).Error
	if err != nil {
		log.Error("查询用户失败", "error", err, "student_id", userPayload.StudentID)
		response.Fail(c, response.ErrDatabase.WithOrigin(err))
		return
	}

	// 验证旧密码
	if !tools.PasswordCompare(req.OldPassword, user.Password) {
		log.Warn("旧密码错误", "student_id", userPayload.StudentID)
		response.Fail(c, response.ErrInvalidPassword)
		return
	}

	// 加密新密码
	newEncryptedPassword := tools.PasswordEncrypt(req.NewPassword)

	// 更新用户密码
	if err := database.DB.Model(&user).Update("password", newEncryptedPassword).Error; err != nil {
		log.Error("更新密码失败", "error", err, "student_id", userPayload.StudentID)
		response.Fail(c, response.ErrDatabase.WithOrigin(err))
		return
	}

	// 记录修改密码成功的日志
	log.Info("用户修改密码成功",
		"student_id", user.StudentID,
		"nick_name", user.NickName,
		"role_id", user.RoleID)

	// 返回成功响应
	response.Success(c, nil)
}
