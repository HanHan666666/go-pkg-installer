package core

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry[string]("TestRegistry")
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if r.name != "TestRegistry" {
		t.Errorf("Expected name 'TestRegistry', got %s", r.name)
	}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry[string]("TestRegistry")

	err := r.Register("key1", "value1")
	if err != nil {
		t.Errorf("Register should not return error: %v", err)
	}

	val, ok := r.Get("key1")
	if !ok {
		t.Error("Get should return true for existing key")
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', got %s", val)
	}
}

func TestRegistryDuplicateKey(t *testing.T) {
	r := NewRegistry[string]("TestRegistry")

	_ = r.Register("key1", "value1")
	err := r.Register("key1", "value2")

	if err == nil {
		t.Error("Register should return error for duplicate key")
	}
}

func TestRegistryMustRegister(t *testing.T) {
	r := NewRegistry[string]("TestRegistry")

	// Should not panic
	r.MustRegister("key1", "value1")

	// Should panic on duplicate
	defer func() {
		if recover() == nil {
			t.Error("MustRegister should panic on duplicate key")
		}
	}()
	r.MustRegister("key1", "value2")
}

func TestRegistryGetNonexistent(t *testing.T) {
	r := NewRegistry[string]("TestRegistry")

	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("Get should return false for nonexistent key")
	}
}

func TestRegistryMustGet(t *testing.T) {
	r := NewRegistry[string]("TestRegistry")
	r.MustRegister("key1", "value1")

	// Should not panic
	val := r.MustGet("key1")
	if val != "value1" {
		t.Errorf("Expected 'value1', got %s", val)
	}

	// Should panic on nonexistent
	defer func() {
		if recover() == nil {
			t.Error("MustGet should panic on nonexistent key")
		}
	}()
	_ = r.MustGet("nonexistent")
}

func TestRegistryHas(t *testing.T) {
	r := NewRegistry[string]("TestRegistry")
	r.MustRegister("key1", "value1")

	if !r.Has("key1") {
		t.Error("Has should return true for existing key")
	}
	if r.Has("nonexistent") {
		t.Error("Has should return false for nonexistent key")
	}
}

func TestRegistryKeys(t *testing.T) {
	r := NewRegistry[string]("TestRegistry")
	r.MustRegister("key1", "value1")
	r.MustRegister("key2", "value2")
	r.MustRegister("key3", "value3")

	keys := r.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Check all keys are present (order not guaranteed)
	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}
	for _, expected := range []string{"key1", "key2", "key3"} {
		if !keySet[expected] {
			t.Errorf("Expected key %s in keys", expected)
		}
	}
}

func TestRegistryClear(t *testing.T) {
	r := NewRegistry[string]("TestRegistry")
	r.MustRegister("key1", "value1")
	r.MustRegister("key2", "value2")

	r.Clear()

	if r.Has("key1") || r.Has("key2") {
		t.Error("Registry should be empty after Clear")
	}
	if len(r.Keys()) != 0 {
		t.Error("Keys should be empty after Clear")
	}
}

func TestIsGoExtension(t *testing.T) {
	tests := []struct {
		typeName string
		expected bool
	}{
		{"go:customTask", true},
		{"go:myScreen", true},
		{"download", false},
		{"license", false},
		{"golang:test", false},
		{"GO:upper", false}, // Case sensitive
	}

	for _, tt := range tests {
		result := IsGoExtension(tt.typeName)
		if result != tt.expected {
			t.Errorf("IsGoExtension(%q) = %v, want %v", tt.typeName, result, tt.expected)
		}
	}
}

func TestStripGoPrefix(t *testing.T) {
	tests := []struct {
		typeName string
		expected string
	}{
		{"go:customTask", "customTask"},
		{"go:myScreen", "myScreen"},
		{"download", "download"},
		{"go:", ""},
	}

	for _, tt := range tests {
		result := StripGoPrefix(tt.typeName)
		if result != tt.expected {
			t.Errorf("StripGoPrefix(%q) = %q, want %q", tt.typeName, result, tt.expected)
		}
	}
}

func TestTaskFactoryRegistry(t *testing.T) {
	// Clear global registry for test
	Tasks.Clear()

	factory := func(config map[string]any, ctx *InstallContext) (Task, error) {
		return nil, nil // Mock factory
	}

	err := Tasks.Register("download", factory)
	if err != nil {
		t.Errorf("Register task factory should not error: %v", err)
	}

	if !Tasks.Has("download") {
		t.Error("download task should be registered")
	}

	_, ok := Tasks.Get("download")
	if !ok {
		t.Error("Should get download factory")
	}
}

func TestGuardFactoryRegistry(t *testing.T) {
	// Clear global registry for test
	Guards.Clear()

	factory := func(config map[string]any) (Guard, error) {
		return nil, nil // Mock factory
	}

	err := Guards.Register("mustAccept", factory)
	if err != nil {
		t.Errorf("Register guard factory should not error: %v", err)
	}

	if !Guards.Has("mustAccept") {
		t.Error("mustAccept guard should be registered")
	}
}

func TestScreenFactoryRegistry(t *testing.T) {
	// Clear global registry for test
	Screens.Clear()

	factory := func(config map[string]any) (Screen, error) {
		return nil, nil // Mock factory
	}

	err := Screens.Register("license", factory)
	if err != nil {
		t.Errorf("Register screen factory should not error: %v", err)
	}

	if !Screens.Has("license") {
		t.Error("license screen should be registered")
	}
}

func TestRegistryThreadSafety(t *testing.T) {
	r := NewRegistry[int]("TestRegistry")
	done := make(chan bool)

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			key := string(rune('a' + i%26))
			_ = r.Register(key, i)
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = r.Has("a")
			_, _ = r.Get("b")
		}
		done <- true
	}()

	<-done
	<-done
}
