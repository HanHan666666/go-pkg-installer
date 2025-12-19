package ui

import (
	"fmt"

	. "modernc.org/tk9.0"

	"github.com/anthropics/go-pkg-installer/pkg/core"
)

// SummaryScreen renders an installation summary screen.
type SummaryScreen struct {
	step      *core.StepConfig
	launchVar *VariableOpt
}

// NewSummaryScreen creates a summary screen renderer.
func NewSummaryScreen(step *core.StepConfig) ScreenRenderer {
	return &SummaryScreen{step: step}
}

// Render creates the summary screen UI.
func (s *SummaryScreen) Render(parent *TFrameWidget, ctx *core.InstallContext, bus *core.EventBus) error {
	// Check installation status
	isComplete, _ := ctx.Get("install.complete")
	success := isComplete == true

	// Title
	titleText := s.step.Screen.Title
	if titleText == "" {
		if success {
			titleText = "Installation Complete"
		} else {
			titleText = "Installation Summary"
		}
	}
	titleText = ctx.Render(titleText)

	title := parent.TLabel(Txt(titleText), Font("TkHeadingFont"))
	Pack(title, Pady("20"), Side("top"))

	// Status icon/message
	var statusText string
	if success {
		statusText = "✓ The installation completed successfully!"
	} else {
		statusText = "✗ The installation encountered errors."
	}

	statusLabel := parent.TLabel(Txt(statusText))
	Pack(statusLabel, Pady("10"), Side("top"))

	// Description
	desc := s.step.Screen.Description
	if desc == "" && success {
		productName := ctx.RenderOrDefault("product.name", "The application")
		desc = fmt.Sprintf("%s has been installed on your computer.", productName)
	}
	if desc != "" {
		desc = ctx.Render(desc)
		descLabel := parent.TLabel(Txt(desc), Wraplength("600"))
		Pack(descLabel, Pady("10"), Side("top"))
	}

	// Summary frame
	summaryFrame := parent.TFrame()
	Pack(summaryFrame, Fill("x"), Pady("20"))

	// Show installation details
	if installDir, ok := ctx.Get("install_dir"); ok {
		dirFrame := summaryFrame.TFrame()
		Pack(dirFrame, Fill("x"), Pady("5"))

		dirLbl := dirFrame.TLabel(Txt("Installed to:"))
		Pack(dirLbl, Side("left"))

		dirVal := dirFrame.TLabel(Txt(fmt.Sprintf(" %v", installDir)))
		Pack(dirVal, Side("left"))
	}

	// Spacer
	spacer := parent.TFrame()
	Pack(spacer, Fill("both"), Expand(true))

	// Optional: Launch application checkbox
	if success {
		launchFrame := parent.TFrame()
		Pack(launchFrame, Pady("10"), Side("bottom"))

		// Check if there's a launch command configured
		if launchCmd, ok := ctx.Get("launch.command"); ok && launchCmd != "" {
			s.launchVar = Variable("1")
			launchCheck := launchFrame.TCheckbutton(
				Txt("Launch application after closing"),
				Variable(s.launchVar),
			)
			Pack(launchCheck, Side("left"))

			// Store the preference
			ctx.Set("launch.on_close", true)
		}
	}

	// Footer
	footerText := "Click 'Close' to exit the installer."
	footer := parent.TLabel(Txt(footerText))
	Pack(footer, Pady("10"), Side("bottom"))

	return nil
}

// Validate validates the summary screen (always valid).
func (s *SummaryScreen) Validate() error {
	return nil
}

// Collect collects data from the summary screen.
func (s *SummaryScreen) Collect(ctx *core.InstallContext) error {
	if s.launchVar != nil {
		ctx.Set("launch.on_close", s.launchVar.Get() == "1")
	}
	return nil
}

// Cleanup cleans up the summary screen resources.
func (s *SummaryScreen) Cleanup() {}
