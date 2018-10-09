package task

import (
	"encoding/json"
	"fmt"
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
		log.Println("[Error] Failed to open file," + err.Error())
		task.Error()
		return
	}

	res, errorMsg := Client.PutFile("/me/drive/root:/"+task.Attr.SavePath+"/"+task.Attr.Objname, r)
	if errorMsg != "" {
		log.Println("[Error] Unload Failed," + errorMsg)
		task.Error()
	}
	fmt.Println(res)

}

//Init 执行Onedrive上传Task
func (task *OneDriveUpload) Init() bool {

	log.SetPrefix("[Task #" + strconv.Itoa(task.Info.sqlInfo.ID) + "]")

	if task.Tried >= 2 {
		log.Print("[ERROR] Failed to get policy #" + strconv.Itoa(task.Attr.PolicyID) + " Info, abandoned")
		return task.Error()
	}

	if task.Attr.Fname == "" {
		err := json.Unmarshal([]byte(task.Info.sqlInfo.Attr.(string)), &task.Attr)
		if err != nil {
			log.Printf("[ERROR] Failed to decode task infomation,  %v ", err.Error())
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

	err := json.Unmarshal([]byte(policyString), &task.Policy)
	if err != nil {
		log.Printf("[ERROR] Failed to decode policy infomation,  %v ", err.Error())
		return task.Error()
	}
	return true

}

func (task *OneDriveUpload) Error() bool {
	return false
}
