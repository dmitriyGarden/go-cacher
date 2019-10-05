package go_cacher

import "time"

type IDependency interface {
	//GetKey return key of dependency counter
	GetKey() string
	//GetValue return value of dependency counter
	GetValue() int64
}

type Dependency struct {
	Key   string
	Value int64
}

//GetKey return key of dependency counter
func (d Dependency) GetKey() string {
	return d.Key
}

//GetValue return value of dependency counter
func (d Dependency) GetValue() int64 {
	return d.Value
}

type ICache interface {
	//Set sets value of key with ttl and dependencies
	Set(key, value string, ttl *time.Duration, dependency ...IDependency) error
	//Get returns value of key
	Get(key string) (value string, dependencies []Dependency, ok bool, err error)
	//Del delete key from cache
	Del(key string) error
	//IncrDependency increment dependency counter
	IncrDependency(depKey ...string) error
	//GetDependencies returns dependency by key
	GetDependencies(depKey ...string) ([]Dependency, error)
	//SetDependency  sets value of dependency
	SetDependency(dependencies ...IDependency) error
	//Clear clear all cache
	Clear() error
}
