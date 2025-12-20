package builtin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

func TestRemoveDesktopEntryByName(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	appDir := filepath.Join(tmpDir, ".local", "share", "applications")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	target := filepath.Join(appDir, "demo-app.desktop")
	if err := os.WriteFile(target, []byte("content"), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	ctx := core.NewInstallContext()
	ctx.Env.IsRoot = true
	bus := core.NewEventBus()

	task := &RemoveDesktopEntryTask{
		BaseTask: core.BaseTask{
			TaskID:   "remove-desktop",
			TaskType: "removeDesktopEntry",
		},
		Name: "Demo App",
	}

	if err := task.Execute(ctx, bus); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected desktop entry to be removed")
	}
}
