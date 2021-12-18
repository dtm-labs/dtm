package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"gorm.io/gorm"
)

var ctx context.Context = context.Background()

type RedisStore struct {
}

func (s *RedisStore) Ping() error {
	_, err := redisGet().Ping(ctx).Result()
	return err
}

func (s *RedisStore) PopulateData(skipDrop bool) {
	_, err := redisGet().FlushAll(ctx).Result()
	dtmimp.PanicIf(err != nil, err)
}

func (s *RedisStore) FindTransGlobalStore(gid string) *TransGlobalStore {
	r, err := redisGet().Get(ctx, config.Store.RedisPrefix+"_g_"+gid).Result()
	if err == redis.Nil {
		return nil
	}
	dtmimp.E2P(err)
	trans := &TransGlobalStore{}
	dtmimp.MustUnmarshalString(r, trans)
	return trans
}

func (s *RedisStore) ScanTransGlobalStores(position *string, limit int64) []TransGlobalStore {
	lid := uint64(0)
	if *position != "" {
		lid = uint64(dtmimp.MustAtoi(*position))
	}
	keys, cursor, err := redisGet().Scan(ctx, lid, config.Store.RedisPrefix+"_g_*", limit).Result()
	dtmimp.E2P(err)
	globals := []TransGlobalStore{}
	if len(keys) > 0 {
		values, err := redisGet().MGet(ctx, keys...).Result()
		dtmimp.E2P(err)
		for _, v := range values {
			global := TransGlobalStore{}
			dtmimp.MustUnmarshalString(v.(string), &global)
			globals = append(globals, global)
		}
	}
	if cursor > 0 {
		*position = fmt.Sprintf("%d", cursor)
	} else {
		*position = ""
	}
	return globals
}

func (s *RedisStore) FindBranches(gid string) []TransBranchStore {
	sa, err := redisGet().LRange(ctx, config.Store.RedisPrefix+"_b_"+gid, 0, -1).Result()
	dtmimp.E2P(err)
	branches := make([]TransBranchStore, len(sa))
	for k, v := range sa {
		dtmimp.MustUnmarshalString(v, &branches[k])
	}
	return branches
}

func (s *RedisStore) UpdateBranchesSql(branches []TransBranchStore, updates []string) *gorm.DB {
	return nil // not implemented
}

type argList struct {
	List []interface{}
}

func newArgList() *argList {
	a := &argList{}
	return a.AppendRaw(config.Store.RedisPrefix).AppendObject(config.Store.DataExpire)
}

func (a *argList) AppendRaw(v interface{}) *argList {
	a.List = append(a.List, v)
	return a
}

func (a *argList) AppendObject(v interface{}) *argList {
	return a.AppendRaw(dtmimp.MustMarshalString(v))
}

func (a *argList) AppendBranches(branches []TransBranchStore) *argList {
	for _, b := range branches {
		a.AppendRaw(dtmimp.MustMarshalString(b))
	}
	return a
}

func handleRedisResult(ret interface{}, err error) (string, error) {
	dtmimp.Logf("result is: '%v', err: '%v'", ret, err)
	if err != nil && err != redis.Nil {
		return "", err
	}
	s, _ := ret.(string)
	err = map[string]error{
		"NOT_FOUND":       ErrNotFound,
		"UNIQUE_CONFLICT": ErrUniqueConflict,
	}[s]
	return s, err
}

func callLua(args []interface{}, lua string) (string, error) {
	dtmimp.Logf("calling lua. args: %v\nlua:%s", args, lua)
	ret, err := redisGet().Eval(ctx, lua, []string{config.Store.RedisPrefix}, args...).Result()
	return handleRedisResult(ret, err)
}

func (s *RedisStore) MaySaveNewTrans(global *TransGlobalStore, branches []TransBranchStore) error {
	args := newArgList().
		AppendObject(global).
		AppendRaw(global.NextCronTime.Unix()).
		AppendBranches(branches).
		List
	global.Steps = nil
	global.Payloads = nil
	_, err := callLua(args, `-- MaySaveNewTrans
local gs = cjson.decode(ARGV[3])
local g = redis.call('GET', ARGV[1] .. '_g_' .. gs.gid)
if g ~= false then
	return 'UNIQUE_CONFLICT'
end

redis.call('SET', ARGV[1] .. '_g_' .. gs.gid, ARGV[3], 'EX', ARGV[2])
redis.call('ZADD', ARGV[1] .. '_u', ARGV[4], gs.gid)
for k = 5, table.getn(ARGV) do
	redis.call('RPUSH', ARGV[1] .. '_b_' .. gs.gid, ARGV[k])
end
redis.call('EXPIRE', ARGV[1] .. '_b_' .. gs.gid, ARGV[2])
`)
	return err
}

func (s *RedisStore) LockGlobalSaveBranches(gid string, status string, branches []TransBranchStore, branchStart int) {
	args := newArgList().
		AppendObject(&TransGlobalStore{Gid: gid, Status: status}).
		AppendRaw(branchStart).
		AppendBranches(branches).
		List
	_, err := callLua(args, `
local pre = ARGV[1]
local gs = cjson.decode(ARGV[3])
local g = redis.call('GET', pre .. '_g_' .. gs.gid)
if (g == false) then
	return 'NOT_FOUND'
end
local js = cjson.decode(g)
if js.status ~= gs.status then
	return 'NOT_FOUND'
end
local start = ARGV[4]
for k = 5, table.getn(ARGV) do
	if start == "-1" then
		redis.call('RPUSH', pre .. '_b_' .. gs.gid, ARGV[k])
	else
		redis.call('LSET', pre .. '_b_' .. gs.gid, start+k-5, ARGV[k])
	end
end
redis.call('EXPIRE', pre .. '_b_' .. gs.gid, ARGV[2])
	`)
	dtmimp.E2P(err)
}

func (s *RedisStore) ChangeGlobalStatus(global *TransGlobalStore, newStatus string, updates []string, finished bool) {
	old := global.Status
	global.Status = newStatus
	args := newArgList().AppendObject(global).AppendRaw(old).AppendRaw(finished).List
	_, err := callLua(args, `-- ChangeGlobalStatus
local p = ARGV[1]
local gs = cjson.decode(ARGV[3])
local old = redis.call('GET', p .. '_g_' .. gs.gid)
if old == false then
	return 'NOT_FOUND'
end
local os = cjson.decode(old)
if os.status ~= ARGV[4] then
  return 'NOT_FOUND'
end
redis.call('SET', p .. '_g_' .. gs.gid,  ARGV[3], 'EX', ARGV[2])
redis.log(redis.LOG_WARNING, 'finished: ', ARGV[5])
if ARGV[5] == '1' then
	redis.call('ZREM', p .. '_u', gs.gid)
end
`)
	dtmimp.E2P(err)
}

func (s *RedisStore) LockOneGlobalTrans(expireIn time.Duration) *TransGlobalStore {
	expired := time.Now().Add(expireIn).Unix()
	next := time.Now().Add(time.Duration(config.RetryInterval) * time.Second).Unix()
	args := newArgList().AppendRaw(expired).AppendRaw(next).List
	lua := `-- LocakOneGlobalTrans
local k = ARGV[1] .. '_u'
local r = redis.call('ZRANGE', k, 0, 0, 'WITHSCORES')
local gid = r[1]
if gid == nil then
	return 'NOT_FOUND'
end
local g = redis.call('GET', ARGV[1] .. '_g_' .. gid)
redis.log(redis.LOG_WARNING, 'g is: ', g, 'gid is: ', gid)
if g == false then
	redis.call('ZREM', k, gid)
	return 'NOT_FOUND'
end

if tonumber(r[2]) > tonumber(ARGV[3]) then
	return 'NOT_FOUND'
end
redis.call('ZADD', k, ARGV[4], gid)
return g
`
	r, err := callLua(args, lua)
	for err == ErrShouldRetry {
		r, err = callLua(args, lua)
	}
	if err == ErrNotFound {
		return nil
	}
	dtmimp.E2P(err)
	global := &TransGlobalStore{}
	dtmimp.MustUnmarshalString(r, global)
	return global
}

func (s *RedisStore) TouchCronTime(global *TransGlobalStore, nextCronInterval int64) {
	global.NextCronTime = common.GetNextTime(nextCronInterval)
	global.UpdateTime = common.GetNextTime(0)
	global.NextCronInterval = nextCronInterval
	args := newArgList().AppendObject(global).AppendRaw(global.NextCronTime.Unix()).List
	_, err := callLua(args, `-- TouchCronTime
local p = ARGV[1]
local g = cjson.decode(ARGV[3])
local old = redis.call('GET', p .. '_g_' .. g.gid)
if old == false then
	return 'NOT_FOUND'
end
local os = cjson.decode(old)
if os.status ~= g.status then
  return 'NOT_FOUND'
end
redis.call('ZADD', p .. '_u', ARGV[4], g.gid)
redis.call('SET', p .. '_g_' .. g.gid, ARGV[3], 'EX', ARGV[2])
	`)
	dtmimp.E2P(err)
}
