package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/dtm-labs/dtm/dtmutil"
)

// TODO: optimize this, it's very strange to use pointer to dtmutil.Config
var conf = &config.Config

// TODO: optimize this, all function should have context as first parameter
var ctx = context.Background()

// Store is the storage with redis, all transaction information will bachend with redis
type Store struct {
}

// Ping execs ping cmd to redis
func (s *Store) Ping() error {
	_, err := redisGet().Ping(ctx).Result()
	return err
}

// PopulateData populates data to redis
func (s *Store) PopulateData(skipDrop bool) {
	if !skipDrop {
		_, err := redisGet().FlushAll(ctx).Result()
		logger.Infof("call redis flushall. result: %v", err)
		dtmimp.PanicIf(err != nil, err)
	}
}

// FindTransGlobalStore finds GlobalTrans data by gid
func (s *Store) FindTransGlobalStore(gid string) *storage.TransGlobalStore {
	logger.Debugf("calling FindTransGlobalStore: %s", gid)
	r, err := redisGet().Get(ctx, conf.Store.RedisPrefix+"_g_"+gid).Result()
	if err == redis.Nil {
		return nil
	}
	dtmimp.E2P(err)
	trans := &storage.TransGlobalStore{}
	dtmimp.MustUnmarshalString(r, trans)
	return trans
}

// ScanTransGlobalStores lists GlobalTrans data
func (s *Store) ScanTransGlobalStores(position *string, limit int64) []storage.TransGlobalStore {
	logger.Debugf("calling ScanTransGlobalStores: %s %d", *position, limit)
	lid := uint64(0)
	if *position != "" {
		lid = uint64(dtmimp.MustAtoi(*position))
	}
	keys, cursor, err := redisGet().Scan(ctx, lid, conf.Store.RedisPrefix+"_g_*", limit).Result()
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

// FindBranches finds Branch data by gid
func (s *Store) FindBranches(gid string) []storage.TransBranchStore {
	logger.Debugf("calling FindBranches: %s", gid)
	sa, err := redisGet().LRange(ctx, conf.Store.RedisPrefix+"_b_"+gid, 0, -1).Result()
	dtmimp.E2P(err)
	branches := make([]storage.TransBranchStore, len(sa))
	for k, v := range sa {
		dtmimp.MustUnmarshalString(v, &branches[k])
	}
	return branches
}

// UpdateBranches updates branches info
func (s *Store) UpdateBranches(branches []storage.TransBranchStore, updates []string) (int, error) {
	return 0, nil // not implemented
}

type argList struct {
	Keys []string
	List []interface{}
}

func newArgList() *argList {
	a := &argList{}
	return a.AppendRaw(conf.Store.RedisPrefix).AppendObject(conf.Store.DataExpire)
}

func (a *argList) AppendGid(gid string) *argList {
	a.Keys = append(a.Keys, conf.Store.RedisPrefix+"_g_"+gid)
	a.Keys = append(a.Keys, conf.Store.RedisPrefix+"_b_"+gid)
	a.Keys = append(a.Keys, conf.Store.RedisPrefix+"_u")
	a.Keys = append(a.Keys, conf.Store.RedisPrefix+"_s_"+gid)
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

// MaySaveNewTrans creates a new trans
func (s *Store) MaySaveNewTrans(global *storage.TransGlobalStore, branches []storage.TransBranchStore) error {
	a := newArgList().
		AppendGid(global.Gid).
		AppendObject(global).
		AppendRaw(global.NextCronTime.Unix()).
		AppendRaw(global.Gid).
		AppendRaw(global.Status).
		AppendBranches(branches)
	global.Steps = nil
	global.Payloads = nil
	_, err := callLua(a, `-- MaySaveNewTrans
local g = redis.call('GET', KEYS[1])
if g ~= false then
	return 'UNIQUE_CONFLICT'
end

redis.call('SET', KEYS[1], ARGV[3], 'EX', ARGV[2])
redis.call('SET', KEYS[4], ARGV[6], 'EX', ARGV[2])
redis.call('ZADD', KEYS[3], ARGV[4], ARGV[5])
for k = 7, table.getn(ARGV) do
	redis.call('RPUSH', KEYS[2], ARGV[k])
end
redis.call('EXPIRE', KEYS[2], ARGV[2])
`)
	return err
}

// LockGlobalSaveBranches creates branches
func (s *Store) LockGlobalSaveBranches(gid string, status string, branches []storage.TransBranchStore, branchStart int) {
	args := newArgList().
		AppendGid(gid).
		AppendRaw(status).
		AppendRaw(branchStart).
		AppendBranches(branches)
	_, err := callLua(args, `-- LockGlobalSaveBranches
local old = redis.call('GET', KEYS[4])
if old ~= ARGV[3] then
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

// ChangeGlobalStatus changes global trans status
func (s *Store) ChangeGlobalStatus(global *storage.TransGlobalStore, newStatus string, updates []string, finished bool) {
	old := global.Status
	global.Status = newStatus
	args := newArgList().
		AppendGid(global.Gid).
		AppendObject(global).
		AppendRaw(old).
		AppendRaw(finished).
		AppendRaw(global.Gid).
		AppendRaw(newStatus)
	_, err := callLua(args, `-- ChangeGlobalStatus
local old = redis.call('GET', KEYS[4])
if old ~= ARGV[4] then
  return 'NOT_FOUND'
end
redis.call('SET', KEYS[1],  ARGV[3], 'EX', ARGV[2])
redis.call('SET', KEYS[4],  ARGV[7], 'EX', ARGV[2])
if ARGV[5] == '1' then
	redis.call('ZREM', KEYS[3], ARGV[6])
end
`)
	dtmimp.E2P(err)
}

// LockOneGlobalTrans finds GlobalTrans
func (s *Store) LockOneGlobalTrans(expireIn time.Duration) *storage.TransGlobalStore {
	expired := time.Now().Add(expireIn).Unix()
	next := time.Now().Add(time.Duration(conf.RetryInterval) * time.Second).Unix()
	args := newArgList().AppendGid("").AppendRaw(expired).AppendRaw(next)
	lua := `-- LockOneGlobalTrans
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
		if errors.Is(err, storage.ErrNotFound) {
			return nil
		}
		dtmimp.E2P(err)
		global := s.FindTransGlobalStore(r)
		if global != nil {
			return global
		}
	}
}

// TouchCronTime updates cronTime
func (s *Store) TouchCronTime(global *storage.TransGlobalStore, nextCronInterval int64) {
	global.NextCronTime = dtmutil.GetNextTime(nextCronInterval)
	global.UpdateTime = dtmutil.GetNextTime(0)
	global.NextCronInterval = nextCronInterval
	args := newArgList().
		AppendGid(global.Gid).
		AppendObject(global).
		AppendRaw(global.NextCronTime.Unix()).
		AppendRaw(global.Status).
		AppendRaw(global.Gid)
	_, err := callLua(args, `-- TouchCronTime
local old = redis.call('GET', KEYS[4])
if old ~= ARGV[5] then
	return 'NOT_FOUND'
end
redis.call('ZADD', KEYS[3], ARGV[4], ARGV[6])
redis.call('SET', KEYS[1], ARGV[3], 'EX', ARGV[2])
	`)
	dtmimp.E2P(err)
}

var (
	rdb  *redis.Client
	once sync.Once
)

func redisGet() *redis.Client {
	once.Do(func() {
		logger.Debugf("connecting to redis: %v", conf.Store)
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", conf.Store.Host, conf.Store.Port),
			Username: conf.Store.User,
			Password: conf.Store.Password,
		})
	})
	return rdb
}
