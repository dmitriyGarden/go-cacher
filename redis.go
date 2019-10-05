package cacher

import (
	"errors"
	"github.com/go-redis/redis"
	"log"
	"strconv"
	"time"
)

const depKey = "dpn"

//RedisConfig structure is used for configure redis based cache
type RedisConfig struct {
	Client           redis.UniversalClient
	KeyPrefix        string
	DependencyPrefix string
	LogPrefix        string
}

//Redis is implements of ICache methods
type Redis struct {
	conf          *RedisConfig
	dependencyKey string
}

//GetDependencies returns dependency by key
func (r *Redis) GetDependencies(depKey ...string) ([]Dependency, error) {
	m, err := r.getDependencies(depKey)
	if err != nil {
		return nil, err
	}
	res := make([]Dependency, 0, len(m))
	for k, val := range m {
		counter, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			log.Println(r.conf.LogPrefix, err)
		} else {
			res = append(res, Dependency{
				Key:   k,
				Value: counter,
			})
		}
	}
	return res, nil
}

func (r *Redis) prepareKey(key string) string {
	return r.conf.KeyPrefix + key
}

//Set sets value of key with ttl and dependencies
func (r *Redis) Set(key, value string, ttl *time.Duration, dependency ...IDependency) error {
	key = r.prepareKey(key)
	sets := make(map[string]interface{})
	sets["value"] = value
	for i := range dependency {
		sets[dependency[i].GetKey()] = dependency[i].GetValue()
	}
	res := r.conf.Client.Del(key)
	if res.Err() != nil && res.Err() != redis.Nil {
		return res.Err()
	}
	status := r.conf.Client.HMSet(key, sets)
	if status.Err() != nil {
		return status.Err()
	}
	if ttl != nil {
		res := r.conf.Client.Expire(key, *ttl)
		if res.Err() != nil {
			return res.Err()
		}
	}
	return nil
}

//Get returns value of key
func (r *Redis) Get(key string) (value string, dependencies []Dependency, ok bool, err error) {
	v := r.conf.Client.HGetAll(r.prepareKey(key))
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
			counter, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				log.Println(r.conf.LogPrefix, err)
			} else {
				dependencies = append(dependencies, Dependency{
					Key:   k,
					Value: counter,
				})
			}
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
		return "", nil, false, err
	}
	for k, val := range m {
		if k == "value" {
			continue
		}
		if current[k] != val {
			// invalidate
			r.Del(key)
			return "", nil, false, nil
		}
	}
	return value, dependencies, true, nil
}

func (r *Redis) getDependencies(depKeys []string) (map[string]string, error) {
	if len(depKeys) == 0 {
		return map[string]string{}, nil
	}
	v := r.conf.Client.HMGet(r.dependencyKey, depKeys...)
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
func (r *Redis) SetDependency(dependency ...IDependency) error {
	if len(dependency) == 0 {
		return nil
	}
	m := make(map[string]interface{})
	for _, v := range dependency {
		m[v.GetKey()] = v.GetValue()
	}
	res := r.conf.Client.HMSet(r.dependencyKey, m)
	return res.Err()
}

//Clear invalidate all cache
func (r *Redis) Clear() error {
	res := r.conf.Client.Del(r.dependencyKey)
	if res.Err() != nil && res.Err() != redis.Nil {
		return res.Err()
	}
	return nil
}

//NewRedis creat structure which implements ICache interface
func NewRedis(conf *RedisConfig) (ICache, error) {
	if conf == nil {
		return nil, errors.New("config cannot be nil")
	}
	if conf.Client == nil {
		return nil, errors.New("Client required")
	}
	if conf.KeyPrefix == "" {
		return nil, errors.New("KeyPrefix required")
	}
	if conf.DependencyPrefix == "" {
		return nil, errors.New("DependencyPrefix required")
	}
	return &Redis{
		conf:          conf,
		dependencyKey: conf.DependencyPrefix + depKey,
	}, nil
}
