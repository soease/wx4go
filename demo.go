/*
功能：微信个人服务程序
开发：Ease
时间：2017-9-3
修改：2017-9-19
*/

package main

import (
	"encoding/json"
	"fmt"
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
	uuid         string
	err          error
	loginMap     m.LoginMap
	contactMap   map[string]m.User
	groupMap     map[string][]m.User // 关键字为key的，群组数组
	appConfig    AppConfig           //系统配置
	AdminID      string              //进入管理员帐号
	UserInfoList map[string]string   //把用户信息存起来
)

type AppConfig struct {
	AutoReplay_PrivateChat     string //私聊自动回复
	AutoReplay_KeyFilter       string //屏蔽关键词
	AutoReplay_FunKey          string //调戏功能关键词
	AutoReplay_PrivateFunction string //私人功能定义
}

type InfoRet struct {
	Result  int    `json:"result"`
	Content string `json:"content"`
}

// 读取配置文件
func init() {
	conf := goini.SetConfig("./app.conf")
	appConfig.AutoReplay_PrivateChat = conf.GetValue("chat", "PrivateChat")
	appConfig.AutoReplay_KeyFilter = conf.GetValue("chat", "KeyFilter")
	appConfig.AutoReplay_FunKey = conf.GetValue("chat", "FunKey")
	appConfig.AutoReplay_PrivateFunction = conf.GetValue("chat", "PrivateFunction")
	UserInfoList = make(map[string]string)
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

func main() {
	var cmd *exec.Cmd
	var Message string
	var EchoMessage string

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
		fmt.Println("正在验证登陆... ...")
		status, msg := s.CheckLogin(uuid)

		if status == 200 {
			fmt.Println("登陆成功,处理登陆信息...")
			loginMap, err = s.ProcessLoginInfo(msg)
			if err != nil {
				panicErr(err)
			}

			fmt.Println("登陆信息处理完毕,正在初始化微信....")
			err = s.InitWX(&loginMap)
			if err != nil {
				panicErr(err)
			}

			fmt.Println("初始化完毕,通知微信服务器登陆状态变更...")
			err = s.NotifyStatus(&loginMap)
			if err != nil {
				panicErr(err)
			}

			fmt.Println("通知完毕,本次登陆信息：")
			fmt.Println(e.SKey + ": " + loginMap.BaseRequest.SKey)
			fmt.Println(e.PassTicket + ": " + loginMap.PassTicket)
			break
		} else if status == 201 {
			fmt.Println("请在手机上确认")
		} else if status == 408 {
			fmt.Println("请扫描二维码")
		} else {
			fmt.Println(msg)
		}
	}

	cmd.Process.Kill() //关闭显示的二维码图（Win下未生效）

	fmt.Println("开始获取联系人信息...")
	contactMap, err = s.GetAllContact(&loginMap)
	if err != nil {
		panicErr(err)
	}
	fmt.Printf("成功获取 %d个 联系人信息,开始整理群组信息...\n", len(contactMap))

	//fmt.Println(contactMap)
	groupMap = s.MapGroupInfo(contactMap)
	groupSize := 0
	for _, v := range groupMap {
		groupSize += len(v)
	}
	fmt.Printf("整理完毕，共有 %d个 群组是焦点群组，它们是：\n", groupSize)
	for key, v := range groupMap {
		fmt.Println(key)
		for _, user := range v {
			fmt.Println("========>" + user.NickName)
		}
	}

	fmt.Println("开始监听消息响应...")
	var retcode, selector int64
	regAt := regexp.MustCompile(`^@.*@.*(易云辉|ease).*$`) // 群聊时其他人说话时会在前面加上@XXX
	regGroup := regexp.MustCompile(`^@@.+`)
	regAd := regexp.MustCompile(`(朋友圈|点赞)+`)

	for {
		retcode, selector, err = s.SyncCheck(&loginMap)
		if err != nil {
			fmt.Println(retcode, selector)
			printErr(err)
			if retcode == 1101 {
				fmt.Println("帐号已在其他地方登陆，程序将退出。")
				os.Exit(2)
			}
			continue
		}

		if retcode == 0 && selector != 0 {
			//fmt.Printf("selector=%d,有新消息产生,准备拉取...\n", selector)
			wxRecvMsges, err := s.WebWxSync(&loginMap)
			panicErr(err)

			for i := 0; i < wxRecvMsges.MsgCount; i++ {
				if wxRecvMsges.MsgList[i].MsgType == 1 { // 普通文本消息
					if regGroup.MatchString(wxRecvMsges.MsgList[i].FromUserName) {
						Message = getMessage(wxRecvMsges.MsgList[i].Content, wxRecvMsges.MsgList[i].FromUserName)
					} else {
						Message = wxRecvMsges.MsgList[i].Content
					}
					fmt.Println(contactMap[wxRecvMsges.MsgList[i].FromUserName].NickName, ":", Message)
					if regGroup.MatchString(wxRecvMsges.MsgList[i].FromUserName) && regAt.MatchString(wxRecvMsges.MsgList[i].Content) {
						// 有人在群里@我，发个消息回答一下
						wxSendMsg := m.WxSendMsg{}
						wxSendMsg.Type = 1
						wxSendMsg.Content = appConfig.AutoReplay_PrivateChat
						wxSendMsg.FromUserName = wxRecvMsges.MsgList[i].ToUserName
						wxSendMsg.ToUserName = wxRecvMsges.MsgList[i].FromUserName
						wxSendMsg.LocalID = fmt.Sprintf("%d", time.Now().Unix())
						wxSendMsg.ClientMsgId = wxSendMsg.LocalID

						time.Sleep(time.Second) // 加点延时，避免消息次序混乱，同时避免微信侦察到机器人
						go s.SendMsg(&loginMap, wxSendMsg)
						EchoMessage = wxSendMsg.Content
					} else if !regGroup.MatchString(wxRecvMsges.MsgList[i].FromUserName) { //不在群里
						if regAd.MatchString(wxRecvMsges.MsgList[i].Content) {
							// 有人私聊我，并且内容含有「朋友圈」、「点赞」等敏感词，则回复
							wxSendMsg := m.WxSendMsg{}
							wxSendMsg.Type = 1
							wxSendMsg.Content = appConfig.AutoReplay_KeyFilter
							wxSendMsg.FromUserName = wxRecvMsges.MsgList[i].ToUserName
							wxSendMsg.ToUserName = wxRecvMsges.MsgList[i].FromUserName
							wxSendMsg.LocalID = fmt.Sprintf("%d", time.Now().Unix())
							wxSendMsg.ClientMsgId = wxSendMsg.LocalID

							time.Sleep(time.Second)
							go s.SendMsg(&loginMap, wxSendMsg)
							EchoMessage = wxSendMsg.Content
						} else if strings.EqualFold(wxRecvMsges.MsgList[i].Content, appConfig.AutoReplay_FunKey) {
							// 有人私聊我，开启调戏功能
							wxSendMsg := m.WxSendMsg{}
							wxSendMsg.Type = 1
							wxSendMsg.FromUserName = wxRecvMsges.MsgList[i].ToUserName
							wxSendMsg.ToUserName = wxRecvMsges.MsgList[i].FromUserName
							wxSendMsg.LocalID = fmt.Sprintf("%d", time.Now().Unix())
							wxSendMsg.ClientMsgId = wxSendMsg.LocalID
							if AdminID == "" {
								wxSendMsg.Content = fmt.Sprintf(appConfig.AutoReplay_PrivateFunction, "进入")
								AdminID = wxRecvMsges.MsgList[i].FromUserName //设置为管理员
							} else {
								wxSendMsg.Content = fmt.Sprintf(appConfig.AutoReplay_PrivateFunction, "退出")
								AdminID = ""
							}

							time.Sleep(time.Second)
							go s.SendMsg(&loginMap, wxSendMsg)
							EchoMessage = wxSendMsg.Content
						} else if AdminID == wxRecvMsges.MsgList[i].FromUserName { //已进入管理员模式
							wxSendMsg := m.WxSendMsg{}
							wxSendMsg.Type = 1
							wxSendMsg.FromUserName = wxRecvMsges.MsgList[i].ToUserName
							wxSendMsg.ToUserName = wxRecvMsges.MsgList[i].FromUserName
							wxSendMsg.LocalID = fmt.Sprintf("%d", time.Now().Unix())
							wxSendMsg.ClientMsgId = wxSendMsg.LocalID
							wxSendMsg.Content = AI(wxRecvMsges.MsgList[i].Content)

							time.Sleep(time.Second)
							go s.SendMsg(&loginMap, wxSendMsg)
							EchoMessage = wxSendMsg.Content
						}

					}
				} else if wxRecvMsges.MsgList[i].MsgType == 3 {
					fmt.Println(contactMap[wxRecvMsges.MsgList[i].FromUserName].NickName, ": ", strings.Split(wxRecvMsges.MsgList[i].Content, ":<br/>")[0]+": 发了一张图片")
				} else if wxRecvMsges.MsgList[i].MsgType == 34 {
					fmt.Println(contactMap[wxRecvMsges.MsgList[i].FromUserName].NickName, ": ", strings.Split(wxRecvMsges.MsgList[i].Content, ":<br/>")[0]+": 发了一个语音信息")
				} else if wxRecvMsges.MsgList[i].MsgType == 49 {
					fmt.Println(contactMap[wxRecvMsges.MsgList[i].FromUserName].NickName, ": ", strings.Replace(wxRecvMsges.MsgList[i].Content, ":<br/>", ": ", -1), " 发了一条普通链接或应用分享消息")
				} else if wxRecvMsges.MsgList[i].MsgType == 51 {
					fmt.Println(contactMap[wxRecvMsges.MsgList[i].FromUserName].NickName, ": 主人进入微信群")
				} else if wxRecvMsges.MsgList[i].MsgType == 10000 {
					fmt.Println(contactMap[wxRecvMsges.MsgList[i].FromUserName].NickName, ": ", strings.Replace(wxRecvMsges.MsgList[i].Content, ":<br/>", ": ", -1), " 有人发红包吗？")
				} else if wxRecvMsges.MsgList[i].MsgType == 10002 {
					fmt.Println(contactMap[wxRecvMsges.MsgList[i].FromUserName].NickName, ": ", strings.Replace(wxRecvMsges.MsgList[i].Content, ":<br/>", ": ", -1), " 撤回了一条消息")
				} else {
					fmt.Println(wxRecvMsges.MsgList[i].MsgType, contactMap[wxRecvMsges.MsgList[i].FromUserName].NickName+":", wxRecvMsges.MsgList[i].Content)
				}

				if EchoMessage != "" { //显示回复信息
					fmt.Println("我的回复：", EchoMessage)
					EchoMessage = ""
				}
			}
		}
	}
}

//获取用户信息及聊天内容
func getMessage(content string, groupid string) (Message string) {
	var UserNickName string
	UserID := strings.Split(content, ":<br/>")[0]
	if _, ok := UserInfoList[UserID]; ok {
		UserNickName = UserInfoList[UserID]
	} else {
		UserNickName = s.GetGroupUserName(&loginMap, UserID, groupid)
		UserInfoList[UserID] = UserNickName
	}

	Message = strings.Replace(content, UserID, UserNickName, -1)
	Message = strings.Replace(Message, ":<br/>", ": ", -1)

	return
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func printErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
