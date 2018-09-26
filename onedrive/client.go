package onedrive

import (
	"net/http"
	"os"
)

//APIURL Graph API URL
const APIURL string = "https://graph.microsoft.com/v1.0"

const authURL string = "https://login.microsoftonline.com/common/oauth2/v2.0"
const maxTry int = 5

//Client 客户端基本信息
type Client struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	HTTPClient   *http.Client
}

//Response 接口返回内容
type Response struct {
	Code     int
	Content  string
	Error    bool
	ErrorMsg string
}

//Init 初始化客户端
func (client *Client) Init() bool {
	client.HTTPClient = &http.Client{}
	return true
}

//PutFile 上传新文件
func (client *Client) PutFile(path string, file *os.File) (string, string) {
	return client.apiPut(path, file)
}

//apiPut 发送PUT请求
func (client *Client) apiPut(path string, stream *os.File) (string, string) {
	req, err := http.NewRequest("PUT", APIURL+path, stream)
	if err != nil {
		return "", err.Error()
	}

	req.Header.Set("Authorization", "Bearer "+client.AccessToken)

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", err.Error()
	}
	return
	defer res.Body.Close()

}
