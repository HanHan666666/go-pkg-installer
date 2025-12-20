package builtin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

func TestRollbackTask(t *testing.T) {
	RegisterWriteConfigTask()
	RegisterRollbackTask()

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "rollback.txt")

	cfg := map[string]any{
		"type": "rollback",
		"tasks": []any{
			map[string]any{
				"type":    "writeConfig",
				"path":    target,
				"content": "rollback",
			},
		},
	}

	factory, ok := core.Tasks.Get("rollback")
	if !ok {
		t.Fatalf("rollback factory not registered")
	}

	task, err := factory(cfg, core.NewInstallContext())
	if err != nil {
		t.Fatalf("factory failed: %v", err)
	}

	if err := task.Validate(); err != nil {
		t.Fatalf("validate failed: %v", err)
	}

	if err := task.Execute(core.NewInstallContext(), core.NewEventBus()); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected file to not exist before rollback")
	}

	if err := task.Rollback(core.NewInstallContext(), core.NewEventBus()); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected file after rollback, got %v", err)
	}
}
