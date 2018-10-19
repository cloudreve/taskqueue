package task

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"../onedrive"
)

const uploadThreadNum int = 4
const aria2ChunkSize int = 10 * 1024 * 1024

type chunkFile struct {
	ID      int    `json:"id"`
	User    int    `json:"user"`
	CTX     string `json:"ctx"`
	Time    string `json:"time"`
	ObjName string `json:"obj_name"`
	ChunkID int    `json:"chunk_id"`
	Sum     int    `json:"sum"`
}

type oneDriveUploadAttr struct {
	Fname      string      `json:"fname"`
	Path       string      `json:"path"`
	Objname    string      `json:"objname"`
	SavePath   string      `json:"savePath"`
	Fsize      uint64      `json:"fsize"`
	PicInfo    string      `json:"picInfo"`
	PolicyID   int         `json:"policyId"`
	Chunks     []chunkFile `json:"chunks"`
	OriginPath string      `json:"originPath"`
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

//Chunk 文件分片
type Chunk struct {
	Type      int //分片方式，0-已切割为文件片，1-无需切割
	From      int
	To        int
	ChunkPath string
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

	if task.Type == "uploadSingleToOnedrive" || task.Type == "UploadRegularRemoteDownloadFileToOnedrive" {
		task.uploadRegularFile(&Client)
	} else if task.Type == "uploadChunksToOnedrive" || task.Type == "UploadLargeRemoteDownloadFileToOnedrive" {
		task.uploadChunks(&Client)
	}

}

func (task *OneDriveUpload) uploadChunks(Client *onedrive.Client) {
	//获取上传URL
	url, err := Client.CreateUploadSession("/me/drive/root:/" + task.Attr.SavePath + "/" + task.Attr.Objname + ":/createUploadSession")
	if err != "" {
		task.Log("[Error] Failed to create upload session," + err)
		task.Error()
		return
	}

	chunkList, chunkErr := task.buildChunks()
	if chunkErr != "" {
		task.Log("[Error] Failed to upload chunks," + chunkErr)
		task.Error()
		return
	}
	uploaded := 0
	for _, v := range chunkList {
		if task.uploadSingleChunk(v, Client, url) {
			uploaded += (v.To - v.From + 1)
			task.Log(fmt.Sprintf("[Info] Chunk uploaded, From:%d To:%d Total:%d Complete:%.2f", v.From, v.To, task.Attr.Fsize, float32(uploaded)/float32(task.Attr.Fsize)))
		} else {
			task.Info.apiInfo.CancelUploadSession(url)
			return
		}
	}

	addRes := task.Info.apiInfo.SetSuccess(task.Info.sqlInfo.ID)

	if addRes != "" {
		task.Log("[Error] " + addRes)
	}

}

func (task *OneDriveUpload) uploadSingleChunk(chunk Chunk, Client *onedrive.Client, url string) bool {

	var r *os.File
	var err error
	var bfRd *bufio.Reader
	if chunk.Type == 0 {
		r, err = os.Open(chunk.ChunkPath)
		bfRd = nil
	} else {
		r, err = os.Open(chunk.ChunkPath)
		r.Seek(int64(chunk.From), 0)
		bfRd = bufio.NewReader(r)
	}

	if err != nil {
		task.Log("[Error] Failed to open file," + err.Error())
		task.Error()
		return false
	}
	var uploadErr string
	if bfRd == nil {
		_, uploadErr = Client.UploadChunk(url, chunk.From, chunk.To, int(task.Attr.Fsize), r, nil)
	} else {
		_, uploadErr = Client.UploadChunk(url, chunk.From, chunk.To, int(task.Attr.Fsize), r, bfRd)
	}

	if uploadErr != "" {
		task.Log("[Error] Failed to upload chunk," + uploadErr)
		task.Error()
		return false
	}
	r.Close()
	return true
}

func (task *OneDriveUpload) buildChunks() ([]Chunk, string) {

	var chunkType int
	if task.Type == "uploadChunksToOnedrive" {
		chunkType = 0
	} else {
		chunkType = 1
	}

	var chunkList []Chunk
	var offset int

	if chunkType == 0 {

		for _, v := range task.Attr.Chunks {

			var (
				chunkPath string
				chunkSize int
			)
			chunkPath = task.BasePath + "public/uploads/chunks/" + v.ObjName + ".chunk"

			fileInfo, err := os.Stat(chunkPath)
			if os.IsNotExist(err) {
				return chunkList, "Chunk file " + chunkPath + " not exist"
			}
			chunkSize = int(fileInfo.Size())

			chunkList = append(chunkList, Chunk{
				Type:      chunkType,
				From:      offset,
				To:        offset + chunkSize - 1,
				ChunkPath: chunkPath,
			})

			offset += chunkSize
		}

	} else {
		for {

			var (
				chunkPath string
				chunkSize int
			)

			chunkPath = task.Attr.OriginPath
			if uint64(offset+aria2ChunkSize) > task.Attr.Fsize {
				chunkSize = int(task.Attr.Fsize) - offset
			} else {
				chunkSize = aria2ChunkSize
			}

			chunkList = append(chunkList, Chunk{
				Type:      chunkType,
				From:      offset,
				To:        offset + chunkSize - 1,
				ChunkPath: chunkPath,
			})

			offset += chunkSize

			if offset >= int(task.Attr.Fsize) {
				break
			}

		}

	}

	return chunkList, ""

}

func (task *OneDriveUpload) uploadRegularFile(Client *onedrive.Client) {
	var filePath string
	if task.Type == "UploadRegularRemoteDownloadFileToOnedrive" {
		filePath = task.Attr.OriginPath
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
		return
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
