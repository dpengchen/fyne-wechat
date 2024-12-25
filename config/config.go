package config

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"io"
	"log"
	"os"
	"time"
)

type Config struct {
	Contact       []string  `json:"contact"`       //联系人
	KeyWord       []string  `json:"keyWord"`       //关键词
	Interval      int       `json:"interval"`      //回复间隔
	Target        string    `json:"target"`        //目标人群（全部、朋友、群组）
	TomorrowReset bool      `json:"TomorrowReset"` //次日重置
	ReplayText    string    `json:"replayText"`    //回复文本
	ReplayCount   int       `json:"replayCount"`   //回复次数
	PreviousTime  time.Time `json:"previousTime"`  //上一次回复时间
}

// ApplicationConfig 配置信息
var ApplicationConfig = Config{
	Interval:      1,
	Target:        "group",
	TomorrowReset: true,
	PreviousTime:  time.Now(),
}

var ContactMap = make(map[string]bool)
var KeyWordMap = make(map[string]bool)

func init() {
	file, err := os.OpenFile("config.json", os.O_CREATE|os.O_RDONLY, 0644)
	defer file.Close()
	if err != nil {
		log.Println(err.Error())
		return
	}

	configBytes, err := io.ReadAll(file)
	if err != nil {
		log.Println(err.Error())
		return
	}

	err = json.Unmarshal(configBytes, &ApplicationConfig)
	if err != nil {
		log.Println(err.Error())
		return
	}
	configLoadToMap()
}

// SaveConfig 保存配置
func SaveConfig(window *fyne.Window) {

	//打开配置文件
	file, err := os.OpenFile("config.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		dialog.ShowInformation("警告！", "创建config.json失败！", *window)
		return
	}
	bytes, err := json.Marshal(&ApplicationConfig)
	if err != nil {
		dialog.ShowInformation("警告！", "序列化配置失败！", *window)
		return
	}
	file.Write(bytes)

	configLoadToMap()
}

// configLoadToMap 将配置转换为map
func configLoadToMap() {
	clear(ContactMap)
	clear(KeyWordMap)

	for _, contact := range ApplicationConfig.Contact {
		ContactMap[contact] = true
	}
	for _, keyWord := range ApplicationConfig.KeyWord {
		KeyWordMap[keyWord] = true
	}
}
