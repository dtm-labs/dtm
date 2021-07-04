package dtmsvr

type dtmsvrConfig struct {
	TransCronInterval int64 // 单位秒 当事务等待这个时间之后，还没有变化，则进行一轮处理，包括prepared中的任务和commited的任务
	Mysql             map[string]string
}

var config = &dtmsvrConfig{
	TransCronInterval: 10,
}

var dbName = "dtm"
