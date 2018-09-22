package onedrive

import (
	"net/http"
)

//Client 客户端基本信息
type Client struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	HTTPClient   *http.Client
}

//Init 初始化客户端
func (client *Client) Init() bool {
	client.HTTPClient = &http.Client{}
	return true
}
