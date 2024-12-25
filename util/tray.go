package util

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"log"
)

// InitTray 初始化系统托盘
func InitTray(mainWindow *fyne.Window) error {

	//将程序强转为桌面模式
	desk, ok := (*application).(desktop.App)
	if !ok {
		return errors.New("创建系统托盘失败")
	}

	//对主窗口进行关闭拦截
	(*mainWindow).SetCloseIntercept(func() {
		//设置为隐藏这样总会有一个程序存在不导致退出
		(*mainWindow).Hide()
	})

	//设置系统托盘菜单项
	menuItemsSlice := []*fyne.MenuItem{

		//第一项，显示面板
		fyne.NewMenuItem("显示面板", func() {
			(*mainWindow).Show()
		}),
	}

	//创建托盘菜单
	trayMenu := fyne.NewMenu("微信小工具", menuItemsSlice...)

	icon, err := fyne.LoadResourceFromPath("xfj.ico")
	if err != nil {
		log.Println(err.Error())
	}
	//设置应用的icon
	(*application).SetIcon(icon)
	(*mainWindow).SetIcon(icon)
	//设置托盘的icon
	desk.SetSystemTrayIcon(icon)
	//设置系统托盘
	desk.SetSystemTrayMenu(trayMenu)
	return nil
}
