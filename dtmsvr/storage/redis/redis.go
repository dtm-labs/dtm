/*
 * Copyright (c) 2022 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

// package reds implement the storage for reds
package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/logger"
	"github.com/go-redis/redis/v8"
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
func (s *Store) ScanTransGlobalStores(position *string, limit int64, condition storage.TransGlobalScanCondition) []storage.TransGlobalStore {
	logger.Debugf("calling ScanTransGlobalStores: %s %d", *position, limit)
	lid := uint64(0)
	if *position != "" {
		lid = uint64(dtmimp.MustAtoi(*position))
	}
	globals := []storage.TransGlobalStore{}
	redis := redisGet()
	for {
		limit -= int64(len(globals))
		keys, nextCursor, err := redis.Scan(ctx, lid, conf.Store.RedisPrefix+"_g_*", limit).Result()
		logger.Debugf("calling redis scan: SCAN %d MATCH %s COUNT %d ,scan result: nextCursor:%d keys_len:%d", lid, conf.Store.RedisPrefix+"_g_*", limit, nextCursor, len(keys))

		dtmimp.E2P(err)

		if len(keys) > 0 {
			values, err := redis.MGet(ctx, keys...).Result()
			dtmimp.E2P(err)
			for _, v := range values {
				global := storage.TransGlobalStore{}
				dtmimp.MustUnmarshalString(v.(string), &global)
				if (condition.Status == "" || global.Status == condition.Status) &&
					(condition.TransType == "" || global.TransType == condition.TransType) &&
					(condition.CreateTimeStart.IsZero() || global.CreateTime.After(condition.CreateTimeStart)) &&
					(condition.CreateTimeEnd.IsZero() || global.CreateTime.Before(condition.CreateTimeEnd)) {
					globals = append(globals, global)
				}
				// redis.Scan may return more records than limit
				if len(globals) >= int(limit) {
					break
				}
			}
		}

		lid = nextCursor
		if len(globals) >= int(limit) || nextCursor == 0 {
			break
		}
	}

	if lid > 0 {
		*position = fmt.Sprintf("%d", lid)
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
	Keys []string      // 1 global trans, 2 branches, 3 indices, 4 status
	List []interface{} // 1 redis prefix, 2 data expire
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
-- check duplicates for workflow
if start == "-1" then
	local t = cjson.decode(ARGV[5])
	local bs = redis.call('LRANGE', KEYS[2], 0, -1)
	for i = 1, table.getn(bs) do
		local c = cjson.decode(bs[i])
		if t['branch_id'] == c['branch_id'] and t['op'] == c['op'] then
			return 'UNIQUE_CONFLICT'
		end
	end
end
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
		AppendRaw(newStatus).
		AppendObject(conf.Store.FinishedDataExpire)
	_, err := callLua(args, `-- ChangeGlobalStatus
local old = redis.call('GET', KEYS[4])
if old ~= ARGV[4] then
  return 'NOT_FOUND'
end
redis.call('SET', KEYS[1],  ARGV[3], 'EX', ARGV[2])
redis.call('SET', KEYS[4],  ARGV[7], 'EX', ARGV[2])
if ARGV[5] == '1' then
	redis.call('ZREM', KEYS[3], ARGV[6])
	redis.call('EXPIRE', KEYS[1], ARGV[8])
	redis.call('EXPIRE', KEYS[2], ARGV[8])
	redis.call('EXPIRE', KEYS[4], ARGV[8])
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

// ResetCronTime reset nextCronTime
// unfinished transactions need to be retried as soon as possible after business downtime is recovered
func (s *Store) ResetCronTime(after time.Duration, limit int64) (succeedCount int64, hasRemaining bool, err error) {
	next := time.Now().Unix()
	timeoutTimestamp := time.Now().Add(after).Unix()
	args := newArgList().AppendGid("").AppendRaw(timeoutTimestamp).AppendRaw(next).AppendRaw(limit)
	lua := `-- ResetCronTime
local r = redis.call('ZRANGEBYSCORE', KEYS[3], ARGV[3], '+inf', 'LIMIT', 0, ARGV[5]+1)
local i = 0
for score,gid in pairs(r) do
	if i == tonumber(ARGV[5]) then
		i = i + 1
		break
	end
	redis.call('ZADD', KEYS[3], ARGV[4], gid)
	i = i + 1
end
return tostring(i)
`
	r := ""
	r, err = callLua(args, lua)
	dtmimp.E2P(err)
	succeedCount = int64(dtmimp.MustAtoi(r))
	if succeedCount > limit {
		hasRemaining = true
		succeedCount = limit
	}
	return
}

// ResetTransGlobalCronTime reset nextCronTime of one global trans.
func (s *Store) ResetTransGlobalCronTime(global *storage.TransGlobalStore) error {
	now := dtmutil.GetNextTime(0)
	global.NextCronTime = now
	global.UpdateTime = now
	key := conf.Store.RedisPrefix + "_g_" + global.Gid
	_, err := redisGet().Set(ctx, key, dtmimp.MustMarshalString(global), time.Duration(conf.Store.DataExpire)*time.Second).Result()
	return err
}

// TouchCronTime updates cronTime
func (s *Store) TouchCronTime(global *storage.TransGlobalStore, nextCronInterval int64, nextCronTime *time.Time) {
	global.UpdateTime = dtmutil.GetNextTime(0)
	global.NextCronTime = nextCronTime
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

// ScanKV lists KV pairs
func (s *Store) ScanKV(cat string, position *string, limit int64) []storage.KVStore {
	logger.Debugf("calling ScanKV: %s %s %d", cat, *position, limit)
	lid := uint64(0)
	if *position != "" {
		lid = uint64(dtmimp.MustAtoi(*position))
	}

	kvs := []storage.KVStore{}
	redis := redisGet()
	for {
		limit -= int64(len(kvs))
		keys, nextCursor, err := redis.Scan(ctx, lid, conf.Store.RedisPrefix+"_kv_"+cat+"_*", limit).Result()
		logger.Debugf("calling redis scan: SCAN %d MATCH %s COUNT %d ,scan result: nextCursor:%d keys_len:%d", lid, conf.Store.RedisPrefix+"_kv_"+cat+"_*", limit, nextCursor, len(keys))
		dtmimp.E2P(err)
		if len(keys) > 0 {
			values, err := redis.MGet(ctx, keys...).Result()
			dtmimp.E2P(err)
			logger.Debugf("keys: %s values: %s", dtmimp.MustMarshalString(keys), dtmimp.MustMarshalString(values))
			for _, v := range values {
				if v == nil {
					continue
				}
				kv := storage.KVStore{}
				dtmimp.MustUnmarshalString(v.(string), &kv)
				kvs = append(kvs, kv)
			}
		}

		lid = nextCursor
		// for redis, `count` in `scan` command is only a hint, may return more than `count` items
		if len(kvs) >= int(limit) || nextCursor == 0 {
			break
		}
	}

	if lid > 0 {
		*position = fmt.Sprintf("%d", lid)
	} else {
		*position = ""
	}
	return kvs
}

// FindKV finds key-value pairs
func (s *Store) FindKV(cat, key string) []storage.KVStore {
	var keys []string
	pattern := conf.Store.RedisPrefix + "_kv_"
	if cat != "" {
		pattern += cat + "_"
	}
	if key != "" {
		keys = []string{pattern + key}
	} else {
		lid := uint64(0)
		r := redisGet().Scan(ctx, lid, pattern+"*", int64(-1))
		dtmimp.E2P(r.Err())
		keys, _ = r.Val()
	}

	kvs := []storage.KVStore{}
	if len(keys) <= 0 {
		return nil
	}
	values, err := redisGet().MGet(ctx, keys...).Result()
	dtmimp.E2P(err)
	for _, v := range values {
		if v == nil {
			continue
		}
		kv := storage.KVStore{}
		dtmimp.MustUnmarshalString(v.(string), &kv)
		kvs = append(kvs, kv)
	}
	return kvs
}

// UpdateKV updates key-value pair
func (s *Store) UpdateKV(kv *storage.KVStore) error {
	now := time.Now()
	kv.UpdateTime = &now
	oldVersion := kv.Version
	kv.Version = oldVersion + 1

	redisKey := fmt.Sprintf("%s_kv_%s_%s", conf.Store.RedisPrefix, kv.Cat, kv.K)
	args := &argList{}
	args.Keys = append(args.Keys, redisKey)
	args.AppendRaw(oldVersion)
	args.AppendObject(kv)
	_, err := callLua(args, `-- UpdateKV
local oldJson = redis.call('GET', KEYS[1])
if oldJson == false then
	return 'NOT_FOUND'
end
local old = cjson.decode(oldJson)
if tostring(old.version) == ARGV[1] then
	redis.call('SET', KEYS[1], ARGV[2])
else 
	return 'NOT_FOUND'
end
`)
	return err
}

// DeleteKV deletes key-value pair
func (s *Store) DeleteKV(cat, key string) error {
	affected, err := redisGet().Del(ctx, fmt.Sprintf("%s_kv_%s_%s", conf.Store.RedisPrefix, cat, key)).Result()
	if err == nil && affected == 0 {
		return storage.ErrNotFound
	}
	return err
}

// CreateKV creates key-value pair
func (s *Store) CreateKV(cat, key, value string) error {
	now := time.Now()
	kv := &storage.KVStore{
		ModelBase: dtmutil.ModelBase{
			CreateTime: &now,
			UpdateTime: &now,
		},
		Cat:     cat,
		K:       key,
		V:       value,
		Version: 1,
	}
	redisKey := fmt.Sprintf("%s_kv_%s_%s", conf.Store.RedisPrefix, kv.Cat, kv.K)
	args := &argList{}
	args.Keys = append(args.Keys, redisKey)
	args.AppendObject(kv)
	_, err := callLua(args, `-- CreateKV
local key = redis.call('GET', KEYS[1])
if key ~= false then
	return 'UNIQUE_CONFLICT'
end
redis.call('SET', KEYS[1], ARGV[1])
`)
	return err
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
