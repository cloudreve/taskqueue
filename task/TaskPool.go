package task

import (
	"encoding/json"
	"log"

	"../api"
)

//SingleTaskInfo 单个任务
type SingleTaskInfo struct {
	ID       int         `json:"id"`
	TaskName string      `json:"task_name"`
	Attr     interface{} `json:"attr"`
	TaskType string      `json:"type"`
	Status   string      `json:"status"`
	Addtime  string      `json:"addtime"`
}

//Init 初始化任务线程池
func Init(taskListContent string, apiInfo api.ApiInfo, siteInfo map[string]string, threadID int) {
	var taskStringList []SingleTaskInfo
	err := json.Unmarshal([]byte(taskListContent), &taskStringList)
	if err != nil {
		log.Printf("[ERROR] Failed to decode basic infomation,  %v ", err.Error())
	}
	for _, v := range taskStringList {
		log.Printf("[Info][Thread %d] New task: %s", threadID, v.TaskName)
		signleTask := taskInfo{sqlInfo: v, apiInfo: apiInfo, siteInfo: siteInfo}
		signleTask.Run()

	}
}
