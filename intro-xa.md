# 分布式事务深入浅出
### 事务
某些业务要求，一系列操作必须全部执行，而不能仅执行一部分。例如，一个转账操作：  

```
-- 从id=1的账户给id=2的账户转账100元
-- 第一步：将id=1的A账户余额减去100
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
-- 第二步：将id=2的B账户余额加上100
UPDATE accounts SET balance = balance + 100 WHERE id = 2;
```
这两条SQL语句必须全部执行，或者，由于某些原因，如果第一条语句成功，第二条语句失败，就必须全部撤销。

这种把多条语句作为一个整体进行操作的功能，被称为数据库事务。数据库事务可以确保该事务范围内的所有操作都可以全部成功或者全部失败。如果事务失败，那么效果就和没有执行这些SQL一样，不会对数据库数据有任何改动。

[更多事务介绍](https://www.liaoxuefeng.com/wiki/1177760294764384/1179611198786848)


### 微服务

如果一个事务涉及的所有操作能够放在一个服务内部，那么使用各门语言里事务相关的库，可以轻松的实现多个操作作为整体的事务操作。

但是有些服务，例如生成订单涉及做很多操作，包括库存、优惠券、赠送、账户余额等。当系统复杂程度增加时，想要把所有这些操作放到一个服务内实现，会导致耦合度太高，维护成本非常高。

针对复杂的系统，当前流行的微服务架构是非常好的解决方案，该架构能够把复杂系统进行拆分，拆分后形成了大量微服务，独立开发，独立维护。

[更多微服务介绍](https://www.zhihu.com/question/65502802)

虽然服务拆分了，但是订单本身的逻辑需要多个操作作为一个整体，要么全部成功，要么全部失败，这就带来了新的挑战。如何把散落在各个微服务中的本地事务，组成一个大的事务，保证他们作为一个整体，这就是分布式事务需要解决的问题。

### 分布式事务
分布式事务简单的说，就是一次大的操作由不同的小操作组成，这些小的操作分布在不同的服务器上，且属于不同的应用，分布式事务需要保证这些小操作要么全部成功，要么全部失败。本质上来说，分布式事务就是为了保证不同数据库的数据一致性。

[更多分布式事务介绍](https://juejin.cn/post/6844903647197806605)

分布式事务方案包括：
  * xa
  * tcc
  * saga
  * 可靠消息
  
下面我们看看最简单的xa

### XA

XA一共分为两阶段：

第一阶段（prepare）：事务管理器向所有本地资源管理器发起请求，询问是否是 ready 状态，所有参与者都将本事务能否成功的信息反馈发给协调者；

第二阶段 (commit/rollback)：事务管理器根据所有本地资源管理器的反馈，通知所有本地资源管理器，步调一致地在所有分支上提交或者回滚。

目前主流的数据库基本都支持XA事务，包括mysql、oracle、sqlserver、postgre

### xa实践

介绍了这么多，我们来实践完成一个微服务上的xa事务，加深分布式事务的理解，这里将采用[dtm](https://github.com/yedf/dtm.git)作为示例

[安装go](https://golang.org/doc/install)

[安装mysql](https://www.mysql.com/cn/)

获取dtm
```
git clone https://github.com/yedf/dtm.git
cd dtm
```
配置mysql
```
cp conf.sample.yml conf.yml
vi conf.yml
```

运行示例

```
go run app/main xa
```

从日志里，能够找到以下输出
```
# 服务1输出
XA start '4fPqCNTYeSG'
UPDATE `user_account` SET `balance`=balance + 30,`update_time`='2021-06-09 11:50:42.438' WHERE user_id = '1'
XA end '4fPqCNTYeSG'
XA prepare '4fPqCNTYeSG'

# 服务2输出
XA start '4fPqCPijxyC'
UPDATE `user_account` SET `balance`=balance + 30,`update_time`='2021-06-09 11:50:42.493' WHERE user_id = '2'
XA end '4fPqCPijxyC'
XA prepare '4fPqCPijxyC'

# 服务1输出
xa commit '4fPqCNTYeSG'

#服务2输出
xa commit '4fPqCPijxyC'
```


### 总结
在这篇简短的文章里，我们大致介绍了 事务->分布式事务->微服务处理XA事务。有兴趣的同学可以通过[dtm](https://github.com/yedf/dtm)继续研究分布式事务