package window

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/dpengchen/wechat-bot/config"
	"github.com/dpengchen/wechat-bot/util"
	"image/color"
	"strconv"
)

// MainContainerInit 设置主窗口布局
func MainContainerInit(window *fyne.Window) {

	//设置窗口
	w := *window

	//设置窗口大小
	w.Resize(fyne.NewSize(600, 500))

	//不允许更改大小
	w.SetFixedSize(true)

	//调用显示登录容器
	w.SetContent(loginContainer(window))
	//loginSuccessContainer(window, "")

}

// loginContainer 登录显示
func loginContainer(window *fyne.Window) *fyne.Container {

	//初始化wechat机器人
	qrcodeImg := util.InitBot(window, loginSuccessContainer)

	//得到二维码图片
	imageCanvas := canvas.NewImageFromImage(qrcodeImg)
	//设置填充模式，不设置不显示
	imageCanvas.SetMinSize(fyne.NewSize(250, 250))
	imageCanvas.FillMode = canvas.ImageFillContain
	refreshLoginQrcode := widget.NewButton("重新获取二维码", func() {
		//重新初始化机器人
		qrcodeImg = util.InitBot(window, loginSuccessContainer)
		//设置新的二维码，并刷新
		imageCanvas.Image = qrcodeImg
		imageCanvas.Refresh()
	})

	//创建上下结构
	vBox := container.New(layout.NewVBoxLayout(), imageCanvas, refreshLoginQrcode)

	//添加居中布局
	return container.New(layout.NewCenterLayout(), vBox)
}

// loginSuccessContainer 登录成功回调
func loginSuccessContainer(window *fyne.Window, information fyne.CanvasObject) {

	//联系人名
	contactsBind := binding.NewStringList()
	contactsBind.Set(config.ApplicationConfig.Contact)
	contactList := generatorList(&contactsBind)

	//关键词
	keyWordBind := binding.NewStringList()
	keyWordBind.Set(config.ApplicationConfig.KeyWord)
	keyWordList := generatorList(&keyWordBind)

	//获取通道
	infoChan := util.GetChain()

	//创建添加按钮
	addListenContact := widget.NewButton("添加监听人", func() {
		contactsBind.Append("")
	})
	addListenKeyWord := widget.NewButton("添加关键字", func() {
		keyWordBind.Append("")
	})

	//因为直接使用Box，它会吧组件的宽度设置为最小宽度，所以需要用Border包起来
	rectangle := canvas.NewRectangle(&color.NRGBA{0, 0, 0, 0})
	rectangle.SetMinSize(fyne.NewSize(1, 150))
	contactContainer := container.NewBorder(widget.NewLabel("联系人"), addListenContact, rectangle, nil, contactList)
	keyWordContainer := container.NewBorder(widget.NewLabel("关键词"), addListenKeyWord, rectangle, nil, keyWordList)

	//间隔回复
	intervalBind := binding.NewString()
	intervalBind.Set(fmt.Sprintf("%d", config.ApplicationConfig.Interval))
	//次日重置
	tomorrowResetBind := binding.NewBool()
	tomorrowResetBind.Set(config.ApplicationConfig.TomorrowReset)
	//针对人群
	targetSelectWidget := widget.NewSelect([]string{"all", "friend", "group"}, func(key string) {
		config.ApplicationConfig.Target = key
	})
	targetSelectWidget.Selected = config.ApplicationConfig.Target

	//回复信息
	replayTextBind := binding.NewString()
	replayContainer := container.NewBorder(nil, nil, widget.NewLabel("回复信息"), nil, widget.NewEntryWithData(replayTextBind))

	//保存配置应用
	saveConfigBtn := widget.NewButtonWithIcon("应用配置", theme.DocumentSaveIcon(), func() {
		//将配置内容提交到结构体中
		contact, _ := contactsBind.Get()
		config.ApplicationConfig.Contact = contact
		keyWord, _ := keyWordBind.Get()
		config.ApplicationConfig.KeyWord = keyWord
		interval, _ := intervalBind.Get()
		intervalNumber, _ := strconv.Atoi(interval)
		config.ApplicationConfig.Interval = intervalNumber
		tomorrowReset, _ := tomorrowResetBind.Get()
		config.ApplicationConfig.TomorrowReset = tomorrowReset
		replayText, _ := replayTextBind.Get()
		config.ApplicationConfig.ReplayText = replayText

		//保存配置
		config.SaveConfig(window)
	})

	otherOption := container.NewHBox(
		//开启间隔
		widget.NewLabel("回复间隔"), widget.NewEntryWithData(intervalBind),
		//针对目标
		widget.NewLabel("针对人群"),
		targetSelectWidget,
		//次日重置
		container.NewPadded(),
		widget.NewLabel("次日重置"),
		widget.NewCheckWithData("", tomorrowResetBind),
		//保存配置按钮
		saveConfigBtn,
	)

	//将监听人和关键字并列
	columns := container.NewGridWithColumns(2, contactContainer, keyWordContainer)

	//日志内容
	journalEntry := widget.NewMultiLineEntry()
	journalEntry.SetMinRowsVisible(4)
	journalEntry.TextStyle.TabWidth = 1

	//打印日志
	go func() {
		for {
			select {
			case info := <-*infoChan:
				if info == "机器人错误退出" {
					dialog.ShowError(errors.New(info), *window)
					loginContainer(window)
					return
				}
				journalEntry.Text = fmt.Sprintf("%s\n%s", info, journalEntry.Text)
				journalEntry.Refresh()
			}
		}
	}()

	//添加到容器中
	box := container.NewVBox(
		information,
		columns, container.NewPadded(),
		replayContainer,
		container.NewPadded(),
		otherOption,
		container.NewPadded(),
		journalEntry)
	scroll := container.NewVScroll(box)
	(*window).SetContent(scroll)

}

// generatorList 生成列表
func generatorList(list *binding.StringList) *widget.List {
	stringList := *list

	contacts := widget.NewListWithData(stringList, func() fyne.CanvasObject {
		//输入框
		entry := widget.NewEntry()
		//删除按钮
		removeBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {})
		HBox := container.NewBorder(nil, nil, nil, removeBtn, entry)
		return HBox
	}, func(item binding.DataItem, object fyne.CanvasObject) {
		//转换为
		bindingStr := item.(binding.String)
		val, _ := bindingStr.Get()
		//获取容器中的元素
		objects := object.(*fyne.Container).Objects
		entry := objects[0].(*widget.Entry)
		//进行双向绑定
		entry.Bind(bindingStr)
		removeBtn := objects[1].(*widget.Button)
		//删除按钮事件
		removeBtn.OnTapped = func() {
			stringList.Remove(val)
		}
	})

	return contacts
}
