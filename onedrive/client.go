package onedrive

import (
	"bytes"
	"encoding/json"
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
	Success   bool
	Error     ErrorResponse
	ResString string
}

type uploadSessionResponse struct {
	DataContxt         string   `json:"@odata.context"`
	ExpirationDateTime string   `json:"expirationDateTime"`
	NextExpectedRanges []string `json:"nextExpectedRanges"`
	UploadURL          string   `json:"uploadUrl"`
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
		return res.ResString, ""
	}
	return "", res.Error.Error.Message
}

//CreateUploadSession 创建分片上传会话
func (client *Client) CreateUploadSession(path string) (string, string) {
	res := client.apiPost(path, []byte(""))
	if res.Success {
		response := uploadSessionResponse{}
		err := json.Unmarshal([]byte(res.ResString), &response)
		if err != nil {
			return "", err.Error()
		}
		return response.UploadURL, ""
	}

	return "", res.Error.Error.Message
}

//apiPost 发送POST请求
func (client *Client) apiPost(path string, jsonStr []byte) Response {
	if client.Tried > maxTry {
		return buildResponseResult("PUT failed, reached the maximum number of attempts.", 0)
	}

	req, err := http.NewRequest("POST", APIURL+path, bytes.NewBuffer(jsonStr))
	if err != nil {
		return buildResponseResult(err.Error(), 0)
	}

	req.Header.Set("Authorization", "Bearer "+client.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		client.Tried++
		return client.apiPost(path, jsonStr)
	}
	defer res.Body.Close()

	r, _ := ioutil.ReadAll(res.Body)
	return client.praseResponse(string(r), res.StatusCode)

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
	if code != 200 && code != 202 {
		errorRes := ErrorResponse{}
		json.Unmarshal([]byte(res), &errorRes)
		return Response{
			Success: false,
			Error:   errorRes,
		}
	}
	return Response{
		Success:   true,
		ResString: res,
	}
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
