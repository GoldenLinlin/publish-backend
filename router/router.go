package router

import (
	"BIT-Helper/controller"
	"BIT-Helper/middleware"

	"github.com/gin-gonic/gin"
)

// 配置路由
func SetRouter(router *gin.Engine) {
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "Hello BIT-Helper!"})
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

}
