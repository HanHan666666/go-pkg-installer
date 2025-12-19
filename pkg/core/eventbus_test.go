package core

import (
	"sync"
	"testing"
)

func TestNewEventBus(t *testing.T) {
	eb := NewEventBus()
	if eb == nil {
		t.Fatal("NewEventBus returned nil")
	}
	if eb.handlers == nil {
		t.Error("handlers map should be initialized")
	}
}

func TestEventBusSubscribeAndPublish(t *testing.T) {
	eb := NewEventBus()
	received := false
	var receivedEvent Event

	eb.Subscribe(EventProgress, func(event Event) {
		received = true
		receivedEvent = event
	})

	eb.Publish(Event{
		Type: EventProgress,
		Payload: ProgressPayload{
			TaskID:   "task1",
			Progress: 0.5,
			Message:  "Halfway done",
		},
	})

	if !received {
		t.Error("Handler should have been called")
	}
	if receivedEvent.Type != EventProgress {
		t.Errorf("Expected EventProgress, got %v", receivedEvent.Type)
	}

	payload, ok := receivedEvent.Payload.(ProgressPayload)
	if !ok {
		t.Fatal("Payload should be ProgressPayload")
	}
	if payload.TaskID != "task1" {
		t.Errorf("Expected task1, got %s", payload.TaskID)
	}
	if payload.Progress != 0.5 {
		t.Errorf("Expected 0.5, got %f", payload.Progress)
	}
}

func TestEventBusMultipleSubscribers(t *testing.T) {
	eb := NewEventBus()
	count := 0
	var mu sync.Mutex

	for i := 0; i < 3; i++ {
		eb.Subscribe(EventLog, func(event Event) {
			mu.Lock()
			count++
			mu.Unlock()
		})
	}

	eb.Publish(Event{Type: EventLog})

	if count != 3 {
		t.Errorf("Expected 3 handlers to be called, got %d", count)
	}
}

func TestEventBusSubscribeAll(t *testing.T) {
	eb := NewEventBus()
	count := 0

	eb.SubscribeAll(func(event Event) {
		count++
	})

	eb.Publish(Event{Type: EventProgress})
	eb.Publish(Event{Type: EventLog})
	eb.Publish(Event{Type: EventStepChange})

	if count != 3 {
		t.Errorf("Expected 3 events, got %d", count)
	}
}

func TestEventBusNoMatchingHandler(t *testing.T) {
	eb := NewEventBus()
	called := false

	eb.Subscribe(EventProgress, func(event Event) {
		called = true
	})

	// Publish a different event type
	eb.Publish(Event{Type: EventLog})

	if called {
		t.Error("Handler should not be called for different event type")
	}
}

func TestEventBusPublishProgress(t *testing.T) {
	eb := NewEventBus()
	var received ProgressPayload

	eb.Subscribe(EventProgress, func(event Event) {
		received = event.Payload.(ProgressPayload)
	})

	eb.PublishProgress("download", 0.75, "Downloading...")

	if received.TaskID != "download" {
		t.Errorf("Expected 'download', got %s", received.TaskID)
	}
	if received.Progress != 0.75 {
		t.Errorf("Expected 0.75, got %f", received.Progress)
	}
	if received.Message != "Downloading..." {
		t.Errorf("Expected 'Downloading...', got %s", received.Message)
	}
}

func TestEventBusPublishLog(t *testing.T) {
	eb := NewEventBus()
	var received LogPayload

	eb.Subscribe(EventLog, func(event Event) {
		received = event.Payload.(LogPayload)
	})

	eb.PublishLog(LogWarn, "Low disk space")

	if received.Level != LogWarn {
		t.Errorf("Expected LogWarn, got %v", received.Level)
	}
	if received.Message != "Low disk space" {
		t.Errorf("Expected 'Low disk space', got %s", received.Message)
	}
}

func TestEventBusPublishStepChange(t *testing.T) {
	eb := NewEventBus()
	var received StepChangePayload

	eb.Subscribe(EventStepChange, func(event Event) {
		received = event.Payload.(StepChangePayload)
	})

	eb.PublishStepChange("welcome", "license")

	if received.FromStep != "welcome" {
		t.Errorf("Expected 'welcome', got %s", received.FromStep)
	}
	if received.ToStep != "license" {
		t.Errorf("Expected 'license', got %s", received.ToStep)
	}
}

func TestEventBusPublishTaskEvents(t *testing.T) {
	eb := NewEventBus()
	startCalled := false
	completeCalled := false
	errorCalled := false

	eb.Subscribe(EventTaskStart, func(event Event) {
		startCalled = true
	})
	eb.Subscribe(EventTaskComplete, func(event Event) {
		completeCalled = true
	})
	eb.Subscribe(EventTaskError, func(event Event) {
		errorCalled = true
	})

	eb.PublishTaskStart("task1", "download")
	eb.PublishTaskComplete("task1", "download")
	eb.PublishTaskError("task2", "unpack", nil)

	if !startCalled {
		t.Error("TaskStart handler should be called")
	}
	if !completeCalled {
		t.Error("TaskComplete handler should be called")
	}
	if !errorCalled {
		t.Error("TaskError handler should be called")
	}
}

func TestEventBusClear(t *testing.T) {
	eb := NewEventBus()
	called := false

	eb.Subscribe(EventProgress, func(event Event) {
		called = true
	})
	eb.SubscribeAll(func(event Event) {
		called = true
	})

	eb.Clear()
	eb.Publish(Event{Type: EventProgress})

	if called {
		t.Error("Handlers should be cleared")
	}
}

func TestEventBusThreadSafety(t *testing.T) {
	eb := NewEventBus()
	var wg sync.WaitGroup
	count := 0
	var mu sync.Mutex

	eb.Subscribe(EventProgress, func(event Event) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	// Concurrent publishes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eb.Publish(Event{Type: EventProgress})
		}()
	}

	// Concurrent subscribes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eb.Subscribe(EventLog, func(event Event) {})
		}()
	}

	wg.Wait()

	if count < 100 {
		t.Errorf("Expected at least 100 events handled, got %d", count)
	}
}
