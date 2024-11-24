package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"net/http"
	"os"
	"path/filepath"
	"publish-backend/database"
	"strconv"

	"publish-backend/util/wpapi"
)

// 根据用户信息查询用户的wp token
func GetUserPlatformToken(c *gin.Context, platform_id int) []string {
	// Get the user ID from the context
	idInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(500, gin.H{"msg": "用户未登录"})
		return nil
	}

	// Convert user ID to string
	userIDStr, ok := idInterface.(string)
	if !ok || userIDStr == "" {
		c.JSON(500, gin.H{"msg": "用户ID格式错误"})
		return nil
	}

	// Convert user ID to uint
	uid64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(500, gin.H{"msg": "用户ID转换失败"})
		return nil
	}
	uid := uint(uid64)

	// Query UserSocialAccount for the given user_id and platform_id = 1
	var accountId []database.UserSocialAccount
	if err := database.DB.Where("user_id = ? AND platform_id = ? AND state = 1", uid, platform_id).Find(&accountId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "查询账号失败"})
		return nil
	}

	// 提取 account_id 字段
	var accountIDs []uint
	for _, account := range accountId {
		accountIDs = append(accountIDs, account.AccountID)
	}

	// 使用提取的 account_id 列表进行查询
	var wptokens []database.SensitiveAccountInfo
	if err := database.DB.Where("account_id IN (?)", accountIDs).Find(&wptokens).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "查询敏感信息失败"})
		return nil
	}

	var tokens []string
	for _, token := range wptokens {
		tokens = append(tokens, token.AccountToken)
	}

	return tokens
}

// 上传文件到wordpress并传回前端url列表
func UploadFile(c *gin.Context) {
	//其他平台的图片也可以使用相同的方式上传，或者直接上传在wordpress上即可
	//只需要上传一次即可，不用重复上传

	tokens := GetUserPlatformToken(c, 1)
	// multiple files
	file, err := c.FormFile("files")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}

	// Save the uploaded file to a temporary location
	tempFilePath := filepath.Join(os.TempDir(), file.Filename)
	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Upload the file to WordPress
	fileURL, err := wpapi.UploadMedia(tokens[0], tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Remove the temporary file
	err = os.Remove(tempFilePath)
	if err != nil {
		return
	}

	// Return the file URLs to the frontend
	c.JSON(http.StatusOK, gin.H{"url": fileURL})
}

// 获取已发布内容
func GetPublishedPostLList(c *gin.Context) {
	//仅展示wordpress平台的文章，其他平台使用相同的方式接入
	tokens := GetUserPlatformToken(c, 1)
	var formattedPosts []map[string]interface{}
	for _, token := range tokens {
		posts, err := wpapi.GetUserPostList(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, post := range posts {
			formattedPosts = append(formattedPosts, map[string]interface{}{
				"title":     gjson.Get(post, "title.rendered").String(),
				"timestamp": gjson.Get(post, "date").String(),
			})
		}
	}

	c.JSON(http.StatusOK, formattedPosts)
}

// 发布文章
func PublishMessage(c *gin.Context) {
	// Define the request body structure
	var requestBody struct {
		Type      []int    `json:"type"`
		Title     string   `json:"title"`
		Intro     string   `json:"intro"`
		ImageURLs []string `json:"image_urls"`
	}
	// Parse the request body
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//publish post to multi platforms
	for _, platform := range requestBody.Type {
		switch platform {
		case 1:
			// Publish the post to WordPress
			tokens := GetUserPlatformToken(c, 1)
			for _, token := range tokens {
				// use wpapi.PublishPost to publish the post
				err := wpapi.PublishPost(token, "", requestBody.Title, requestBody.Intro, requestBody.ImageURLs)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		case 2:
			// Publish the post to bilibili
		case 3:
			// Publish the post to 小红书
		case 4:
			// Publish the post to 微博
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform type"})

		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post published successfully"})
}
