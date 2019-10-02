package go_cacher

import "time"

type Dependency interface {
	//GetKey return key of dependency counter
	GetKey() string
	//GetValue return value of dependency counter
	GetValue() int64
}

type Cache interface {
	//Set sets value of key with ttl and dependencies
	Set(key, value string, ttl *time.Duration, dependencies ...Dependency) error
	//Get returns value of key
	Get(key string) (value string, ok bool, err error)
	//Del delete key from cache
	Del(key string) error
	//IncrDependency increment dependency counter
	IncrDependency(depKey ...string) error
	//SetDependency  sets value of dependency
	SetDependency(dependencies ...Dependency) error
	//Clear clear all cache
	Clear() error
}
