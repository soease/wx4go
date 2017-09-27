/*
功能：微信个人服务程序
开发：Ease
时间：2017-9-3
修改：2017-9-23
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
	"github.com/Comdex/imgo"
	"github.com/op/go-logging"
	e "github.com/soease/wx4go/enum"
	m "github.com/soease/wx4go/model"
	s "github.com/soease/wx4go/service"
	"github.com/widuu/goini"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var (
	err          error
	appConfig    AppConfig         //系统配置
	UserInfoList map[string]string //把用户信息存起来
	chatString   chan string
	logFile      *os.File
	log          = logging.MustGetLogger("example")
	format       = logging.MustStringFormatter(`%{color}%{time:2006-01-02 15:04:05} %{level:.4s} %{id:03x}%{color:reset} %{message}`)
	logFileName  = "log.txt"
	QRFile       = "./qrcode.jpg"
)

type AppConfig struct {
	AutoReplay_PrivateChat     string //私聊自动回复
	AutoReplay_KeyFilter       string //屏蔽关键词
	AutoReplay_FunKey          string //调戏功能关键词
	AutoReplay_PrivateFunction string //私人功能定义
	COMScreen                  bool   //串口屏上显示二维码
}

type InfoRet struct { //AI返回信息解析
	Result  int    `json:"result"`
	Content string `json:"content"`
}

// 读取配置文件
func init() {
	// 日志输出及格式
	logFile, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE, 0666)
	panicErr(err)
	backend1 := logging.NewLogBackend(logFile, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(backend1Formatter, backend2Formatter)

	//系统配置
	conf := goini.SetConfig("./app.conf")
	appConfig.AutoReplay_PrivateChat = conf.GetValue("chat", "PrivateChat")
	appConfig.AutoReplay_KeyFilter = conf.GetValue("chat", "KeyFilter")
	appConfig.AutoReplay_FunKey = conf.GetValue("chat", "FunKey")
	appConfig.AutoReplay_PrivateFunction = conf.GetValue("chat", "PrivateFunction")
	UserInfoList = make(map[string]string)

	chatString = make(chan string)
	go getChat()
}

func main() {
	var (
		cmd          *exec.Cmd
		Message      string
		UserNickName string
		EchoMessage  string
		loginMap     m.LoginMap
		uuid         string
		contactMap   map[string]m.User
		AdminID      string //进入管理员帐号
		ToUserName   string
		FromUserName string
		MeToUser     string //主动发送信息到用户
	)

	Log(logging.INFO, "系统启动中,运行环境:", runtime.GOOS, runtime.GOARCH)
	// 从微信服务器获取UUID
	uuid, err = s.GetUUIDFromWX()
	panicErr(err)

	Log(logging.INFO, "获取二维码...")
	// 根据UUID获取二维码
	err = s.DownloadImagIntoDir(e.QRCODE_URL+uuid, ".")
	panicErr(err)

	Log(logging.INFO, "显示二维码...")
	if runtime.GOOS == "linux" {
		if runtime.GOARCH == "mipsle" {
			// 在MT7688中运行
			//go Command("stty -F /dev/ttyS0 raw 115200")  //速率设置
			Log(logging.INFO, "二维码处理中,速度较慢请稍候...")
			var img [][][]uint8
			go Command(`echo "CLS(0);\r\n"> /dev/ttyS0`)
			img, _ = imgo.ResizeForMatrix(QRFile, 200, 195)
			Log(logging.INFO, "二维码显示中...")
			var buf bytes.Buffer
			for i := 10; i < 190; i++ {
				for n := 10; n < 190; n++ {
					if img[i][n][1] > 200 {
						buf.WriteString(fmt.Sprintf("PS(%d,%d,15);", n-10, i-10))
					}
					if len(buf.String()) > 900 {
						Command("echo \"" + buf.String() + "\r\n\"> /dev/ttyS0")
						buf.Reset() //清空缓存
					}
				}
			}
			cmd = exec.Command("echo") //没有实际用途，仅是为了统一处理后面的Kill
		} else {
			// 在普通的Linux环境下运行
			cmd = exec.Command("eog", QRFile)
			/*
				//在命令行显示二维码，需要把字符界面缩很小
				var img [][][]uint8
				img, _ = imgo.ResizeForMatrix(QRFile, 280, 180)
				for i := 0; i < 180; i++ {
					for n := 0; n < 280; n++ {
						if img[i][n][1] > 200 {
							fmt.Print("█")
						} else {
							fmt.Print(" ")
						}
					}
					fmt.Println("")
				}
			*/
		}
	} else if runtime.GOOS == "windows" {
		// 在windows环境中运行
		cmd = exec.Command("cmd", "/c start "+QRFile)
	}
	cmd.Start()

	// 轮询服务器判断二维码是否扫过暨是否登陆了
	for {
		Log(logging.INFO, "正在验证登陆... ...")
		status, msg := s.CheckLogin(uuid)

		if status == 200 {
			Log(logging.INFO, "登陆成功,处理登陆信息...")
			loginMap, err = s.ProcessLoginInfo(msg)
			panicErr(err)

			Log(logging.INFO, "登陆信息处理完毕,正在初始化微信....")
			err = s.InitWX(&loginMap)
			panicErr(err)

			Log(logging.INFO, "初始化完毕,通知微信服务器登陆状态变更...")
			err = s.NotifyStatus(&loginMap)
			panicErr(err)

			Log(logging.INFO, "通知完毕.")
			Log(logging.DEBUG, e.SKey, loginMap.BaseRequest.SKey)
			Log(logging.DEBUG, e.PassTicket, loginMap.PassTicket)
			break
		} else if status == 201 {
			Log(logging.INFO, "请在手机上确认")
		} else if status == 408 {
			Log(logging.INFO, "请扫描二维码")
		} else {
			Log(logging.ERROR, msg)
		}
	}

	if runtime.GOARCH == "mipsle" {
		go Command(`echo "CLS(0);\r\n"> /dev/ttyS0`)
	}
	cmd.Process.Kill() //关闭显示的二维码图（Win下未生效）
	Log(logging.INFO, "开始获取联系人信息...")
	contactMap, err = s.GetAllContact(&loginMap)
	panicErr(err)
	Log(logging.INFO, fmt.Sprintf("成功获取 %d个 联系人信息,开始整理群组信息...", len(contactMap)))

	Log(logging.INFO, "开始监听消息响应...")
	var retcode, selector int64
	regAt := regexp.MustCompile(`^@.*@.*(易云辉|ease).*$`) // 群聊时其他人说话时会在前面加上@XXX
	regGroup := regexp.MustCompile(`^@@.+`)
	regAd := regexp.MustCompile(`(朋友圈|点赞)+`)

	for {

		//控制台命令
		select {
		case MeChatString := <-chatString:
			if strings.ToUpper(MeChatString) == "U" {
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
			} else if strings.ToUpper(MeChatString) == "QUIT" {
				Log(logging.INFO, "系统退出.")
				close(chatString)
				os.Exit(1)
			} else if strings.HasPrefix(MeChatString, "@") { //指定发送人
				if MeChatString[5:6] == ":" {
					for i, _ := range contactMap {
						if strings.HasPrefix(i[:5], MeChatString[0:5]) {
							MeToUser = i
							EchoMessage = Chat(&loginMap, 1, loginMap.SelfUserName, MeToUser, MeChatString[6:])
							break
						}
					}
				}
			} else {
				if MeToUser != "" {
					_ = Chat(&loginMap, 1, loginMap.SelfUserName, MeToUser, MeChatString)
					Log(logging.INFO, "我的消息：", MeChatString)
				} else {
					Log(logging.INFO, "不知道消息发向何处: ", MeChatString)
				}
			}
		default:
		}

		retcode, selector, err = s.SyncCheck(&loginMap)
		if err != nil {
			//log.Error(retcode, selector)
			printErr(err)
			if retcode == 1101 {
				Log(logging.INFO, "帐号已在其他地方登陆，程序将退出。")
				os.Exit(2)
			}
			continue
		}

		if retcode == 0 && selector != 0 {
			//fmt.Printf("selector=%d,有新消息产生,准备拉取...\n", selector)
			wxRecvMsges, err := s.WebWxSync(&loginMap)
			panicErr(err)

			for i := 0; i < wxRecvMsges.MsgCount; i++ {
				ToUserName = wxRecvMsges.MsgList[i].ToUserName
				FromUserName = wxRecvMsges.MsgList[i].FromUserName

				if regGroup.MatchString(wxRecvMsges.MsgList[i].FromUserName) { //是否在群中
					UserNickName, Message = getMessage(&loginMap, wxRecvMsges.MsgList[i].Content, wxRecvMsges.MsgList[i].FromUserName)
				} else {
					Message = wxRecvMsges.MsgList[i].Content
				}
				if wxRecvMsges.MsgList[i].MsgType == 1 { // 普通文本消息
					Log(logging.INFO, contactMap[wxRecvMsges.MsgList[i].FromUserName].NickName, ":", Message)

					if regGroup.MatchString(wxRecvMsges.MsgList[i].FromUserName) && regAt.MatchString(wxRecvMsges.MsgList[i].Content) {
						// 有人在群里@我，发个消息回答一下
						EchoMessage = Chat(&loginMap, 1, ToUserName, FromUserName, appConfig.AutoReplay_PrivateChat)
					} else if !regGroup.MatchString(wxRecvMsges.MsgList[i].FromUserName) { //不在群里
						if regAd.MatchString(wxRecvMsges.MsgList[i].Content) {
							// 有人私聊我，并且内容含有「朋友圈」、「点赞」等敏感词，则回复
							EchoMessage = Chat(&loginMap, 1, ToUserName, FromUserName, appConfig.AutoReplay_KeyFilter)
						} else if strings.EqualFold(wxRecvMsges.MsgList[i].Content, appConfig.AutoReplay_FunKey) {
							// 有人私聊我，开启调戏功能
							if AdminID == "" {
								EchoMessage = fmt.Sprintf(appConfig.AutoReplay_PrivateFunction, "进入")
								AdminID = wxRecvMsges.MsgList[i].FromUserName //设置为调戏开启用户
							} else {
								EchoMessage = fmt.Sprintf(appConfig.AutoReplay_PrivateFunction, "退出")
								AdminID = ""
							}

							EchoMessage = Chat(&loginMap, 1, ToUserName, FromUserName, EchoMessage)
						} else if AdminID == wxRecvMsges.MsgList[i].FromUserName { //已进入调戏模式
							EchoMessage = Chat(&loginMap, 1, ToUserName, FromUserName, AI(wxRecvMsges.MsgList[i].Content))
						}

					}
				} else if wxRecvMsges.MsgList[i].MsgType == 3 {
					picXML := m.PicInfo{}
					err = xml.Unmarshal([]byte(Html2Txt(Message)), &picXML)
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", UserNickName, ": 发了一张图片 ", PicUrl(wxRecvMsges.MsgList[i].MsgId))
				} else if wxRecvMsges.MsgList[i].MsgType == 34 {
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", strings.Split(wxRecvMsges.MsgList[i].Content, ":<br/>")[0]+": 发了一个语音信息")
				} else if wxRecvMsges.MsgList[i].MsgType == 43 {
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", strings.Split(wxRecvMsges.MsgList[i].Content, ":<br/>")[0]+": 发了一个视频")
				} else if wxRecvMsges.MsgList[i].MsgType == 47 {
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", strings.Split(wxRecvMsges.MsgList[i].Content, ":<br/>")[0]+": 发了一个发情")
				} else if wxRecvMsges.MsgList[i].MsgType == 49 {
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", strings.Replace(wxRecvMsges.MsgList[i].Content, ":<br/>", ": ", -1), " 发了一条普通链接或应用分享消息")
				} else if wxRecvMsges.MsgList[i].MsgType == 51 {
					if strings.HasPrefix(ToUserName, "@@") {
						Log(logging.INFO, contactMap[FromUserName].NickName, ": 客户端进入微信群", s.GetUserName(&loginMap, ToUserName))
					} else {
						Log(logging.INFO, contactMap[FromUserName].NickName, ": 客户端进入微信", s.GetUserName(&loginMap, ToUserName))
					}
				} else if wxRecvMsges.MsgList[i].MsgType == 10000 { //系统信息
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", strings.Replace(wxRecvMsges.MsgList[i].Content, ":<br/>", ": ", -1))
				} else if wxRecvMsges.MsgList[i].MsgType == 10002 {
					Log(logging.INFO, contactMap[FromUserName].NickName, ": ", strings.Replace(wxRecvMsges.MsgList[i].Content, ":<br/>", ": ", -1), " 撤回了一条消息")
				} else {
					Log(logging.INFO, fmt.Sprintf("%d", wxRecvMsges.MsgList[i].MsgType), contactMap[FromUserName].NickName+":", wxRecvMsges.MsgList[i].Content)
				}

				if EchoMessage != "" { //显示回复信息
					Log(logging.INFO, "我的回复：", EchoMessage)
					EchoMessage = ""
				}
			}
		}
	}
}

// 获取图片
func PicUrl(msgid string) string {
	return "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetmsgimg?&type=slave&MsgID=" + msgid
}

//转换网页代码
func Html2Txt(content string) string {
	c := strings.Replace(content, "<br/>", "\n", -1)
	c = strings.Replace(c, "&lt;", "<", -1)
	c = strings.Replace(c, "&gt;", ">", -1)
	c = strings.Replace(c, "&amp;", "&", -1)
	c = strings.Replace(c, "&nbsp;", " ", -1)
	c = strings.Replace(c, "&quot;", "\"", -1)
	return c
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

// 聊天
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

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func printErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func getChat() {
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

func FilterName(name string) []byte {
	var nameRegexp = regexp.MustCompile("\\<[\\S\\s]+?\\>")
	return nameRegexp.ReplaceAll([]byte(name), []byte(""))
}

func Command(cmd string) {
	c := exec.Command("ash", "-c", cmd)
	c.Start()
}

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
	if runtime.GOARCH == "mipsle" {
		var enc mahonia.Encoder
		enc = mahonia.NewEncoder("gbk")
		if ret, ok := enc.ConvertStringOK(strings.Join(args, "")); ok {
			go Command(`echo "BOXF(0,0,220,40,0);BS16(1,1,220,3,'` + ret + `',15);\r\n"> /dev/ttyS0`)
		}
	}

}
