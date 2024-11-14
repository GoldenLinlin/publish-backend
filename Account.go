package controller

import (
	"BIT-Helper/database"
	"BIT-Helper/util/config"
	"BIT-Helper/util/jwt"

	"github.com/gin-gonic/gin"
)

// 绑定用户账号
func BindAccount(c *gin.Context) {
	var query struct {
		Token        string `json:"token" binding:"required"`
		PlatformID   int    `json:"platform_id" binding:"required"`
		AccountName  string `json:"account_name" binding:"required"`
		AccountToken string `json:"account_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}
	// 验证用户令牌
	userID, isValid := ValidateUserToken(query.Token)
	if !isValid {
		c.JSON(401, gin.H{"msg": "无效的令牌"})
		return
	}
	// 保存账号信息
	newAccount := database.UserSocialAccount{
		UserID:      userID,
		PlatformID:  query.PlatformID,
		AccountName: query.AccountName,
	}
	if err := database.DB.Create(&newAccount).Error; err != nil {
		c.JSON(500, gin.H{"msg": "账号绑定失败"})
		return
	}
	newSensitiveInfo := database.SensitiveAccountInfo{
		AccountID:    newAccount.AccountID,
		AccountToken: query.AccountToken,
	}
	if err := database.DB.Create(&newSensitiveInfo).Error; err != nil {
		c.JSON(500, gin.H{"msg": "敏感信息保存失败"})
		return
	}
	c.JSON(200, gin.H{"msg": "账号绑定成功"})
}

// 删除用户账号
func DeleteAccount(c *gin.Context) {
	var query struct {
		Token       string `json:"token" binding:"required"`
		PlatformID  int    `json:"platform_id" binding:"required"`
		AccountName string `json:"account_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}
	// 验证用户令牌
	userID, isValid := ValidateUserToken(query.Token)
	if !isValid {
		c.JSON(401, gin.H{"msg": "无效的令牌"})
		return
	}
	// 删除账号信息
	if err := database.DB.Where("user_id = ? AND platform_id = ? AND account_name = ?", userID, query.PlatformID, query.AccountName).Delete(&database.UserSocialAccount{}).Error; err != nil {
		c.JSON(500, gin.H{"msg": "删除账号失败"})
		return
	}
	c.JSON(200, gin.H{"msg": "账号删除成功"})
}

// 列出用户账号
func ListAccounts(c *gin.Context) {
	var query struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}
	// 验证用户令牌
	userID, isValid := ValidateUserToken(query.Token)
	if !isValid {
		c.JSON(401, gin.H{"msg": "无效的令牌"})
		return
	}
	// 获取用户的所有账号
	var accounts []database.UserSocialAccount
	if err := database.DB.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		c.JSON(500, gin.H{"msg": "查询账号失败"})
		return
	}
	c.JSON(200, gin.H{"accounts": accounts})
}

// 验证用户令牌的有效性
func ValidateUserToken(token string) (uint, bool) {
	userID, isValid := jwt.VerifyUserToken(token, config.Config.Key)
	if !isValid {
		return 0, false
	}
	return userID, true
}
