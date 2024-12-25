package main

import (
	"github.com/dpengchen/wechat-bot/util"
	"github.com/dpengchen/wechat-bot/window"
	"log"
	"os"
)

func init() {
	file, err := os.OpenFile("log.log", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("创建日志文件失败")
	}
	log.SetFlags(log.Llongfile)
	log.SetOutput(file)
}

// 微信机器人
func main() {
	application := util.InitWindowUtil()

	mainWindow := (*application).NewWindow("Rocc微信机器人")

	//初始化系统托盘
	if err := util.InitTray(&mainWindow); err != nil {
		log.Panicln(err.Error())
	}

	window.MainContainerInit(&mainWindow)

	//运行主窗口
	mainWindow.ShowAndRun()
}
