package ui

import (
	"testing"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

func TestResolveLocale(t *testing.T) {
	ctx := core.NewInstallContext()
	ctx.Set("meta.lang", "zh_CN")
	if got := resolveLocale(ctx); got != "zh" {
		t.Fatalf("expected zh, got %s", got)
	}
}

func TestTranslate(t *testing.T) {
	ctx := core.NewInstallContext()
	ctx.Set("meta.lang", "zh")
	if got := tr(ctx, "button.cancel", "Cancel"); got == "Cancel" {
		t.Fatalf("expected translation, got fallback")
	}
}
