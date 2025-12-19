package builtin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anthropics/go-pkg-installer/pkg/core"
)

func TestRegisterDesktopEntryTask(t *testing.T) {
	RegisterDesktopEntryTask()

	if !core.Tasks.Has("desktopEntry") {
		t.Error("expected desktopEntry task to be registered")
	}
}

func TestDesktopEntryTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    *DesktopEntryTask
		wantErr bool
	}{
		{
			name:    "valid",
			task:    &DesktopEntryTask{Name: "My App", Exec: "/usr/bin/myapp"},
			wantErr: false,
		},
		{
			name:    "missing name",
			task:    &DesktopEntryTask{Exec: "/usr/bin/myapp"},
			wantErr: true,
		},
		{
			name:    "missing exec",
			task:    &DesktopEntryTask{Name: "My App"},
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

func TestDesktopEntryTaskExecute(t *testing.T) {
	tmpDir := t.TempDir()
	desktopFile := filepath.Join(tmpDir, "myapp.desktop")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &DesktopEntryTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-desktop",
			TaskType: "desktopEntry",
		},
		Name:        "My App",
		Exec:        "/usr/bin/myapp %U",
		Icon:        "/usr/share/icons/myapp.png",
		Comment:     "A test application",
		Categories:  []string{"Development", "Utility"},
		Terminal:    false,
		EntryType:   "Application",
		MimeTypes:   []string{"text/plain"},
		Keywords:    []string{"app", "test"},
		Destination: desktopFile,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(desktopFile)
	if err != nil {
		t.Fatalf("failed to read desktop file: %v", err)
	}

	content := string(data)

	// Check required fields
	if !strings.Contains(content, "[Desktop Entry]") {
		t.Error("missing [Desktop Entry] header")
	}
	if !strings.Contains(content, "Name=My App") {
		t.Error("missing Name field")
	}
	if !strings.Contains(content, "Exec=/usr/bin/myapp %U") {
		t.Error("missing Exec field")
	}
	if !strings.Contains(content, "Icon=/usr/share/icons/myapp.png") {
		t.Error("missing Icon field")
	}
	if !strings.Contains(content, "Comment=A test application") {
		t.Error("missing Comment field")
	}
	if !strings.Contains(content, "Categories=Development;Utility;") {
		t.Error("missing or incorrect Categories field")
	}
	if !strings.Contains(content, "Terminal=false") {
		t.Error("missing Terminal field")
	}
	if !strings.Contains(content, "Type=Application") {
		t.Error("missing Type field")
	}
	if !strings.Contains(content, "MimeType=text/plain;") {
		t.Error("missing MimeType field")
	}
	if !strings.Contains(content, "Keywords=app;test;") {
		t.Error("missing Keywords field")
	}
}

func TestDesktopEntryTaskMinimal(t *testing.T) {
	tmpDir := t.TempDir()
	desktopFile := filepath.Join(tmpDir, "minimal.desktop")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &DesktopEntryTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-desktop",
			TaskType: "desktopEntry",
		},
		Name:        "Minimal App",
		Exec:        "/usr/bin/minimal",
		EntryType:   "Application",
		Destination: desktopFile,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	data, _ := os.ReadFile(desktopFile)
	content := string(data)

	if !strings.Contains(content, "Name=Minimal App") {
		t.Error("missing Name field")
	}
	if !strings.Contains(content, "Exec=/usr/bin/minimal") {
		t.Error("missing Exec field")
	}
}

func TestDesktopEntryTaskWithTerminal(t *testing.T) {
	tmpDir := t.TempDir()
	desktopFile := filepath.Join(tmpDir, "terminal.desktop")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &DesktopEntryTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-desktop",
			TaskType: "desktopEntry",
		},
		Name:        "Terminal App",
		Exec:        "/usr/bin/termapp",
		Terminal:    true,
		EntryType:   "Application",
		Destination: desktopFile,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	data, _ := os.ReadFile(desktopFile)
	if !strings.Contains(string(data), "Terminal=true") {
		t.Error("missing Terminal=true field")
	}
}

func TestDesktopEntryTaskRollback(t *testing.T) {
	tmpDir := t.TempDir()
	desktopFile := filepath.Join(tmpDir, "myapp.desktop")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &DesktopEntryTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-desktop",
			TaskType: "desktopEntry",
		},
		Name:        "My App",
		Exec:        "/usr/bin/myapp",
		EntryType:   "Application",
		Destination: desktopFile,
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

	// File should be removed
	if _, err := os.Stat(desktopFile); !os.IsNotExist(err) {
		t.Error("desktop file should have been removed during rollback")
	}
}

func TestDesktopEntryTaskFactory(t *testing.T) {
	RegisterDesktopEntryTask()

	ctx := core.NewInstallContext()
	ctx.Set("appPath", "/opt/myapp/bin/myapp")
	ctx.Set("iconPath", "/opt/myapp/icon.png")

	factory, ok := core.Tasks.Get("desktopEntry")
	if !ok {
		t.Fatal("desktopEntry factory not registered")
	}

	tmpDir := t.TempDir()
	task, err := factory(map[string]any{
		"id":          "desktop-1",
		"name":        "My Application",
		"exec":        "{{.appPath}}",
		"icon":        "{{.iconPath}}",
		"categories":  []interface{}{"Development"},
		"destination": filepath.Join(tmpDir, "myapp.desktop"),
	}, ctx)

	if err != nil {
		t.Fatalf("factory error = %v", err)
	}

	deTask, ok := task.(*DesktopEntryTask)
	if !ok {
		t.Fatal("expected *DesktopEntryTask")
	}

	if deTask.Exec != "/opt/myapp/bin/myapp" {
		t.Errorf("expected exec to be rendered, got %q", deTask.Exec)
	}
	if deTask.Icon != "/opt/myapp/icon.png" {
		t.Errorf("expected icon to be rendered, got %q", deTask.Icon)
	}
}

func TestDesktopEntryTaskCreatesParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	desktopFile := filepath.Join(tmpDir, "applications", "myapp.desktop")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &DesktopEntryTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-desktop",
			TaskType: "desktopEntry",
		},
		Name:        "My App",
		Exec:        "/usr/bin/myapp",
		EntryType:   "Application",
		Destination: desktopFile,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if _, err := os.Stat(desktopFile); err != nil {
		t.Errorf("desktop file should exist: %v", err)
	}
}
