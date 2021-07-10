[English](https://github.com/yedf/dtm/blob/master/README-en.md)

# GO语言分布式事务管理服务
DTM是首款golang的分布式事务管理器，优雅的解决了幂等、空补偿、悬挂等分布式事务难题。在微服务架构中，提供了高性能和简单易用的分布式事务解决方案。

受邀参加中国数据库大会分享[多语言环境下分布式事务实践](http://dtcc.it168.com/yicheng.html#b9)
## 亮点

* 稳定可靠
  + 经过生产环境考验，单元测试覆盖率90%以上
* 使用简单
  + 接口简单，开发者不再担心悬挂、空补偿、幂等各类问题，框架层代为处理
* 跨语言
  + 可适合多语言栈的公司使用。协议支持http。方便go、python、php、nodejs、ruby各类语言使用。
* 社区活跃
  + 任何问题都快速响应
* 易部署、易扩展
  + 仅依赖mysql，部署简单，易集群化，易水平扩展
* 多种分布式事务协议支持
  + TCC: Try-Confirm-Cancel
  + SAGA:
  + 可靠消息
  + XA

### 文档与介绍(更新中)
  * [分布式事务简介](https://zhuanlan.zhihu.com/p/387487859)
  * 分布式事务模式
    - [XA事务模式](https://zhuanlan.zhihu.com/p/384756957)
    - [SAGA事务模式](https://zhuanlan.zhihu.com/p/385594256)
    - [TCC事务模式](https://zhuanlan.zhihu.com/p/388357329)
    - 可靠消息事务模式
  * [子事务屏障](https://zhuanlan.zhihu.com/p/388444465)
  * FAQ
  * 部署指南

# 快速开始
### 安装
`go get github.com/yedf/dtm`
### dtm依赖于mysql

配置mysql：  

`cp conf.sample.yml conf.yml # 修改conf.yml`  

### 启动并运行saga示例
`go run app/main.go saga`

# 开始使用

### 使用
``` go
  // 具体业务微服务地址
  const qsBusi = "http://localhost:8081/api/busi_saga"
	req := &gin.H{"amount": 30} // 微服务的载荷
	// DtmServer为DTM服务的地址，是一个url
	saga := dtmcli.NewSaga("http://localhost:8080/api/dtmsvr").
		// 添加一个TransOut的子事务，正向操作为url: qsBusi+"/TransOut"， 补偿操作为url: qsBusi+"/TransOutCompensate"
		Add(qsBusi+"/TransOut", qsBusi+"/TransOutCompensate", req).
		// 添加一个TransIn的子事务，正向操作为url: qsBusi+"/TransOut"， 补偿操作为url: qsBusi+"/TransInCompensate"
		Add(qsBusi+"/TransIn", qsBusi+"/TransInCompensate", req)
	// 提交saga事务，dtm会完成所有的子事务/回滚所有的子事务
  err := saga.Submit()
```
### 完整示例
参考[examples/quick_start.go](./examples/quick_start.go)

### 交流群
请加 yedf2008 好友或者扫码加好友，验证回复 dtm 按照指引进群  

![yedf2008](http://service.ivydad.com/cover/dubbingb6b5e2c0-2d2a-cd59-f7c5-c6b90aceb6f1.jpeg)

如果您觉得此项目不错，或者对您有帮助，请赏颗星吧！

### 谁在使用
<div style='vertical-align: middle'>
    <img alt='常青藤爸爸' height='40'  src='https://www.ivydad.com/_nuxt/img/header-logo.2645ad5.png'  /img>
    <img alt='镜小二' height='40'  src='https://img.epeijing.cn/official-website/assets/logo.png'  /img>
</div>
