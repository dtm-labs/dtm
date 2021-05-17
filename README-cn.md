### 轻量级分布式事务管理服务

## 配置rabbitmq和mysql

dtm依赖于rabbitmq和mysql，请搭建好rabbitmq和mysql，并修改dtm.yml

## 启动tc

```go run dtm-svr/svr```

## 启动例子saga的tm+rm

```go run example/saga```

## 或者启动例子tcc的tm+rm

```go run example/tcc```