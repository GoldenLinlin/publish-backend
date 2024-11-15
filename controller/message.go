package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"net/http"
	"os"
	"path/filepath"

	"publish-backend/util/wpapi"
)

// 上传文件到wordpress并传回前端url列表
func UploadFile(c *gin.Context) {
	// multiple files
	form, _ := c.MultipartForm()
	files := form.File["upload[]"]
	var fileURLs []map[string]string

	for _, file := range files {
		// Save the uploaded file to a temporary location
		tempFilePath := filepath.Join(os.TempDir(), file.Filename)
		if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Upload the file to WordPress
		fileURL, err := wpapi.UploadMedia("your_token_here", tempFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Append the file URL to the list
		fileURLs = append(fileURLs, map[string]string{"url": fileURL})

		// Remove the temporary file
		err = os.Remove(tempFilePath)
		if err != nil {
			return
		}
	}

	// Return the file URLs to the frontend
	c.JSON(http.StatusOK, fileURLs)
}

// 获取已发布内容
func GetPublishedPostLList(c *gin.Context) {
	posts, err := wpapi.GetUserPostList("your_token_here")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var formattedPosts []map[string]interface{}
	for _, post := range posts {
		formattedPosts = append(formattedPosts, map[string]interface{}{
			"title":     gjson.Get(post, "title").String(),
			"timestamp": gjson.Get(post, "date").Int(),
		})
	}

	c.JSON(http.StatusOK, formattedPosts)
}

// 发布文章
func PublishMessage(c *gin.Context) {
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

	// Publish the post
	err := wpapi.PublishPost("your_token_here", "", requestBody.Title, requestBody.Intro, requestBody.ImageURLs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post published successfully"})
}
