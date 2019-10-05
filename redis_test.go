package go_cacher

import (
	"github.com/go-redis/redis"
	"reflect"
	"testing"
	"time"
)

func TestNewRedis(t *testing.T) {
	r, err := NewRedis(nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if r != nil {
		t.Error("Expected nil, got", r)
	}
	conf := &RedisConfig{
		Client:           nil,
		KeyPrefix:        "",
		DependencyPrefix: "",
		LogPrefix:        "",
	}
	r, err = NewRedis(conf)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if r != nil {
		t.Error("Expected nil, got", r)
	}
	conf.Client = getRedisClient()
	r, err = NewRedis(conf)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if r != nil {
		t.Error("Expected nil, got", r)
	}
	conf.KeyPrefix = "test-key"
	r, err = NewRedis(conf)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if r != nil {
		t.Error("Expected nil, got", r)
	}
	conf.DependencyPrefix = "test-dep"
	r, err = NewRedis(conf)
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	if r == nil {
		t.Error("Expected *Redis, got nil")
	}

}

func TestRedis(t *testing.T) {
	fake, err := getFakeRedis()
	if err != nil {
		t.Fatal("Expected nil, got", err)
	}
	r, err := getRedis()
	if err != nil {
		t.Fatal("Expected nil, got", err)
	}

	dep1 := Dependency{
		Key:   "dep-1",
		Value: 0,
	}
	dep2 := Dependency{
		Key:   "dep-2",
		Value: 0,
	}
	//Fake
	err = fake.SetDependency(dep1, dep2)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	//Real
	err = r.SetDependency()
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	err = r.SetDependency(dep1, dep2)
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	//Fake
	deps, err := fake.GetDependencies(dep1.GetKey(), dep2.GetKey())
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if deps != nil {
		t.Error("Expected nil, got", deps)
	}
	//Real
	deps, err = r.GetDependencies(dep1.GetKey(), dep2.GetKey())
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	if !compareDeps(deps, []Dependency{dep1, dep2}) {
		t.Errorf("Expected %#v, got %#v", []Dependency{dep1, dep2}, deps)
	}
	duration := time.Second * 2
	key := "test-cahce-key"
	val := "my value"
	//Fake
	err = fake.Set(key, val, &duration, dep1, dep2)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	//Real
	err = r.Set(key, val, &duration, dep1, dep2)
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	//Fake
	value, deps, ok, err := fake.Get(key)
	if value != "" {
		t.Error("Expected \"\", got", value)
	}
	if deps != nil {
		t.Error("Expected nil, got", deps)
	}
	if ok {
		t.Error("Expected false, got true")
	}
	if err == nil {
		t.Error("Expected error, got nil")
	}
	//Real
	//Get cache data
	value, deps, ok, err = r.Get(key)
	if value != val {
		t.Error("Expected", val, "got", value)
	}
	if !compareDeps(deps, []Dependency{dep1, dep2}) {
		t.Errorf("Expected %#v, got %#v", []Dependency{dep1, dep2}, deps)
	}
	if !ok {
		t.Error("Expected true, got false")
	}
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	//wait ttl
	time.Sleep(time.Second * 3)
	//Get invalidate data
	value, deps, ok, err = r.Get(key)
	if value != "" {
		t.Error("Expected \"\", got", value)
	}
	if deps != nil {
		t.Error("Expected nil, got", deps)
	}
	if ok {
		t.Error("Expected false, got true")
	}
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	//Fake
	err = fake.Del(key)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	//Real delete missing key
	err = r.Del(key)
	if err != nil {
		t.Error("Expected nil, got", err)
	}

	err = r.Set(key, val, &duration, dep1, dep2)
	if err != nil {
		t.Error("Expected nil, got", err)
	}

	value, deps, ok, err = r.Get(key)
	if value != val {
		t.Error("Expected", val, "got", value)
	}
	if !compareDeps(deps, []Dependency{dep1, dep2}) {
		t.Errorf("Expected %#v, got %#v", []Dependency{dep1, dep2}, deps)
	}
	if !ok {
		t.Error("Expected true, got false")
	}
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	err = r.Del(key)
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	value, deps, ok, err = r.Get(key)
	if value != "" {
		t.Error("Expected \"\", got", value)
	}
	if deps != nil {
		t.Error("Expected nil, got", deps)
	}
	if ok {
		t.Error("Expected false, got true")
	}
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	err = r.Set(key, val, &duration, dep1, dep2)
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	//invalidate dependency
	//Fake
	err = fake.IncrDependency(dep1.GetKey())
	if err == nil {
		t.Error("Expected error, got nil")
	}
	//Real
	err = r.IncrDependency()
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	err = r.IncrDependency(dep1.GetKey(), dep2.GetKey())
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	value, deps, ok, err = r.Get(key)
	if value != "" {
		t.Error("Expected \"\", got", value)
	}
	if deps != nil {
		t.Error("Expected nil, got", deps)
	}
	if ok {
		t.Error("Expected false, got true")
	}
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	err = r.Set(key, val, &duration, dep1, dep2)
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	deps, err = r.GetDependencies(dep1.GetKey(), dep2.GetKey())
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	dep1.Value += 1
	dep2.Value += 1
	if !compareDeps(deps, []Dependency{dep1, dep2}) {
		t.Errorf("Expected %#v, got %#v", []Dependency{dep1, dep2}, deps)
	}
	//Fake
	err = fake.Clear()
	if err == nil {
		t.Error("Expected error, got nil")
	}
	err = r.Clear()
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	deps, err = r.GetDependencies(dep1.GetKey(), dep2.GetKey())
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	if len(deps) != 0 {
		t.Error("Expected 0, got", len(deps))
	}
}

func compareDeps(deps []Dependency, deps1 []Dependency) bool {
	m := make(map[string]int64)
	m1 := make(map[string]int64)
	for _, d := range deps {
		m[d.GetKey()] = d.GetValue()
	}
	for _, d := range deps1 {
		m1[d.GetKey()] = d.GetValue()
	}
	return reflect.DeepEqual(m, m1)
}

func getRedis() (ICache, error) {
	return NewRedis(&RedisConfig{
		Client:           getRedisClient(),
		KeyPrefix:        "test-key",
		DependencyPrefix: "test-dep",
		LogPrefix:        "[R]",
	})
}
func getFakeRedis() (ICache, error) {
	return NewRedis(&RedisConfig{
		Client:           getFakeRedisClient(),
		KeyPrefix:        "test-key",
		DependencyPrefix: "test-dep",
		LogPrefix:        "[R]",
	})
}

func getFakeRedisClient() redis.UniversalClient {
	return redis.NewClient(&redis.Options{
		Addr: "fake:6379",
	})
}

func getRedisClient() redis.UniversalClient {
	c := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "vagrant",
	})
	return c
}
