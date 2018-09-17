package task

import (
	"encoding/json"
	"log"
	"strconv"
)

//Task 任务执行接口
type Task interface {
	Excute()
	Init() bool
	Error() bool
}

type oneDriveUploadAttr struct {
	Fname    string `json:"fname"`
	Path     string `json:"path"`
	Objname  string `json:"objname"`
	SavePath string `json:"savePath"`
	Fsize    uint64 `json:"fsize"`
	PicInfo  string `json:"picInfo"`
	PolicyID int    `json:"policyId"`
}

//OneDriveUpload OneDrive上传类型Task
type OneDriveUpload struct {
	Info  *taskInfo
	Tried int
	Attr  oneDriveUploadAttr
}

//Excute 执行Onedrive上传Task
func (task *OneDriveUpload) Excute() {

}

//Init 执行Onedrive上传Task
func (task *OneDriveUpload) Init() bool {
	if task.Tried >= 2 {
		log.Print("[ERROR] Failed to get policy #" + strconv.Itoa(task.Attr.PolicyID) + " Info, abandoned")
		return task.Error()
	}
	if task.Attr.Fname == "" {
		err := json.Unmarshal([]byte(task.Info.sqlInfo.Attr.(string)), &task.Attr)
		if err != nil {
			log.Printf("[ERROR] Failed to decode policy infomation,  %v ", err.Error())
			return task.Error()
		}
	}
	var policyString string
	policyString = task.Info.apiInfo.GetPolicy(task.Attr.PolicyID)
	if policyString == "" {
		task.Tried++
		log.Print("[ERROR] Failed to get policy #" + strconv.Itoa(task.Attr.PolicyID) + " Info, retring...")
		return task.Init()
	}
	return true
}

func (task *OneDriveUpload) Error() bool {
	return false
}
