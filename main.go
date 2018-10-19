package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"./api"
	"./task"

	"gopkg.in/yaml.v2"
)

type taskConfig struct {
	TOKEN    string `yaml:"token"`
	APIURL   string `yaml:"api"`
	TASKNUM  int    `yaml:"taskNum"`
	DURATION int    `yaml:"Duration"`
}

func (c *taskConfig) getConf() (*taskConfig, error) {

	yamlFile, err := ioutil.ReadFile("conf.yaml")
	if err != nil {
		log.Printf("[ERROR] yamlFile.Get err   #%v ", err)
		return nil, errors.New("Cant not read config.yaml")
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, errors.New("Cant not prase config.yaml")
	}

	return c, nil
}

func main() {

	fmt.Println("Cloudreve Queue Go Version")
	fmt.Println("Author: AaronLiu <abslant@foxmail.com>")
	fmt.Println("")
	var config taskConfig
	_, err := config.getConf()
	if err == nil {

		log.Printf("[INFO] Config information:  %v ", config)
		api := api.ApiInfo{
			TOKEN:  config.TOKEN,
			APIURL: config.APIURL,
			Lock:   new(sync.Mutex),
		}
		basicInfo := api.GetBasicInfo()

		if basicInfo != "" {

			log.Printf("[INFO] Basic Info:  %v ", basicInfo)
			var siteInfo map[string]string
			err := json.Unmarshal([]byte(basicInfo), &siteInfo)
			if err != nil {
				log.Printf("[ERROR] Failed to decode basic infomation,  %v ", err.Error())
			}

			var wg sync.WaitGroup
			for i := 0; i < config.TASKNUM; i++ {
				wg.Add(1)
				log.Printf("[Info] Thread %d start", i+1)
				threadID := i
				go func() {
					for {

						api.Lock.Lock()
						taskListContent := api.GetTaskList(1)
						api.Lock.Unlock()

						if taskListContent != "none" {
							task.Init(taskListContent, api, siteInfo, threadID)
						}
						time.Sleep(time.Duration(config.DURATION) * time.Second)
					}

				}()
				time.Sleep(time.Duration(1) * time.Second)
			}

			wg.Wait()

		}

	}

}
