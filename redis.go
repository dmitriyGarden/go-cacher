package go_cacher

import (
	"errors"
	"github.com/go-redis/redis"
	"time"
)

type RedisConfig struct {
	Client           redis.UniversalClient
	KeyPrefix        string
	DependencyPrefix string
}

type Redis struct {
	conf          *RedisConfig
	dependencyKey string
}

func (r *Redis) prepareKey(key string) string {
	return r.conf.KeyPrefix + key
}

//Set sets value of key with ttl and dependencies
func (r *Redis) Set(key, value string, ttl *time.Duration, dependencies ...Dependency) error {
	key = r.prepareKey(key)
	sets := make(map[string]interface{})
	sets["value"] = value
	for i := range dependencies {
		sets[dependencies[i].GetKey()] = dependencies[i].GetValue()
	}
	pipe := r.conf.Client.TxPipeline()
	pipe.Del(key)
	pipe.HMSet(key, sets)
	if ttl != nil {
		pipe.Expire(key, *ttl)
	}
	_, err := pipe.Exec()
	return err
}

//Get returns value of key
func (r *Redis) Get(key string) (value string, ok bool, err error) {
	v := r.conf.Client.HGetAll(r.prepareKey(key))
	if v.Err() == redis.Nil {
		return
	}
	if v.Err() != nil {
		err = v.Err()
		return
	}
	if len(v.Val()) == 0 {
		//bad value
		err = r.Del(key)
		return
	}

	m := v.Val()
	deps := make([]string, 0, len(m))
	for k, val := range m {
		if k == "value" {
			value = val
			ok = true
		} else {
			deps = append(deps, k)
		}
	}
	if !ok {
		// bad value
		err = r.Del(key)
		return
	}
	if len(deps) == 0 {
		return
	}
	// validate dependencies
	current, err := r.getDependencies(deps)
	if err != nil {
		return "", false, err
	}
	for k, val := range m {
		if current[k] != val {
			// invalidate
			r.Del(key)
			return "", false, nil
		}
	}
	return value, true, nil
}

func (r *Redis) getDependencies(depKeys []string) (map[string]string, error) {
	if len(depKeys) == 0 {
		return map[string]string{}, nil
	}
	v := r.conf.Client.HMGet(r.dependencyKey, depKeys...)
	if v.Err() == redis.Nil {
		return map[string]string{}, nil
	}
	if v.Err() != nil {
		return nil, v.Err()
	}
	res := make(map[string]string)
	for i, val := range v.Val() {
		if val == nil {
			continue
		}
		d, ok := val.(string)
		if ok {
			res[depKeys[i]] = d
		}
	}
	return res, nil
}

//Del delete key from cache
func (r *Redis) Del(key string) error {
	res := r.conf.Client.Del(r.prepareKey(key))
	if res.Err() == redis.Nil {
		return nil
	}
	return res.Err()
}

//IncrDependency increment dependency counter
func (r *Redis) IncrDependency(depKey ...string) error {
	if len(depKey) == 0 {
		return nil
	}
	if len(depKey) == 1 {
		res := r.conf.Client.HIncrBy(r.dependencyKey, depKey[0], 1)
		return res.Err()
	}
	pipe := r.conf.Client.TxPipeline()
	for _, key := range depKey {
		pipe.HIncrBy(r.dependencyKey, key, 1)
	}
	_, err := pipe.Exec()
	return err
}

//SetDependency  sets value of dependency
func (r *Redis) SetDependency(dependencies ...Dependency) error {
	if len(dependencies) == 0 {
		return nil
	}
	m := make(map[string]interface{})
	for _, v := range dependencies {
		m[v.GetKey()] = v.GetValue()
	}
	res := r.conf.Client.HMSet(r.dependencyKey, m)
	return res.Err()
}

//Clear clear all cache
func (r *Redis) Clear() error {
	res := r.conf.Client.Del(r.dependencyKey)
	if res.Err() != nil && res.Err() != redis.Nil {
		return res.Err()
	}
	cursor := uint64(0)
	for {
		res := r.conf.Client.Scan(cursor, r.conf.KeyPrefix+"*", 100)
		keys, cursor, err := res.Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			res := r.conf.Client.Del(keys...)
			if res.Err() != nil {
				return res.Err()
			}
		}
		if cursor == 0 {
			break
		}
	}
	return nil
}

func NewRedis(conf *RedisConfig) (Cache, error) {
	if conf == nil {
		return nil, errors.New("config cannot be nil")
	}
	if conf.KeyPrefix == "" {
		return nil, errors.New("KeyPrefix required")
	}
	if conf.DependencyPrefix == "" {
		return nil, errors.New("DependencyPrefix required")
	}
	if conf.Client == nil {
		return nil, errors.New("Client required")
	}
	return &Redis{
		conf:          conf,
		dependencyKey: conf.KeyPrefix + "dpn",
	}, nil
}
