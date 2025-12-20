package builtin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

func TestRemovePathUserDataSkip(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "data.txt")
	if err := os.WriteFile(target, []byte("data"), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	ctx := core.NewInstallContext()
	ctx.Env.IsRoot = true
	ctx.Set("uninstall.keepUserData", true)
	bus := core.NewEventBus()

	task := &RemovePathTask{
		BaseTask: core.BaseTask{
			TaskID:   "remove-userdata",
			TaskType: "removePath",
		},
		Path:     target,
		UserData: true,
	}

	if err := task.Execute(ctx, bus); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected file to remain, got %v", err)
	}
}
