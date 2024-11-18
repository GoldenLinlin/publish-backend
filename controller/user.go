package controller

import (
	"fmt"
	"publish-backend/database"
	"publish-backend/util/config"
	"publish-backend/util/jwt"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// 用户信息请求结构
type UserInfoQuery struct {
	UserID   uint   `json:"user_id"`  // 用户ID
	Phone    string `json:"phone"`    // 手机号
	Username string `json:"username"` // 用户名
	Email    string `form:"email"`    // 邮箱
	Avatar   string `form:"avatar"`   // 头像URL
}

// 转换用户信息
func CleanUser(old_user database.User, old_user_secret database.UserSensitiveInfo) UserInfoQuery {
	var user = UserInfoQuery{
		UserID:   old_user.UserID,
		Phone:    old_user_secret.Phone,
		Username: old_user.Username,

		Email:  old_user_secret.Email,
		Avatar: GetImageUrl(old_user.Avatar),
	}
	return user
}

// 获取用户信息
func GetUserAPI(uid int) UserInfoQuery {
	return GetUserAPIMap(map[int]bool{uid: true})[uid]
}

// 批量获取用户信息
func GetUserAPIMap(uid_map map[int]bool) map[int]UserInfoQuery {
	out := make(map[int]UserInfoQuery)
	uid_list := make([]int, 0)

	// 从 uid_map 提取所有的 user_id
	for uid := range uid_map {
		uid_list = append(uid_list, uid)
	}

	// 1. 查询 users 表，获取基础信息
	var users []database.User
	if err := database.DB.Where("user_id IN ?", uid_list).Find(&users).Error; err != nil {
		return nil
	}

	// 2. 查询 user_sensitive_infos 表，获取敏感信息
	var userSecrets []database.UserSensitiveInfo
	if err := database.DB.Where("user_id IN ?", uid_list).Find(&userSecrets).Error; err != nil {
		return nil
	}

	// 将两个查询结果关联起来
	for _, user := range users {
		// 在 userSecrets 切片中找到与 user.UserID 对应的敏感信息
		var userSecret *database.UserSensitiveInfo
		for _, secret := range userSecrets {
			if secret.UserID == user.UserID {
				userSecret = &secret
				break
			}
		}

		// 如果找到了对应的敏感信息，合并到 UserInfoQuery 中
		if userSecret != nil {
			out[int(user.UserID)] = CleanUser(user, *userSecret)
		}
	}

	return out
}

// 注册请求结构
type UserRegisterQuery struct {
	Username string `json:"username" binding:"required"` // 用户名
	Phone    string `json:"phone" binding:"required"`    // 手机号
	Password string `json:"password" binding:"required"` // 密码
}

// 用户登录请求结构
type UserLoginQuery struct { // 用户ID（可选）
	User     string `json:"user"`                        // 手机号（可选）
	Password string `json:"password" binding:"required"` // 密码
}

// 用户登录
func UserLogin(c *gin.Context) {
	var query UserLoginQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		fmt.Println(query)
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	// 查找用户敏感信息表中的用户信息
	var sensitiveInfo database.UserSensitiveInfo
	var err error

	// 根据传入的字段来选择查询方式：用户ID或手机号
	if query.User != 0 {
		// 使用用户ID查找
		err = database.DB.Where("user_id = ?", query.User).First(&sensitiveInfo).Error
	} else if query.User != "" {
		// 使用手机号查找
		err = database.DB.Where("phone = ?", query.User).First(&sensitiveInfo).Error
	} else {
		// 如果未提供用户ID和手机号
		c.JSON(400, gin.H{"msg": "请提供用户ID或手机号"})
		return
	}

	if err != nil {
		c.JSON(404, gin.H{"msg": "用户不存在Orz"})
		return
	}

	// 校验密码
	err = bcrypt.CompareHashAndPassword([]byte(sensitiveInfo.Password), []byte(query.Password))
	if err != nil {
		c.JSON(401, gin.H{"msg": "密码错误Orz"})
		return
	}

	// 获取 user_id，用于生成 token
	userID := sensitiveInfo.UserID

	// 密码验证成功，生成并返回用户令牌
	token, err := jwt.GetUserToken(fmt.Sprint(userID), config.Config.LoginExpire, config.Config.Key, int(userID))
	if err != nil {
		c.JSON(500, gin.H{"msg": "生成用户令牌失败"})
		return
	}
	// 设置 Token 到 Cookie 中
	c.SetCookie("fake-cookie", token, int(config.Config.LoginExpire), "/", "localhost", false, true)

	c.JSON(200, gin.H{"msg": "登录成功OvO", "user_id": query.UserID})
}

// 用户注册
func UserRegister(c *gin.Context) {
	var query UserRegisterQuery
	fmt.Println(c)
	if err := c.ShouldBindJSON(&query); err != nil {
		fmt.Println(err)
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	// 检查手机号是否已注册
	var existingUser database.UserSensitiveInfo
	if err := database.DB.Where("phone = ?", query.Phone).First(&existingUser).Error; err == nil {
		c.JSON(409, gin.H{"msg": "手机号已被注册Orz"})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(query.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"msg": "密码加密失败Orz"})
		return
	}

	// 创建新用户条目
	newUser := database.User{
		Username: query.Username,       // 用户名
		Identity: 0,                    // 默认身份为普通用户，可以自定义调整
		Avatar:   "default_avatar.jpg", // 默认头像，可根据需求自定义
	}
	if err := database.DB.Create(&newUser).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}

	// 创建新用户的敏感信息条目
	newSensitiveInfo := database.UserSensitiveInfo{
		UserID:   newUser.UserID,
		Password: string(hashedPassword),
		Phone:    query.Phone,
	}
	if err := database.DB.Create(&newSensitiveInfo).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}

	// 注册成功，返回用户ID
	c.JSON(200, gin.H{"msg": "注册成功OvO", "user_id": newUser.UserID})
}

// UserGetInfo 获取用户信息
func UserGetInfo(c *gin.Context) {
	// 从上下文中获取 user_id
	idInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(500, gin.H{"msg": "错误"})
		return
	}

	// 尝试将 user_id 转换为字符串
	userIDStr, ok := idInterface.(string)
	if !ok || userIDStr == "" {
		c.JSON(500, gin.H{"msg": "用户ID格式错误"})
		return
	}
	fmt.Println("User ID:", userIDStr)

	// 将字符串形式的 user_id 转换为 uint
	uid64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(500, gin.H{"msg": "用户ID转换失败"})
		return
	}
	uid := uint(uid64)

	// 查询数据库中的用户信息
	var user database.User
	if err := database.DB.Limit(1).Find(&user, "user_id = ?", uid).Error; err != nil {
		c.JSON(500, gin.H{"msg": "查询用户信息失败"})
		return
	}
	if user.UserID == 0 {
		c.JSON(404, gin.H{"msg": "用户不存在Orz"})
		return
	}

	// 返回用户信息
	c.JSON(200, GetUserAPI(int(uid)))
}

// 修改用户信息请求结构
type UpdateUserInfoForm struct {
	Email  string `form:"email"`  // 邮箱
	Avatar string `form:"avatar"` // 头像URL
}

func UpdateUserInfo(c *gin.Context) {
	// 从上下文中获取 user_id
	idInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(500, gin.H{"msg": "错误"})
		return
	}

	// 尝试将 user_id 转换为字符串
	userIDStr, ok := idInterface.(string)
	if !ok || userIDStr == "" {
		c.JSON(500, gin.H{"msg": "用户ID格式错误"})
		return
	}
	fmt.Println("User ID:", userIDStr)

	// 将字符串形式的 user_id 转换为 uint
	uid64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(500, gin.H{"msg": "用户ID转换失败"})
		return
	}
	uid := uint(uid64)
	// 绑定表单数据
	var form UpdateUserInfoForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误", "error": err.Error()})
		return
	}

	// 查找用户和用户敏感信息记录
	var user database.User
	var sensitiveInfo database.UserSensitiveInfo

	// 通过 user_id 查找用户
	if err := database.DB.Where("user_id = ?", uid).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"msg": "用户不存在"})
		return
	}
	if err := database.DB.Where("user_id = ?", uid).First(&sensitiveInfo).Error; err != nil {
		c.JSON(404, gin.H{"msg": "用户敏感信息不存在"})
		return
	}

	// 更新字段，如果表单中字段为空，则保留原值
	if form.Email != "" {
		sensitiveInfo.Email = form.Email
	}
	if form.Avatar != "" {
		user.Avatar = form.Avatar
	}

	// 保存更改
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"msg": "更新用户信息失败"})
		return
	}
	if err := database.DB.Save(&sensitiveInfo).Error; err != nil {
		c.JSON(500, gin.H{"msg": "更新用户敏感信息失败"})
		return
	}

	c.JSON(200, gin.H{"msg": "用户信息更新成功"})
}
