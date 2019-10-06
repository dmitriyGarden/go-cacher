package cacher

import (
	"testing"
	"time"
)

func TestDependency_GetKey(t *testing.T) {
	d := Dependency{
		Key:   "dep-key",
		Value: 0,
	}
	if d.GetKey() != "dep-key" {
		t.Error("Expected dep-key, got", d.GetKey())
	}
}

func TestDependency_GetValue(t *testing.T) {
	d := Dependency{
		Key:   "",
		Value: 100,
	}
	if d.GetValue() != 100 {
		t.Error("Expected 100, got", d.GetValue())
	}
}

func testFakeCache(t *testing.T, fake ICache) {

	dep1 := Dependency{
		Key:   "dep-1",
		Value: 0,
	}
	dep2 := Dependency{
		Key:   "dep-2",
		Value: 0,
	}

	err := fake.SetDependency(nil, dep1, dep2)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	deps, err := fake.GetDependencies(dep1.GetKey(), dep2.GetKey())
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if deps != nil {
		t.Error("Expected nil, got", deps)
	}
	duration := time.Second * 2
	key := "test-cahce-key"
	val := "my value"
	err = fake.Set(key, val, &duration, dep1, dep2)
	if err == nil {
		t.Error("Expected error, got nil")
	}
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
	err = fake.Del(key)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	err = fake.IncrDependency(nil, dep1.GetKey())
	if err == nil {
		t.Error("Expected error, got nil")
	}
	err = fake.Clear()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func testCache(t *testing.T, r ICache) {
	t.Log("Start test cache")
	dep1 := Dependency{
		Key:   "dep-1",
		Value: 0,
	}
	dep2 := Dependency{
		Key:   "dep-2",
		Value: 0,
	}
	t.Log("Set empty dependencies")
	err := r.SetDependency(nil)
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	t.Log("Set dependencies")
	ttl := time.Second * 10
	err = r.SetDependency(&ttl, dep1, dep2)
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	t.Log("Get dependencies")
	deps, err := r.GetDependencies(dep1.GetKey(), dep2.GetKey())
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	if !compareDeps(deps, []Dependency{dep1, dep2}) {
		t.Errorf("Expected %#v, got %#v", []Dependency{dep1, dep2}, deps)
	}
	duration := time.Second * 2
	key := "test-cahce-key"
	val := "my value"

	t.Log("Set cache data")
	err = r.Set(key, val, &duration, dep1, dep2)
	if err != nil {
		t.Error("Expected nil, got", err)
	}

	t.Log("Get cache data")
	value, deps, ok, err := r.Get(key)
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
	t.Log("Wait for ttl")
	time.Sleep(time.Second * 3)
	//Get invalidate data
	t.Log("Get invalidated data")
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

	t.Log("Delete missing key")
	err = r.Del(key)
	if err != nil {
		t.Error("Expected nil, got", err)
	}

	t.Log("Set cache data")
	err = r.Set(key, val, &duration, dep1, dep2)
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	t.Log("Get cache data")
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
	t.Log("Delete cache data")
	err = r.Del(key)
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	t.Log("Get missing cache data")
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
	t.Log("Set cache data")
	err = r.Set(key, val, &duration, dep1, dep2)
	if err != nil {
		t.Error("Expected nil, got", err)
	}

	t.Log("Invalidate empty dependency")
	err = r.IncrDependency(nil)
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	t.Log("Invalidate dependency")
	err = r.IncrDependency(&duration, dep1.GetKey(), dep2.GetKey())
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	t.Log("Get invalidated cache data")
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
	t.Log("Set cache data")
	err = r.Set(key, val, &duration, dep1, dep2)
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	t.Log("Get dependencies")
	deps, err = r.GetDependencies(dep1.GetKey(), dep2.GetKey())
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	dep1.Value++
	dep2.Value++
	if !compareDeps(deps, []Dependency{dep1, dep2}) {
		t.Errorf("Expected %#v, got %#v", []Dependency{dep1, dep2}, deps)
	}

	t.Log("Delete whole cache")
	err = r.Clear()
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	t.Log("Get  missing dependencies")
	deps, err = r.GetDependencies(dep1.GetKey(), dep2.GetKey())
	if err != nil {
		t.Error("Expected nil, got", err)
	}
	if len(deps) != 0 {
		t.Error("Expected 0, got", len(deps))
	}
}
