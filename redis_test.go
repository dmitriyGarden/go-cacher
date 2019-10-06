package cacher

import (
	"github.com/go-redis/redis"
	"reflect"
	"testing"
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

func TestFakeRedis(t *testing.T) {
	fake, err := getFakeRedis()
	if err != nil {
		t.Fatal("Expected nil, got", err)
	}
	testFakeCache(t, fake)
}

func TestRedis(t *testing.T) {
	r, err := getRedis(false)
	if err != nil {
		t.Fatal("Expected nil, got", err)
	}
	testCache(t, r)
}

func TestCompressedRedis(t *testing.T) {
	r, err := getRedis(true)
	if err != nil {
		t.Fatal("Expected nil, got", err)
	}
	testCache(t, r)
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

func getRedis(compress bool) (ICache, error) {
	return NewRedis(&RedisConfig{
		Client:             getRedisClient(),
		KeyPrefix:          "test-key",
		DependencyPrefix:   "test-dep",
		LogPrefix:          "[R]",
		EnableDataCompress: compress,
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
