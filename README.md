![license](https://img.shields.io/github/license/yedf/dtm)
[![Build Status](https://travis-ci.com/yedf/dtm.svg?branch=main)](https://travis-ci.com/yedf/dtm)
[![Coverage Status](https://coveralls.io/repos/github/yedf/dtm/badge.svg?branch=main)](https://coveralls.io/github/yedf/dtm?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/yedf/dtm)](https://goreportcard.com/report/github.com/yedf/dtm)
[![Go Reference](https://pkg.go.dev/badge/github.com/yedf/dtm.svg)](https://pkg.go.dev/github.com/yedf/dtm)

[中文文档](https://dtm.pub)

# A lightweight distributed transaction manager

DTM is the first golang open source distributed transaction project. It elegantly resolve the problem of execution out of order. In the microservice architecture, a high-performance and easy-to-use distributed transaction solution is provided.

## Characteristic

* Stable and reliable
  + Tested in the production environment, unit test coverage is over 90%
* Simple to use
  + The interface is simple, developers no longer worry about suspension, null compensation, idempotence, etc., and the framework layer takes care of them.
* Multi language supported
  + It is suitable for companies with multi-language stacks. The protocol supports http. Convenient to use in various languages ​​such as go, python, php, nodejs, ruby.
* Easy to deploy, easy to expand
  + Only rely on mysql, easy to deploy, easy to expand horizontally
* Multiple distributed transaction protocol supported
  + TCC: Try-Confirm-Cancel
  + SAGA:
  + Reliable news
  + XA

## dtm vs other

At present, the open source distributed transaction framework has not yet seen a mature framework for non-Java languages. There are many Java projects, such as Ali's SEATA, Huawei's ServiceComb-Pack, JD's shardingsphere, and himly, tcc-transaction, ByteTCC, etc. Among them, seata is the most widely used.

The following is a comparison of the main features of dtm and seata:

| Features | DTM | SEATA | Remarks |
|:-----:|:----:|:----:|:----:|
| Supported languages ​​| <font color=green>Golang, python, php and others</font> | <font color=orange>Java</font> |dtm can easily support a new language|
|Exception Handling| <font color=green>[Sub-transaction barrier automatic processing](./doc/barrier-en.md)</font>|<font color=orange>Manual processing</font> | dtm solves idempotence, suspension, and null compensation|
| TCC| <font color=green>✓</font>|<font color=green>✓</font>||
| XA|<font color=green>✓</font>|<font color=green>✓</font>||
|AT |<font color=red>✗</font>|<font color=green>✓</font>|AT is similar to XA, with better performance but dirty rollback|
| SAGA | <font color=orange>simple mode</font> | <font color=green>state machine complex mode</font> |dtm's state machine mode is under planning|
|Transaction message|<font color=green>✓</font>|<font color=red>✗</font>|dtm provides transaction messages similar to rocketmq|
|Communication protocol|HTTP|dubbo and other protocols, no HTTP|dtm will support grpc protocol in the future|
|star number|<img src="https://img.shields.io/github/stars/yedf/dtm.svg?style=social" alt="github stars"/>|<img src="https:/ /img.shields.io/github/stars/seata/seata.svg?style=social" alt="github stars"/>|dtm releases 0.1 from 20210604, fast development|

From the features of the comparison above, if your language stack includes languages ​​other than Java, then dtm is your first choice. If your language stack is Java, you can also choose to access dtm and use sub-transaction barrier technology to simplify your business writing.

# Quick start
### Installation
`git clone https://github.com/yedf/dtm`
### DTM depends on mysql

Configure mysql:

`cp conf.sample.yml conf.yml # Modify conf.yml`

### Start and run the saga example
`go run app/main.go`

# Start using

### Use
``` go
  // business microservice address
  const qsBusi = "http://localhost:8081/api/busi_saga"
  req := &gin.H{"amount": 30} // Microservice payload
  // The address where DtmServer serves DTM, which is a url
  DtmServer := "http://localhost:8080/api/dtmsvr"
  saga := dtmcli.NewSaga(DtmServer, dtmcli.MustGenGid(DtmServer)).
    // Add a TransOut sub-transaction, the operation is url: qsBusi+"/TransOut"，
    // compensation operation is url: qsBusi+"/TransOutCompensate"
    Add(qsBusi+"/TransOut", qsBusi+"/TransOutCompensate", req).
    // Add a TransIn sub-transaction, the operation is url: qsBusi+"/TransOut"，
    // compensation operation is url: qsBusi+"/TransInCompensate"
    Add(qsBusi+"/TransIn", qsBusi+"/TransInCompensate", req)
  // Submit the saga transaction, dtm will complete all sub-transactions/rollback all sub-transactions
  err := saga.Submit()
```
### Complete example

Refer to [examples/quick_start.go](./examples/quick_start.go)

### [sub-transaction barrier](./doc/barrier-en.md)

### [protocol](./doc/protocol-en.md)

### Wechat Group

Please add wechat yedf2008 friends or scan the code to add friends

![yedf2008](http://service.ivydad.com/cover/dubbingb6b5e2c0-2d2a-cd59-f7c5-c6b90aceb6f1.jpeg)

If you think this project is good, or helpful to you, please give a star!

### Who is using
<div style='vertical-align: middle'>
    <img alt='ivydad' height='40'  src='https://www.ivydad.com/_nuxt/img/header-logo.2645ad5.png'  /img>
    <img alt='eglass' height='40'  src='https://img.epeijing.cn/official-website/assets/logo.png'  /img>
</div>

### Following is keyword for SEO

分布式事务框架

事务消息

可靠消息

微服务

