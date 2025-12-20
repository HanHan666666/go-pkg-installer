package core

import "testing"

func TestGetPrivilegeStrategy(t *testing.T) {
	ctx := NewInstallContext()
	ctx.Set("meta.privilegeStrategy", "sudo")
	if got := GetPrivilegeStrategy(ctx); got != PrivilegeSudo {
		t.Fatalf("expected %s, got %s", PrivilegeSudo, got)
	}
}

func TestEnsurePrivilege(t *testing.T) {
	ctx := NewInstallContext()
	ctx.Env.IsRoot = false
	if err := EnsurePrivilege(ctx, true); err == nil {
		t.Fatalf("expected error when not root")
	}

	ctx.Env.IsRoot = true
	if err := EnsurePrivilege(ctx, true); err != nil {
		t.Fatalf("expected no error when root, got %v", err)
	}
}

func TestNeedsPrivilege(t *testing.T) {
	cfg := &Config{
		Flows: map[string]*FlowConfig{
			"install": {
				Entry: "start",
				Steps: []*StepConfig{
					{
						ID:    "start",
						Title: "Start",
						Tasks: []TaskConfig{
							{
								Type:   "copy",
								Params: map[string]any{"requirePrivilege": true},
							},
							{
								Type: "systemdService",
							},
						},
					},
				},
			},
		},
	}

	if !NeedsPrivilege(cfg, "install") {
		t.Fatalf("expected privilege requirement to be detected")
	}
	if NeedsPrivilege(cfg, "missing") {
		t.Fatalf("expected false for unknown flow")
	}
}
