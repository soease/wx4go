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
	BaseRet  BaseResponse     `json:"BaseResponse"`
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
	VerifyFlag int8   `json:"VerifyFlag"`
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

type GroupList struct { //组列表，用于查询组信息
	UserName        string `json:"UserName"`
	EncryChatRoomId string `json:"EncryChatRoomId"`
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

//用户信息
type Contact struct {
	Uin              int                  `json:"Uin"`
	UserName         string               `json:"UserName"`
	NickName         string               `json:"NickName"`
	HeadImgUrl       string               `json:"HeadImgUrl"`
	ContactFlag      int                  `json:"ContactFlag"`
	MemberCount      int                  `json:"MemberCount"`
	MemberList       []Contact_MemberList `json:"MemberList"`
	RemarkName       string               `json:"RemarkName"`
	HideInputBarFlag int                  `json:"HideInputBarFlag"`
	Sex              int                  `json:"Sex"`
	Signature        string               `json:"Signature"`
	VerifyFlag       int                  `json:"VerifyFlag"`
	OwnerUin         int                  `json:"OwnerUin"`
	PYInitial        string               `json:"PYInitial"`
	PYQuanPin        string               `json:"PYQuanPin"`
	RemarkPYInitial  string               `json:"RemarkPYInitial"`
	RemarkPYQuanPin  string               `json:"RemarkPYQuanPin"`
	StarFriend       int                  `json:"StarFriend"`
	AppAccountFlag   int                  `json:"AppAccountFlag"`
	Statues          int                  `json:"Statues"`
	AttrStatus       int64                `json:"AttrStatus"`
	Province         string               `json:"Province"`
	City             string               `json:"City"`
	Alias            string               `json:"Alias"`
	SnsFlag          int                  `json:"SnsFlag"`
	UniFriend        int                  `json:"UniFriend"`
	DisplayName      string               `json:"DisplayName"`
	ChatRoomId       int                  `json:"ChatRoomId"`
	KeyWord          string               `json:"KeyWord"`
	EncryChatRoomId  string               `json:"EncryChatRoomId"`
	IsOwner          int                  `json:"IsOwner"`
}

//当获取群组信息时，会附加群内的用户信息
type Contact_MemberList struct {
	Uin             int    `json:"Uin"`
	UserName        string `json:"UserName"`
	NickName        string `json:"NickName"`
	AttrStatus      int    `json:"AttrStatus"`
	PYInitial       string `json:"PYInitial"`
	PYQuanPin       string `json:"PYQuanPin"`
	RemarkPYInitial string `json:"RemarkPYInitial"`
	RemarkPYQuanPin string `json:"RemarkPYQuanPin"`
	MemberStatus    int    `json:"MemberStatus"`
	DisplayName     string `json:"DisplayName"`
	KeyWord         string `json:"KeyWord"`
}

type BaseResponse struct {
	Ret     int    `json:"Ret"`
	Message string `json:"ErrMsg"`
}

//图文信息
/*
 <msg>
     <appmsg appid="" sdkver="0">
             <title><![CDATA[惊呆了！动动手指，就能免费领走百兆流量及超值大礼！]]></title>
			 <des><![CDATA[流量任性送，大礼随性兑！]]></des>
			 <action></action>
			 <type>5</type>
			 <showtype>1</showtype>
			 <content><![CDATA[]]></content>
			 <contentattr>0</contentattr>
             <url><![CDATA[http://mp.weixin.qq.com/s?__biz=MjM5MjUzODk0MA==&amp;mid=2653382533&amp;idx=1&amp;sn=b573add6e281c01c15f40a8f7d338718&amp;chksm=bd770fd68a0086c03ab113f81c830c9825e2cc0025caaa6517112da9787b1e529e07714a9b80&amp;scene=0#rd]]></url>
             <lowurl><![CDATA[]]></lowurl>
             <appattach>
                <totallen>0</totallen>
                <attachid></attachid>
                <fileext></fileext>
             </appattach>
             <extinfo></extinfo>
             <mmreader>
                 <category type="20" count="5">
                      <name><![CDATA[建设银行四川省分行]]></name>
                      <topnew>
                          <cover><![CDATA[http://mmbiz.qpic.cn/mmbiz_jpg/Tz4w2jickxxALcWmE0n5fwiazibUgbPsZBOCrYIYibtosMnicNw5MuCb5G4OM9ibzJPSH3VmSJ4ZfEVrnWvCYaMbYIgg/640?wxtype=jpeg&amp;wxfrom=0]]></cover>
                          <width>0</width>
                          <height>0</height>
                          <digest><![CDATA[]]></digest>
                      </topnew>
                      <item>
                          <itemshowtype>0</itemshowtype>
                          <title><![CDATA[惊呆了！动动手指，就能免费领走百兆流量及超值大礼！]]></title>
                          <url><![CDATA[http://mp.weixin.qq.com/s?__biz=MjM5MjUzODk0MA==&amp;mid=2653382533&amp;idx=1&amp;sn=b573add6e281c01c15f40a8f7d338718&amp;chksm=bd770fd68a0086c03ab113f81c830c9825e2cc0025caaa6517112da9787b1e529e07714a9b80&amp;scene=0#rd]]></url>
                          <shorturl><![CDATA[]]></shorturl>
                          <longurl><![CDATA[]]></longurl>
                          <pub_time>1505883109</pub_time>
                          <cover><![CDATA[http://mmbiz.qpic.cn/mmbiz_jpg/Tz4w2jickxxALcWmE0n5fwiazibUgbPsZBOCrYIYibtosMnicNw5MuCb5G4OM9ibzJPSH3VmSJ4ZfEVrnWvCYaMbYIgg/640?wxtype=jpeg&amp;wxfrom=0]]></cover>
                          <tweetid></tweetid>
                          <digest><![CDATA[流量任性送，大礼随性兑！]]></digest>
                          <fileid>505898884</fileid>
                          <sources>
                              <source>
                                   <name><![CDATA[建设银行四川省分行]]></name>
                              </source>
                          </sources>
                          <styles></styles>
                          <native_url></native_url>
						  <del_flag>0</del_flag>
						  <contentattr>0</contentattr>
						  <play_length>0</play_length>
                      </item>
                      <item>
                          <itemshowtype>0</itemshowtype>
                          <title><![CDATA[信用卡 约定账户自动扣款，一切就这么简单]]></title>
                          <url><![CDATA[http://mp.weixin.qq.com/s?__biz=MjM5MjUzODk0MA==&amp;mid=2653382533&amp;idx=2&amp;sn=3c2060cf622b5c45854b6ff311b044dc&amp;chksm=bd770fd68a0086c0ff91271d2621c52eae00cf24090c16f252f5eb2603847f2c3c66e8b7b12b&amp;scene=0#rd]]></url>
                          <shorturl><![CDATA[]]></shorturl>
                          <longurl><![CDATA[]]></longurl>
                          <pub_time>1505883109</pub_time>
                          <cover><![CDATA[http://mmbiz.qpic.cn/mmbiz_jpg/Tz4w2jickxxALcWmE0n5fwiazibUgbPsZBOeRDTKjhVYg31OMP6TFGrOMFsmRGQnZzxLFtL6ibjwL5JktbfYK1j4hg/300?wxtype=jpeg&amp;wxfrom=0]]></cover>
                          <tweetid></tweetid>
                          <digest><![CDATA[约定账户还款，就是这么便捷！]]></digest>
                          <fileid>505898865</fileid>
                          <sources>
                             <source>
                                <name><![CDATA[建设银行四川省分行]]></name>
                             </source>
                          </sources>
                          <styles></styles>
                          <native_url></native_url>
                          <del_flag>0</del_flag>
                          <contentattr>0</contentattr>
                          <play_length>0</play_length>
                      </item>
                      <item>
                          <itemshowtype>0</itemshowtype>
                          <title><![CDATA[【金普月】长知识——原来你是这样的人民币]]></title>
                          <url><![CDATA[http://mp.weixin.qq.com/s?__biz=MjM5MjUzODk0MA==&amp;mid=2653382533&amp;idx=3&amp;sn=647861eb51d8b0dfcb4c2b88ef3dc478&amp;chksm=bd770fd68a0086c0080bbbdf245d387d4a733228dd6db40f80b4fb91b77ccd988e957999c2ba&amp;scene=0#rd]]></url>
                          <shorturl><![CDATA[]]></shorturl>
                          <longurl><![CDATA[]]></longurl>
                          <pub_time>1505883109</pub_time>
                          <cover><![CDATA[http://mmbiz.qpic.cn/mmbiz_jpg/Tz4w2jickxxALcWmE0n5fwiazibUgbPsZBOPTaEPuPVeRCDdrC5ib2ClJt9xXia8gRCnrbV2f5Mbk8RLOXQgibRV0C4Q/300?wxtype=jpeg&amp;wxfrom=0]]></cover>
                          <tweetid></tweetid>
                          <digest><![CDATA[用心了解一下我吧，我的名字叫人民币。]]></digest>
                          <fileid>505898868</fileid>
                          <sources>
                              <source>
                                  <name><![CDATA[建设银行四川省分行]]></name>
                              </source>
                          </sources>
                          <styles></styles>
                          <native_url></native_url>
                          <del_flag>0</del_flag>
                          <contentattr>0</contentattr>
                          <play_length>0</play_length>
                      </item>
                 </category>
                 <publisher>
                    <username><![CDATA[gh_ad399acfadb9]]></username>
                    <nickname><![CDATA[建设银行四川省分行]]></nickname>
                 </publisher>
                 <template_header></template_header>
                 <template_detail></template_detail>
                 <forbid_forward>0</forbid_forward>
             </mmreader>
             <thumburl><![CDATA[http://mmbiz.qpic.cn/mmbiz_jpg/Tz4w2jickxxALcWmE0n5fwiazibUgbPsZBOCrYIYibtosMnicNw5MuCb5G4OM9ibzJPSH3VmSJ4ZfEVrnWvCYaMbYIgg/640?wxtype=jpeg&amp;wxfrom=0]]></thumburl>
     </appmsg>
     <fromusername><![CDATA[gh_ad399acfadb9]]></fromusername>
     <appinfo>
         <version></version>
         <appname><![CDATA[建设银行四川省分行]]></appname>
         <isforceupdate>1</isforceupdate>
     </appinfo>
 </msg>
*/

//图文信息
type PicTxtInfo struct {
	XMLName      xml.Name           `xml:"msg"` /* 根节点定义 */
	AppMsg       PicTxtInfo_AppMsg  `xml:"appmsg"`
	FromUserName string             `xml:"fromusername"`
	AppInfo      PicTxtInfo_AppInfo `xml:"appinfo"`
}

type PicTxtInfo_AppMsg struct {
	Title     string              `xml:"title"`
	Desc      string              `xml:"des"`
	Action    string              `xml:"action"`
	Type      int                 `xml:"type"`
	ShowType  int                 `xml:"showtype"`
	Content   string              `xml:"content"`
	Url       string              `xml:"url"`
	LowUrl    string              `xml:"lowurl"`
	AppAttach string              `xml:"appattach"`
	ExtInfo   string              `xml:"extinfo"`
	MMReader  PicTxtInfo_MMReader `xml:"mmreader"`
}

type PicTxtInfo_MMReader struct {
	Category       string               `xml:"category"` //这个还没完成
	Publisher      PicTxtInfo_Publisher `xml:"publisher"`
	TemplateHeader string               `xml:"template_header"`
	TemplateDetail string               `xml:"template_detail"`
	ForbidForward  int                  `xml:"forbid_forward"`
}

type PicTxtInfo_Publisher struct {
	UserName string `xml:"username"`
	NickName string `xml:"nickname"`
}

type PicTxtInfo_AppInfo struct {
	Version     string `xml:"version"`
	AppName     string `xml:"appname"`
	ForceUpdate int    `xml:"isforceupdate"`
}

type PicTxtInfo_Item struct {
}

//视频信息
/*
<?xml version="1.0"?>
<msg>
     <videomsg aeskey="f08b0a7edbba4dae9efc9f5c80a60223"
               cdnthumbaeskey="f08b0a7edbba4dae9efc9f5c80a60223"
               cdnvideourl="30500201000449304702010002042ac5cbcf02032dcd010204390a96b6020459c1f30c0422373236323639303035304063686174726f6f6d323731315f313530353838323839320204010800040201000400"
               cdnthumburl="30500201000449304702010002042ac5cbcf02032dcd010204390a96b6020459c1f30c0422373236323639303035304063686174726f6f6d323731315f313530353838323839320204010800040201000400"
               length="1065951"
               playlength="10"
               cdnthumblength="9507"
               cdnthumbwidth="176"
               cdnthumbheight="320"
               fromusername="wxid_x6ny9w9bvgo422"
               md5="9e703a36878eb9a8f4c949c03227376d"
               newmd5="07ffa2560fec17da55706d4ecdb32688"
               isad="0" />
</msg>
*/

type VideoInfo struct {
}

//表情信息
/*
<msg>
  <emoji fromusername="gain_8884495"
         tousername="1030168696@chatroom"
         type="1"
         idbuffer="media:0_0"
         md5="d4f697ebb9e00d22390faa922efd0924"
         len="18384"
         productid=""
         androidmd5="d4f697ebb9e00d22390faa922efd0924"
         androidlen="18384"
         s60v3md5="d4f697ebb9e00d22390faa922efd0924"
         s60v3len="18384"
         s60v5md5="d4f697ebb9e00d22390faa922efd0924"
         s60v5len="18384"
         cdnurl="http://emoji.qpic.cn/wx_emoji/g1MVz8aKUGHMNXuXJ5ylBbBGAMib7abvJjjMMmnQicukHc29m88N7R2w/"
         designerid=""
         thumburl=""
         encrypturl="http://emoji.qpic.cn/wx_emoji/JPeZw4ufGRibMA6QF23JIa0NRpn8WYljhKwo1fu51eYLLyNmOokRZAA/"
         aeskey="a68cee7ec1ec28a7f36d4bdbf1c4e4ac"
         externurl="http://emoji.qpic.cn/wx_emoji/fQYHtIzNCmQM66lkuQU172xtrk6mZFANCjb6ic61mL483ftXwbZuNOtWcTGVYLh5p/"
         externmd5="38abe9cbe59fb7b8257247de3067ef8d"
         width="200"
         height="70">
  </emoji>
</msg>
*/

type EmojiInfo struct {
	XMLName xml.Name        `xml:"msg"` /* 根节点定义 */
	Emoji   EmojiInfo_Emoji `xml:"emoji"`
}

type EmojiInfo_Emoji struct {
	Url string `xml:"cdnurl,attr"`
}

//图片信息
/*
<msg>
    <img aeskey="550c6c62c07f4efc8cc27559f8310f0c"
    encryver="0"
    cdnthumbaeskey="550c6c62c07f4efc8cc27559f8310f0c"
    cdnthumburl="305002010004493047020100020413621d0b02030f488102044bb78cb6020459c20d790425617570696d675f303564343639326234303065613639345f313530353838393635353937330201000201000400"
    cdnthumblength="16994"
    cdnthumbheight="120"
    cdnthumbwidth="120"
    cdnmidheight="0"
    cdnmidwidth="0"
    cdnhdheight="0"
    cdnhdwidth="0"
    cdnmidimgurl="305002010004493047020100020413621d0b02030f488102044bb78cb6020459c20d790425617570696d675f303564343639326234303065613639345f313530353838393635353937330201000201000400"
    length="80973"
    cdnbigimgurl="305002010004493047020100020413621d0b02030f488102044bb78cb6020459c20d790425617570696d675f303564343639326234303065613639345f313530353838393635353937330201000201000400"
    hdlength="81223"
    md5="fd3cbc3f0706073a124fcbe45cf74321" />
</msg>
*/

type PicInfo struct {
	XMLName xml.Name    `xml:"msg"` /* 根节点定义 */
	Image   PicInfo_Img `xml:"img"`
}

type PicInfo_Img struct {
	Url string `xml:"cdnthumburl,attr"`
	Key string `xml:"cdnthumbaeskey,attr"`
}

//链接分享信息
/*
<msg>
   <appmsg appid="" sdkver="0">
     <title>九月，再见；十月，你好！</title>
     <des>诗词天地，倡导诗意生活态度。</des>
     <action></action>
     <type>5</type>
     <showtype>0</showtype>
     <soundtype>0</soundtype>
     <mediatagname></mediatagname>
     <messageext></messageext>
     <messageaction></messageaction>
     <content></content>
     <contentattr>0</contentattr>
     <url>http://mp.weixin.qq.com/s?__biz=MzI2NDI3ODk0OQ==&amp;amp;mid=2247499056&amp;amp;idx=1&amp;amp;sn=fc565800b9c44797d07cca03d718cf76&amp;amp;chksm=eaadbbfaddda32ece08af9acb79640cbd5745881b25e4e6f22f54d4ba0d3a269dcc672cde877&amp;amp;mpshare=1&amp;amp;scene=1&amp;amp;srcid=0930e7NeLvI1ohyYTmRSnlgO#rd</url>
     <lowurl></lowurl>
     <dataurl></dataurl>
     <lowdataurl></lowdataurl>
     <appattach>
       <totallen>0</totallen>
       <attachid></attachid>
       <emoticonmd5></emoticonmd5>
       <fileext></fileext>
       <cdnthumburl>304c02010004453043020100020428a112bf02030f48810204a0b28cb6020459cee81e0421777869645f3435726c3830763571736d6f32323733335f313530363733323035330201000201000400</cdnthumburl>
       <cdnthumbmd5>6d57e7c7bd5bf8f2de9256799af8e8d9</cdnthumbmd5>
       <cdnthumblength>7052</cdnthumblength>
       <cdnthumbwidth>160</cdnthumbwidth>
       <cdnthumbheight>160</cdnthumbheight>
       <cdnthumbaeskey>62363135353164326432666634643764</cdnthumbaeskey>
       <aeskey>62363135353164326432666634643764</aeskey>
       <encryver>0</encryver>
     </appattach>
     <extinfo></extinfo>
     <sourceusername>gh_99766370ece6</sourceusername>
     <sourcedisplayname>诗词天地</sourcedisplayname>
     <thumburl>http://mmbiz.qpic.cn/mmbiz_jpg/r5ibYXKmn8p1qaoR9Wmj52bjic2KlF8rvCJ1Au9mesTYUGhEGkDhYmAOlGGj41B18QePRLndBMCbokGYYdlAJoibw/300?wx_fmt=jpeg&amp;amp;wxfrom=1</thumburl>
     <md5></md5>
     <statextstr></statextstr>
   </appmsg>

   <fromusername>wxid_ugbdc6082hjx11</fromusername>
   <scene>0</scene>
   <appinfo>
     <version>1</version>
     <appname></appname>
   </appinfo>
   <commenturl></commenturl>
 </msg>
*/

type LinkInfo struct {
	XMLName  xml.Name     `xml:"msg"` /* 根节点定义 */
	Msg      LinkInfo_msg `xml:"appmsg"`
	FromUser string       `xml:"fromusername"`
}

type LinkInfo_msg struct {
	Title    string `xml:"title"`
	Desc     string `xml:"des"`
	InfoType int    `xml:"type"`
	Url      string `xml:"url"`
}
