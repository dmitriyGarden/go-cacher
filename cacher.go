package cacher

import "time"

//IDependency is an abstract dependency data
type IDependency interface {
	//GetKey return key of dependency counter
	GetKey() string
	//GetValue return value of dependency counter
	GetValue() int64
}

//Dependency is a dependency data
type Dependency struct {
	Key   string `json:"k"`
	Value int64  `json:"v"`
}

//GetKey return key of dependency counter
func (d Dependency) GetKey() string {
	return d.Key
}

//GetValue return value of dependency counter
func (d Dependency) GetValue() int64 {
	return d.Value
}

//ICache is an abstract cache, which describe methods
type ICache interface {
	//Set sets value of key with ttl and dependencies
	Set(key, value string, ttl *time.Duration, dependency ...IDependency) error
	//Get returns value of key
	Get(key string) (value string, dependencies []Dependency, ok bool, err error)
	//Del delete key from cache
	Del(key string) error
	//IncrDependency increment dependency counter
	IncrDependency(ttl *time.Duration, depKey ...string) error
	//GetDependencies returns dependency by key
	GetDependencies(depKey ...string) ([]Dependency, error)
	//SetDependency  sets value of dependency
	SetDependency(ttl *time.Duration, dependencies ...IDependency) error
	//Clear clear all cache
	Clear() error
}
