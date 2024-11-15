package middleware

import (
	"github.com/gin-gonic/gin"
	"publish-backend/util/config"
	"publish-backend/util/jwt"
	"strconv"
)

// 验证用户是否登录
func CheckLogin(strict bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("fake-cookie")
		uid, ok, _ := jwt.VerifyUserToken(token, config.Config.Key)
		if ok {
			_, err := strconv.ParseUint(uid, 10, 32)
			if err != nil {
				c.JSON(500, gin.H{"msg": "获取用户ID错误Orz"})
				c.Abort()
				return
			}
			c.Set("user_id", uid)
		} else if strict {
			c.JSON(401, gin.H{"msg": "请先登录awa"})
			c.Abort()
		}
	}
}
