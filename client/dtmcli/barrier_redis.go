package dtmcli

import (
	"context"
	"fmt"

	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/logger"
	"github.com/redis/go-redis/v9"
)

// RedisCheckAdjustAmount check the value of key is valid and >= amount. then adjust the amount
func (bb *BranchBarrier) RedisCheckAdjustAmount(rd redis.Cmdable, key string, amount int, barrierExpire int) error {
	bid := bb.newBarrierID()
	bkey1 := fmt.Sprintf("%s-%s-%s-%s", bb.Gid, bb.BranchID, bb.Op, bid)
	originOp := map[string]string{
		dtmimp.OpCancel:     dtmimp.OpTry,
		dtmimp.OpCompensate: dtmimp.OpAction,
	}[bb.Op]
	bkey2 := fmt.Sprintf("%s-%s-%s-%s", bb.Gid, bb.BranchID, originOp, bid)
	v, err := rd.Eval(context.Background(), ` -- RedisCheckAdjustAmount
local v = redis.call('GET', KEYS[1])
local e1 = redis.call('GET', KEYS[2])

if v == false or v + ARGV[1] < 0 then
	return 'FAILURE'
end

if e1 ~= false then
	return 'DUPLICATE'
end

redis.call('SET', KEYS[2], 'op', 'EX', ARGV[3])

if ARGV[2] ~= '' then
	local e2 = redis.call('GET', KEYS[3])
	if e2 == false then
		redis.call('SET', KEYS[3], 'rollback', 'EX', ARGV[3])
		return
	end
end
redis.call('INCRBY', KEYS[1], ARGV[1])
`, []string{key, bkey1, bkey2}, amount, originOp, barrierExpire).Result()
	logger.Debugf("lua return v: %v err: %v", v, err)
	if err == redis.Nil {
		err = nil
	}
	if err == nil && bb.Op == dtmimp.MsgDoOp && v == "DUPLICATE" { // msg DoAndSubmit should be rejected when duplicate
		return ErrDuplicated
	}
	if err == nil && v == ResultFailure {
		err = ErrFailure
	}
	return err
}

// RedisQueryPrepared query prepared for redis
func (bb *BranchBarrier) RedisQueryPrepared(rd redis.Cmdable, barrierExpire int) error {
	bkey1 := fmt.Sprintf("%s-%s-%s-%s", bb.Gid, dtmimp.MsgDoBranch0, dtmimp.MsgDoOp, dtmimp.MsgDoBarrier1)
	v, err := rd.Eval(context.Background(), ` -- RedisQueryPrepared
local v = redis.call('GET', KEYS[1])
if v == false then
	redis.call('SET', KEYS[1], 'rollback', 'EX', ARGV[1])
	v = 'rollback'
end
if v == 'rollback' then
	return 'FAILURE'
end
`, []string{bkey1}, barrierExpire).Result()
	logger.Debugf("lua return v: %v err: %v", v, err)
	if err == redis.Nil {
		err = nil
	}
	if err == nil && v == ResultFailure {
		err = ErrFailure
	}
	return err
}
