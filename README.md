# 轻量级分布式事务管理服务
DTM 是一款跨语言的分布式事务管理方案，在各类微服务架构中，提供高性能和简单易用的分布式事务服务。
# 特色
### 跨语言
语言无关，任何语言实现了http方式的服务，都可以接入DTM，用来管理分布式事务。支持go、python、php、nodejs、ruby
### [分布式事务简介](./intro-xa.md)
### 多种分布式事务协议支持
  * TCC: Try-Confirm-Cancel
  * SAGA:
  * 可靠消息
  * XA 需要底层数据库支持XA
### 高可用
基于数据库实现，易集群化，已水平扩展
# 快速开始
### 安装
`go get github.com/yedf/dtm`
### dtm依赖于mysql

使用已有的mysql：  

`cp conf.sample.yml conf.yml # 修改conf.yml`  

或者通过docker安装mysql  

`docker-compose -f compose.mysql.yml up`
### 启动并运行saga示例
`go run app/main.go`

# 开始使用

### 使用
``` go
const DtmServer = "http://localhost:8080/api/dtmsvr"
const startBusi = "http://localhost:8081/api/busi_saga"
gid := common.GenGid() // 生成事务id
req := &gin.H{"amount": 30} // 微服务的负荷
// 生成dtm的saga对象
saga := dtm.SagaNew(DtmServer, gid).
  // 添加两个子事务
  Add(startBusi+"/TransOut", startBusi+"/TransOutCompensate", req).
  Add(startBusi+"/TransIn", startBusi+"/TransInCompensate", req)
  // 提交saga事务
err := saga.Commit()
```
### 完整示例
参考[examples/quick_start.go](./examples/quick_start.go)

### 交流群
请加 yedf2008 好友或者扫码加好友，验证回复 dtm 按照指引进群  

![yedf2008](http://service.ivydad.com/cover/dubbingd9af238e-a2a7-e9fa-1267-cc757c83e834.jpeg)