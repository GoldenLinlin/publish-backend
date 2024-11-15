package saver

import (
	"path/filepath"
	"publish-backend/util/config"
)

// 保存文件 返回url
func Save(path string, content []byte) (string, error) {
	SaveLocal(path, content)
	return GetUrl(path), nil
}

// 通过文件路径获取url
func GetUrl(path string) string {
	return config.Config.Saver.Url + filepath.Join("/", path)
}
