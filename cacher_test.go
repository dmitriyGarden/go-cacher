package cacher

import "testing"

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
