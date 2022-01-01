![license](https://img.shields.io/github/license/dtm-labs/dtm)
![Build Status](https://github.com/dtm-labs/dtm/actions/workflows/tests.yml/badge.svg?branch=main)
[![codecov](https://codecov.io/gh/dtm-labs/dtm/branch/main/graph/badge.svg?token=UKKEYQLP3F)](https://codecov.io/gh/dtm-labs/dtm)
[![Go Report Card](https://goreportcard.com/badge/github.com/dtm-labs/dtm)](https://goreportcard.com/report/github.com/dtm-labs/dtm)
[![Go Reference](https://pkg.go.dev/badge/github.com/dtm-labs/dtm.svg)](https://pkg.go.dev/github.com/dtm-labs/dtm)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go#database)

# [English Docs](https://en.dtm.pub)
# 跨语言分布式事务管理器

DTM是一款golang开发的分布式事务管理器，解决了跨数据库、跨服务、跨语言栈更新数据的一致性问题。

他优雅的解决了幂等、空补偿、悬挂等分布式事务难题，提供了简单易用、高性能、易水平扩展的解决方案。

## 谁在使用DTM(仅列出部分)
[Tencent 腾讯](https://dtm.pub/other/using.html#tencent)

[Ivydad 常青藤爸爸](https://dtm.pub/other/using.html#ivydad)

[Eglass 视咖镜小二](https://dtm.pub/other/using.html)

[极欧科技](https://dtm.pub/other/using.html)

## 亮点

* 极易接入
  - 零配置启动服务，提供非常简单的HTTP接口，极大降低上手分布式事务的难度，新手也能快速接入
* 跨语言
  - 可适合多语言栈的公司使用。方便go、python、php、nodejs、ruby、c# 各类语言使用。
* 使用简单
  - 开发者不再担心悬挂、空补偿、幂等各类问题，首创子事务屏障技术代为处理
* 易部署、易扩展
  - 依赖mysql|redis，部署简单，易集群化，易水平扩展
* 多种分布式事务协议支持
  - TCC、SAGA、XA、二阶段消息，一站式解决所有分布式事务问题

## 与其他框架对比

非Java语言类的，暂未看到除dtm之外的成熟框架，因此这里将DTM和Java中最成熟的Seata对比：

|  特性| DTM | SEATA |备注|
|:-----:|:----:|:----:|:----:|
| [支持语言](https://dtm.pub/other/opensource.html#lang) |<span style="color:green">Go、Java、python、php、c#...</span>|<span style="color:orange">Java</span>|dtm可轻松接入一门新语言|
|[异常处理](https://dtm.pub/other/opensource.html#exception)| <span style="color:green"> 子事务屏障自动处理 </span>|<span style="color:orange">手动处理</span> |dtm解决了幂等、悬挂、空补偿|
| [TCC事务](https://dtm.pub/other/opensource.html#tcc)| <span style="color:green">✓</span>|<span style="color:green">✓</span>||
| [XA事务](https://dtm.pub/other/opensource.html#xa)|<span style="color:green">✓</span>|<span style="color:green">✓</span>||
|[AT事务](https://dtm.pub/other/opensource.html#at)|<span style="color:orange">建议使用XA</span>|<span style="color:green">✓</span>|AT与XA类似，性能更好，但有脏回滚|
| [SAGA事务](https://dtm.pub/other/opensource.html#saga) |<span style="color:green">支持并发</span> |<span style="color:green">状态机模式</span> ||
|[二阶段消息](https://dtm.pub/other/opensource.html#msg)|<span style="color:green">✓</span>|<span style="color:red">✗</span>|dtm提供类似rocketmq的事务消息|
|[单服务多数据源](https://dtm.pub/other/opensource.html#multidb)|<span style="color:green">✓</span>|<span style="color:red">✗</span>||
|[通信协议](https://dtm.pub/other/opensource.html#protocol)|HTTP、gRPC、go-zero|dubbo等协议|dtm对云原生更加友好|
|[star数量](https://dtm.pub/other/opensource.html#star)|<img src="https://img.shields.io/github/stars/dtm-labs/dtm.svg?style=social" alt="github stars"/>|<img src="https://img.shields.io/github/stars/seata/seata.svg?style=social" alt="github stars"/>|dtm从20210604发布0.1，发展快|

从上面对比的特性来看，如果您的语言栈包含了Java之外的语言，那么dtm是您的首选。如果您的语言栈是Java，您也可以选择接入dtm，使用子事务屏障技术，简化您的业务编写。

详细的对比可以点击特性中的链接，跳到相关文档
## [性能测试报告](https://dtm.pub/other/performance.html)

## [教程与文档](https://dtm.pub)

## [各语言客户端及示例](https://dtm.pub/summary/code.html#go)

## 微服务框架支持
- [go-zero](https://github.com/zeromicro/go-zero)：一开源就非常火爆的微服务框架，首家接入dtm的微服务框架。感谢go-zero作者[kevwan](https://github.com/kevwan)的大力支持
- [polaris](https://github.com/polarismesh/polaris): 腾讯开源的注册发现组件，以及在其上构建的微服务框架。感谢腾讯同学[ychensha](https://github.com/ychensha)的PR
- 其他：看用户需求量，择机接入，参见[微服务支持](https://dtm.pub/protocol/intro.html)

## 快速开始

如果您不是Go语言，可以跳转[各语言客户端及示例](https://dtm.pub/summary/code.html#go)，里面有相关的快速开始示例

### 运行dtm

``` bash
git clone https://github.com/dtm-labs/dtm && cd dtm
go run main.go
```

### 启动并运行一个saga示例
下面运行一个类似跨行转账的示例，包括两个事务分支：资金转出（TransOut)、资金转入（TransIn)。DTM保证TransIn和TransOut要么全部成功，要么全部回滚，保证最终金额的正确性。

`go run qs/main.go`

## 接入详解

### 接入代码
``` GO
  // 具体业务微服务地址
  const qsBusi = "http://localhost:8081/api/busi_saga"
  req := &gin.H{"amount": 30} // 微服务的载荷
  // DtmServer为DTM服务的地址，是一个url
  DtmServer := "http://localhost:36789/api/dtmsvr"
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

### 失败情况
在实际的业务中，子事务可能出现失败，例如转入的子账号被冻结导致转账失败。我们对业务代码进行修改，让TransIn的正向操作失败，然后看看结果

``` go
	app.POST(qsBusiAPI+"/TransIn", common.WrapHandler(func(c *gin.Context) (interface{}, error) {
		return M{"dtm_result": "FAILURE"}, nil
	}))
```

再运行这个例子，整个事务最终失败，时序图如下：

![saga_rollback](https://pic3.zhimg.com/80/v2-8d8f1476be8a1e2e09ce97a89b4116c2_1440w.jpg)

在转入操作失败的情况下，TransIn和TransOut的补偿操作被执行，保证了最终的余额和转账前是一样的。

### 更多示例
参考[dtm-labs/dtm-examples](https://github.com/dtm-labs/dtm-examples)

## 联系我们
### 公众号
dtm官方公众号：分布式事务，大量干货分享，以及dtm的最新消息
### 交流群
请加 yedf2008 好友或者扫码加好友，验证回复 dtm 按照指引进群

![yedf2008](http://service.ivydad.com/cover/dubbingb6b5e2c0-2d2a-cd59-f7c5-c6b90aceb6f1.jpeg)

### github
作者github: [https://github.com/yedf2](https://github.com/yedf2)

欢迎使用[dtm](https://github.com/dtm-labs/dtm)，或者通过dtm学习实践分布式事务相关知识，欢迎star支持我们

