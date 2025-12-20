package ui

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	. "modernc.org/tk9.0"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

// LicenseScreen renders a license agreement screen.
type LicenseScreen struct {
	step          *core.StepConfig
	accepted      bool
	acceptVar     *VariableOpt
	text          *TextWidget
	requireScroll bool
	ctx           *core.InstallContext
}

// NewLicenseScreen creates a license screen renderer.
func NewLicenseScreen(step *core.StepConfig) ScreenRenderer {
	return &LicenseScreen{step: step}
}

// Render creates the license screen UI.
func (s *LicenseScreen) Render(parent *TFrameWidget, ctx *core.InstallContext, bus *core.EventBus) error {
	s.ctx = ctx
	// Title
	titleText := s.step.Screen.Title
	if titleText == "" {
		titleText = s.step.Title
	}
	if titleText == "" {
		titleText = tr(ctx, "title.license", "License Agreement")
	}
	titleText = ctx.Render(titleText)

	title := parent.TLabel(Txt(titleText), Font("TkHeadingFont"))
	Pack(title, Pady("10"), Side("top"))

	// Description
	if desc := s.step.Screen.Description; desc != "" {
		desc = ctx.Render(desc)
		descLabel := parent.TLabel(Txt(desc), Wraplength("600"))
		Pack(descLabel, Pady("5"), Side("top"))
	}

	// License text frame with scrollbar
	textFrame := parent.TFrame()
	Pack(textFrame, Fill("both"), Expand(true), Pady("10"))

	// Create text widget with scrollbar
	scrollbar := textFrame.TScrollbar()
	Pack(scrollbar, Side("right"), Fill("y"))

	text := textFrame.Text(
		Width(80),
		Height(15),
		Wrap("word"),
		Yscrollcommand(func(e *Event) { e.ScrollSet(scrollbar) }),
	)
	scrollbar.Configure(Command(func(e *Event) { e.Yview(text) }))
	Pack(text, Side("left"), Fill("both"), Expand(true))
	s.text = text
	s.requireScroll = s.step.Screen.RequireScrollToEnd

	// Load license content
	licenseText := ""
	if s.step.Screen.Content != "" {
		licenseText = ctx.Render(s.step.Screen.Content)
	} else if s.step.Screen.Source != "" {
		source := ctx.Render(s.step.Screen.Source)
		if strings.ContainsAny(source, "\n\r") {
			licenseText = source
		} else if data, err := os.ReadFile(source); err == nil {
			licenseText = string(data)
		} else {
			licenseText = source
		}
	} else if s.step.Screen.ContentFile != "" {
		filePath := ctx.Render(s.step.Screen.ContentFile)
		if data, err := os.ReadFile(filePath); err == nil {
			licenseText = string(data)
		} else {
			licenseText = fmt.Sprintf("License content would be loaded from: %s", filePath)
		}
	}
	if licenseText == "" {
		licenseText = "License content is empty."
	}

	// Insert license text
	text.Configure(State("normal"))
	text.Insert("1.0", licenseText)
	text.Configure(State("disabled"))

	// Acceptance checkbox
	acceptFrame := parent.TFrame()
	Pack(acceptFrame, Pady("10"), Side("bottom"))

	s.acceptVar = Variable("")
	checkbox := acceptFrame.TCheckbutton(
		Txt(tr(ctx, "label.accept", "I accept the terms of the license agreement")),
		s.acceptVar,
		Command(func() {
			s.syncAccepted()
		}),
	)
	Pack(checkbox, Side("left"))

	return nil
}

// Validate validates that the license is accepted.
func (s *LicenseScreen) Validate() error {
	s.syncAccepted()
	if s.requireScroll && !s.scrolledToEnd() {
		return errors.New(tr(s.ctx, "msg.scroll_end", "please scroll to the end of the license to continue"))
	}
	if !s.accepted {
		return errors.New(tr(s.ctx, "msg.accept_license", "you must accept the license agreement to continue"))
	}
	return nil
}

// Collect collects the acceptance status.
func (s *LicenseScreen) Collect(ctx *core.InstallContext) error {
	s.syncAccepted()
	ctx.Set("license.accepted", s.accepted)
	return nil
}

// Cleanup cleans up the license screen resources.
func (s *LicenseScreen) Cleanup() {}

func (s *LicenseScreen) syncAccepted() {
	if s.acceptVar == nil {
		return
	}
	value := strings.ToLower(strings.TrimSpace(s.acceptVar.Get()))
	switch value {
	case "1", "true", "yes", "on":
		s.accepted = true
	default:
		s.accepted = false
	}
}

func (s *LicenseScreen) scrolledToEnd() bool {
	if s.text == nil {
		return true
	}
	return scrollReachedEnd(s.text.Yview())
}

func scrollReachedEnd(yview string) bool {
	parts := strings.Fields(yview)
	if len(parts) < 2 {
		return false
	}
	last, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return false
	}
	return last >= 0.999
}
