![license](https://img.shields.io/github/license/dtm-labs/dtm)
![Build Status](https://github.com/dtm-labs/dtm/actions/workflows/tests.yml/badge.svg?branch=main)
[![codecov](https://codecov.io/gh/dtm-labs/dtm/branch/main/graph/badge.svg?token=UKKEYQLP3F)](https://codecov.io/gh/dtm-labs/dtm)
[![Go Report Card](https://goreportcard.com/badge/github.com/dtm-labs/dtm)](https://goreportcard.com/report/github.com/dtm-labs/dtm)
[![Go Reference](https://pkg.go.dev/badge/github.com/dtm-labs/dtm.svg)](https://pkg.go.dev/github.com/dtm-labs/dtm)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go#database)

简体中文 | [English](https://github.com/dtm-labs/dtm/blob/main/helper/README-en.md)

# 跨语言分布式事务管理器

DTM是一款变革性的分布式事务框架，提供了傻瓜式的使用方式，极大的降低了分布式事务的使用门槛，改变了“能不用分布式事务就不用”的行业现状，优雅的解决了服务间的数据一致性问题。

## 谁在使用DTM(仅列出部分)
[Tencent 腾讯](https://dtm.pub/other/using.html#tencent)

[Bytedance 字节](https://dtm.pub/other/using.html#bytedance)

[Ivydad 常青藤爸爸](https://dtm.pub/other/using.html#ivydad)

[更多](https://dtm.pub/other/using.html)

如果贵公司也已使用 dtm，欢迎在 [登记地址](https://github.com/dtm-labs/dtm/issues/7) 登记，仅仅为了推广，不做其它用途。

## 特性
* 支持多种语言：支持Go、Java、PHP、C#、Python、Nodejs 各种语言的SDK
* 支持多种事务模式：SAGA、TCC、XA
* 支持消息最终一致性：二阶段消息，比本地消息表更优雅的方案
* 未支持 AT 事务模式，建议使用XA，详情参见[XA vs AT](https://dtm.pub/practice/at)
* 支持多种数据库事务：Mysql、Redis、MongoDB、Postgres、TDSQL等
* 支持多种存储引擎：Mysql（常用）、Redis（高性能）、MongoDB（规划中）
* 支持多种微服务架构：[go-zero](https://github.com/zeromicro/go-zero)、go-kratos/kratos、polarismesh/polaris
* 支持高可用，易水平扩展

## 应用场景：
DTM 可以应用于大量的场景下的数据一致性问题，以下是几个常见场景
* [缓存管理](https://dtm.pub/app/cache.html)：彻底保证缓存最终一致及强一致
* [秒杀扣库存](https://dtm.pub/app/flash.html)：极端情况下，也能保证Redis中精准的库存，和最终创建的订单完全一致，无需手动调整
* [非单体的订单系统](https://dtm.pub/app/order.html)： 大幅简化架构
* [事件发布/订阅](https://dtm.pub/practice/msg.html)：更好的发件箱模式

## [性能测试报告](https://dtm.pub/other/performance.html)

## [教程与文档](https://dtm.pub)

## [各语言客户端及示例](https://dtm.pub/ref/sdk.html#go)

## 快速开始
如果您不是Go语言，可以跳转[各语言客户端及示例](https://dtm.pub/ref/sdk.html#go)，里面有相关的快速开始示例

喜欢视频教程的朋友，可以访问[分布式事务教程-快速开始](https://www.bilibili.com/video/BV1fS4y1h7Tj/)

### 运行dtm

``` bash
git clone https://github.com/dtm-labs/dtm && cd dtm
go run main.go
```

### 启动并运行一个saga示例
下面运行一个类似跨行转账的示例，包括两个事务分支：资金转出（TransOut)、资金转入（TransIn)。DTM保证TransIn和TransOut要么全部成功，要么全部回滚，保证最终金额的正确性。

``` bash
git clone https://github.com/dtm-labs/dtmcli-go-sample && cd dtmcli-go-sample
go run main.go
```

## 接入详解

### 接入代码
``` GO
  // 具体业务微服务地址
  const qsBusi = "http://localhost:8081/api/busi_saga"
  req := &gin.H{"amount": 30} // 微服务的载荷
  // DtmServer为DTM服务的地址，是一个url
  DtmServer := "http://localhost:36789/api/dtmsvr"
  saga := dtmcli.NewSaga(DtmServer, shortuuid.New()).
    // 添加一个TransOut的子事务，正向操作为url: qsBusi+"/TransOut"， 补偿操作为url: qsBusi+"/TransOutCom"
    Add(qsBusi+"/TransOut", qsBusi+"/TransOutCom", req).
    // 添加一个TransIn的子事务，正向操作为url: qsBusi+"/TransIn"， 补偿操作为url: qsBusi+"/TransInCom"
    Add(qsBusi+"/TransIn", qsBusi+"/TransInCom", req)
  // 提交saga事务，dtm会完成所有的子事务/回滚所有的子事务
  err := saga.Submit()
```

成功运行后，可以看到TransOut、TransIn依次被调用，完成了整个分布式事务

### 时序图

上述saga分布式事务的时序图如下：

<img src="https://pic3.zhimg.com/80/v2-b7d98659093c399e182a0173a8e549ca_1440w.jpg" height=428 />

### 失败情况
在实际的业务中，子事务可能出现失败，例如转入的子账号被冻结导致转账失败。我们对业务代码进行修改，让TransIn的正向操作失败，然后看看结果

``` go
	app.POST(qsBusiAPI+"/TransIn", func(c *gin.Context) {
		logger.Infof("TransIn")
		c.JSON(409, "") // Status 409 表示失败，不再重试，直接回滚
	})
```

再运行这个例子，整个事务最终失败，时序图如下：

<img src="https://pic3.zhimg.com/80/v2-8d8f1476be8a1e2e09ce97a89b4116c2_1440w.jpg"  height=528 />

在转入操作失败的情况下，TransIn和TransOut的补偿操作被执行，保证了最终的余额和转账前是一样的。

### 更多示例
上述示例主要演示了分布式事务的流程，更多的内容，包括如何与实际的数据库对接，如何做补偿，如何做回滚等实际的例子，请参考[dtm-labs/dtm-examples](https://github.com/dtm-labs/dtm-examples)

## 联系我们
### 微信交流群

如果您希望更快的获得反馈，或者更多的了解其他用户在使用过程中的各种反馈，欢迎加入我们的微信交流群

请加作者的微信 yedf2008 好友或者扫码加好友，备注 `dtm` 按照指引进群

![yedf2008](http://service.ivydad.com/cover/dubbingb6b5e2c0-2d2a-cd59-f7c5-c6b90aceb6f1.jpeg)

欢迎使用[dtm](https://github.com/dtm-labs/dtm)，或者通过dtm学习实践分布式事务相关知识，欢迎star支持我们

