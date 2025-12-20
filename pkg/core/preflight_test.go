package core

import "testing"

func TestDetectEnv(t *testing.T) {
	ctx := NewInstallContext()
	DetectEnv(ctx)

	if ctx.Env.Arch == "" {
		t.Fatalf("expected arch to be set")
	}
	if ctx.Env.Distro == "" {
		t.Fatalf("expected distro to be set")
	}
	if ctx.Env.Desktop == "" {
		t.Fatalf("expected desktop to be set")
	}
	if ctx.Env.DiskFreeMB < 0 {
		t.Fatalf("expected disk free to be non-negative")
	}
}
