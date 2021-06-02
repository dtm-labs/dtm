### 轻量级分布式事务管理服务
  跨语言--语言无关，基于http协议
  支持xa、tcc、saga
## 快速开始
  场景描述：
    假设您实现了一个转账功能，分为两个微服务：转入、转出
      转出：服务地址为 http://example.com/api/busi_saga/transOut?gid=xxx POST 参数为 {"uid": 2, "amount":30}
      转入：服务地址为 http://example.com/api/busi_saga/transIn?gid=xxx POST 参数为 {"uid": 1, "amount":30}
    在saga模式下，有对应的补偿微服务
      转出：服务地址为 http://example.com/api/busi_saga/transOutCompensate?gid=xxx POST 参数为 {"uid": 2, "amount":30}
      转入：服务地址为 http://example.com/api/busi_saga/transInCompensate?gid=xxx POST 参数为 {"uid": 1, "amount":30}
  HTTP协议方式
    curl -d '{"gid":"xxx","trans_type":"saga","steps":[{"action":"http://example.com/api/busi_saga/TransOut","compensate":"http://example.com/api/busi_saga/TransOutCompensate","data":"{\"amount\":30}"},{"action":"http://localhost:8081/api/busi_saga/TransIn","compensate":"http://localhost:8081/api/busi_saga/TransInCompensate","data":"{\"amount\":30}"}]}' 8.140.124.252/api/dtm/commit
    此请求向dtm提交了一个saga事务，dtm会按照saga模式，请求transIn/transOut，并且在出错情况下，保证抵用相关的补偿api
  go客户端方式
    // 事务参与者的服务地址
    const startBusiPort = 8084
    const startBusiApi = "/api/busi_start"

    var startBusi = fmt.Sprintf("http://localhost:%d%s", startBusiPort, startBusiApi)
    err := dtm.SagaNew(DtmServer, gid).Add(startBusi+"/TransOut", startBusi+"/TransOutCompensate", &gin.H{
      "amount":         30,
      "uid": 2,
    }).Add(startBusi+"/TransIn", startBusi+"/TransInCompensate", &gin.H{
      "amount":         30,
      "uid": 1
    }).Commit()
  
  本地启动方式
    需要安装docker，和docker-compose
    curl localhost:8080/api/initMysql
    go run examples/app/main saga

  其他
