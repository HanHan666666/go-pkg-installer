package builtin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

func TestPermissionTaskMode(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(target, []byte("data"), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	ctx := core.NewInstallContext()
	ctx.Env.IsRoot = true
	bus := core.NewEventBus()

	task := &PermissionTask{
		BaseTask: core.BaseTask{
			TaskID:   "perm-test",
			TaskType: "permission",
		},
		Path:             target,
		Mode:             0600,
		RequirePrivilege: false,
	}

	if err := task.Execute(ctx, bus); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("expected mode 0600, got %v", info.Mode().Perm())
	}
}
