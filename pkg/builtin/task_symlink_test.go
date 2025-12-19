package builtin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anthropics/go-pkg-installer/pkg/core"
)

func TestRegisterSymlinkTask(t *testing.T) {
	RegisterSymlinkTask()

	if !core.Tasks.Has("symlink") {
		t.Error("expected symlink task to be registered")
	}
}

func TestSymlinkTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    *SymlinkTask
		wantErr bool
	}{
		{
			name:    "valid",
			task:    &SymlinkTask{Target: "/usr/bin/app", LinkPath: "/usr/local/bin/app"},
			wantErr: false,
		},
		{
			name:    "missing target",
			task:    &SymlinkTask{LinkPath: "/usr/local/bin/app"},
			wantErr: true,
		},
		{
			name:    "missing link path",
			task:    &SymlinkTask{Target: "/usr/bin/app"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.task.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestSymlinkTaskExecute(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target.txt")
	link := filepath.Join(tmpDir, "link.txt")

	os.WriteFile(target, []byte("content"), 0644)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &SymlinkTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-symlink",
			TaskType: "symlink",
		},
		Target:   target,
		LinkPath: link,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify symlink was created
	linkTarget, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}
	if linkTarget != target {
		t.Errorf("expected target %q, got %q", target, linkTarget)
	}
}

func TestSymlinkTaskOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	target1 := filepath.Join(tmpDir, "target1.txt")
	target2 := filepath.Join(tmpDir, "target2.txt")
	link := filepath.Join(tmpDir, "link.txt")

	os.WriteFile(target1, []byte("content1"), 0644)
	os.WriteFile(target2, []byte("content2"), 0644)
	os.Symlink(target1, link)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	t.Run("without overwrite", func(t *testing.T) {
		task := &SymlinkTask{
			Target:    target2,
			LinkPath:  link,
			Overwrite: false,
		}

		err := task.Execute(ctx, bus)
		if err == nil {
			t.Error("expected error when link exists and overwrite is false")
		}
	})

	t.Run("with overwrite", func(t *testing.T) {
		task := &SymlinkTask{
			BaseTask: core.BaseTask{
				TaskID:   "test-symlink",
				TaskType: "symlink",
			},
			Target:    target2,
			LinkPath:  link,
			Overwrite: true,
		}

		err := task.Execute(ctx, bus)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		linkTarget, _ := os.Readlink(link)
		if linkTarget != target2 {
			t.Errorf("expected target %q, got %q", target2, linkTarget)
		}
	})
}

func TestSymlinkTaskRollback(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target.txt")
	link := filepath.Join(tmpDir, "link.txt")

	os.WriteFile(target, []byte("content"), 0644)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &SymlinkTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-symlink",
			TaskType: "symlink",
		},
		Target:   target,
		LinkPath: link,
	}

	// Execute
	if err := task.Execute(ctx, bus); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify can rollback
	if !task.CanRollback() {
		t.Error("expected CanRollback to return true")
	}

	// Rollback
	if err := task.Rollback(ctx, bus); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	// Link should be removed
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Error("symlink should have been removed during rollback")
	}
}

func TestSymlinkTaskRollbackRestoresOld(t *testing.T) {
	tmpDir := t.TempDir()
	target1 := filepath.Join(tmpDir, "target1.txt")
	target2 := filepath.Join(tmpDir, "target2.txt")
	link := filepath.Join(tmpDir, "link.txt")

	os.WriteFile(target1, []byte("content1"), 0644)
	os.WriteFile(target2, []byte("content2"), 0644)
	os.Symlink(target1, link)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &SymlinkTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-symlink",
			TaskType: "symlink",
		},
		Target:    target2,
		LinkPath:  link,
		Overwrite: true,
	}

	// Execute (overwrites old link)
	if err := task.Execute(ctx, bus); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Rollback should restore old link
	if err := task.Rollback(ctx, bus); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	// Link should point to original target
	linkTarget, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}
	if linkTarget != target1 {
		t.Errorf("expected rollback to restore target %q, got %q", target1, linkTarget)
	}
}

func TestSymlinkTaskCreatesParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target.txt")
	link := filepath.Join(tmpDir, "subdir", "link.txt")

	os.WriteFile(target, []byte("content"), 0644)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &SymlinkTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-symlink",
			TaskType: "symlink",
		},
		Target:   target,
		LinkPath: link,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify symlink was created
	if _, err := os.Lstat(link); err != nil {
		t.Errorf("symlink should exist: %v", err)
	}
}
