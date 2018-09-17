package task

import (
	"sync"

	"../api"
)

type taskInfo struct {
	sqlInfo SingleTaskInfo
	apiInfo api.ApiInfo
}

//Run 根据任务类型处理任务
func (task *taskInfo) Run(wg *sync.WaitGroup) {
	var newTask Task
	switch {
	case task.sqlInfo.TaskType == "uploadSingleToOnedrive":
		newTask = &OneDriveUpload{Info: task, Tried: 0}
		newTask.Init()
	}
	newTask.Excute()
	wg.Done()
}
