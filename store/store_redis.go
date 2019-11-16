package store

import (
	"errors"
	"time"
	"utask/app"
	"utask/log"
	"utask/pkg/randstr"
	"utask/task"

	"github.com/go-redis/redis"
)

// RedisKey 是redisStore队列key
const RedisKey = "UTask"

// RedisStore 是redis实现的taskStore
var (
	// DefaultRedisStore redis store
	DefaultRedisStore = NewRedisStore()
)

// RedisStore 是redis实现的taskStore
type RedisStore struct {
	redis *redis.Client
}

func (s *RedisStore) LPop() (task task.Tasker, err error) {
	t, err := s.redis.LPop(RedisKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	task, err = Decode([]byte(t))
	if err != nil {
		_, errPush := s.redis.RPush(RedisKey, t).Result()
		if errPush != nil {
			//这条记录需要重发
			log.Error("task pop err, task : ", t, " data pop err: ", err, " push back err: ", errPush)
		}
		return nil, err
	}
	log.Info("task pop: ", task)
	return task, nil
}

func (s *RedisStore) RPush(task task.Tasker) (bool, error) {
	data, err := Encode(task)
	if err != nil {
		return false, err
	}
	_, err = s.redis.RPush(RedisKey, string(data)).Result()
	if err != nil {
		return false, err
	}
	log.Info("task push: ", task)
	return true, nil
}

// Generate 生成一个token
func (s *RedisStore) Generate(tid string, lifetime time.Duration) (token string, err error) {
	token = randstr.New(32)
	_, err = s.redis.Set(tid, token, lifetime).Result()
	if err != nil {
		return "", err
	}
	return token, nil
}

// Check 校验并删除一个token
func (s *RedisStore) Check(tid, token string) (ok bool, err error) {
	t, err := s.redis.GetSet(tid, "").Result()
	if err != nil {
		return false, nil
	}
	_, _ = s.redis.Del(tid).Result()
	return t == token, nil
}

// NewRedisStore redis construct
func NewRedisStore() *RedisStore {
	return &RedisStore{redis.NewClient(&redis.Options{
		Addr:         app.Config.Redis.Addr,
		Password:     app.Config.Redis.Password,
		DB:           app.Config.Redis.Db,
		IdleTimeout:  20 * time.Second,
		MinIdleConns: 1,
	})}
}
