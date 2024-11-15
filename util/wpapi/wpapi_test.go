package wpapi

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_getWPJWTToken(t *testing.T) {
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// Test cases
		{
			name: "Test case 1",
			args: args{
				username: "root",
				password: "123456",
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getWPJWTToken(tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("getWPJWTToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(got)
			if got != tt.want {
				t.Errorf("getWPJWTToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_publishPost(t *testing.T) {
	type args struct {
		token   string
		postID  string
		title   string
		content string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// Test cases
		{
			name: "Test case 1",
			args: args{
				token:   "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vMTgyLjkyLjE5Mi4xOTY6ODA4MCIsImlhdCI6MTczMTU1OTA3NiwibmJmIjoxNzMxNTU5MDc2LCJleHAiOjE3MzIxNjM4NzYsImRhdGEiOnsidXNlciI6eyJpZCI6IjEifX19.tvNHf0BumGvs9QEhkUpIuGcKwWZu6-BFfqaxuCCVdbE",
				postID:  "1",
				title:   "go api test",
				content: "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PublishPost(tt.args.token, tt.args.postID, tt.args.title, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("publishPost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_uploadMedia(t *testing.T) {
	type args struct {
		token    string
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// Test cases
		{
			name: "Test case 1",
			args: args{
				token:    "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vMTgyLjkyLjE5Mi4xOTY6ODA4MCIsImlhdCI6MTczMTU1OTA3NiwibmJmIjoxNzMxNTU5MDc2LCJleHAiOjE3MzIxNjM4NzYsImRhdGEiOnsidXNlciI6eyJpZCI6IjEifX19.tvNHf0BumGvs9QEhkUpIuGcKwWZu6-BFfqaxuCCVdbE",
				filePath: "E:\\easy-publish-backend\\lisbon-8275994_1280.jpg",
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UploadMedia(tt.args.token, tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("uploadMedia() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("uploadMedia() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUserPostList(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{

		{
			name: "Test case 1",
			args: args{
				token: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vMTgyLjkyLjE5Mi4xOTY6ODA4MCIsImlhdCI6MTczMTU1OTA3NiwibmJmIjoxNzMxNTU5MDc2LCJleHAiOjE3MzIxNjM4NzYsImRhdGEiOnsidXNlciI6eyJpZCI6IjEifX19.tvNHf0BumGvs9QEhkUpIuGcKwWZu6-BFfqaxuCCVdbE",
			},
			want:    []string{"go api test"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserPostList(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserPostList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUserPostList() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_verifyToken(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// Test cases
		{
			name: "Test case 1",
			args: args{
				token: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vMTgyLjkyLjE5Mi4xOTY6ODA4MCIsImlhdCI6MTczMTU1OTA3NiwibmJmIjoxNzMxNTU5MDc2LCJleHAiOjE3MzIxNjM4NzYsImRhdGEiOnsidXNlciI6eyJpZCI6IjEifX19.tvNHf0BumGvs9QEhkUpIuGcKwWZu6-BFfqaxuCCVdbE",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := verifyToken(tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("verifyToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
