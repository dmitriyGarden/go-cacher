Consul leader election
[![Build Status](https://travis-ci.org/dmitriyGarden/go-cacher.svg?branch=master)](https://travis-ci.org/dmitriyGarden/go-cahcer)
[![codecov](https://codecov.io/gh/dmitriyGarden/go-cacher/branch/master/graph/badge.svg)](https://codecov.io/gh/dmitriyGarden/go-cacher)
======================

This package provides dependency cache

 How to use
 ==========
 
 
 See [example](https://github.com/dmitriyGarden/go-cacher/blob/master/example/cache.go)
 
 
 See [documentation](https://godoc.org/github.com/dmitriyGarden/go-cacher)
 
 It is safe to use in multiple goroutines
 
 
 Supports:
 
 
 - [Redis](https://redis.io)
 - [Redis sentinel](https://redis.io/topics/sentinel)
 - [Redis cluster](https://redis.io/topics/cluster-tutorial)
 
 
 TODO:
 
 
 - Memory
 - Memcached
 - MongoDB
 
 
 We use [go-redis/redis](https://github.com/go-redis/redis)
 

