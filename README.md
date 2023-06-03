# easyio
a simple netpoll implemented using go

### 说明
目前Go圈有很多款异步的网络框架:
- https://github.com/tidwall/evio
- https://github.com/lesismal/nbio
- https://github.com/panjf2000/gnet
- https://github.com/cloudwego/netpoll
- .......

排名不分先后。

这里面最早的实现是evio。evio也存在一些问题，之前也写过[evio](https://www.syst.top/posts/go/evio/)文章介绍过。
其他比如nbio和gnet也写过一些源码分析。


为什么会出现这些框架？之前也提到过，由于标准库netpoll的一些特性:
- 一个conn一个goroutine导致利用率低
- 用户无法感知conn状态
- .....

这些框架在应用层上做了很多优化，比如:Worker Pool,Buffer,Ring Buffer,NoCopy......。


**那为什么还会有这个库?**

借鉴(模仿)上面框架的实现，用最少的代码实现一个最小化的Non-blocking IO库，然后写一个0到1实现easyio的小课程， 帮助小白理解一些原理。
然后在这个基础上去扩展(多平台)，去优化(阅读上面框架的代码，参考别人是咋么设计的)来达到学习的效果。

最后，让这个世界充满爱～～
