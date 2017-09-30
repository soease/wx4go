/*
功能：微信个人服务程序
开发：Ease
时间：2017-9-3
修改：2017-9-28
备注：env GOOS=linux GOARCH=mipsle go build -ldflags "-s -w" .
*/

package main

import (
	"bufio"
	"bytes"
	"code.google.com/p/mahonia"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/nfnt/resize"
	"github.com/op/go-logging"
	e "github.com/soease/wx4go/enum"
	m "github.com/soease/wx4go/model"
	s "github.com/soease/wx4go/service"
	"github.com/soease/wx4go/tools"
	"github.com/widuu/goini"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	err          error
	QRFile       string    //二维码文件
	appConfig    AppConfig //系统配置
	chatString   chan string
	logFile      *os.File
	MeToUser     string //主动发送信息到用户
	log          = logging.MustGetLogger("example")
	UserInfoList = make(map[string]string) //把用户信息存起来
)

type AppConfig struct { //配置
	AutoReplay_PrivateChat     string //私聊自动回复
	AutoReplay_FilterKey       string //屏蔽关键词自动回复
	AutoReplay_FunKey          string //调戏功能关键词
	AutoReplay_PrivateFunction string //私人功能定义
	LogFile                    string //日志文件
	ComDisplay                 string //串口屏位置
	FilterKey                  string //屏蔽关键词
	ChatAt                     string //聊天对象
}

type InfoRet struct { //AI返回信息解析
	Result  int    `json:"result"`
	Content string `json:"content"`
}

//读取配置文件
func init() {
	confInit()   //系统配置
	logInit()    //日志初始化
	go getChat() //命令行输入信息
	QuitSystem() //系统退出处理
}

func main() {
	var (
		cmd          *exec.Cmd
		Message      string
		UserNickName string
		EchoMessage  string
		loginMap     m.LoginMap
		contactMap   map[string]m.User
		ToUserName   string
		FromUserName string
		MsgType      int                                    //信息类型
		Content      string                                 //信息内容
		retcode      int64                                  //微信信息反馈
		selector     int64                                  //微信信息反馈
		FunUser      = make(map[string]string)              //进入调戏模式用户
		regAt        = regexp.MustCompile(appConfig.ChatAt) // 群聊时其他人说话时会在前面加上@XXX
		regGroup     = regexp.MustCompile(`^@@.+`)
		regAd        = regexp.MustCompile(appConfig.FilterKey)
	)

	Log(logging.INFO, "系统启动中,运行环境:", runtime.GOOS, runtime.GOARCH)

	// 从微信服务器获取UUID  ------------------------------------------------------
	uuid, err := s.GetUUIDFromWX()
	panicErr(false, err)

	Log(logging.INFO, "获取二维码...")
	QRFile, err = s.DownloadImagIntoDir(e.QRCODE_URL + uuid) // 根据UUID获取二维码
	panicErr(false, err)

	Log(logging.INFO, "显示二维码...")
	if runtime.GOOS == "linux" {
		if runtime.GOARCH == "mipsle" { // 在MT7688中运行
			ComDisplayQRCode(appConfig.ComDisplay)
			cmd = exec.Command("echo") //没有实际用途，仅是为了统一处理后面的Kill
		} else { // 在普通的Linux环境下运行
			cmd = exec.Command("eog", QRFile)
		}
	} else if runtime.GOOS == "windows" { // 在windows环境中运行
		cmd = exec.Command("cmd", "/c start "+QRFile)
	}
	cmd.Start()

	// 轮询服务器判断二维码是否扫过暨是否登陆  ----------------------------------------
	for {
		Log(logging.NOTICE, "正在验证登陆...")
		status, msg := s.CheckLogin(uuid)

		if status == 200 {
			Log(logging.INFO, "登陆成功,处理登陆信息...")
			loginMap, err = s.ProcessLoginInfo(msg)
			panicErr(false, err)

			Log(logging.INFO, "登陆信息处理完毕,正在初始化微信....")
			err = s.InitWX(&loginMap)
			panicErr(true, err) //登陆出现错误，直接退出系统

			Log(logging.INFO, "初始化完毕,通知微信服务器登陆状态变更...")
			err = s.NotifyStatus(&loginMap)
			panicErr(false, err)

			Log(logging.INFO, "通知完毕.")
			Log(logging.DEBUG, e.SKey, ": ", loginMap.BaseRequest.SKey)
			Log(logging.DEBUG, e.PassTicket, ": ", loginMap.PassTicket)
			break
		} else if status == 201 {
			Log(logging.NOTICE, "请在手机上确认")
		} else if status == 408 {
			Log(logging.NOTICE, "请扫描二维码")
		} else {
			Log(logging.ERROR, msg)
		}
	}

	if runtime.GOARCH == "mipsle" {
		go Command(fmt.Sprintf(`echo "CLS(0);\r\n"> %s`, appConfig.ComDisplay))
	}
	cmd.Process.Kill() //关闭显示的二维码图（Win下未生效）

	// 扫码成功 ----------------------------------------------
	Log(logging.INFO, "开始获取联系人信息...")
	contactMap, err = s.GetAllContact(&loginMap)
	panicErr(false, err)
	Log(logging.INFO, fmt.Sprintf("成功获取 %d个 联系人信息。", len(contactMap)))
	Log(logging.INFO, "开始监听消息响应...")

	for {
		CommandControl(&loginMap, contactMap) //控制台命令

		// 获取微信信息 ----------------------------------------------------
		retcode, selector, err = s.SyncCheck(&loginMap)
		if err != nil {
			Log(logging.ERROR, err.Error())
			if retcode == 1101 {
				Log(logging.INFO, "帐号已在其他地方登陆,程序将退出.")
				os.Exit(2)
			}
			continue
		}

		if retcode == 0 && selector != 0 { // 有新消息产生
			wxRecvMsges, err := s.WebWxSync(&loginMap)
			printErr(err)

			for i := 0; i < wxRecvMsges.MsgCount; i++ {
				ToUserName = wxRecvMsges.MsgList[i].ToUserName     //接收人
				FromUserName = wxRecvMsges.MsgList[i].FromUserName //发送人
				MsgType = wxRecvMsges.MsgList[i].MsgType           //信息类型
				Content = wxRecvMsges.MsgList[i].Content           //内容

				if regGroup.MatchString(FromUserName) { //是否在群中
					UserNickName, Message = getMessage(&loginMap, Content, FromUserName)
				} else {
					Message = Content
				}
				if MsgType == 1 { // 普通文本消息
					Log(logging.INFO, contactMap[FromUserName].NickName, ":", Message)

					if regGroup.MatchString(wxRecvMsges.MsgList[i].FromUserName) && regAt.MatchString(Content) { // 有人在群里@我，发个消息回答一下
						EchoMessage = Chat(&loginMap, 1, ToUserName, FromUserName, appConfig.AutoReplay_PrivateChat)
					} else if !regGroup.MatchString(FromUserName) { //不在群里私聊我
						_, ok := FunUser[FromUserName]
						if regAd.MatchString(Content) { // 有人私聊我，并且内容含有「朋友圈」、「点赞」等敏感词，则回复
							EchoMessage = Chat(&loginMap, 1, ToUserName, FromUserName, appConfig.AutoReplay_FilterKey)
						} else if strings.EqualFold(Content, appConfig.AutoReplay_FunKey) { // 有人私聊我，开启调戏功能
							if ok {
								EchoMessage = fmt.Sprintf(appConfig.AutoReplay_PrivateFunction, "退出")
								delete(FunUser, FromUserName)
							} else {
								EchoMessage = fmt.Sprintf(appConfig.AutoReplay_PrivateFunction, "进入")
								FunUser[FromUserName] = FromUserName
							}
							EchoMessage = Chat(&loginMap, 1, ToUserName, FromUserName, EchoMessage)
						} else if ok { //已进入调戏模式
							EchoMessage = Chat(&loginMap, 1, ToUserName, FromUserName, AI(Content))
						}

					}
				} else if MsgType == 3 { //图片消息
					picXML := m.PicInfo{}
					err = xml.Unmarshal([]byte(tools.Html2Txt(Message)), &picXML)
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", UserNickName, ": 发了一张图片 ", PicUrl(wxRecvMsges.MsgList[i].MsgId))
				} else if MsgType == 34 { //语音消息
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", strings.Split(Content, ":<br/>")[0], ": 发了一个语音信息")
				} else if MsgType == 42 { //共享名片

				} else if MsgType == 43 { //视频消息
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", strings.Split(Content, ":<br/>")[0], ": 发了一个视频")
				} else if MsgType == 47 { //动画表情
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", strings.Split(Content, ":<br/>")[0], ": 发了一个发情")
				} else if MsgType == 48 { //位置消息

				} else if MsgType == 49 { //分享链接
					linkXML := m.LinkInfo{}
					err = xml.Unmarshal([]byte(tools.Html2Txt(Message)), &linkXML)
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", linkXML.FromUser, " 发了一条分享链接,标题:", linkXML.Msg.Title, "说明:", linkXML.Msg.Desc)
				} else if MsgType == 51 { //微信初始化消息
					if strings.HasPrefix(ToUserName, "@@") {
						Log(logging.INFO, contactMap[FromUserName].NickName, ": 客户端进入微信群", s.GetUserName(&loginMap, ToUserName))
					} else {
						Log(logging.INFO, contactMap[FromUserName].NickName, ": 客户端进入微信", s.GetUserName(&loginMap, ToUserName))
					}
				} else if MsgType == 62 { //小视频

				} else if MsgType == 10000 { //系统信息
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", strings.Replace(Content, ":<br/>", ": ", -1))
				} else if MsgType == 10002 { //撤回消息
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", strings.Replace(Content, ":<br/>", ": ", -1), " 撤回了一条消息")
				} else {
					Log(logging.INFO, fmt.Sprintf("%d", MsgType), contactMap[FromUserName].NickName+":", Content)
				}

				if EchoMessage != "" { //显示回复信息
					Log(logging.INFO, "我的回复：", EchoMessage)
					EchoMessage = ""
				}
			}
		}
	}
}

// -------------------------------------------------------------------------------------------------------------

// 获取图片
func PicUrl(msgid string) string {
	return "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetmsgimg?&type=slave&MsgID=" + msgid
}

//获取用户信息及聊天内容
func getMessage(logMap *m.LoginMap, content string, groupid string) (UserNickName string, Message string) {
	UserID := strings.Split(content, ":<br/>")[0]
	if _, ok := UserInfoList[UserID]; ok {
		UserNickName = UserInfoList[UserID]
	} else {
		UserNickName = s.GetGroupUserName(logMap, UserID, groupid)
		UserInfoList[UserID] = UserNickName
	}

	Message = strings.Replace(content, UserID, UserNickName, -1)
	Message = strings.Replace(Message, ":<br/>", ": ", -1)

	return
}

//获取用户信息
func getUserName(logMap *m.LoginMap, userID string) (nickName string) {
	if _, ok := UserInfoList[userID]; ok {
		nickName = UserInfoList[userID]
	} else {
		nickName = s.GetUserName(logMap, userID)
		UserInfoList[userID] = nickName
	}
	return
}

// AI机器人
func AI(q string) string {
	resp, err := http.Get("http://api.qingyunke.com/api.php?key=free&appid=0&msg=" + q)
	if err != nil {
		return ""
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return ""
		} else {
			atr := InfoRet{}
			json.Unmarshal(body, &atr)
			return strings.Replace(atr.Content, "{br}", "\n", -1)
		}
	}
}

//聊天
func Chat(logMap *m.LoginMap, chatType int, FromUser, ToUser, Content string) string {
	wxSendMsg := m.WxSendMsg{}
	wxSendMsg.Type = chatType
	wxSendMsg.Content = Content
	wxSendMsg.FromUserName = FromUser
	wxSendMsg.ToUserName = ToUser
	wxSendMsg.LocalID = fmt.Sprintf("%d", time.Now().Unix())
	wxSendMsg.ClientMsgId = wxSendMsg.LocalID

	time.Sleep(time.Second)
	go s.SendMsg(logMap, wxSendMsg)
	return wxSendMsg.Content
}

//抛异常
func panicErr(exit bool, err error) {
	if err != nil {
		if exit {
			Log(logging.WARNING, err.Error())
			os.Exit(1)
		} else {
			panic(err)
		}
	}
}

//显示异常
func printErr(err error) {
	if err != nil {
		Log(logging.ERROR, err.Error())
	}
}

//获取命令行输入
func getChat() {
	chatString = make(chan string)

	reader := bufio.NewReader(os.Stdin)
	for {
		data, _, _ := reader.ReadLine()
		chatString <- string(data)
	}
}

func iif(sour bool, ret1 string, ret2 string) string {
	if sour {
		return ret1
	} else {
		return ret2
	}
}

//过滤用户昵称中的符号
func FilterName(name string) []byte {
	var nameRegexp = regexp.MustCompile("\\<[\\S\\s]+?\\>")
	return nameRegexp.ReplaceAll([]byte(name), []byte(""))
}

//Linux Shell
func Command(cmd string) {
	c := exec.Command("ash", "-c", cmd)
	c.Start()
}

//日志功能
func Log(logType logging.Level, args ...string) {
	new := make([]interface{}, len(args))
	for i, v := range args {
		new[i] = interface{}(v)
	}
	if logType == logging.DEBUG {
		log.Debug(new...)
	} else if logType == logging.ERROR {
		log.Error(new...)
	} else {
		log.Info(new...)
	}
	if runtime.GOARCH == "mipsle" && logType != logging.NOTICE { //当为logging.NOTICE时不显示到串口屏,避免影响当前显示的内容
		var enc mahonia.Encoder
		enc = mahonia.NewEncoder("gbk")
		if ret, ok := enc.ConvertStringOK(strings.Join(args, "")); ok {
			go Command(fmt.Sprintf(`echo "BOXF(0,0,220,52,0);BS16(1,1,220,3,'`+ret+`',15);\r\n"> %s`, appConfig.ComDisplay))
		}
	}

}

//载入图片并解析
func LoadImage(path string) (img image.Image, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	img, err = jpeg.Decode(file)
	return
}

//退出系统
func QuitSystem() {
	c := make(chan os.Signal)                                                          //创建监听退出chan
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) //监听指定信号 ctrl+c kill
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				Log(logging.INFO, "请求退出系统", s.String())
				os.Remove(QRFile) //删除二维码文件
				logFile.Close()   //关闭日志文件
				Log(logging.INFO, "处理完毕,退出系统\n\n")
				os.Exit(0)
			default:
				Log(logging.INFO, s.String())
			}
		}
	}()
}

//日志初始化
func logInit() {
	format := logging.MustStringFormatter(`%{color}%{time:2006-01-02 15:04:05} %{level:.4s} %{id:03x}%{color:reset} %{message}`) //日志输出及格式

	logFile, err = os.OpenFile(appConfig.LogFile, os.O_APPEND|os.O_CREATE, 0666)
	panicErr(false, err)
	backend1 := logging.NewLogBackend(logFile, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(backend1Formatter, backend2Formatter)
}

//配置初始化
func confInit() {
	conf := goini.SetConfig("./app.conf")
	appConfig.AutoReplay_PrivateChat = conf.GetValue("chat", "RetPrivateChat")
	appConfig.AutoReplay_FilterKey = conf.GetValue("chat", "RetFilterKey")
	appConfig.AutoReplay_FunKey = conf.GetValue("chat", "FunKey")
	appConfig.AutoReplay_PrivateFunction = conf.GetValue("chat", "PrivateFunction")
	appConfig.LogFile = conf.GetValue("chat", "LogFile")
	appConfig.ComDisplay = conf.GetValue("chat", "ComDisplay")
	appConfig.FilterKey = conf.GetValue("chat", "FilterKey")
	appConfig.ChatAt = conf.GetValue("chat", "ChatAt")
}

//通过串口屏显示二维码
func ComDisplayQRCode(com string) {
	time.Sleep(time.Second)
	go Command(fmt.Sprintf(`stty -F %s raw 115200; echo "CLS(0);\r\n"> %`, com, com))
	var buf bytes.Buffer
	var r uint32

	Log(logging.NOTICE, "二维码处理中...")
	src, err := LoadImage(QRFile)
	panicErr(false, err)
	dst := resize.Resize(200, 200, src, resize.Lanczos2) // 缩略图的大小。我的串口屏只有220*176

	for i := 10; i < 190; i++ {
		for n := 10; n < 190; n++ {
			r, _, _, _ = dst.At(i, n).RGBA() //获取某点颜色
			if r > 50000 {                   // 简单判断是否是有色点
				buf.WriteString(fmt.Sprintf("PS(%d,%d,15);", n-10, i-10))
			}
			if len(buf.String()) > 900 {
				go Command("echo \"" + buf.String() + "\r\n\"> " + com)
				buf.Reset() //清空缓存
			}
		}
	}
}

//聊天信息中的命令行控制
func CommandControl(loginMap *m.LoginMap, contactMap map[string]m.User) {
	select {
	case MeChatString := <-chatString:
		if strings.ToUpper(MeChatString) == "U" { // 列出用户
			Log(logging.INFO, "列出用户.")
			for i, n := range contactMap {
				if strings.HasPrefix(i, "@@") == true { //先列出群
					Log(logging.INFO, fmt.Sprintf("%s %-30s", i[:5], FilterName(n.NickName)))
				}
			}
			for i, n := range contactMap {
				if strings.HasPrefix(i, "@@") == false && contactMap[i].VerifyFlag != 24 { //再列出用户,公众号不显示
					Log(logging.INFO, fmt.Sprintf("%s %-30s %s %s %s", i[:5], FilterName(n.NickName), iif(n.Sex == 1, "男", "女"), n.Province, n.City))
				}
			}
		} else if strings.ToUpper(MeChatString) == "QUIT" { // 退出系统
			Log(logging.INFO, "系统退出.")
			close(chatString)
			os.Exit(1)
		} else if strings.HasPrefix(MeChatString, "@") { //指定发送人
			if MeChatString[5:6] == ":" {
				for i, _ := range contactMap {
					if strings.HasPrefix(i[:5], MeChatString[0:5]) {
						MeToUser = i
						em := Chat(loginMap, 1, loginMap.SelfUserName, MeToUser, MeChatString[6:])
						Log(logging.INFO, fmt.Sprintf("我发给%s的消息：%s", getUserName(loginMap, MeToUser), em))
						break
					}
				}
			}
		} else {
			if MeToUser != "" {
				em := Chat(loginMap, 1, loginMap.SelfUserName, MeToUser, MeChatString)
				Log(logging.INFO, fmt.Sprintf("我发给%s的消息：%s", getUserName(loginMap, MeToUser), em))
			} else {
				Log(logging.INFO, "不知道消息发向何处: ", MeChatString)
			}
		}
		return
	default:
		return
	}
}
