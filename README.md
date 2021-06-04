## 轻量级分布式事务管理服务
  * 跨语言
    - 语言无关，基于http协议
  * 多种分布式协议支持
    - 支持xa、tcc、saga
## 运行示例
### dtm依赖于mysql

使用已有的mysql：  

`cp conf.sample.yml conf.yml # 修改conf.yml`  

或者通过docker安装mysql  

`docker-compose up -f compose.mysql.yml`
### 启动并运行saga示例
`go run app/main.go`

## 开始使用

### 安装
`go get github.com/yedf/dtm`
### 使用
``` go
gid := common.GenGid()
req := &gin.H{"amount": 30}
saga := dtm.SagaNew(DtmServer, gid).
  Add(startBusi+"/TransOut", startBusi+"/TransOutCompensate", req).
  Add(startBusi+"/TransIn", startBusi+"/TransInCompensate", req)
err := saga.Commit()
```
### 完整示例
参考examples/quick_start.go

### 交流群
请加 yedf2008 好友或者扫码加好友，验证回复 dtm 按照指引进群  

![yedf2008](http://service.ivydad.com/cover/dubbingd9af238e-a2a7-e9fa-1267-cc757c83e834.jpeg)