package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/yedf/dtm/common"
	"github.com/yedf/dtm/dtmcli/dtmimp"
	"github.com/yedf/dtm/dtmcli/logger"
	"github.com/yedf/dtm/dtmsvr/storage"
)

var config = &common.Config

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

func (s *RedisStore) FindTransGlobalStore(gid string) *storage.TransGlobalStore {
	r, err := redisGet().Get(ctx, config.Store.RedisPrefix+"_g_"+gid).Result()
	if err == redis.Nil {
		return nil
	}
	dtmimp.E2P(err)
	trans := &storage.TransGlobalStore{}
	dtmimp.MustUnmarshalString(r, trans)
	return trans
}

func (s *RedisStore) ScanTransGlobalStores(position *string, limit int64) []storage.TransGlobalStore {
	lid := uint64(0)
	if *position != "" {
		lid = uint64(dtmimp.MustAtoi(*position))
	}
	keys, cursor, err := redisGet().Scan(ctx, lid, config.Store.RedisPrefix+"_g_*", limit).Result()
	dtmimp.E2P(err)
	globals := []storage.TransGlobalStore{}
	if len(keys) > 0 {
		values, err := redisGet().MGet(ctx, keys...).Result()
		dtmimp.E2P(err)
		for _, v := range values {
			global := storage.TransGlobalStore{}
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

func (s *RedisStore) FindBranches(gid string) []storage.TransBranchStore {
	sa, err := redisGet().LRange(ctx, config.Store.RedisPrefix+"_b_"+gid, 0, -1).Result()
	dtmimp.E2P(err)
	branches := make([]storage.TransBranchStore, len(sa))
	for k, v := range sa {
		dtmimp.MustUnmarshalString(v, &branches[k])
	}
	return branches
}

func (s *RedisStore) UpdateBranchesSql(branches []storage.TransBranchStore, updates []string) *gorm.DB {
	return nil // not implemented
}

type argList struct {
	Keys []string
	List []interface{}
}

func newArgList() *argList {
	a := &argList{}
	return a.AppendRaw(config.Store.RedisPrefix).AppendObject(config.Store.DataExpire)
}

func (a *argList) AppendGid(gid string) *argList {
	a.Keys = append(a.Keys, config.Store.RedisPrefix+"_g_"+gid)
	a.Keys = append(a.Keys, config.Store.RedisPrefix+"_b_"+gid)
	a.Keys = append(a.Keys, config.Store.RedisPrefix+"_u")
	return a
}

func (a *argList) AppendRaw(v interface{}) *argList {
	a.List = append(a.List, v)
	return a
}

func (a *argList) AppendObject(v interface{}) *argList {
	return a.AppendRaw(dtmimp.MustMarshalString(v))
}

func (a *argList) AppendBranches(branches []storage.TransBranchStore) *argList {
	for _, b := range branches {
		a.AppendRaw(dtmimp.MustMarshalString(b))
	}
	return a
}

func handleRedisResult(ret interface{}, err error) (string, error) {
	logger.Debugf("result is: '%v', err: '%v'", ret, err)
	if err != nil && err != redis.Nil {
		return "", err
	}
	s, _ := ret.(string)
	err = map[string]error{
		"NOT_FOUND":       storage.ErrNotFound,
		"UNIQUE_CONFLICT": storage.ErrUniqueConflict,
	}[s]
	return s, err
}

func callLua(a *argList, lua string) (string, error) {
	logger.Debugf("calling lua. args: %v\nlua:%s", a, lua)
	ret, err := redisGet().Eval(ctx, lua, a.Keys, a.List...).Result()
	return handleRedisResult(ret, err)
}

func (s *RedisStore) MaySaveNewTrans(global *storage.TransGlobalStore, branches []storage.TransBranchStore) error {
	a := newArgList().
		AppendGid(global.Gid).
		AppendObject(global).
		AppendRaw(global.NextCronTime.Unix()).
		AppendBranches(branches)
	global.Steps = nil
	global.Payloads = nil
	_, err := callLua(a, `-- MaySaveNewTrans
local gs = cjson.decode(ARGV[3])
local g = redis.call('GET', KEYS[1])
if g ~= false then
	return 'UNIQUE_CONFLICT'
end

redis.call('SET', KEYS[1], ARGV[3], 'EX', ARGV[2])
redis.call('ZADD', KEYS[3], ARGV[4], gs.gid)
for k = 5, table.getn(ARGV) do
	redis.call('RPUSH', KEYS[2], ARGV[k])
end
redis.call('EXPIRE', KEYS[2], ARGV[2])
`)
	return err
}

func (s *RedisStore) LockGlobalSaveBranches(gid string, status string, branches []storage.TransBranchStore, branchStart int) {
	args := newArgList().
		AppendGid(gid).
		AppendObject(&storage.TransGlobalStore{Gid: gid, Status: status}).
		AppendRaw(branchStart).
		AppendBranches(branches)
	_, err := callLua(args, `
local gs = cjson.decode(ARGV[3])
local g = redis.call('GET', KEYS[1])
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
		redis.call('RPUSH', KEYS[2], ARGV[k])
	else
		redis.call('LSET', KEYS[2], start+k-5, ARGV[k])
	end
end
redis.call('EXPIRE', KEYS[2], ARGV[2])
	`)
	dtmimp.E2P(err)
}

func (s *RedisStore) ChangeGlobalStatus(global *storage.TransGlobalStore, newStatus string, updates []string, finished bool) {
	old := global.Status
	global.Status = newStatus
	args := newArgList().AppendGid(global.Gid).AppendObject(global).AppendRaw(old).AppendRaw(finished)
	_, err := callLua(args, `-- ChangeGlobalStatus
local gs = cjson.decode(ARGV[3])
local old = redis.call('GET', KEYS[1])
if old == false then
	return 'NOT_FOUND'
end
local os = cjson.decode(old)
if os.status ~= ARGV[4] then
  return 'NOT_FOUND'
end
redis.call('SET', KEYS[1],  ARGV[3], 'EX', ARGV[2])
redis.log(redis.LOG_WARNING, 'finished: ', ARGV[5])
if ARGV[5] == '1' then
	redis.call('ZREM', KEYS[3], gs.gid)
end
`)
	dtmimp.E2P(err)
}

func (s *RedisStore) LockOneGlobalTrans(expireIn time.Duration) *storage.TransGlobalStore {
	expired := time.Now().Add(expireIn).Unix()
	next := time.Now().Add(time.Duration(config.RetryInterval) * time.Second).Unix()
	args := newArgList().AppendGid("").AppendRaw(expired).AppendRaw(next)
	lua := `-- LocakOneGlobalTrans
local r = redis.call('ZRANGE', KEYS[3], 0, 0, 'WITHSCORES')
local gid = r[1]
if gid == nil then
	return 'NOT_FOUND'
end

if tonumber(r[2]) > tonumber(ARGV[3]) then
	return 'NOT_FOUND'
end
redis.call('ZADD', KEYS[3], ARGV[4], gid)
return gid
`
	for {
		r, err := callLua(args, lua)
		if err == storage.ErrNotFound {
			return nil
		}
		dtmimp.E2P(err)
		global := s.FindTransGlobalStore(r)
		if global != nil {
			return global
		}
	}
}

func (s *RedisStore) TouchCronTime(global *storage.TransGlobalStore, nextCronInterval int64) {
	global.NextCronTime = common.GetNextTime(nextCronInterval)
	global.UpdateTime = common.GetNextTime(0)
	global.NextCronInterval = nextCronInterval
	args := newArgList().AppendGid(global.Gid).AppendObject(global).AppendRaw(global.NextCronTime.Unix())
	_, err := callLua(args, `-- TouchCronTime
local g = cjson.decode(ARGV[3])
local old = redis.call('GET', KEYS[1])
if old == false then
	return 'NOT_FOUND'
end
local os = cjson.decode(old)
if os.status ~= g.status then
  return 'NOT_FOUND'
end
redis.call('ZADD', KEYS[3], ARGV[4], g.gid)
redis.call('SET', KEYS[1], ARGV[3], 'EX', ARGV[2])
	`)
	dtmimp.E2P(err)
}

func redisGet() *redis.Client {
	return common.RedisGet()
}
