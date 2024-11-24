package controller

import (
	"fmt"
	"publish-backend/database"
	"publish-backend/util/wpapi"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 绑定用户账号
func BindAccount(c *gin.Context) {
	var query struct {
		PlatformName string `json:"platform_name" binding:"required"`
		AccountName  string `json:"name" binding:"required"`
		Account      string `json:"account" binding:"required"`
		Password     string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}
	// 验证用户令牌
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
	platformMap := make(map[string]uint)
	var platforms []database.Platform
	if err := database.DB.Find(&platforms).Error; err != nil {
		c.JSON(500, gin.H{"msg": "查询平台信息失败"})
		return
	}
	for _, platform := range platforms {
		platformMap[platform.PlatformName] = platform.PlatformID
	}
	PlatformID := int(platformMap[query.PlatformName])
	// 保存账号信息
	newAccount := database.UserSocialAccount{
		UserID:      uid,
		PlatformID:  PlatformID,
		AccountName: query.AccountName,
		State:       1,
	}
	var account database.UserSocialAccount
	if err := database.DB.Where("user_id = ? AND platform_id = ? AND account_name = ?", uid, PlatformID, query.AccountName).First(&account).Error; err == nil {
		c.JSON(404, gin.H{"msg": "账户已存在，请更换账号"})
		return
	}
	if err := database.DB.Omit("account_id").Create(&newAccount).Error; err != nil {
		c.JSON(500, gin.H{"msg": "账号绑定失败", "error": err.Error()})
		return
	}
	var AccountToken string
	var bl error
	AccountToken, bl = wpapi.GetWPJWTToken(query.Account, query.Password)
	fmt.Println(AccountToken)
	if bl != nil {
		c.JSON(500, gin.H{"msg": "账号未注册或密码用户名错误"})
		return
	}
	newSensitiveInfo := database.SensitiveAccountInfo{
		AccountID:    newAccount.AccountID,
		AccountToken: AccountToken,
	}
	if err := database.DB.Create(&newSensitiveInfo).Error; err != nil {
		c.JSON(500, gin.H{"msg": "敏感信息保存失败"})
		return
	}
	c.JSON(200, gin.H{"msg": "账号绑定成功"})
}

// 转换状态
func ToggleAccountState(c *gin.Context) {
	var query struct {
		PlatformName string `json:"platform_name" binding:"required"`
		AccountName  string `json:"account_name" binding:"required"`
	}

	// 解析前端请求的 JSON 数据
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	// 验证用户令牌
	idInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(500, gin.H{"msg": "用户未登录"})
		return
	}

	// 将 user_id 转换为字符串
	userIDStr, ok := idInterface.(string)
	if !ok || userIDStr == "" {
		c.JSON(500, gin.H{"msg": "用户ID格式错误"})
		return
	}

	// 将字符串形式的 user_id 转换为 uint
	uid64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(500, gin.H{"msg": "用户ID转换失败"})
		return
	}
	uid := uint(uid64)

	// 查询对应的平台信息以获取 PlatformID
	var platform database.Platform
	if err := database.DB.Where("platform_name = ?", query.PlatformName).First(&platform).Error; err != nil {
		c.JSON(404, gin.H{"msg": "平台不存在"})
		return
	}

	// 查找目标账户
	var account database.UserSocialAccount
	if err := database.DB.Where("user_id = ? AND platform_id = ? AND account_name = ?", uid, platform.PlatformID, query.AccountName).First(&account).Error; err != nil {
		c.JSON(404, gin.H{"msg": "账户不存在"})
		return
	}

	// 确定新的状态
	newState := 1
	if account.State == 1 {
		newState = 0
	}

	// 更新账户状态
	if err := database.DB.Model(&account).Update("state", newState).Error; err != nil {
		c.JSON(500, gin.H{"msg": "账户状态切换失败"})
		return
	}

	// 返回切换后的状态
	c.JSON(200, gin.H{
		"new_state": newState,
	})
}

// 删除用户账号
func DeleteAccount(c *gin.Context) {
	var query struct {
		PlatformName string `json:"platform_name" binding:"required"`
		AccountName  string `json:"account_name" binding:"required"`
	}
	// Query platform information
	var platforms []database.Platform
	if err := database.DB.Find(&platforms).Error; err != nil {
		c.JSON(500, gin.H{"msg": "查询平台信息失败"})
		return
	}

	// Create a map of platform ID to platform name for quick lookup
	platformMap := make(map[string]uint)
	for _, platform := range platforms {
		platformMap[platform.PlatformName] = platform.PlatformID
	}
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}
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
	PlatformID := uint(platformMap[query.PlatformName])
	fmt.Println("PlatformID", PlatformID)
	// 删除账号信息
	if err := database.DB.Where("user_id = ? AND platform_id = ? AND account_name = ?", uid, PlatformID, query.AccountName).Delete(&database.UserSocialAccount{}).Error; err != nil {
		c.JSON(500, gin.H{"msg": "删除账号失败"})
		return
	}

	c.JSON(200, gin.H{"msg": "账号删除成功"})
}

// 列出用户账号
func ListAccounts(c *gin.Context) {
	// Get the user ID from the context
	idInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(500, gin.H{"msg": "用户未登录"})
		return
	}

	// Convert user ID to string
	userIDStr, ok := idInterface.(string)
	if !ok || userIDStr == "" {
		c.JSON(500, gin.H{"msg": "用户ID格式错误"})
		return
	}

	// Convert user ID to uint
	uid64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(500, gin.H{"msg": "用户ID转换失败"})
		return
	}
	uid := uint(uid64)

	// Query user's social accounts
	var accounts []database.UserSocialAccount
	if err := database.DB.Where("user_id = ?", uid).Find(&accounts).Error; err != nil {
		c.JSON(500, gin.H{"msg": "查询账号失败"})
		return
	}

	// Query platform information
	var platforms []database.Platform
	if err := database.DB.Find(&platforms).Error; err != nil {
		c.JSON(500, gin.H{"msg": "查询平台信息失败"})
		return
	}

	// Create a map of platform ID to platform name for quick lookup
	platformMap := make(map[uint]string)
	for _, platform := range platforms {
		platformMap[platform.PlatformID] = platform.PlatformName
	}

	// Organize accounts by platform
	accountMenus := make(map[string][]gin.H)
	loggedInCount := 0
	// Pre-fill accountMenus with empty lists for each platform
	for _, platform := range platforms {
		platformMap[platform.PlatformID] = platform.PlatformName
		accountMenus[platform.PlatformName] = []gin.H{} // Initialize each platform with an empty list
	}

	for _, account := range accounts {
		platformName, exists := platformMap[uint(account.PlatformID)]
		if !exists {
			continue
		}

		item := gin.H{
			"name":     account.AccountName,
			"loggedIn": account.State, // Assume all accounts in the database are logged in
		}
		if account.State == 1 {
			loggedInCount++
		}
		accountMenus[platformName] = append(accountMenus[platformName], item)
	}

	// Format the response
	response := gin.H{
		"totalAccounts":    len(accounts),
		"loggedInAccounts": loggedInCount, //len(accounts), // All accounts are logged in based on DB
		"accountMenus": func() []gin.H {
			menus := []gin.H{}
			for title, items := range accountMenus {
				menus = append(menus, gin.H{
					"title": title,
					"items": items,
				})
			}
			return menus
		}(),
	}

	c.JSON(200, response)
}
