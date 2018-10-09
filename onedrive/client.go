package onedrive

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	Tried        int
}

//ErrorResponse 接口返回错误内容
type ErrorResponse struct {
	Error errorEntity `json:"error"`
}

type errorEntity struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	InnerError interface{} `json:"InnerError"`
}

//Response 接口返回信息
type Response struct {
	Success bool
	Error   ErrorResponse
}

//Init 初始化客户端
func (client *Client) Init() bool {
	client.HTTPClient = &http.Client{}
	return true
}

//PutFile 上传新文件
func (client *Client) PutFile(path string, file *os.File) (string, string) {
	res := client.apiPut(path, file)
	if res.Success {
		fmt.Println("成功")
		return "", ""
	}
	return "", res.Error.Error.Message
}

//apiPut 发送PUT请求
func (client *Client) apiPut(path string, stream *os.File) Response {
	if client.Tried > maxTry {
		return buildResponseResult("PUT failed, reached the maximum number of attempts.", 0)
	}

	req, err := http.NewRequest("PUT", APIURL+path, stream)
	if err != nil {
		return buildResponseResult(err.Error(), 0)
	}

	req.Header.Set("Authorization", "Bearer "+client.AccessToken)

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		client.Tried++
		return client.apiPut(path, stream)
	}
	defer res.Body.Close()

	r, _ := ioutil.ReadAll(res.Body)
	return client.praseResponse(string(r), res.StatusCode)
}

func (client *Client) praseResponse(res string, code int) Response {
	if code != 200 {
		errorRes := ErrorResponse{}
		json.Unmarshal([]byte(res), &errorRes)
		return Response{
			Success: false,
			Error:   errorRes,
		}
	}
	return Response{Success: true}
}

func buildResponseResult(msg string, code int) Response {
	return Response{
		Success: false,
		Error: ErrorResponse{
			Error: errorEntity{
				Message: msg,
			},
		},
	}
}
