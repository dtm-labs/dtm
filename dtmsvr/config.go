package dtmsvr

type dtmsvrConfig struct {
	PreparedExpire  int64 // 单位秒，处于prepared中的任务，过了这个时间，查询结果还是PENDING的话，则会被cancel
	JobCronInterval int64 // 单位秒 当事务等待这个时间之后，还没有变化，则进行一轮处理，包括prepared中的任务和commited的任务
	Mysql           map[string]string
}

var config = &dtmsvrConfig{
	PreparedExpire:  60,
	JobCronInterval: 20,
}

var dbName = "dtm"
