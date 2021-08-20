## abnormal situations
For a distributed system, because the state of each node is uncontrollable, downtime may occur; the network is uncontrollable, and network failure may occur. Therefore, all distributed systems need to deal with the above-mentioned failure conditions. For example, a failed operation needs to be retried to ensure that the final operation is completed.

When a service receives a retry request, the request may be the first normal request, or it may be a repeated request (after the previous request is processed, the returned response has a network problem). In a distributed system, the business needs to successfully process repeated requests, which is called idempotent.

Distributed transactions not only involve the above-mentioned distributed problems, but also involve transactions, and the situation is more complicated. Let's take a look at the three common problems.

## dangling compensation, idempotent, dangling action

The following describes these abnormal situations with TCC transactions:

- Dangling compensation: The second-stage Cancel method is called without calling the Try method of the TCC resource. The Cancel method needs to recognize that this is an empty rollback, and then directly returns success.
- Idempotence: Since any request may have network abnormalities and repeated requests, all distributed transaction branches need to be idempotent.
- Dangling action: For a distributed transaction, when the Try method is executed, the second phase Cancel interface has been executed before. The Try method needs to recognize that this is a dangling action and return directly to failure.

Let’s take a look at a sequence diagram of network abnormalities to better understand the above-mentioned problems.

![abnormal network image](https://pic2.zhimg.com/80/v2-04c577b69ab7145ab493a8158a048a08_1440w.png)

- When business processing step 4, Cancel is executed before Try, and empty rollback needs to be processed.
- When business processing step 6, Cancel is executed repeatedly and needs to be idempotent.
- When business processing step 8, Try is executed after Cancel and needs to be processed.

For the above-mentioned complex network abnormalities, all distributed transaction systems currently recommend that business developers use unique keys to query whether the associated operations have been completed. If it is completed, it returns directly to success. The related logic is complex, error-prone, and the business burden is heavy.

## Sub-transaction barrier technology

We propose a seed transaction barrier technology, using this technology, can achieve this effect, see the schematic diagram:

![sub-transaction barrier image](https://pic3.zhimg.com/80/v2-b5e742b74fe5a3ccedb11ae444613e3c_1440w.png)

All these requests, after reaching the sub-transaction barrier: abnormal requests will be filtered; normal requests will pass the barrier. After the developer uses the sub-transaction barrier, the various exceptions mentioned above are all properly handled, and the business developer only needs to pay attention to the actual business logic, and the burden is greatly reduced.

The external interface of the sub-transaction barrier is very simple. It only provides a method ThroughBarrierCall. The prototype of the method (the Go language is currently completed, and other languages ​​are under development):

`func ThroughBarrierCall(db *sql.DB, transInfo *TransInfo, busiCall BusiFunc)`

Business developers write their own related logic in busiCall and call this function. ThroughBarrierCall guarantees that busiCall will not be called in scenarios such as dangling compensation, repeated request, dangling action, ensuring that the actual business processing logic is only submitted once.

Sub-transaction barrier will manage TCC, SAGA, transaction messages, etc., and can also be extended to other areas


## Principle of Subtransaction Barrier

The principle of the sub-transaction barrier technology is to create a branch transaction status table sub_trans_barrier in the local database. The unique key is the global transaction id-sub-transaction id-sub-transaction branch type (try|confirm|cancel). The table creation statement is as follows:

``` SQL
CREATE TABLE `barrier` (
  `id` int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `trans_type` varchar(45) DEFAULT'',
  `gid` varchar(128) DEFAULT'',
  `branch_id` varchar(128) DEFAULT'',
  `branch_type` varchar(45) DEFAULT'',
  `create_time` datetime DEFAULT now(),
  `update_time` datetime DEFAULT now(),
  UNIQUE KEY `gid` (`gid`,`branch_id`,`branch_type`)
);
```

The internal steps of ThroughBarrierCall are as follows:

- Open transaction
- If it is a Try branch, then insert ignore is inserted into gid-branchid-try, if it is successfully inserted, then the logic in the barrier is called
- If it is a Confirm branch, then insert ignore inserts gid-branchid-confirm, if it is successfully inserted, call the logic in the barrier
- If it is a Cancel branch, insert ignore into gid-branchid-try, and then insert gid-branchid-cancel, if the try is not inserted and cancel is inserted successfully, the logic in the barrier is called
- The logic within the barrier returns success, commits the transaction, and returns success
- The logic inside the barrier returns an error, rolls back the transaction, and returns an error

Under this mechanism, problems related to network abnormalities are solved

- Dangling compensation control: If Try is not executed and Cancel is executed directly, then Cancel will be inserted into gid-branchid-try successfully, and the logic inside the barrier will not be followed, ensuring empty compensation control
- Idempotent control: No single key can be inserted repeatedly in any branch, which ensures that it will not be executed repeatedly
- Dangling action control: Try is to be executed after Cancel, then the inserted gid-branchid-try is not successful, it will not be executed

Let's take a look at the example in http_tcc_barrier.go in dtm:

``` GO
func tccBarrierTransInTry(c *gin.Context) (interface{}, error) {
	req := reqFrom(c) // 去重构一下，改成可以重复使用的输入
	barrier := MustBarrierFromGin(c)
	return dtmcli.ResultSuccess, barrier.Call(txGet(), func(db dtmcli.DB) error {
		return adjustTrading(db, transInUID, req.Amount)
	})
}
```

The Try in the TransIn business only needs one barrier.Call call to handle the above abnormal situation, which greatly simplifies the work of business developers. For SAGA transactions, reliable messages, etc., a similar mechanism can also be used.

## summary
The sub-transaction barrier technology proposed in this project systematically solves the problem of network disorder in distributed transactions and greatly reduces the difficulty of sub-transaction disorder processing.

Other development languages ​​can also quickly access the technology