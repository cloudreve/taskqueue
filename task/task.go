package task

import (
	"../api"
)

type taskInfo struct {
	sqlInfo  SingleTaskInfo
	apiInfo  api.ApiInfo
	siteInfo map[string]string
}

type policy struct {
	ID            int         `json:"id"`
	PolicyName    string      `json:"policy_name"`
	PolicyType    string      `json:"policy_type"`
	Server        string      `json:"server"`
	BucketName    string      `json:"bucketname"`
	BucketPrivate int         `json:"bucket_private"`
	URL           string      `json:"url"`
	AK            string      `json:"ak"`
	SK            interface{} `json:"sk"`
	OpName        string      `json:"op_name"`
	OpPwd         string      `json:"op_pwd"`
	FileType      string      `json:"file_type"`
	MIMEType      string      `json:"mimetype"`
	MaxSize       uint64      `json:"max_size"`
	AutoName      int         `json:"autoname"`
	DirRule       string      `json:"dirrule"`
	NameRule      string      `json:"namerule"`
	OriginLink    int         `json:"origin_link"`
}

//Task 任务执行接口
type Task interface {
	Excute()
	Init() bool
	Error() bool
}

//Run 根据任务类型处理任务
func (task *taskInfo) Run() {
	var newTask Task
	switch {
	case task.sqlInfo.TaskType == "uploadSingleToOnedrive":
		newTask = &OneDriveUpload{
			Info:     task,
			Tried:    0,
			Type:     task.sqlInfo.TaskType,
			BasePath: task.siteInfo["basePath"],
		}
		newTask.Init()
	case task.sqlInfo.TaskType == "uploadChunksToOnedrive":
		newTask = &OneDriveUpload{
			Info:     task,
			Tried:    0,
			Type:     task.sqlInfo.TaskType,
			BasePath: task.siteInfo["basePath"],
		}
		newTask.Init()
	case task.sqlInfo.TaskType == "UploadRegularRemoteDownloadFileToOnedrive":
		newTask = &OneDriveUpload{
			Info:     task,
			Tried:    0,
			Type:     task.sqlInfo.TaskType,
			BasePath: task.siteInfo["basePath"],
		}
		newTask.Init()
	case task.sqlInfo.TaskType == "UploadLargeRemoteDownloadFileToOnedrive":
		newTask = &OneDriveUpload{
			Info:     task,
			Tried:    0,
			Type:     task.sqlInfo.TaskType,
			BasePath: task.siteInfo["basePath"],
		}
		newTask.Init()
	}
	newTask.Excute()
}
