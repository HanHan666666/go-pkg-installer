// Package ui provides the tk9-based UI adapter for the installer.
package ui

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	. "modernc.org/tk9.0"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

// InstallerWindow represents the main installer window.
type InstallerWindow struct {
	mu sync.Mutex

	ctx      *core.InstallContext
	bus      *core.EventBus
	workflow *core.Workflow

	// UI state
	mainFrame    *TFrameWidget
	sidebarFrame *TFrameWidget
	contentFrame *TFrameWidget
	navFrame     *TFrameWidget

	// Navigation buttons
	backBtn   *TButtonWidget
	nextBtn   *TButtonWidget
	cancelBtn *TButtonWidget

	// Current screen
	currentScreen ScreenRenderer

	// Screen registry
	screenRenderers map[string]ScreenRendererFactory

	// Callbacks
	onComplete func()
	onCancel   func()
}

// ScreenRenderer interface for rendering different screen types.
type ScreenRenderer interface {
	// Render creates the screen UI in the given frame
	Render(parent *TFrameWidget, ctx *core.InstallContext, bus *core.EventBus) error
	// Validate checks if the screen input is valid
	Validate() error
	// Collect gathers input data into the context
	Collect(ctx *core.InstallContext) error
	// Cleanup removes any screen-specific resources
	Cleanup()
}

// ScreenRendererFactory creates a ScreenRenderer for a step configuration.
type ScreenRendererFactory func(step *core.StepConfig) ScreenRenderer

// NewInstallerWindow creates a new installer window.
func NewInstallerWindow(ctx *core.InstallContext, workflow *core.Workflow, bus *core.EventBus) *InstallerWindow {
	w := &InstallerWindow{
		ctx:             ctx,
		bus:             bus,
		workflow:        workflow,
		screenRenderers: make(map[string]ScreenRendererFactory),
	}

	// Register built-in screen renderers
	w.RegisterScreenRenderer("welcome", NewWelcomeScreen)
	w.RegisterScreenRenderer("license", NewLicenseScreen)
	w.RegisterScreenRenderer("directory", NewDirectoryScreen)
	w.RegisterScreenRenderer("pathPicker", NewDirectoryScreen)
	w.RegisterScreenRenderer("progress", NewProgressScreen)
	w.RegisterScreenRenderer("richtext", NewRichtextScreen)
	w.RegisterScreenRenderer("summary", NewSummaryScreen)
	w.RegisterScreenRenderer("form", NewFormScreen)
	w.RegisterScreenRenderer("options", NewOptionsScreen)
	w.RegisterScreenRenderer("installType", NewOptionsScreen)
	w.RegisterScreenRenderer("finish", NewSummaryScreen)

	return w
}

// RegisterScreenRenderer registers a screen renderer factory.
func (w *InstallerWindow) RegisterScreenRenderer(screenType string, factory ScreenRendererFactory) {
	w.screenRenderers[screenType] = factory
}

// OnComplete sets the callback for installation completion.
func (w *InstallerWindow) OnComplete(fn func()) {
	w.onComplete = fn
}

// OnCancel sets the callback for installation cancellation.
func (w *InstallerWindow) OnCancel(fn func()) {
	w.onCancel = fn
}

// Run initializes and runs the installer UI.
func (w *InstallerWindow) Run() error {
	// Lock to OS thread for tk9
	runtime.LockOSThread()

	// Set up the main window
	w.setupWindow()

	// Subscribe to events
	w.subscribeEvents()

	// Render the first screen
	w.renderCurrentStep()

	// Start the main loop
	App.Wait()

	return nil
}

func (w *InstallerWindow) setupWindow() {
	// Get product info from context
	productName, _ := w.ctx.Get("product.name")
	if productName == nil {
		productName = "Installer"
	}

	// Configure main window title and size
	App.WmTitle(fmt.Sprintf("%v", productName))
	WmGeometry(App, "700x500")
	WmMinSize(App, 700, 500)

	// Create main container frame
	w.mainFrame = TFrame(Padding("10"))
	Pack(w.mainFrame, Fill("both"), Expand(true))

	// Create sidebar frame (for steps/branding)
	w.sidebarFrame = w.mainFrame.TFrame(Width(180), Padding("5"))
	Pack(w.sidebarFrame, Side("left"), Fill("y"), Padx("5"))

	// Create content wrapper frame (content + nav)
	contentWrapper := w.mainFrame.TFrame()
	Pack(contentWrapper, Side("right"), Fill("both"), Expand(true))

	// Create content frame (for screen content)
	w.contentFrame = contentWrapper.TFrame(Padding("5"))
	Pack(w.contentFrame, Fill("both"), Expand(true), Side("top"))

	// Create separator
	separator := contentWrapper.TSeparator()
	Pack(separator, Fill("x"), Pady("10"))

	// Create navigation frame
	w.navFrame = contentWrapper.TFrame()
	Pack(w.navFrame, Fill("x"), Side("bottom"))

	// Create navigation buttons
	w.cancelBtn = w.navFrame.TButton(Txt(tr(w.ctx, "button.cancel", "Cancel")), Command(w.handleCancel))
	Pack(w.cancelBtn, Side("left"), Padx("5"))

	w.nextBtn = w.navFrame.TButton(Txt(tr(w.ctx, "button.continue", "Continue")), Command(w.handleNext))
	Pack(w.nextBtn, Side("right"), Padx("5"))

	w.backBtn = w.navFrame.TButton(Txt(tr(w.ctx, "button.back", "Go Back")), Command(w.handleBack))
	Pack(w.backBtn, Side("right"), Padx("5"))
}

func (w *InstallerWindow) subscribeEvents() {
	// Subscribe to progress events
	w.bus.Subscribe(core.EventProgress, func(e core.Event) {
		// Update progress bar if on progress screen
		if ps, ok := w.currentScreen.(*ProgressScreen); ok {
			if p := e.ProgressPayload(); p != nil {
				ps.UpdateProgress(p.Progress*100, p.Message)
			}
		}
	})

	// Subscribe to log events
	w.bus.Subscribe(core.EventLog, func(e core.Event) {
		if ps, ok := w.currentScreen.(*ProgressScreen); ok {
			if p := e.LogPayload(); p != nil {
				ps.AddLogMessage(p.Message)
			}
		}
	})

	// Subscribe to step change events
	w.bus.Subscribe(core.EventStepChange, func(e core.Event) {
		// Render the new step
		w.renderCurrentStep()
	})
}

func (w *InstallerWindow) renderCurrentStep() {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Clean up current screen
	if w.currentScreen != nil {
		w.currentScreen.Cleanup()
	}

	// Clear content frame by destroying children
	children := WinfoChildren(w.contentFrame.Window)
	for _, child := range children {
		Destroy(child)
	}

	// Get current step
	step := w.workflow.CurrentStep()
	if step == nil || step.Config == nil {
		return
	}

	stepConfig := step.Config

	// Get screen type
	screenType := ""
	if stepConfig.Screen != nil {
		screenType = stepConfig.Screen.Type
	}
	if core.IsGoExtension(screenType) {
		screenType = core.StripGoPrefix(screenType)
	}

	factory, ok := w.screenRenderers[screenType]
	if !ok {
		// Fallback to richtext for unknown types
		factory = NewRichtextScreen
	}

	// Create and render screen
	w.currentScreen = factory(stepConfig)
	if err := w.currentScreen.Render(w.contentFrame, w.ctx, w.bus); err != nil {
		w.ctx.AddLog(core.LogError, fmt.Sprintf("Failed to render screen: %v", err))
	}

	// Update sidebar and navigation buttons
	w.renderSidebar()
	w.updateNavButtons()
}

func (w *InstallerWindow) renderSidebar() {
	if w.sidebarFrame == nil {
		return
	}

	children := WinfoChildren(w.sidebarFrame.Window)
	for _, child := range children {
		Destroy(child)
	}

	productName := w.ctx.RenderOrDefault("product.name", "Installer")
	title := w.sidebarFrame.TLabel(Txt(productName), Font("TkHeadingFont"), Anchor("w"))
	Pack(title, Side("top"), Fill("x"), Pady("5"))

	steps := w.workflow.Steps()
	for _, step := range steps {
		status := w.workflow.StepStatus(step.ID)
		prefix := "[ ]"
		switch status {
		case core.StepCurrent:
			prefix = "[>]"
		case core.StepCompleted:
			prefix = "[x]"
		case core.StepDisabled:
			prefix = "[-]"
		}
		text := fmt.Sprintf("%s %s", prefix, step.Title)
		label := w.sidebarFrame.TLabel(Txt(text), Anchor("w"), Wraplength("160"))
		Pack(label, Side("top"), Fill("x"), Padx("5"), Pady("2"))
	}
}

func (w *InstallerWindow) updateNavButtons() {
	step := w.workflow.CurrentStep()
	if step == nil || step.Config == nil {
		return
	}

	stepConfig := step.Config

	// Back button - disable on first step
	if w.workflow.CanGoBack() {
		w.backBtn.Configure(State("normal"))
	} else {
		w.backBtn.Configure(State("disabled"))
	}

	// Next button - change text based on screen type
	screenType := ""
	if stepConfig.Screen != nil {
		screenType = stepConfig.Screen.Type
	}

	if w.workflow.IsLastStep() {
		w.nextBtn.Configure(Txt(tr(w.ctx, "button.install", "Install")))
	} else if screenType == "progress" {
		w.nextBtn.Configure(Txt(tr(w.ctx, "button.finish", "Finish")))
	} else if screenType == "summary" {
		w.nextBtn.Configure(Txt(tr(w.ctx, "button.close", "Close")))
	} else {
		w.nextBtn.Configure(Txt(tr(w.ctx, "button.continue", "Continue")))
	}
}

func (w *InstallerWindow) handleNext() {
	w.mu.Lock()
	screen := w.currentScreen
	w.mu.Unlock()

	// Validate current screen
	if screen != nil {
		if err := screen.Validate(); err != nil {
			MessageBox(Icon("error"), Msg(err.Error()), Title(tr(w.ctx, "dialog.validation.title", "Validation Error")))
			return
		}

		// Collect data from screen
		if err := screen.Collect(w.ctx); err != nil {
			MessageBox(Icon("error"), Msg(err.Error()), Title(tr(w.ctx, "dialog.error.title", "Error")))
			return
		}
	}

	// Get current step config
	step := w.workflow.CurrentStep()
	screenType := ""
	if step != nil && step.Config != nil && step.Config.Screen != nil {
		screenType = step.Config.Screen.Type
	}

	// Check if this is the last step or summary
	if w.workflow.IsLastStep() || screenType == "summary" {
		if w.onComplete != nil {
			w.onComplete()
		}
		Destroy(App)
		return
	}

	_, err := w.workflow.Next()
	if err != nil {
		MessageBox(Icon("error"), Msg(err.Error()), Title(tr(w.ctx, "dialog.error.title", "Navigation Error")))
		return
	}

	// Render will be triggered by the step-change event.
	if w.bus == nil {
		w.renderCurrentStep()
	}
}

func (w *InstallerWindow) handleBack() {
	if !w.workflow.CanGoBack() {
		return
	}

	_, err := w.workflow.Prev()
	if err != nil {
		MessageBox(Icon("error"), Msg(err.Error()), Title(tr(w.ctx, "dialog.error.title", "Navigation Error")))
		return
	}

	// Render will be triggered by the step-change event.
	if w.bus == nil {
		w.renderCurrentStep()
	}
}

func (w *InstallerWindow) handleCancel() {
	result := MessageBox(
		Icon("question"),
		Msg(tr(w.ctx, "dialog.cancel.msg", "Are you sure you want to cancel the installation?")),
		Title(tr(w.ctx, "dialog.cancel.title", "Cancel Installation")),
		Type("yesno"),
	)

	if result == "yes" {
		if w.onCancel != nil {
			w.onCancel()
		}
		Destroy(App)
		os.Exit(0)
	}
}

// after schedules a function to run after a delay.
func after(d time.Duration, fn func()) {
	go func() {
		time.Sleep(d)
		fn()
	}()
}
