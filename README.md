这是一个利用网页版微信协议的应用。

原出处github.com/newflydd/itchat4go，作了一些扩充。

这是一个Python版 github.com/littlecodersh/ItChat

这里有Python代码和协议分析 github.com/Urinx/WeixinBot

代码不够漂亮，先运行，后整理。

运行：go run demo.go

MT7688+串口屏上的效果：

![](https://github.com/soease/wx4go/blob/master/other/MT7688.jpg)

---

## 功能

## 引用
- 日志功能 github.com/op/go-logging
- 配置文件 github.com/widuu/goini
- UTF8转GBK code.google.com/p/mahonia
- 图片缩放 github.com/nfnt/resize
- 网页分析 github.com/PuerkitoBio/goquery (用于AI等)

## 更新
- 2017.9.30 分享链接信息的处理
- 2017.9.30 完善配置文件
- 2017.9.29 代码整理.将二维码文件放入临时文件夹中
- 2017.9.29 处理Ctrl+C等系统中断，并善后处理
- 2017.9.26 将引用包压缩到项目中
- 2017.9.26 支持运行在MT7688，显示二维码到串口屏
- 2017.9.23 支持对指定群或用户聊天（@xxxx:xxxx,指定对象后可直接输入聊天内容)
- 2017.9.20 记录日志
- 2017.9.19 显示我的回复内容
- 2017.9.19 增加AI功能（调戏AI）
- 2017.9.19 微信群中的网友昵称显示
- 2017.9.18 弹出二维码自动关闭（Windows中暂未生效）
- 2017.9.17 让程序符合Go的风格--全系统使用
- 2017.9.17 除外文字聊天的判断（将进一步完善）

## 计划
- 2017.9.30 不扫码，直接在手机上确认
- 2017.9.30 自动通过添加好友
- 2017.9.29 记事及提醒功能
- 2017.9.26 串口屏信息显示优化
- 2017.9.20 群内私聊时，返回私聊信息
- 2017.9.20 其它协议的解析
- 2017.9.20 减少外部包的引用
- 2017.9.19 除文字聊天以外的功能
- 2017.9.19 部份微信扫码时出错不能使用，查找原因
- 2017.9.19 在Pi中运行，可以用于家电控制等
- 
## 其它
- 首次学习开源，多鼓励，少放气，多建议。高手太多，请绕行。
- 闲暇之余持续更新
- 我的微信，可以探讨、学习，当然更欢迎打赏鼓励。

![](http://wyyyh.3322.org:88/static/upload/bigpic/20170919/1505787805515811601.jpg)

```
    //在命令行显示二维码，需要把字符界面缩小
    "github.com/Comdex/imgo"
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

    // 也可以用下面这种，速度更快。有变形，但二维码还是能扫描。
    var r uint32
    src, err := LoadImage(QRFile)
    panicErr(err)
    // 缩略图的大小
    dst := resize.Resize(430, 430, src, resize.Lanczos2)
    for i := 0; i < 430; i++ {
        for n := 0; n < 430; n++ {
            r, _, _, _ = dst.At(i, n).RGBA()

            if r > 50000 {
                fmt.Print("█")
            } else {
                fmt.Print(" ")
            }
        }
        fmt.Print("\n")
    }
    cmd = exec.Command("echo")

```                

- 将imgo库换为原装库,虽然看起来二维码有些变化,但好在没有影响扫码.而速度在MT7688上减少了六十多秒的图片缩小处理时间.
