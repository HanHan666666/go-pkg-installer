package core

import (
	"testing"
)

func TestNewInstallContext(t *testing.T) {
	ctx := NewInstallContext()

	if ctx == nil {
		t.Fatal("NewInstallContext returned nil")
	}
	if ctx.UserInput == nil {
		t.Error("UserInput should be initialized")
	}
	if ctx.Meta == nil {
		t.Error("Meta should be initialized")
	}
	if ctx.Runtime.Logs == nil {
		t.Error("Runtime.Logs should be initialized")
	}
}

func TestContextSetAndGet(t *testing.T) {
	ctx := NewInstallContext()

	// Test simple set/get
	ctx.Set("install.dir", "/opt/myapp")
	val, ok := ctx.Get("install.dir")
	if !ok {
		t.Error("Get should return true for existing key")
	}
	if val != "/opt/myapp" {
		t.Errorf("Expected '/opt/myapp', got %v", val)
	}

	// Test GetString
	s := ctx.GetString("install.dir")
	if s != "/opt/myapp" {
		t.Errorf("GetString expected '/opt/myapp', got %s", s)
	}
}

func TestContextNestedPath(t *testing.T) {
	ctx := NewInstallContext()

	// Set nested value
	ctx.Set("license.accepted", true)
	ctx.Set("options.desktop.shortcut", true)
	ctx.Set("options.desktop.autostart", false)

	// Verify nested structure was created
	if !ctx.GetBool("license.accepted") {
		t.Error("license.accepted should be true")
	}
	if !ctx.GetBool("options.desktop.shortcut") {
		t.Error("options.desktop.shortcut should be true")
	}
	if ctx.GetBool("options.desktop.autostart") {
		t.Error("options.desktop.autostart should be false")
	}
}

func TestContextGetNonexistent(t *testing.T) {
	ctx := NewInstallContext()

	_, ok := ctx.Get("nonexistent.path")
	if ok {
		t.Error("Get should return false for nonexistent key")
	}

	s := ctx.GetString("nonexistent")
	if s != "" {
		t.Errorf("GetString should return empty string for nonexistent key, got %s", s)
	}

	b := ctx.GetBool("nonexistent")
	if b {
		t.Error("GetBool should return false for nonexistent key")
	}

	i := ctx.GetInt("nonexistent")
	if i != 0 {
		t.Errorf("GetInt should return 0 for nonexistent key, got %d", i)
	}
}

func TestContextMeta(t *testing.T) {
	ctx := NewInstallContext()

	ctx.SetMeta("payloadUrl", "https://example.com/app.tar.gz")
	ctx.SetMeta("payloadSha256", "abc123")

	url := ctx.GetString("payloadUrl")
	if url != "https://example.com/app.tar.gz" {
		t.Errorf("Expected payload URL, got %s", url)
	}
}

func TestContextEnvFields(t *testing.T) {
	ctx := NewInstallContext()
	ctx.Env = EnvInfo{
		Distro:        "ubuntu",
		DistroVersion: "22.04",
		Arch:          "x86_64",
		Desktop:       "gnome",
		IsRoot:        false,
		HasPolkit:     true,
		HasSudo:       true,
		DiskFreeMB:    50000,
	}

	// Test direct field access
	distro := ctx.GetString("distro")
	if distro != "ubuntu" {
		t.Errorf("Expected 'ubuntu', got %s", distro)
	}

	// Test with env. prefix
	distro2 := ctx.GetString("env.distro")
	if distro2 != "ubuntu" {
		t.Errorf("Expected 'ubuntu' with env prefix, got %s", distro2)
	}

	// Test boolean field
	if ctx.GetBool("isRoot") {
		t.Error("isRoot should be false")
	}
	if !ctx.GetBool("hasPolkit") {
		t.Error("hasPolkit should be true")
	}

	// Test numeric field
	diskFree := ctx.GetInt("diskFreeMB")
	if diskFree != 50000 {
		t.Errorf("Expected 50000, got %d", diskFree)
	}
}

func TestContextRender(t *testing.T) {
	ctx := NewInstallContext()
	ctx.Set("install.dir", "/opt/myapp")
	ctx.SetMeta("version", "1.0.0")
	ctx.Env.Arch = "x86_64"

	tests := []struct {
		template string
		expected string
	}{
		{
			template: "Installing to ${install.dir}",
			expected: "Installing to /opt/myapp",
		},
		{
			template: "Version: ${version}, Arch: ${arch}",
			expected: "Version: 1.0.0, Arch: x86_64",
		},
		{
			template: "No placeholder here",
			expected: "No placeholder here",
		},
		{
			template: "Unknown: ${unknown.path}",
			expected: "Unknown: ${unknown.path}",
		},
		{
			template: "${install.dir}/bin/app",
			expected: "/opt/myapp/bin/app",
		},
	}

	for _, tt := range tests {
		result := ctx.Render(tt.template)
		if result != tt.expected {
			t.Errorf("Render(%q) = %q, want %q", tt.template, result, tt.expected)
		}
	}
}

func TestContextGetInt(t *testing.T) {
	ctx := NewInstallContext()

	ctx.Set("intVal", 42)
	ctx.Set("int64Val", int64(100))
	ctx.Set("floatVal", 3.14)
	ctx.Set("stringVal", "hello")

	if ctx.GetInt("intVal") != 42 {
		t.Error("GetInt failed for int value")
	}
	if ctx.GetInt("int64Val") != 100 {
		t.Error("GetInt failed for int64 value")
	}
	if ctx.GetInt("floatVal") != 3 {
		t.Error("GetInt failed for float value")
	}
	if ctx.GetInt("stringVal") != 0 {
		t.Error("GetInt should return 0 for string value")
	}
}

func TestContextLogging(t *testing.T) {
	ctx := NewInstallContext()

	ctx.AddLog(LogInfo, "Starting installation")
	ctx.AddLog(LogWarn, "Low disk space")
	ctx.AddLog(LogError, "Failed to copy file")

	if len(ctx.Runtime.Logs) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(ctx.Runtime.Logs))
	}
	if ctx.Runtime.Logs[0].Level != LogInfo {
		t.Error("First log should be INFO")
	}
	if ctx.Runtime.Logs[0].Message != "Starting installation" {
		t.Error("First log message mismatch")
	}
}

func TestContextErrors(t *testing.T) {
	ctx := NewInstallContext()

	ctx.AddError(nil)
	if len(ctx.Runtime.Errors) != 1 {
		t.Errorf("Expected 1 error entry, got %d", len(ctx.Runtime.Errors))
	}
}

func TestContextProgress(t *testing.T) {
	ctx := NewInstallContext()

	ctx.SetProgress(0.5)
	if ctx.Runtime.Progress != 0.5 {
		t.Errorf("Expected progress 0.5, got %f", ctx.Runtime.Progress)
	}

	ctx.SetProgress(1.0)
	if ctx.Runtime.Progress != 1.0 {
		t.Errorf("Expected progress 1.0, got %f", ctx.Runtime.Progress)
	}
}

func TestContextThreadSafety(t *testing.T) {
	ctx := NewInstallContext()
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			ctx.Set("counter", i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = ctx.GetInt("counter")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			ctx.AddLog(LogInfo, "test")
		}
		done <- true
	}()

	<-done
	<-done
	<-done
}
