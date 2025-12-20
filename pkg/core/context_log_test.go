package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAddLogWritesAndPublishes(t *testing.T) {
	ctx := NewInstallContext()
	bus := NewEventBus()
	ctx.SetEventBus(bus)

	logPath := filepath.Join(t.TempDir(), "installer.log")
	if err := ctx.SetLogFile(logPath); err != nil {
		t.Fatalf("failed to set log file: %v", err)
	}

	done := make(chan LogPayload, 1)
	bus.Subscribe(EventLog, func(e Event) {
		if payload := e.LogPayload(); payload != nil {
			done <- *payload
		}
	})

	ctx.AddLog(LogWarn, "test message")

	select {
	case payload := <-done:
		if payload.Level != LogWarn || payload.Message != "test message" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected log event")
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("expected log file to contain data")
	}
}
