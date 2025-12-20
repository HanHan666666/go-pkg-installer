package builtin

import (
	"os/exec"
	"testing"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

func TestDbusServiceTaskExecute(t *testing.T) {
	prev := execCommand
	defer func() { execCommand = prev }()

	var got []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		got = append([]string{name}, args...)
		return exec.Command("sh", "-c", "exit 0")
	}

	ctx := core.NewInstallContext()
	ctx.Env.IsRoot = true
	bus := core.NewEventBus()

	task := &DbusServiceTask{
		BaseTask: core.BaseTask{
			TaskID:   "dbus-test",
			TaskType: "dbusService",
		},
		Name:   "demo",
		Action: "restart",
	}

	if err := task.Execute(ctx, bus); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	expected := []string{"systemctl", "restart", "demo.service"}
	if len(got) != len(expected) {
		t.Fatalf("expected %d args, got %d", len(expected), len(got))
	}
	for i, item := range expected {
		if got[i] != item {
			t.Fatalf("expected %q at %d, got %q", item, i, got[i])
		}
	}
}
