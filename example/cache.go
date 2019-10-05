package main

import (
	cacher "github.com/dmitriyGarden/go-cacher"
	"github.com/go-redis/redis"
	"log"
	"time"
)

func main() {
	// Single redis example
	cache := singleRedis()

	//Redis Sentinel
	//cache := redisSentinel()

	//Redis Cluster
	//cache := redisCluster()

	cacheExample(cache)
}

func redisCluster() cacher.ICache {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{":7000", ":7001", ":7002", ":7003", ":7004", ":7005"},
	})
	conf := getRedisConfig(client)
	c, err := cacher.NewRedis(conf)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func redisSentinel() cacher.ICache {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    "mymaster",
		SentinelAddrs: []string{":7000", ":7001", ":7002"},
		Password:      "",
	})
	conf := getRedisConfig(client)
	c, err := cacher.NewRedis(conf)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func singleRedis() cacher.ICache {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
	})
	conf := getRedisConfig(client)
	c, err := cacher.NewRedis(conf)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func getRedisConfig(client redis.UniversalClient) *cacher.RedisConfig {
	return &cacher.RedisConfig{
		Client:           client,
		KeyPrefix:        "test-key",
		DependencyPrefix: "test-dep",
		LogPrefix:        "[R]",
	}
}

func cacheExample(cache cacher.ICache) {
	//Create some dependencies
	dep1 := cacher.Dependency{
		Key:   "dep-1",
		Value: 0,
	}
	dep2 := cacher.Dependency{
		Key:   "dep-2",
		Value: 0,
	}

	//Put dependencies into storage
	err := cache.SetDependency(dep1, dep2)
	if err != nil {
		log.Fatal(err)
	}

	//Add data to cache
	duration := time.Duration(time.Hour * 5)
	err = cache.Set("my-key", "my value", &duration, dep1, dep2)
	if err != nil {
		log.Fatal(err)
	}

	// Do something ...

	//Get cache data
	val, deps, ok, err := cache.Get("my-key")
	if err != nil {
		log.Fatal(err)
	}

	//Check is data exists
	if ok {
		log.Println(val, deps)
	} else {
		log.Println("Key does not exist")
	}

	//Invalidate data by dependency dep1
	err = cache.IncrDependency(dep1.GetKey())
	if err != nil {
		log.Fatal(err)
	}

	//Delete key from cache
	err = cache.Del("my-key")
	if err != nil {
		log.Fatal(err)
	}

	//Invalidate all cache
	err = cache.Clear()
	if err != nil {
		log.Fatal(err)
	}
}
