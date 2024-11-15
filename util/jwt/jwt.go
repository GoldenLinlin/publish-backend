package jwt

import (
	"publish-backend/database"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("your-secret-key")

// UserClaims 定义 JWT 的自定义声明结构
type UserClaims struct {
	UserID  string `json:"user_id"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// GetUserToken 生成用户 Token
func GetUserToken(userID string, expireTime int64, key string, identity int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(expireTime) * time.Second)

	// 创建 UserClaims，确保类型匹配
	claims := UserClaims{
		UserID:  userID,
		IsAdmin: identity == database.Identity_Admin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// 使用自定义的 UserClaims 创建 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyUserToken 验证并解析用户 Token
func VerifyUserToken(tokenStr, key string) (string, bool, bool) {
	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})

	// 验证 token 和 claims
	if err != nil || !token.Valid {
		return "", false, false
	}

	// 返回 userID、是否有效、是否为管理员
	return claims.UserID, true, claims.IsAdmin
}
