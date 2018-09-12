package task

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

//ApiInfo Api请求配置
type ApiInfo struct {
	TOKEN    string
	APIURL   string
	BASEPATH string
}

//GetBasicInfo 获取目标站点基本信息
func (apiInfo *ApiInfo) GetBasicInfo() string {
	return apiInfo.apiGet("basicInfo")
}

//GetTaskList 获取待处理任务列表
func (apiInfo *ApiInfo) GetTaskList(num int) string {
	return apiInfo.apiGet("getList?num=" + strconv.Itoa(num))
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
