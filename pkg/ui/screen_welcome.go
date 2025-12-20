package ui

import (
	"fmt"

	. "modernc.org/tk9.0"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

// WelcomeScreen renders a welcome screen with product info.
type WelcomeScreen struct {
	step *core.StepConfig
}

// NewWelcomeScreen creates a welcome screen renderer.
func NewWelcomeScreen(step *core.StepConfig) ScreenRenderer {
	return &WelcomeScreen{step: step}
}

// Render creates the welcome screen UI.
func (s *WelcomeScreen) Render(parent *TFrameWidget, ctx *core.InstallContext, bus *core.EventBus) error {
	// Get product info
	productName := ctx.RenderOrDefault("product.name", "Application")
	version := ctx.RenderOrDefault("product.version", "")

	// Title
	titleText := s.step.Screen.Title
	if titleText == "" {
		titleText = fmt.Sprintf("Welcome to %s", productName)
	}
	titleText = ctx.Render(titleText)

	title := parent.TLabel(Txt(titleText), Font("TkHeadingFont"))
	Pack(title, Pady("20"), Side("top"))

	// Description
	description := s.step.Screen.Description
	if description == "" {
		description = fmt.Sprintf("This will install %s on your computer.", productName)
		if version != "" {
			description = fmt.Sprintf("This will install %s version %s on your computer.", productName, version)
		}
	}
	description = ctx.Render(description)

	descLabel := parent.TLabel(Txt(description), Wraplength("600"))
	Pack(descLabel, Pady("10"), Side("top"))

	// Optional banner or logo (if configured)
	if bannerPath := s.step.Screen.BannerPath; bannerPath != "" {
		bannerPath = ctx.Render(bannerPath)
		// Note: Image loading would be done here if tk9 supports it
	}

	// Spacer
	spacer := parent.TFrame()
	Pack(spacer, Fill("both"), Expand(true))

	// Footer message
	footer := parent.TLabel(Txt("Click 'Continue' to proceed with the installation."))
	Pack(footer, Pady("20"), Side("bottom"))

	return nil
}

// Validate validates the welcome screen (always valid).
func (s *WelcomeScreen) Validate() error {
	return nil
}

// Collect collects data from the welcome screen (nothing to collect).
func (s *WelcomeScreen) Collect(ctx *core.InstallContext) error {
	return nil
}

// Cleanup cleans up the welcome screen resources.
func (s *WelcomeScreen) Cleanup() {}
