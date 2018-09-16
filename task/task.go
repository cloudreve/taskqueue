package task

import (
	"fmt"
	"sync"
)

type taskInfo struct {
	sqlInfo SingleTaskInfo
}

//Run 根据任务类型处理任务
func (task *taskInfo) Run(wg *sync.WaitGroup) {
	switch {
	case task.sqlInfo.TaskType == "uploadSingleToOnedrive":
		fmt.Println("上传但恩建")
	}
	fmt.Println(task.sqlInfo.TaskType)
	wg.Done()
}
