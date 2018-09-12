package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"./task"

	"gopkg.in/yaml.v2"
)

type taskConfig struct {
	TOKEN   string `yaml:"token"`
	APIURL  string `yaml:"api"`
	TASKNUM int    `yaml:"taskNum"`
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

	var config taskConfig
	_, err := config.getConf()
	if err == nil {
		log.Printf("[INFO] Config information:  %v ", config)
		api := task.ApiInfo{TOKEN: config.TOKEN, APIURL: config.APIURL}
		basicInfo := api.GetBasicInfo()
		if basicInfo != "" {
			log.Printf("[INFO] Basic Info:  %v ", basicInfo)
			var siteInfo map[string]string
			err := json.Unmarshal([]byte(basicInfo), &siteInfo)
			if err != nil {
				log.Printf("[ERROR] Failed to decode basic infomation,  %v ", err.Error())
			}
			for {
				taskListContent := api.GetTaskList(config.TASKNUM)
				task.Init(taskListContent, api)
				break
			}
		}

	}

}
