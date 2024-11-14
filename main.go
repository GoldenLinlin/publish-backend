package main

import (
	"BIT-Helper/database"
	"BIT-Helper/router"
	"BIT-Helper/util/config"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	// limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
)

// 服务，启动！
func main() {
	config.Init()
	database.Init()

	if config.Config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	app := gin.Default()
	app.MaxMultipartMemory = 10 << 24  // 10MB
	// app.Use(limits.RequestSizeLimiter(config.Config.Saver.MaxSize << 20))
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Content-Type", "fake-cookie"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		MaxAge: 12 * time.Hour,
	}))
	
	router.SetRouter(app)
	
	fmt.Println("run on port " + config.Config.Port)
	app.Run(":" + config.Config.Port)
}
