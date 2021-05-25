package dtmsvr

type dtmsvrConfig struct {
	PreparedExpire int64 `json:"prepare_expire"` // 单位秒，当prepared的状态超过该时间，才能够转变成canceled，避免cancel了之后，才进入prepared
	Mysql          map[string]string
}

var config = &dtmsvrConfig{
	PreparedExpire: 60,
}
