package model

import (
	"encoding/xml"
	"fmt"
)

/*
 * <error>
 *   <ret>0</ret>
 *   <message></message>
 *   <skey>@crypt_3aaab8d5_aa9febb1c57122a4569c1b1dc4772eac</skey>
 *   <wxsid>vjqCszEkQQw9jep1</wxsid>
 *   <wxuin>154158775</wxuin>
 *   <pass_ticket>wbFO7Vqg%2BpADuIcrQPDM1e0KjmNvgsH8jYAEoT0FtSY%3D</pass_ticket>
 *   <isgrayscale>1</isgrayscale>
 * </error>
 */
type LoginCallbackXMLResult struct {
	XMLName     xml.Name `xml:"error"` /* 根节点定义 */
	Ret         string   `xml:"ret"`
	Message     string   `xml:"message"`
	SKey        string   `xml:"skey"`
	WXSid       string   `xml:"wxsid"`
	WXUin       string   `xml:"wxuin"`
	PassTicket  string   `xml:"pass_ticket"`
	IsGrayscale string   `xml:"isgrayscale"`
}

type BaseRequest struct {
	Uin      string `json:"Uin"`
	Sid      string `json:"Sid"`
	SKey     string `json:"Skey"`
	DeviceID string `json:"DeviceID"`
}

/* 微信初始化时返回的大JSON，选择性地提取一些关键数据 */
type InitInfo struct {
	User     User             `json:"User"`
	SyncKeys SyncKeysJsonData `json:"SyncKey"`
}

/* 微信获取所有联系人列表时返回的大JSON */
type ContactList struct {
	MemberCount int    `json:"MemberCount"`
	MemberList  []User `json:"MemberList"`
}

/* 微信通用User结构，可根据需要扩展 */
type User struct {
	Uin        int64  `json:"Uin"`
	UserName   string `json:"UserName"`
	NickName   string `json:"NickName"`
	RemarkName string `json:"RemarkName"`
	Sex        int8   `json:"Sex"`
	Province   string `json:"Province"`
	City       string `json:"City"`
}

type SyncKeysJsonData struct {
	Count    int       `json:"Count"`
	SyncKeys []SyncKey `json:"List"`
}

type SyncKey struct {
	Key int64 `json:"Key"`
	Val int64 `json:"Val"`
}

type RequestUserList struct {
	List []UserList `json:"List"`
}

type UserList struct {
	UserName string `json:"UserName"`
	RoomId   string `json:"EncryChatRoomId"`
}

/* 设计一个构造成字符串的结构体方法 */
func (sks SyncKeysJsonData) ToString() string {
	resultStr := ""

	for i := 0; i < sks.Count; i++ {
		resultStr = resultStr + fmt.Sprintf("%d_%d|", sks.SyncKeys[i].Key, sks.SyncKeys[i].Val)
	}

	return resultStr[:len(resultStr)-1]
}

/* 微信消息对象 */
type WxRecvMsges struct {
	MsgCount int              `json:"AddMsgCount"`
	MsgList  []WxRecvMsg      `json:"AddMsgList"`
	SyncKeys SyncKeysJsonData `json:"SyncKey"`
}

/* 微信接受消息对象元素 */
type WxRecvMsg struct {
	MsgId        string `json:"MsgId"`
	FromUserName string `json:"FromUserName"`
	ToUserName   string `json:"ToUserName"`
	MsgType      int    `json:"MsgType"`
	Content      string `json:"Content"`
	CreateTime   int64  `json:"CreateTime"`
}

/**
 * "Type":1,
 * "Content":"1",
 * "FromUserName":"@9499e6e8dfd2c1020ecb6cc727982bef",
 * "ToUserName":"@9499e6e8dfd2c1020ecb6cc727982bef",
 * "LocalID":"15046739462870976",
 * "ClientMsgId":"15046739462870976"
 * 微信发送消息对象元素
 */
type WxSendMsg struct {
	Type         int    `json:"Type"`
	Content      string `json:"Content"`
	FromUserName string `json:"FromUserName"`
	ToUserName   string `json:"ToUserName"`
	LocalID      string `json:"LocalID"`
	ClientMsgId  string `json:"ClientMsgId"`
}

type UserInfo struct {
	BaseRet     BaseResponse `json:"BaseResponse"`
	Count       int          `json:"Count"`
	ContactList []Contact    `json:"ContactList"`
}

type Contact struct {
	Uin              int      `json:"Uin"`
	UserName         string   `json:"UserName"`
	NickName         string   `json:"NickName"`
	HeadImgUrl       string   `json:"HeadImgUrl"`
	ContactFlag      int      `json:"ContactFlag"`
	MemberCount      int      `json:"MemberCount"`
	MemberList       []string `json:"MemberList"`
	RemarkName       string   `json:"RemarkName"`
	HideInputBarFlag int      `json:"HideInputBarFlag"`
	Sex              int      `json:"Sex"`
	Signature        string   `json:"Signature"`
	VerifyFlag       int      `json:"VerifyFlag"`
	OwnerUin         int      `json:"OwnerUin"`
	PYInitial        string   `json:"PYInitial"`
	PYQuanPin        string   `json:"PYQuanPin"`
	RemarkPYInitial  string   `json:"RemarkPYInitial"`
	RemarkPYQuanPin  string   `json:"RemarkPYQuanPin"`
	StarFriend       int      `json:"StarFriend"`
	AppAccountFlag   int      `json:"AppAccountFlag"`
	Statues          int      `json:"Statues"`
	AttrStatus       int64    `json:"AttrStatus"`
	Province         string   `json:"Province"`
	City             string   `json:"City"`
	Alias            string   `json:"Alias"`
	SnsFlag          int      `json:"SnsFlag"`
	UniFriend        int      `json:"UniFriend"`
	DisplayName      string   `json:"DisplayName"`
	ChatRoomId       int      `json:"ChatRoomId"`
	KeyWord          string   `json:"KeyWord"`
	EncryChatRoomId  string   `json:"EncryChatRoomId"`
	IsOwner          int      `json:"IsOwner"`
}

type BaseResponse struct {
	Ret     int    `json:"Ret"`
	Message string `json:"ErrMsg"`
}
