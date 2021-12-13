![license](https://img.shields.io/github/license/yedf/dtm)
[![Build Status](https://travis-ci.com/yedf/dtm.svg?branch=main)](https://travis-ci.com/yedf/dtm)
[![Coverage Status](https://coveralls.io/repos/github/yedf/dtm/badge.svg?branch=main)](https://coveralls.io/github/yedf/dtm?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/yedf/dtm)](https://goreportcard.com/report/github.com/yedf/dtm)
[![Go Reference](https://pkg.go.dev/badge/github.com/yedf/dtm.svg)](https://pkg.go.dev/github.com/yedf/dtm)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go#database)

# [English Docs](https://en.dtm.pub)
# 跨语言分布式事务管理器

DTM是一款golang开发的分布式事务管理器，解决了跨数据库、跨服务、跨语言栈更新数据的一致性问题。

他优雅的解决了幂等、空补偿、悬挂等分布式事务难题，提供了简单易用、高性能、易水平扩展的解决方案。

作者受邀参加中国数据库大会分享[多语言环境下分布式事务实践](http://dtcc.it168.com/yicheng.html#b9)

## 谁在使用dtm
[Tencent 腾讯](https://www.tencent.com/)

[Ivydad 常青藤爸爸](https://ivydad.com)

[Eglass 视咖镜小二](https://epeijing.cn)

[极欧科技](http://jiou.me)

[金数智联]()

## 亮点

* 极易接入
  - 支持HTTP，提供非常简单的接口，极大降低上手分布式事务的难度，新手也能快速接入
* 使用简单
  - 开发者不再担心悬挂、空补偿、幂等各类问题，框架层代为处理
* 跨语言
  - 可适合多语言栈的公司使用。方便go、python、php、nodejs、ruby、c# 各类语言使用。
* 易部署、易扩展
  - 仅依赖mysql，部署简单，易集群化，易水平扩展
* 多种分布式事务协议支持
  - TCC、SAGA、XA、事务消息

## 与其他框架对比

目前开源的分布式事务框架，Java的框架较多，有大厂开源的SEATA、ServiceComb-Pack，shardingsphere，以及个人开源的himly，tcc-transaction，ByteTCC等等，其中以Seata的应用最为广泛。

非Java语言类的，暂未看到除dtm之外的成熟框架，因此这里仅将DTM和Java中最成熟的Seata对比：

|  特性| DTM | SEATA |备注|
|:-----:|:----:|:----:|:----:|
| 支持语言 |<span style="color:green">Go、Java、python、php、c#...</span>|<span style="color:orange">Java</span>|dtm可轻松接入一门新语言|
|异常处理| <span style="color:green"> [子事务屏障自动处理](https://zhuanlan.zhihu.com/p/388444465) </span>|<span style="color:orange">手动处理</span> |dtm解决了幂等、悬挂、空补偿|
| TCC事务| <span style="color:green">✓</span>|<span style="color:green">✓</span>||
| XA事务|<span style="color:green">✓</span>|<span style="color:green">✓</span>||
|AT事务|<span style="color:orange">建议使用XA</span>|<span style="color:green">✓</span>|AT与XA类似，性能更好，但有脏回滚|
| SAGA事务 |<span style="color:green">支持并发</span> |<span style="color:green">状态机模式</span> ||
|事务消息|<span style="color:green">✓</span>|<span style="color:red">✗</span>|dtm提供类似rocketmq的事务消息|
|单服务多数据源|<span style="color:green">✓</span>|<span style="color:red">✗</span>||
|通信协议|HTTP、gRPC、go-zero|dubbo等协议|dtm对云原生更加友好|
|star数量|<img src="https://img.shields.io/github/stars/yedf/dtm.svg?style=social" alt="github stars"/>|<img src="https://img.shields.io/github/stars/seata/seata.svg?style=social" alt="github stars"/>|dtm从20210604发布0.1，发展快|

从上面对比的特性来看，如果您的语言栈包含了Java之外的语言，那么dtm是您的首选。如果您的语言栈是Java，您也可以选择接入dtm，使用子事务屏障技术，简化您的业务编写。

## [性能测试报告](https://dtm.pub/other/performance.html)

## [教程与文档](https://dtm.pub)

## [各语言客户端及示例](https://dtm.pub/summary/code.html#go)

## 微服务框架支持
- [go-zero](https://github.com/zeromicro/go-zero)：一开源就非常火爆的微服务框架，首家接入dtm的微服务框架。感谢go-zero作者[kevwan](https://github.com/kevwan)的大力支持
- [polaris](https://github.com/polarismesh/polaris): 腾讯开源的注册发现组件，以及在其上构建的微服务框架。感谢腾讯同学[ychensha](https://github.com/ychensha)的PR
- 其他：看用户需求量，择机接入

具体微服务接入使用，参见[微服务支持](https://dtm.pub/protocol/intro.html)
## 快速开始

### 获取代码

`git clone https://github.com/yedf/dtm && cd dtm`

### dtm依赖于mysql

安装[docker 20.04+](https://docs.docker.com/get-docker/)之后

`docker-compose -f helper/compose.mysql.yml up`

> 您也可以配置使用现有的mysql，需要高级权限，允许dtm创建数据库
>
> `cp conf.sample.yml conf.yml # 修改conf.yml`

### 启动并运行saga示例
`go run app/main.go qs`

## 开始使用

### 使用
``` GO
  // 具体业务微服务地址
  const qsBusi = "http://localhost:8081/api/busi_saga"
  req := &gin.H{"amount": 30} // 微服务的载荷
  // DtmServer为DTM服务的地址，是一个url
  DtmServer := "http://localhost:8080/api/dtmsvr"
  saga := dtmcli.NewSaga(DtmServer, dtmcli.MustGenGid(DtmServer)).
    // 添加一个TransOut的子事务，正向操作为url: qsBusi+"/TransOut"， 补偿操作为url: qsBusi+"/TransOutCompensate"
    Add(qsBusi+"/TransOut", qsBusi+"/TransOutCompensate", req).
    // 添加一个TransIn的子事务，正向操作为url: qsBusi+"/TransIn"， 补偿操作为url: qsBusi+"/TransInCompensate"
    Add(qsBusi+"/TransIn", qsBusi+"/TransInCompensate", req)
  // 提交saga事务，dtm会完成所有的子事务/回滚所有的子事务
  err := saga.Submit()
```

成功运行后，可以看到TransOut、TransIn依次被调用，完成了整个分布式事务

### 时序图

上述saga分布式事务的时序图如下：

<img src="https://pic3.zhimg.com/80/v2-b7d98659093c399e182a0173a8e549ca_1440w.jpg" height=428 />

### 完整示例
参考[examples/quick_start.go](./examples/quick_start.go)

## 公众号
您可以关注公众号：分布式事务，及时跟踪dtm的最新内容
## 交流群
请加 yedf2008 好友或者扫码加好友，验证回复 dtm 按照指引进群

![yedf2008](http://service.ivydad.com/cover/dubbingb6b5e2c0-2d2a-cd59-f7c5-c6b90aceb6f1.jpeg)

欢迎使用dtm，或者通过dtm学习实践分布式事务相关知识，欢迎star支持我们

## 谁在使用
<div style='vertical-align: middle'>
    <img alt='腾讯' height='80'  src='https://dtm.pub/assets/tencent.4b87bfd8.jpeg'  /img>
    <img alt='常青藤爸爸' height='80'  src='https://dtm.pub/assets/ivydad.d0f58a94.png'  /img>
    <img alt='镜小二' height='80'  src='https://img.epeijing.cn/official-website/assets/logo.png'  /img>
    <img alt='极欧科技' height='80'  src='https://dtm.pub/assets/jiou.5bed10c2.png'  /img>
    <img alt='金数智联' height='80'  src='https://dtm.pub/assets/gdci.214d305a.png'  /img>
</div>
