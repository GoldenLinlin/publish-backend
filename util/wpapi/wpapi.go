package wpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/tidwall/gjson"
)

// 获取wp的jwt令牌
func GetWPJWTToken(username, password string) (string, error) {
	url := "http://182.92.192.196:8080/wp-json/jwt-auth/v1/token"

	// 构造请求数据
	data := map[string]string{
		"username": username,
		"password": password,
	}
	reqBody, _ := json.Marshal(data)

	// 构造请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("获取JWT令牌失败，状态码：%d，响应：%s", resp.StatusCode, string(body))
	}

	// 解析token
	token := gjson.Get(string(body), "token").String()
	if token == "" {
		return "", fmt.Errorf("响应中未找到token字段")
	}

	return token, nil
}

// 验证令牌有效性
func verifyToken(token string) error {
	url := "http://182.92.192.196:8080//wp-json/jwt-auth/v1/token/validate"

	// 构造请求
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("验证token失败，状态码：%d，响应：%s", resp.StatusCode, string(body))
	}

	return nil
}

// Post 文章结构体
type Post struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

// 发布文章
func PublishPost(token, postID, title, content string, urls []string) error {
	url := "http://182.92.192.196:8080/wp-json/wp/v2/posts/" + postID

	// 构造文章结构体
	post := Post{
		Title:   title,
		Content: generateHTMLContent(content, urls),
		Status:  "publish",
	}

	reqBody, _ := json.Marshal(post)

	// 构造请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// 解析响应
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("发布文章失败，状态码：%d，响应：%s", resp.StatusCode, string(body))
	}

	return nil
}

// 生成HTML内容
func generateHTMLContent(content string, urls []string) string {
	htmlTemplate := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
	</head>
	<body>
		%s
		%s
	</body>
	</html>
	`

	mediaContent := ""
	if content == "" {
		for _, url := range urls {
			mediaContent += fmt.Sprintf(`<br><video src="%s" controls="controls"></video>`, url)
		}
	} else {
		for _, url := range urls {
			mediaContent += fmt.Sprintf(`<br><img src="%s" />`, url)
		}
	}

	return fmt.Sprintf(htmlTemplate, content, mediaContent)
}

// 媒体上传（文件名不能包含中文）
func UploadMedia(token, filePath string) (string, error) {
	url := "http://182.92.192.196:8080/wp-json/wp/v2/media"

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)
	// Create a buffer to write our multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	// Create a form file field
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return "", err
	}
	// Copy the file content to the form file field
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}
	// Close the multipart writer to set the terminating boundary
	err = writer.Close()
	if err != nil {
		return "", err
	}

	// 构造请求
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// 解析响应
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("上传媒体失败，状态码：%d，响应：%s", resp.StatusCode, string(body))
	}

	// 解析媒体URL
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	mediaURL := gjson.Get(string(body), "guid.rendered").String()
	if mediaURL == "" {
		return "", fmt.Errorf("响应中未找到媒体url字段")
	}

	return mediaURL, nil
}

// PostInList 文章结构
type PostInList struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Time  string `json:"time"`
	Link  string `json:"link"`
}

// 获取用户发布文章记录列表
func GetUserPostList(token string) ([]string, error) {
	url := "http://182.92.192.196:8080/wp-json/wp/v2/posts"

	// 构造请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// 解析响应
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取文章列表失败，状态码：%d，响应：%s", resp.StatusCode, string(body))
	}

	// 解析文章列表
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	posts := gjson.ParseBytes(body).Array()
	var postList []string
	for _, post := range posts {
		postList = append(postList, post.String())
	}

	return postList, nil

}

//// 获取文章详情
//func getPostDetail(token, postID string) (string, error) {
//	url := "http://182.92.192.196:8080/wp-json/wp/v2/posts/" + postID
//
//	// 构造请求
//	req, err := http.NewRequest("GET", url, nil)
//	if err != nil {
//		return "", err
//	}
//	req.Header.Set("Authorization", "Bearer "+token)
//	// 发送请求
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return "", err
//	}
//	defer func(Body io.ReadCloser) {
//		err := Body.Close()
//		if err != nil {
//
//		}
//	}(resp.Body)
//
//	// 解析响应
//	if resp.StatusCode != http.StatusOK {
//		body, _ := io.ReadAll(resp.Body)
//		return "", fmt.Errorf("获取文章详情失败，状态码：%d，响应：%s", resp.StatusCode, string(body))
//	}
//
//	// 解析文章详情
//	body, err := io.ReadAll(resp.Body)
//	if err != nil {
//		return "", err
//	}
//
//	return gjson.Get(string(body), "content.rendered").String(), nil
//
//}

// 修改（更新）文章
func updatePost(token, postID, title, content string, urls []string) error {
	return PublishPost(token, postID, title, content, urls)
}

// 删除文章
func deletePost(token, postID string) error {
	url := "http://182.92.192.196:8080/wp-json/wp/v2/posts/" + postID

	// 构造请求
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// 解析响应
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除文章失败，状态码：%d，响应：%s", resp.StatusCode, string(body))
	}

	return nil
}
