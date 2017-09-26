这是一个利用网页版微信协议的应用。

原出处github.com/newflydd/itchat4go，作了一些扩充。

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
- 图片功能 github.com/Comdex/imgo

## 更新
- 2017.9.26 将引用包压缩到项目中
- 2017.9.26 支持运行在MT7688，显示二维码到串口屏
- 2017.9.23 支持对指定群或用户聊天（@xxxx:xxxx,指定对象后可直接输入聊天内容)
- 2017.9.20 记录日志
- 2017.9.19 显示我的回复内容
- 2017.9.19 增加AI功能（调戏AI）
- 2017.9.19 微信群中的网友昵称显示
- 2017.9.18 弹出二维码自动关闭（Windows中暂未生效）
- 2017.9.18 将部份迁移至配置文件
- 2017.9.17 允许Linux中使用
- 2017.9.17 除外文字聊天的判断（将进一步完善）

## 计划
- 2017.9.26 串口屏信息显示优化
- 2017.9.26 加快图片缩小处理
- 2017.9.26 串口屏配置
- 2017.9.23 日志不能写入文件的处理
- 2017.9.20 手机端进群时提示群名称
- 2017.9.20 群内私聊时，返回私聊信息
- 2017.9.20 其它协议的解析
- 2017.9.20 调戏功能针对每一个网友单独开启
- 2017.9.20 减少外部包的引用
- 2017.9.19 除文字聊天以外的功能
- 2017.9.19 部份微信扫码时出错不能使用，查找原因
- 2017.9.19 运行到MT7688中，并用串口屏显示出来(研究go如何读取jpg中.PS(x,y,color)可以在串口屏中画点)
- 2017.9.19 添加命令行聊天功能
- 2017.9.19 在Pi中运行，可以用于家电控制等
### 其它
- 首次学习开源，多鼓励，少放气，多建议。高手太多，请绕行。
- 闲暇之余持续更新
- 我的微信，可以探讨、学习，当然更欢迎打赏鼓励。

![](http://wyyyh.3322.org:88/static/upload/bigpic/20170919/1505787805515811601.jpg)
