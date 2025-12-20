package core

import "testing"

type testGoTask struct {
	BaseTask
	ran *bool
}

func (t *testGoTask) Validate() error { return nil }
func (t *testGoTask) Execute(ctx *InstallContext, bus *EventBus) error {
	*t.ran = true
	return nil
}
func (t *testGoTask) Rollback(ctx *InstallContext, bus *EventBus) error { return nil }
func (t *testGoTask) CanRollback() bool                                 { return false }

func TestQueueConfigGoExtension(t *testing.T) {
	var ran bool
	name := "customGoTaskTest"
	_ = Tasks.Register(name, func(config map[string]any, ctx *InstallContext) (Task, error) {
		return &testGoTask{
			BaseTask: BaseTask{
				TaskID:   "go-task",
				TaskType: name,
			},
			ran: &ran,
		}, nil
	})

	ctx := NewInstallContext()
	bus := NewEventBus()
	runner := NewTaskRunner(ctx, bus)

	err := runner.QueueConfig(TaskConfig{
		Type:   "go:" + name,
		Params: map[string]any{},
	})
	if err != nil {
		t.Fatalf("QueueConfig failed: %v", err)
	}

	if err := runner.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !ran {
		t.Fatalf("expected task to run")
	}
}
