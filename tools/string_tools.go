package tools

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

type InfoRet struct { //AI返回信息解析
	Result  int    `json:"result"`
	Content string `json:"content"`
}

/**
 * 有序(或者无序)地从一个map中按照index的顺序构造URL中的params
 * 加上有序的目的是为了防止有些环境下需要params根据key的ASC大小排序后进行签名加密
 */
func GetURLParams(values ...interface{}) string {
	var result = "?"
	if len(values) == 1 {
		maap := values[0].(map[string]string)
		for key, value := range maap {
			if key != "" && value != "" {
				result += fmt.Sprintf("%s=%s&", key, url.QueryEscape(value))
			}
		}
	} else if len(values) == 2 {
		index := values[1].([]string)
		maap := values[0].(map[string]string)
		for _, key := range index {
			if key != "" && maap[key] != "" {
				result += fmt.Sprintf("%s=%s&", key, url.QueryEscape(maap[key]))
			}
		}
	}

	return result[:len(result)-1]
}

/**
 *  生成随机字符串
 *  index：取随机序列的前index个
 *  0-9:10
 *  0-9a-z:10+24
 *  0-9a-zA-Z:10+24+24
 *  length：需要生成随机字符串的长度
 */
func GetRandomString(index int, length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(index)])
	}
	return string(result)
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

//Linux Shell
func Command(cmd string) {
	c := exec.Command("ash", "-c", cmd)
	c.Start()
}

func Iif(sour bool, ret1 string, ret2 string) string {
	if sour {
		return ret1
	} else {
		return ret2
	}
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

//成都限行信息
func ChengDuCar() (ret string) {
	url := "https://www.baidu.com/s?wd=%E6%88%90%E9%83%BD%E9%99%90%E8%A1%8C"
	client := http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0")
	response, err := client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode == 200 {
		doc, err := goquery.NewDocumentFromResponse(response)
		if err != nil {
			return
		}
		doc.Find(".c-border .op_traffic_time .op_traffic_left").Each(func(i int, s *goquery.Selection) {
			ret = "成都" + strings.Replace(s.Find(".op_traffic_title").Text(), "限行", s.Find(".op_traffic_off").Text(), -1)
		})
	}
	return
}
