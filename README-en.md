![license](https://img.shields.io/github/license/yedf/dtm)
[![Build Status](https://travis-ci.com/yedf/dtm.svg?branch=main)](https://travis-ci.com/yedf/dtm)
[![Coverage Status](https://coveralls.io/repos/github/yedf/dtm/badge.svg?branch=main)](https://coveralls.io/github/yedf/dtm?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/yedf/dtm)](https://goreportcard.com/report/github.com/yedf/dtm)
[![Go Reference](https://pkg.go.dev/badge/github.com/yedf/dtm.svg)](https://pkg.go.dev/github.com/yedf/dtm)

[中文版](https://github.com/yedf/dtm/blob/main/README.md)

# a golang distributed transaction manager
DTM is the first golang open source distributed transaction project. It elegantly resolve the problem of execution out of order. In the microservice architecture, a high-performance and easy-to-use distributed transaction solution is provided.

## characteristic

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

# Quick start
### installation
`git clone https://github.com/yedf/dtm`
### dtm depends on mysql

Configure mysql：  

`cp conf.sample.yml conf.yml # 修改conf.yml`  

### Start and run the saga example
`go run app/main.go`

# start using

### use
``` go
  // business microservice address
  const qsBusi = "http://localhost:8081/api/busi_saga"
	req := &gin.H{"amount": 30} // Microservice payload
	// The address where DtmServer serves DTM, which is a url
	saga := dtmcli.NewSaga("http://localhost:8080/api/dtmsvr").
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

### Wechat Group
Please add wechat yedf2008 friends or scan the code to add friends  

![yedf2008](http://service.ivydad.com/cover/dubbingb6b5e2c0-2d2a-cd59-f7c5-c6b90aceb6f1.jpeg)

If you think this project is good, or helpful to you, please give a star!

### Who is using
<div style='vertical-align: middle'>
    <img alt='常青藤爸爸' height='40'  src='https://www.ivydad.com/_nuxt/img/header-logo.2645ad5.png'  /img>
    <img alt='镜小二' height='40'  src='https://img.epeijing.cn/official-website/assets/logo.png'  /img>
</div>
