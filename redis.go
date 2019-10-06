package cacher

import (
	"errors"
	"github.com/go-redis/redis"
	"log"
	"reflect"
	"strconv"
	"time"
)

//RedisConfig structure is used for configure redis based cache
type RedisConfig struct {
	Client             redis.UniversalClient
	KeyPrefix          string
	DependencyPrefix   string
	LogPrefix          string
	EnableDataCompress bool
}

//Redis is implements of ICache methods
type Redis struct {
	conf *RedisConfig
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

func (r *Redis) prepareDepKey(key string) string {
	return r.conf.DependencyPrefix + key
}

//Set sets value of key with ttl and dependencies
func (r *Redis) Set(key, value string, ttl *time.Duration, dependency ...IDependency) error {
	data, err := marshalData(value, dependency...)
	if err != nil {
		return err
	}
	key = r.prepareKey(key)
	sets := map[string]interface{}{}
	if r.conf.EnableDataCompress {
		data, err = compressData(data)
		if err != nil {
			return err
		}
		sets["c"] = "1"
	} else {
		sets["c"] = "0"
	}
	sets["v"] = string(data[:])
	status := r.conf.Client.HMSet(key, sets)
	if status.Err() != nil {
		return status.Err()
	}
	if ttl != nil {
		res := r.conf.Client.Expire(key, *ttl)
		if res.Err() != nil {
			return res.Err()
		}
		r.setDepTTL(*ttl*2, dependency...)
	}
	return nil
}

func (r *Redis) setDepTTL(ttl time.Duration, dependency ...IDependency) {
	for _, dep := range dependency {
		res := r.conf.Client.Expire(r.prepareDepKey(dep.GetKey()), ttl)
		if res.Err() != nil {
			log.Println(r.conf.LogPrefix, res.Err())
		}
	}
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
	var data string
	data, ok = m["v"]
	if !ok {
		// bad value
		err = r.Del(key)
		return
	}
	//decompress value
	if m["c"] == "1" {
		d := make([]byte, 0)
		d, err = decompressData([]byte(data))
		if err != nil {
			ok = false
			return
		}
		data = string(d[:])
	}

	res := &storageData{}
	res, err = unmarshalData([]byte(data))
	if err != nil {
		ok = false
		return
	}
	value = res.Value
	dependencies = res.Dependency
	if len(res.Dependency) == 0 {
		return
	}

	// validate dependencies
	deps := make([]string, 0, len(res.Dependency))
	oldDeps := map[string]string{}
	for _, dep := range res.Dependency {
		deps = append(deps, dep.GetKey())
		oldDeps[dep.GetKey()] = strconv.FormatInt(dep.GetValue(), 10)
	}
	current, err := r.getDependencies(deps)
	if err != nil {
		return "", nil, false, err
	}
	if !reflect.DeepEqual(current, oldDeps) {
		// invalidate
		r.Del(key)
		return "", nil, false, nil
	}
	return value, dependencies, true, nil
}

func (r *Redis) getDependencies(depKeys []string) (map[string]string, error) {
	if len(depKeys) == 0 {
		return map[string]string{}, nil
	}
	keys := make([]string, len(depKeys))
	for i := range depKeys {
		keys[i] = r.prepareDepKey(depKeys[i])
	}
	v := r.conf.Client.MGet(keys...)
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
func (r *Redis) IncrDependency(ttl *time.Duration, depKey ...string) error {
	if len(depKey) == 0 {
		return nil
	}
	if len(depKey) == 1 {
		res := r.conf.Client.Incr(r.prepareDepKey(depKey[0]))
		return res.Err()
	}
	pipe := r.conf.Client.TxPipeline()
	for _, key := range depKey {
		k := r.prepareDepKey(key)
		pipe.Incr(k)
		if ttl != nil {
			pipe.Expire(k, *ttl)
		}

	}
	_, err := pipe.Exec()
	return err
}

//SetDependency  sets value of dependency
func (r *Redis) SetDependency(ttl *time.Duration, dependency ...IDependency) error {
	if len(dependency) == 0 {
		return nil
	}
	pairs := make([]interface{}, 0, len(dependency)*2)
	keys := make([]string, 0, len(dependency))
	pipe := r.conf.Client.TxPipeline()
	for _, dep := range dependency {
		key := r.prepareDepKey(dep.GetKey())
		if ttl != nil {
			keys = append(keys, key)
		}
		pairs = append(pairs, key, dep.GetValue())
	}
	pipe.MSet(pairs...)
	for _, key := range keys {
		if ttl != nil {
			pipe.Expire(key, *ttl)
		}
	}
	_, err := pipe.Exec()
	return err
}

//Clear invalidate all cache
func (r *Redis) Clear() error {
	err := r.clearDeps()
	if err != nil {
		return err
	}
	return r.clearCache()
}

func (r *Redis) clearDeps() error {
	cursor := uint64(0)
	for {
		res := r.conf.Client.Scan(cursor, r.conf.DependencyPrefix+"*", 100)
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
			return nil
		}
	}
}

func (r *Redis) clearCache() error {
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
			return nil
		}
	}
}

//NewRedis creat structure which implements ICache interface
func NewRedis(conf *RedisConfig) (ICache, error) {
	if conf == nil {
		return nil, errors.New("config cannot be nil")
	}
	if conf.Client == nil {
		return nil, errors.New("client required")
	}
	if conf.KeyPrefix == "" {
		return nil, errors.New("KeyPrefix required")
	}
	if conf.DependencyPrefix == "" {
		return nil, errors.New("DependencyPrefix required")
	}
	return &Redis{
		conf: conf,
	}, nil
}
