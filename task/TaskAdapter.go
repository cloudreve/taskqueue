package task

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"../onedrive"
)

const uploadThreadNum int = 4

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
	Fname    string      `json:"fname"`
	Path     string      `json:"path"`
	Objname  string      `json:"objname"`
	SavePath string      `json:"savePath"`
	Fsize    uint64      `json:"fsize"`
	PicInfo  string      `json:"picInfo"`
	PolicyID int         `json:"policyId"`
	Chunks   []chunkFile `json:"chunks"`
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

	if task.Type == "uploadSingleToOnedrive" {
		task.uploadRegularFile(&Client)
	} else if task.Type == "uploadChunksToOnedrive" {
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
	fmt.Println(url)

	chunkList, chunkErr := task.buildChunks()
	if chunkErr != "" {
		task.Log("[Error] Failed to upload chunks," + chunkErr)
		task.Error()
		return
	}

	var wg sync.WaitGroup
	ch := make(chan Chunk)
	isFailed := false

	for index := 0; index < uploadThreadNum; index++ {
		wg.Add(1)
		go task.uploadSingleChunk(&wg, ch, &isFailed)
	}

	for _, v := range chunkList {
		if isFailed {
			close(ch)
			break
		}
		ch <- v
	}
	close(ch)
	wg.Wait()

}

func (task *OneDriveUpload) uploadSingleChunk(wg *sync.WaitGroup, ch chan Chunk, isFailed *bool) {
	for {
		chunk, opened := <-ch
		if !opened {
			fmt.Println("quit")
			wg.Done()
			return
		}
		fmt.Println(chunk.From)
	}
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

	for _, v := range task.Attr.Chunks {

		chunkPath := task.BasePath + "public/uploads/chunks/" + v.ObjName + ".chunk"

		fileInfo, err := os.Stat(chunkPath)
		if os.IsNotExist(err) {
			return chunkList, "Chunk file " + chunkPath + " not exist"
		}
		chunkSize := int(fileInfo.Size())

		chunkList = append(chunkList, Chunk{
			Type:      chunkType,
			From:      offset,
			To:        offset + chunkSize - 1,
			ChunkPath: chunkPath,
		})
		offset += chunkSize
	}

	return chunkList, ""

}

func (task *OneDriveUpload) uploadRegularFile(Client *onedrive.Client) {
	var filePath string
	filePath = task.BasePath + "public/uploads/" + task.Attr.SavePath + "/" + task.Attr.Objname
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
