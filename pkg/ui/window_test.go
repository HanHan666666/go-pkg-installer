package ui

import (
	"testing"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

// Note: Full UI testing would require a display and running Tk.
// These tests focus on the data structures and factory registration.

func TestNewInstallerWindow(t *testing.T) {
	ctx := core.NewInstallContext()
	bus := core.NewEventBus()
	workflow := core.NewWorkflow(ctx, bus)

	window := NewInstallerWindow(ctx, workflow, bus)

	if window == nil {
		t.Fatal("NewInstallerWindow returned nil")
	}

	if window.ctx != ctx {
		t.Error("context not set correctly")
	}

	if window.bus != bus {
		t.Error("event bus not set correctly")
	}

	if window.workflow != workflow {
		t.Error("workflow not set correctly")
	}
}

func TestInstallerWindowScreenRenderers(t *testing.T) {
	ctx := core.NewInstallContext()
	bus := core.NewEventBus()
	workflow := core.NewWorkflow(ctx, bus)

	window := NewInstallerWindow(ctx, workflow, bus)

	// Check built-in screen renderers are registered
	renderers := []string{
		"welcome",
		"license",
		"directory",
		"pathPicker",
		"progress",
		"richtext",
		"summary",
		"form",
		"options",
		"finish",
	}

	for _, screenType := range renderers {
		if _, ok := window.screenRenderers[screenType]; !ok {
			t.Errorf("screen renderer '%s' not registered", screenType)
		}
	}
}

func TestRegisterScreenRenderer(t *testing.T) {
	ctx := core.NewInstallContext()
	bus := core.NewEventBus()
	workflow := core.NewWorkflow(ctx, bus)

	window := NewInstallerWindow(ctx, workflow, bus)

	// Register a custom screen renderer
	customFactory := func(step *core.StepConfig) ScreenRenderer {
		return &WelcomeScreen{step: step}
	}

	window.RegisterScreenRenderer("custom", customFactory)

	if _, ok := window.screenRenderers["custom"]; !ok {
		t.Error("custom screen renderer not registered")
	}
}

func TestOnCompleteCallback(t *testing.T) {
	ctx := core.NewInstallContext()
	bus := core.NewEventBus()
	workflow := core.NewWorkflow(ctx, bus)

	window := NewInstallerWindow(ctx, workflow, bus)

	called := false
	window.OnComplete(func() {
		called = true
	})

	if window.onComplete == nil {
		t.Error("OnComplete callback not set")
	}

	// Trigger callback directly
	window.onComplete()

	if !called {
		t.Error("OnComplete callback was not called")
	}
}

func TestOnCancelCallback(t *testing.T) {
	ctx := core.NewInstallContext()
	bus := core.NewEventBus()
	workflow := core.NewWorkflow(ctx, bus)

	window := NewInstallerWindow(ctx, workflow, bus)

	called := false
	window.OnCancel(func() {
		called = true
	})

	if window.onCancel == nil {
		t.Error("OnCancel callback not set")
	}

	// Trigger callback directly
	window.onCancel()

	if !called {
		t.Error("OnCancel callback was not called")
	}
}

func TestNewWelcomeScreen(t *testing.T) {
	step := &core.StepConfig{
		ID:    "welcome",
		Title: "Welcome",
		Screen: &core.ScreenConfig{
			Type:        "welcome",
			Title:       "Welcome Test",
			Description: "Test description",
		},
	}

	screen := NewWelcomeScreen(step)

	if screen == nil {
		t.Fatal("NewWelcomeScreen returned nil")
	}

	welcomeScreen, ok := screen.(*WelcomeScreen)
	if !ok {
		t.Fatal("NewWelcomeScreen did not return *WelcomeScreen")
	}

	if welcomeScreen.step != step {
		t.Error("step not set correctly")
	}
}

func TestNewLicenseScreen(t *testing.T) {
	step := &core.StepConfig{
		ID:    "license",
		Title: "License",
		Screen: &core.ScreenConfig{
			Type:    "license",
			Content: "License text here",
		},
	}

	screen := NewLicenseScreen(step)

	if screen == nil {
		t.Fatal("NewLicenseScreen returned nil")
	}

	licenseScreen, ok := screen.(*LicenseScreen)
	if !ok {
		t.Fatal("NewLicenseScreen did not return *LicenseScreen")
	}

	if licenseScreen.step != step {
		t.Error("step not set correctly")
	}

	// Initially not accepted
	if licenseScreen.accepted {
		t.Error("license should not be accepted initially")
	}
}

func TestNewDirectoryScreen(t *testing.T) {
	step := &core.StepConfig{
		ID:    "directory",
		Title: "Select Directory",
		Screen: &core.ScreenConfig{
			Type: "directory",
		},
	}

	screen := NewDirectoryScreen(step)

	if screen == nil {
		t.Fatal("NewDirectoryScreen returned nil")
	}

	dirScreen, ok := screen.(*DirectoryScreen)
	if !ok {
		t.Fatal("NewDirectoryScreen did not return *DirectoryScreen")
	}

	if dirScreen.step != step {
		t.Error("step not set correctly")
	}
}

func TestNewProgressScreen(t *testing.T) {
	step := &core.StepConfig{
		ID:    "progress",
		Title: "Installing",
		Screen: &core.ScreenConfig{
			Type: "progress",
		},
	}

	screen := NewProgressScreen(step)

	if screen == nil {
		t.Fatal("NewProgressScreen returned nil")
	}

	progressScreen, ok := screen.(*ProgressScreen)
	if !ok {
		t.Fatal("NewProgressScreen did not return *ProgressScreen")
	}

	if progressScreen.step != step {
		t.Error("step not set correctly")
	}

	if progressScreen.isComplete {
		t.Error("progress should not be complete initially")
	}
}

func TestNewSummaryScreen(t *testing.T) {
	step := &core.StepConfig{
		ID:    "summary",
		Title: "Summary",
		Screen: &core.ScreenConfig{
			Type: "summary",
		},
	}

	screen := NewSummaryScreen(step)

	if screen == nil {
		t.Fatal("NewSummaryScreen returned nil")
	}

	summaryScreen, ok := screen.(*SummaryScreen)
	if !ok {
		t.Fatal("NewSummaryScreen did not return *SummaryScreen")
	}

	if summaryScreen.step != step {
		t.Error("step not set correctly")
	}
}

func TestNewFormScreen(t *testing.T) {
	step := &core.StepConfig{
		ID:    "form",
		Title: "Configuration",
		Screen: &core.ScreenConfig{
			Type: "form",
			Fields: []core.FieldConfig{
				{
					Type:     "text",
					Label:    "Username",
					Variable: "user.name",
					Required: true,
				},
			},
		},
	}

	screen := NewFormScreen(step)

	if screen == nil {
		t.Fatal("NewFormScreen returned nil")
	}

	formScreen, ok := screen.(*FormScreen)
	if !ok {
		t.Fatal("NewFormScreen did not return *FormScreen")
	}

	if formScreen.step != step {
		t.Error("step not set correctly")
	}
}

func TestNewRichtextScreen(t *testing.T) {
	step := &core.StepConfig{
		ID:    "info",
		Title: "Information",
		Screen: &core.ScreenConfig{
			Type:    "richtext",
			Content: "Some rich text content",
		},
	}

	screen := NewRichtextScreen(step)

	if screen == nil {
		t.Fatal("NewRichtextScreen returned nil")
	}

	richtextScreen, ok := screen.(*RichtextScreen)
	if !ok {
		t.Fatal("NewRichtextScreen did not return *RichtextScreen")
	}

	if richtextScreen.step != step {
		t.Error("step not set correctly")
	}
}

func TestLicenseScreenValidation(t *testing.T) {
	step := &core.StepConfig{
		Screen: &core.ScreenConfig{Type: "license"},
	}

	screen := &LicenseScreen{step: step, accepted: false}

	// Should fail validation when not accepted
	if err := screen.Validate(); err == nil {
		t.Error("Validate should return error when license not accepted")
	}

	// Should pass when accepted
	screen.accepted = true
	if err := screen.Validate(); err != nil {
		t.Errorf("Validate should pass when accepted: %v", err)
	}
}

func TestWelcomeScreenValidation(t *testing.T) {
	step := &core.StepConfig{
		Screen: &core.ScreenConfig{Type: "welcome"},
	}

	screen := &WelcomeScreen{step: step}

	// Welcome screen always validates
	if err := screen.Validate(); err != nil {
		t.Errorf("Welcome screen Validate should always pass: %v", err)
	}
}

func TestSummaryScreenValidation(t *testing.T) {
	step := &core.StepConfig{
		Screen: &core.ScreenConfig{Type: "summary"},
	}

	screen := &SummaryScreen{step: step}

	// Summary screen always validates
	if err := screen.Validate(); err != nil {
		t.Errorf("Summary screen Validate should always pass: %v", err)
	}
}

func TestRichtextScreenValidation(t *testing.T) {
	step := &core.StepConfig{
		Screen: &core.ScreenConfig{Type: "richtext"},
	}

	screen := &RichtextScreen{step: step}

	// Richtext screen always validates
	if err := screen.Validate(); err != nil {
		t.Errorf("Richtext screen Validate should always pass: %v", err)
	}
}

func TestLicenseScreenCollect(t *testing.T) {
	step := &core.StepConfig{
		Screen: &core.ScreenConfig{Type: "license"},
	}

	ctx := core.NewInstallContext()
	screen := &LicenseScreen{step: step, accepted: true}

	if err := screen.Collect(ctx); err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	accepted, ok := ctx.Get("license.accepted")
	if !ok {
		t.Error("license.accepted not set in context")
	}
	if accepted != true {
		t.Error("license.accepted should be true")
	}
}
