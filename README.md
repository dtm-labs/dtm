![license](https://img.shields.io/github/license/dtm-labs/dtm)
![Build Status](https://github.com/dtm-labs/dtm/actions/workflows/tests.yml/badge.svg?branch=main)
[![codecov](https://codecov.io/gh/dtm-labs/dtm/branch/main/graph/badge.svg?token=UKKEYQLP3F)](https://codecov.io/gh/dtm-labs/dtm)
[![Go Report Card](https://goreportcard.com/badge/github.com/dtm-labs/dtm)](https://goreportcard.com/report/github.com/dtm-labs/dtm)
[![Go Reference](https://pkg.go.dev/badge/github.com/dtm-labs/dtm.svg)](https://pkg.go.dev/github.com/dtm-labs/dtm)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go#database)

English | [简体中文](https://github.com/dtm-labs/dtm/blob/main/helper/README-cn.md)

# Distributed Transactions Manager

## What is DTM

DTM is a distributed transaction framework which provides cross-service eventual data consistency. It provides saga, tcc, xa, 2-phase message, outbox, workflow patterns for a variety of application scenarios. It also supports multiple languages and multiple store engine to form up a transaction as following:

<img alt="function-picture" src="https://en.dtm.pub/assets/function.7d5618f8.png" height=250 />

## Who's using DTM (partial)

[Tencent](https://en.dtm.pub/other/using.html#tencent)

[Bytedance](https://en.dtm.pub/other/using.html#bytedance)

[Ivydad](https://en.dtm.pub/other/using.html#ivydad)

[More](https://en.dtm.pub/other/using.html)

## Features
* Support for multiple transaction modes: SAGA, TCC, XA, Workflow, Outbox
* Multiple languages support: SDK for Go, Java, PHP, C#, Python, Nodejs
* Better Outbox: 2-phase messages, a more elegant solution than Outbox, support multi-databases
* Multiple database transaction support: Mysql, Redis, MongoDB, Postgres, TDSQL, etc.
* Support for multiple storage engines: Mysql (common), Redis (high performance), BoltDB (dev&test), MongoDB (under planning)
* Support for multiple microservices architectures: [go-zero](https://github.com/zeromicro/go-zero), go-kratos/kratos, polarismesh/polaris
* Support for high availability and easy horizontal scaling

## Application scenarios.
DTM can be applied to data consistency issues in a large number of scenarios, here are a few common ones
* [cache management](https://en.dtm.pub/app/cache.html): thoroughly guarantee the cache final consistency and strong consistency
* [flash-sales to deduct inventory](https://en.dtm.pub/app/flash.html): in extreme cases, it is also possible to ensure that the precise inventory in Redis is exactly the same as the final order created, without the need for manual adjustment
* [Non-monolithic order system](https://en.dtm.pub/app/order.html): Dramatically simplifies the architecture
* [Event publishing/subscription](https://en.dtm.pub/practice/msg.html): better outbox pattern

## [Cook Book](https://en.dtm.pub)

## Quick start

### run dtm

``` bash
git clone https://github.com/dtm-labs/dtm && cd dtm
go run main.go
```

### Start an example
Suppose we want to perform an inter-bank transfer. The operations of transfer out (TransOut) and transfer in (TransIn) are coded in separate micro-services.

Here is an example to illustrate a solution of dtm to this problem:

``` bash
git clone https://github.com/dtm-labs/quick-start-sample.git && cd quick-start-sample/workflow-grpc
go run main.go
```

## Code

### Usage
``` go
wfName := "workflow-grpc"
err = workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
  // ...
  // Define a transaction branch for TransOut
  wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
    // compensation for TransOut
    _, err := busiCli.TransOutRevert(wf.Context, &req)
    return err
  })
  _, err = busiCli.TransOut(wf.Context, &req)
  // check error

  // Define another transaction branch for TransIn
  wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
    _, err := busiCli.TransInRevert(wf.Context, &req)
    return err
  })
  _, err = busiCli.TransIn(wf.Context, &req)
  return err
}

// ...
req := busi.BusiReq{Amount: 30, TransInResult: ""}
data, err := proto.Marshal(&req)

// Execute workflow
_, err = workflow.ExecuteCtx(wfName, shortuuid.New(), data)
logger.Infof("result of workflow.Execute is: %v", err)

```

When the above code runs, we can see in the console that services `TransOut`, `TransIn` has been called.

#### Rollback upon failure
If any forward operation fails, DTM invokes the corresponding compensating operation of each sub-transaction to roll back, after which the transaction is successfully rolled back.

Let's purposely trigger the failure of the second sub-transaction and watch what happens

``` go
// req := busi.BusiReq{Amount: 30, TransInResult: ""}
req := busi.BusiReq{Amount: 30, TransInResult: "FAILURE"}
})
```

we can see in the console that services `TransOut`, `TransIn`, `TransOutRevert` has been called

## More examples
If you want more quick start examples, please refer to [dtm-labs/quick-start-sample](https://github.com/dtm-labs/quick-start-sample)

The above example mainly demonstrates the flow of a distributed transaction. More on this, including practical examples of how to interact with an actual database, how to do compensation, how to do rollback, etc. please refer to [dtm-examples](https://github.com/dtm-labs/dtm-examples) for more examples.

## Give a star! ⭐

If you think this project is interesting, or helpful to you, please give a star!
