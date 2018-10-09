package task

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"../onedrive"
)

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
	Info     *taskInfo
	Tried    int
	Attr     oneDriveUploadAttr
	Policy   policy
	Type     string
	BasePath string
}

//OnedriveState OneDrive鉴权状态
type OnedriveState struct {
	RedirectURI string `json:"redirect_uri"`
	Token       struct {
		Obtained int `json:"obtained"`
		Data     struct {
			TokenType    string `json:"token_type"`
			Scope        string `json:"scope"`
			ExpiresIn    int    `json:"expires_in"`
			ExtExpiresIn int    `json:"ext_expires_in"`
			AccessToken  string `json:"access_token"`
		} `json:"data"`
	} `json:"token"`
}

//Excute 执行Onedrive上传Task
func (task *OneDriveUpload) Excute() {
	authState := OnedriveState{}
	json.Unmarshal([]byte(task.Policy.SK.(string)), &authState)
	Client := onedrive.Client{
		ClientID:     task.Policy.BucketName,
		ClientSecret: task.Policy.AK,
		AccessToken:  authState.Token.Data.AccessToken,
		Tried:        0,
	}
	Client.Init()
	var filePath string
	if task.Type == "UploadRegularRemoteDownloadFileToOnedrive" {

	} else {
		filePath = task.BasePath + "public/uploads/" + task.Attr.SavePath + "/" + task.Attr.Objname
	}

	r, err := os.Open(filePath)
	defer r.Close()
	if err != nil {
		task.Log("[Error] Failed to open file," + err.Error())
		task.Error()
		return
	}

	_, errorMsg := Client.PutFile("/me/drive/root:/"+task.Attr.SavePath+"/"+task.Attr.Objname+":/content", r)
	if errorMsg != "" {
		task.Log("[Error] Upload Failed," + errorMsg)
		task.Error()
	}

	addRes := task.Info.apiInfo.SetSuccess(task.Info.sqlInfo.ID)

	if addRes != "" {
		task.Log("[Error] " + addRes)
	}

}

//Init 执行Onedrive上传Task
func (task *OneDriveUpload) Init() bool {
	if task.Tried >= 2 {
		task.Log("[ERROR] Failed to get policy #" + strconv.Itoa(task.Attr.PolicyID) + " Info, abandoned")
		return task.Error()
	}

	if task.Attr.Fname == "" {
		err := json.Unmarshal([]byte(task.Info.sqlInfo.Attr.(string)), &task.Attr)
		if err != nil {
			task.Log("[ERROR] Failed to decode task infomation,  " + err.Error())
			return task.Error()
		}
	}

	var policyString string
	policyString = task.Info.apiInfo.GetPolicy(task.Attr.PolicyID)
	if policyString == "" {
		task.Tried++
		task.Log("[ERROR] Failed to get policy #" + strconv.Itoa(task.Attr.PolicyID) + " Info, retring...")
		return task.Init()
	}

	err := json.Unmarshal([]byte(policyString), &task.Policy)
	if err != nil {
		task.Log("[ERROR] Failed to decode policy infomation,  " + err.Error())
		return task.Error()
	}
	return true

}

func (task *OneDriveUpload) Error() bool {
	task.Info.apiInfo.SetError(task.Info.sqlInfo.ID)
	return false
}

//Log 日志
func (task *OneDriveUpload) Log(msg string) {
	log.Print("[Task #" + strconv.Itoa(task.Info.sqlInfo.ID) + "]" + msg)
}
