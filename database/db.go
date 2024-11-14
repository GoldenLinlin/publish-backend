package database

import (
	"BIT-Helper/util/config"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// 用户身份常量（与数据库中定义一致）
const (
	Identity_Normal = iota
	Identity_Admin
)

// 基本模型
type Base struct {
	ID uint `gorm:"primarykey" json:"id"`
}

type Base1 struct {
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"create_time"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"update_time"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"delete_time"`
}

type Image struct {
	Base
	Mid  string `gorm:"not null;uniqueIndex;size:233" json:"mid"`
	Size uint   `gorm:"not null" json:"size"`
	Uid  uint   `gorm:"not null" json:"uid"`
}

// 1. 用户管理

// 用户基础信息表

type User struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement" json:"user_id"` // 自增主键
	Username  string    `gorm:"not null;Index;size:50" json:"username"`
	Avatar    string    `json:"avatar"`
	Identity  int       `gorm:"not null" json:"identity"` // 0=普通用户, 1=管理员
	CreatedAt time.Time `gorm:"autoCreateTime" json:"create_time"`
}

// 用户敏感信息表，存储加密后的密码、邮箱和手机号
type UserSensitiveInfo struct {
	UserID   uint   `gorm:"primaryKey;not null;Index" json:"user_id"`
	Password string `gorm:"not null" json:"password"` // bcrypt 加密
	Email    string `gorm:"not null" json:"email"`    // AES 加密
	Phone    string `gorm:"not null" json:"phone"`    // 未加密
}

// 用户密钥表，存储 AES 对称密钥
type UserSecretKey struct {
	UserID    uint   `gorm:"primaryKey;not null;Index" json:"user_id"`
	SecretKey string `gorm:"not null" json:"secret_key"`
}

// 2. 账号管理

// 社交账号基础信息表
type UserSocialAccount struct {
	UserID      uint   `gorm:"not null" json:"user_id"`
	PlatformID  int    `gorm:"not null" json:"platform_id"`
	AccountName string `gorm:"not null;size:50" json:"account_name"`
	AccountID   uint   `gorm:"primaryKey;autoIncrement" json:"account_id"`
}

// 账号敏感信息表，存储加密的 API 访问令牌
type SensitiveAccountInfo struct {
	AccountID    uint   `gorm:"primaryKey;not null;Index" json:"account_id"`
	AccountToken string `gorm:"not null" json:"account_token"` // AES 加密
}

// 3. 消息管理

// 内容基础信息表
type Content struct {
	UserID    uint      `gorm:"not null" json:"user_id"`
	Title     string    `gorm:"not null;size:100" json:"title"` // AES 加密
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	ContentID uint      `gorm:"primaryKey;autoIncrement" json:"content_id"`
}

// 内容敏感信息表，存储加密的正文和媒体 URL
type SensitiveContentInfo struct {
	ContentID uint   `gorm:"primaryKey;not null;Index" json:"content_id"`
	Body      string `gorm:"not null" json:"body"`               // AES 加密
	MediaURL  string `gorm:"not null;size:255" json:"media_url"` // AES 加密
}

// 4. 令牌管理

// 用户令牌表，用于会话管理
type TokenManagement struct {
	UserID    uint      `gorm:"primaryKey;not null;Index" json:"user_id"`
	Token     string    `gorm:"not null" json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// 5. 日志管理

// 日志表，用于记录系统的操作日志
type AuditLog struct {
	Base
	UserID     uint      `gorm:"not null" json:"user_id"`
	Message    string    `gorm:"not null" json:"message"`
	PlatformID int       `gorm:"not null" json:"platform_id"`
	Time       time.Time `gorm:"autoCreateTime" json:"time"`
}

// 6. 平台管理

// 平台表，用于存储不同社交平台的基本信息
type Platform struct {
	PlatformID   uint   `gorm:"primaryKey;autoIncrement" json:"platform_id"`
	PlatformName string `gorm:"not null;uniqueIndex;size:255" json:"platform_name"`
}

// 初始化数据库连接
func Init() {
	dsn := config.Config.Dsn
	fmt.Println("Connecting to database with DSN:", dsn)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	DB = db

	// 强制重新运行数据库迁移

	// 启动服务器
	// 自动迁移数据表
	err = db.AutoMigrate(
		&User{}, &UserSensitiveInfo{}, &UserSecretKey{},
		&UserSocialAccount{}, &SensitiveAccountInfo{},
		&Content{}, &SensitiveContentInfo{},
		&TokenManagement{}, &AuditLog{}, &Platform{}, &Image{},
	)
	if err != nil {
		panic("Failed to auto-migrate tables: " + err.Error())
	}

	fmt.Println("Database connected and tables migrated successfully.")
	DB.AutoMigrate(&User{}, &UserSensitiveInfo{}, &Image{})

}
