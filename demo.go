/*
功能：微信个人服务程序
开发：Ease
时间：2017-9-3
修改：2017-9-23
*/

package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
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
	log          = logging.MustGetLogger("example")
	chatString   chan string
)

type AppConfig struct {
	AutoReplay_PrivateChat     string //私聊自动回复
	AutoReplay_KeyFilter       string //屏蔽关键词
	AutoReplay_FunKey          string //调戏功能关键词
	AutoReplay_PrivateFunction string //私人功能定义
}

type InfoRet struct { //AI返回信息解析
	Result  int    `json:"result"`
	Content string `json:"content"`
}

// 读取配置文件
func init() {
	// 日志输出及格式

	var format = logging.MustStringFormatter(`%{color}%{time:2006-01-02 15:04:05} [%{level:.4s}] %{id:03x}%{color:reset} %{message}`)
	file, err := os.OpenFile("./chat.log", os.O_APPEND, 0666)
	if err != nil {
		printErr(err)
	}
	defer file.Close()
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(file, "", 0)
	backendFormatter1 := logging.NewBackendFormatter(backend1, format)
	backendFormatter2 := logging.NewBackendFormatter(backend2, format)
	logging.SetBackend(backendFormatter1, backendFormatter2)

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

	// 从微信服务器获取UUID
	uuid, err = s.GetUUIDFromWX()
	if err != nil {
		panicErr(err)
	}

	// 根据UUID获取二维码
	err = s.DownloadImagIntoDir(e.QRCODE_URL+uuid, ".")
	panicErr(err)
	if runtime.GOOS == "linux" {
		cmd = exec.Command("eog", "./qrcode.jpg")
	} else {
		cmd = exec.Command(`cmd`, `/c start ./qrcode.jpg`)
	}
	cmd.Start()

	panicErr(err)

	// 轮询服务器判断二维码是否扫过暨是否登陆了
	for {
		log.Info("正在验证登陆... ...")
		status, msg := s.CheckLogin(uuid)

		if status == 200 {
			log.Info("登陆成功,处理登陆信息...")
			loginMap, err = s.ProcessLoginInfo(msg)
			if err != nil {
				panicErr(err)
			}

			log.Info("登陆信息处理完毕,正在初始化微信....")
			err = s.InitWX(&loginMap)
			if err != nil {
				panicErr(err)
			}

			log.Info("初始化完毕,通知微信服务器登陆状态变更...")
			err = s.NotifyStatus(&loginMap)
			if err != nil {
				panicErr(err)
			}

			log.Debug("通知完毕.")
			log.Debug(e.SKey + ": " + loginMap.BaseRequest.SKey)
			log.Debug(e.PassTicket + ": " + loginMap.PassTicket)
			break
		} else if status == 201 {
			log.Notice("请在手机上确认")
		} else if status == 408 {
			log.Notice("请扫描二维码")
		} else {
			log.Notice(msg)
		}
	}

	cmd.Process.Kill() //关闭显示的二维码图（Win下未生效）

	log.Info("开始获取联系人信息...")
	contactMap, err = s.GetAllContact(&loginMap)
	if err != nil {
		panicErr(err)
	}
	log.Info(fmt.Sprintf("成功获取 %d个 联系人信息,开始整理群组信息...", len(contactMap)))

	log.Info("开始监听消息响应...")
	var retcode, selector int64
	regAt := regexp.MustCompile(`^@.*@.*(易云辉|ease).*$`) // 群聊时其他人说话时会在前面加上@XXX
	regGroup := regexp.MustCompile(`^@@.+`)
	regAd := regexp.MustCompile(`(朋友圈|点赞)+`)

	for {

		//控制台命令
		select {
		case MeChatString := <-chatString:
			if strings.ToUpper(MeChatString) == "U" {
				log.Info("列出用户.")
				for i, n := range contactMap {
					if strings.HasPrefix(i, "@@") == true { //先列出群
						log.Info(fmt.Sprintf("%s %-30s", i[:5], FilterName(n.NickName)))
					}
				}
				for i, n := range contactMap {
					if strings.HasPrefix(i, "@@") == false && contactMap[i].VerifyFlag != 24 { //再列出用户,公众号不显示
						log.Info(fmt.Sprintf("%s %-30s %s %s %s", i[:5], FilterName(n.NickName), iif(n.Sex == 1, "男", "女"), n.Province, n.City))
					}
				}
			} else if strings.ToUpper(MeChatString) == "QUIT" {
				log.Info("系统退出.")
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
					log.Info("我的消息：", MeChatString)
				} else {
					log.Info("不知道消息发向何处: ", MeChatString)
				}
			}
		default:
		}

		retcode, selector, err = s.SyncCheck(&loginMap)
		if err != nil {
			//log.Error(retcode, selector)
			printErr(err)
			if retcode == 1101 {
				log.Error("帐号已在其他地方登陆，程序将退出。")
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
					log.Info(contactMap[wxRecvMsges.MsgList[i].FromUserName].NickName, ":", Message)

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
					log.Info(contactMap[FromUserName].NickName, ": ", UserNickName, ": 发了一张图片 ", PicUrl(wxRecvMsges.MsgList[i].MsgId))
				} else if wxRecvMsges.MsgList[i].MsgType == 34 {
					log.Info(contactMap[FromUserName].NickName, ": ", strings.Split(wxRecvMsges.MsgList[i].Content, ":<br/>")[0]+": 发了一个语音信息")
				} else if wxRecvMsges.MsgList[i].MsgType == 43 {
					log.Info(contactMap[FromUserName].NickName, ": ", strings.Split(wxRecvMsges.MsgList[i].Content, ":<br/>")[0]+": 发了一个视频")
				} else if wxRecvMsges.MsgList[i].MsgType == 47 {
					log.Info(contactMap[FromUserName].NickName, ": ", strings.Split(wxRecvMsges.MsgList[i].Content, ":<br/>")[0]+": 发了一个发情")
				} else if wxRecvMsges.MsgList[i].MsgType == 49 {
					log.Info(contactMap[FromUserName].NickName, ": ", strings.Replace(wxRecvMsges.MsgList[i].Content, ":<br/>", ": ", -1), " 发了一条普通链接或应用分享消息")
				} else if wxRecvMsges.MsgList[i].MsgType == 51 {
					if strings.HasPrefix(ToUserName, "@@") {
						log.Info(contactMap[FromUserName].NickName, ": 客户端进入微信群", s.GetUserName(&loginMap, ToUserName))
					} else {
						log.Info(contactMap[FromUserName].NickName, ": 客户端进入微信", s.GetUserName(&loginMap, ToUserName))
					}
				} else if wxRecvMsges.MsgList[i].MsgType == 10000 { //系统信息
					log.Info(contactMap[FromUserName].NickName, ": ", strings.Replace(wxRecvMsges.MsgList[i].Content, ":<br/>", ": ", -1))
				} else if wxRecvMsges.MsgList[i].MsgType == 10002 {
					log.Info(contactMap[FromUserName].NickName, ": ", strings.Replace(wxRecvMsges.MsgList[i].Content, ":<br/>", ": ", -1), " 撤回了一条消息")
				} else {
					log.Info(wxRecvMsges.MsgList[i].MsgType, contactMap[FromUserName].NickName+":", wxRecvMsges.MsgList[i].Content)
				}

				if EchoMessage != "" { //显示回复信息
					log.Info("我的回复：", EchoMessage)
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
		log.Error(err)
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
