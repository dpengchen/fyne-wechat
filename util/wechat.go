package util

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/dpengchen/wechat-bot/config"
	"github.com/eatmoreapple/openwechat"
	"image"
	"log"
	"net/http"
	"strings"
	"time"
)

var wechat *openwechat.Bot
var infoChan = make(chan string, 10)

var msgTypeMap = map[int]string{
	1:     "文本消息",
	3:     "图片消息",
	34:    "语音消息",
	37:    "认证消息",
	40:    "好友推荐消息",
	42:    "名片消息",
	43:    "视频消息",
	47:    "表情消息",
	48:    "地理位置消息",
	49:    "APP消息",
	50:    "VOIP消息",
	52:    "VOIP结束消",
	53:    "VOIP邀请",
	62:    "小视频消息",
	10000: "系统消息",
	10002: "消息撤回",
}

func InitBot(window *fyne.Window, callback func(window *fyne.Window, information fyne.CanvasObject)) image.Image {
	if wechat != nil {
		wechat.Context().Done()
	}
	//以桌面方式运行
	wechat = openwechat.DefaultBot(openwechat.Desktop)

	//准备UUID通道
	var uuidChan = make(chan string)
	defer close(uuidChan)
	//获取二维码url
	wechat.UUIDCallback = func(uuid string) {
		//将uuid写入通道中
		uuidChan <- uuid
	}

	//设置登录成功回调
	wechat.LoginCallBack = func(body openwechat.CheckLoginResponse) {
		information, state := getSelfUserInformation(window)
		if state {
			callback(window, information)
		}
	}

	//设置消息处理
	wechat.MessageHandler = msgHandle
	wechat.MessageErrorHandler = errHandle

	//热登录，将登录信息进行保存
	//尝试使用热登录
	go func() {
		err := wechat.Login()
		if err != nil {
			dialog.ShowInformation("警告", "微信机器人登录超时", *window)
		}
	}()

	//等待获取得到uuid
	qrcodeUrl := openwechat.GetQrcodeUrl(<-uuidChan)

	//获取二维码图片
	resp, err := http.Get(qrcodeUrl)
	if err != nil {
		log.Panicln("获取微信二维码失败", err.Error())
	}

	//关闭资源
	defer resp.Body.Close()

	//解析获取图片
	qrcodeImg, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Panicln("获取微信二维码失败", err.Error())
	}

	return qrcodeImg
}

// msgHandle 信息处理
func msgHandle(msg *openwechat.Message) {
	defer func() {
		//捕捉错误
		if err := recover(); err != nil {
			infoChan <- fmt.Sprint("错误信息：", err)
		}
	}()

	cfg := config.ApplicationConfig
	nowTime := time.Now().Format("2006-01-02 15:04:05")

	//是否为朋友发送信息
	if msg.IsSendByFriend() {
		if cfg.Target == "all" || cfg.Target == "friend" {
			friend, _ := msg.Sender()
			if friend == nil {
				return
			}
			checkNameAndKeyWord(friend, msg)
			infoChan <- fmt.Sprintf("[%v]朋友：%v  =>  信息类型：%v", nowTime, friend.NickName, msgTypeMap[int(msg.MsgType)])
			log.Printf("[%v]朋友：%v  =>  信息类型：%v", nowTime, friend.NickName, msgTypeMap[int(msg.MsgType)])
		}
	} else if msg.IsSendByGroup() {
		if cfg.Target == "all" || cfg.Target == "group" {
			//是群组的信息
			group, _ := msg.SenderInGroup()
			if group == nil {
				return
			}
			checkNameAndKeyWord(group, msg)
			infoChan <- fmt.Sprintf("[%v]群组：%v  =>  信息类型：%v", nowTime, group.NickName, msgTypeMap[int(msg.MsgType)])
			log.Printf("[%v]群组：%v  =>  信息类型：%v", nowTime, group.NickName, msgTypeMap[int(msg.MsgType)])
		}
	} else {
		//其他信息
		infoChan <- fmt.Sprintf("[%v]其他信息： =>  信息类型：%v", nowTime, msgTypeMap[int(msg.MsgType)])
		log.Printf("[%v]其他信息： =>  信息类型：%v", nowTime, msgTypeMap[int(msg.MsgType)])
	}

}

func checkNameAndKeyWord(user *openwechat.User, msg *openwechat.Message) {
	//只处理文本信息
	if !msg.IsText() {
		return
	}
	_, nickNameOk := config.ContactMap[user.NickName]
	_, remarkNameOk := config.ContactMap[user.RemarkName]
	_, ok := config.KeyWordMap[msg.Content]
	if (nickNameOk || remarkNameOk) && ok {

		//当回复次数被间隔次数整除则可以回复
		if config.ApplicationConfig.ReplayCount%(config.ApplicationConfig.Interval+1) == 0 {
			msg.ReplyText(config.ApplicationConfig.ReplayText)
		} else if config.ApplicationConfig.TomorrowReset && config.ApplicationConfig.PreviousTime.Day() != time.Now().Day() {
			//不是今天的日期
			msg.ReplyText(config.ApplicationConfig.ReplayText)
		}

		//回复次数++，设置时间
		config.ApplicationConfig.PreviousTime = time.Now()
		config.ApplicationConfig.ReplayCount++
	}
}

// errHandle 错误处理
func errHandle(err error) error {
	infoChan <- fmt.Sprintf("[%v]错误日志：%v", time.Now().Format("2006-01-02 15:04:05"), err.Error())
	log.Println(err.Error())
	if !wechat.Alive() {
		infoChan <- "机器人错误退出" + err.Error()
	}
	if strings.Contains(err.Error(), "cookie invalid") {
		infoChan <- "微信在其他地方登录" + err.Error()
	}
	return nil
}

// GetChain 获取通道
func GetChain() *chan string {
	return &infoChan
}

// getSelfUserInformation 获得头像信息
func getSelfUserInformation(window *fyne.Window) (fyne.CanvasObject, bool) {

	//获取当前用户
	user, err := wechat.GetCurrentUser()
	if err != nil {
		dialog.ShowError(errors.New("获取当前登录用户失败！"), *window)
		wechat.Logout()
		return nil, false
	}
	//获取当前登录用户朋友
	friends, err := user.Friends()
	if err != nil {
		dialog.ShowError(errors.New("获取当前登录用户通讯录失败！"), *window)
		wechat.Logout()
		return nil, false
	}
	//判断朋友列表是否存在授权账户，判断是否授权条件根据签名来判断
	if exists := friends.Search(1, func(friend *openwechat.Friend) bool {
		return friend.Signature == "金钱只是一串数字，数据只是一行代码"
	}); exists.Count() == 0 {
		dialog.ShowError(errors.New("未被授权，请添加微信授权账号：JNYCIW！"), *window)
		wechat.Logout()
		return nil, false
	}

	//登录成功回调
	resp, err := user.GetAvatarResponse()
	defer resp.Body.Close()
	//获取登录用户头像
	avatarImg, _, _ := image.Decode(resp.Body)

	avatarImgWidget := canvas.NewImageFromImage(avatarImg)
	avatarImgWidget.FillMode = canvas.ImageFillOriginal
	avatarImgWidget.SetMinSize(fyne.NewSize(100, 100))
	return container.NewBorder(nil, nil, avatarImgWidget, nil,
		container.NewVBox(
			widget.NewLabel("用户名："+user.NickName),
			widget.NewLabel(fmt.Sprint("性别：", user.Sex)),
			widget.NewLabel(fmt.Sprint("个性签名：", user.Signature)),
		)), true
}
