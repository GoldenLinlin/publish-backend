package router

import (
	"publish-backend/controller"
	"publish-backend/middleware"

	"github.com/gin-gonic/gin"
)

// 配置路由
func SetRouter(router *gin.Engine) {
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "Hello Easy Publish!"})
	})
	// 用户模块
	user := router.Group("/user")
	{
		user.POST("/login", controller.UserLogin)
		user.POST("/UserRegister", controller.UserRegister)
		user.POST("/UserLogin", controller.UserLogin)
		user.GET("/info", middleware.CheckLogin(false), controller.UserGetInfo)
		user.PUT("/info", middleware.CheckLogin(true), controller.UpdateUserInfo)
	}

	//消息（文章）模块
	message := router.Group("/message")
	{
		message.POST("/UploadAttachment", middleware.CheckLogin(true), controller.UploadFile)
		message.GET("/ListMessages", middleware.CheckLogin(true), controller.GetPublishedPostLList)
		message.POST("/PublishMessage", middleware.CheckLogin(true), controller.PublishMessage)
	}
	account := router.Group(("/account"))
	{
		account.POST("/BindAccount", middleware.CheckLogin(true), controller.BindAccount)
		account.POST("/DeleteAccount", middleware.CheckLogin(true), controller.DeleteAccount)
		account.POST("/LoginAccount", middleware.CheckLogin(true), controller.ToggleAccountState)
		account.GET("/ListAccounts", middleware.CheckLogin(true), controller.ListAccounts)
	}

}
