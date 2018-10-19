package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)

//ApiInfo Api请求配置
type ApiInfo struct {
	TOKEN    string
	APIURL   string
	BASEPATH string
	Lock     *sync.Mutex
}

//APIResponse API响应结果
type APIResponse struct {
	Error bool   `json:"error"`
	Msg   string `json:"msg"`
}

//GetBasicInfo 获取目标站点基本信息
func (apiInfo *ApiInfo) GetBasicInfo() string {
	return apiInfo.apiGet("basicInfo")
}

//GetTaskList 获取待处理任务列表
func (apiInfo *ApiInfo) GetTaskList(num int) string {
	return apiInfo.apiGet("getList?num=" + strconv.Itoa(num))
}

//CancelUploadSession 取消分片上传
func (apiInfo *ApiInfo) CancelUploadSession(url string) {
	apiInfo.apiDelete(url)
}

//GetPolicy 获取上传策略详情
func (apiInfo *ApiInfo) GetPolicy(id int) string {
	return apiInfo.apiGet("getPolicy?id=" + strconv.Itoa(id))
}

//SetSuccess 设置任务完成
func (apiInfo *ApiInfo) SetSuccess(id int) string {
	res := APIResponse{}
	response := apiInfo.apiGet("setSuccess?id=" + strconv.Itoa(id))
	if response != "" {
		json.Unmarshal([]byte(response), &res)
		if res.Error {
			return res.Msg
		}
		return ""
	}
	return ""
}

//SetError 设置为失败状态
func (apiInfo *ApiInfo) SetError(id int) string {
	return apiInfo.apiGet("setError?id=" + strconv.Itoa(id))
}

//apiDelete 发送DELETE请求

func (apiInfo *ApiInfo) apiDelete(controller string) string {
	client := &http.Client{}
	request, err := http.NewRequest("DELETE", controller, nil)
	if err != nil {
		log.Printf("[ERROR] Failed to create DELETE requetst, #%v ", err)
	}
	request.Header.Set("Authorization", "Bearer "+apiInfo.TOKEN)
	response, err := client.Do(request)
	if err != nil {
		log.Printf("[ERROR] Failed to send GET requetst, #%v ", err)
		return ""
	}
	defer response.Body.Close()
	if response.StatusCode == 200 || response.StatusCode == 204 {
		r, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("[ERROR] Failed to get GET requetst body, #%v ", err)
		}
		return string(r)
	} else {
		log.Printf("[ERROR] Failed to get respond, HTTP ERROR %v ", response.StatusCode)
		return ""
	}
}

//apiGet 发送GET请求
func (apiInfo *ApiInfo) apiGet(controller string) string {
	client := &http.Client{}
	request, err := http.NewRequest("GET", apiInfo.APIURL+"/"+controller, nil)
	if err != nil {
		log.Printf("[ERROR] Failed to create GET requetst, #%v ", err)
	}
	request.Header.Set("Authorization", "Bearer "+apiInfo.TOKEN)
	response, err := client.Do(request)
	if err != nil {
		log.Printf("[ERROR] Failed to send GET requetst, #%v ", err)
		return ""
	}
	defer response.Body.Close()
	if response.StatusCode == 200 {
		r, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("[ERROR] Failed to get GET requetst body, #%v ", err)
		}
		return string(r)
	} else if response.StatusCode == 403 {
		log.Printf("[ERROR] Auth failed, please verify your token, HTTP ERROR %v ", response.StatusCode)
		return ""
	} else {
		log.Printf("[ERROR] Failed to get respond, HTTP ERROR %v ", response.StatusCode)
		return ""
	}

}
