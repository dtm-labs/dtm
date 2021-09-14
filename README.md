![license](https://img.shields.io/github/license/yedf/dtm)
[![Build Status](https://travis-ci.com/yedf/dtm.svg?branch=main)](https://travis-ci.com/yedf/dtm)
[![Coverage Status](https://coveralls.io/repos/github/yedf/dtm/badge.svg?branch=main)](https://coveralls.io/github/yedf/dtm?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/yedf/dtm)](https://goreportcard.com/report/github.com/yedf/dtm)
[![Go Reference](https://pkg.go.dev/badge/github.com/yedf/dtm.svg)](https://pkg.go.dev/github.com/yedf/dtm)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go#database)

# [English Docs](https://en.dtm.pub)
# GO语言分布式事务管理服务

## 谁在使用dtm

[Ivydad 常青藤爸爸](https://ivydad.com)

[Eglass 视咖镜小二](https://epeijing.cn)

[极欧科技](http://jiou.me)

[金数智联]()

## 可以做什么

* 一站式解决分布式事务需求
  - 支持TCC、SAGA、XA、事务消息、可靠消息
* 支持跨语言栈事务
  - 适合多语言栈的公司使用。支持Go、Python、PHP、Nodejs、Java、c# 各类语言SDK。
* 自动处理子事务乱序
  - 首创子事务屏障技术，框架层自动处理悬挂、空补偿、幂等网络异常问题
* 支持单服务多数据源访问
  - 通过子事务屏障技术，保证单服务多数据源访问的一致性，也能保证本地长事务拆分多个子事务后的一致性
* 开箱即用，提供云上服务
  - 支持通用的HTTP、gRPC协议接入，云原生友好。支持Mysql接入。提供测试版本的云上服务，方便新用户测试

## 与其他框架对比

目前开源的分布式事务框架，暂未看到非Java语言有成熟的框架。而Java语言的较多，其中以seata应用最为广泛。

下面是dtm和seata的主要特性对比：

|  特性| DTM | SEATA |备注|
|:-----:|:----:|:----:|:----:|
| 支持语言 |<span style="color:green">Go、Java、python、php、c#...</span>|<span style="color:orange">Java</span>|dtm可轻松接入一门新语言|
|异常处理| <span style="color:green">[子事务屏障自动处理](https://dtm.pub/exception/barrier.html)</span>|<span style="color:orange">手动处理</span> |dtm解决了幂等、悬挂、空补偿|
| TCC事务| <span style="color:green">✓</span>|<span style="color:green">✓</span>||
| XA事务|<span style="color:green">✓</span>|<span style="color:green">✓</span>||
|AT事务|<span style="color:red">✗</span>|<span style="color:green">✓</span>|AT与XA类似，性能更好，但有脏回滚|
| SAGA事务 |<span style="color:orange">简单模式</span> |<span style="color:green">状态机复杂模式</span> |dtm的状态机模式在规划中|
|事务消息|<span style="color:green">✓</span>|<span style="color:red">✗</span>|dtm提供类似rocketmq的事务消息|
|通信协议|HTTP、GRPC|dubbo等协议，无HTTP||
|star数量|<img src="https://img.shields.io/github/stars/yedf/dtm.svg?style=social" alt="github stars"/>|<img src="https://img.shields.io/github/stars/seata/seata.svg?style=social" alt="github stars"/>|dtm从20210604发布0.1，发展快|

从上面对比的特性来看，如果您的语言栈包含了Java之外的语言，那么dtm是您的首选。如果您的语言栈是Java，您也可以选择接入dtm，使用子事务屏障技术，简化您的业务编写。

## [教程与文档](https://dtm.pub)

## [各语言客户端及示例](https://dtm.pub/summary/code.html#go)

## 快速开始

### 安装

`git clone https://github.com/yedf/dtm`

### dtm依赖于mysql

配置mysql：

`cp conf.sample.yml conf.yml # 修改conf.yml`

### 启动并运行saga示例
`go run app/main.go saga`

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

## 交流群
请加 yedf2008 好友或者扫码加好友，验证回复 dtm 按照指引进群

![yedf2008](http://service.ivydad.com/cover/dubbingb6b5e2c0-2d2a-cd59-f7c5-c6b90aceb6f1.jpeg)

如果您觉得[dtm](https://github.com/yedf/dtm)不错，或者对您有帮助，请赏颗星吧！

## 谁在使用
<div style='vertical-align: middle'>
    <img alt='常青藤爸爸' height='40'  src='https://www.ivydad.com/_nuxt/img/header-logo.5b3eb96.png'  /img>
    <img alt='镜小二' height='40'  src='https://img.epeijing.cn/official-website/assets/logo.png'  /img>
    <img alt='极欧科技' height='40'  src='http://www.siqitech.com.cn/img/logo.3f6c2914.png'  /img>
    <img alt='金数智联' height='40'  src='https://pic1.zhimg.com/80/v2-dc1d0cef5f7b72be345fc34d768e69e3_1440w.png'  /img>
</div>
