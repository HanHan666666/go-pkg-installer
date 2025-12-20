package ui

import (
	"os"
	"testing"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

func TestLogoKeyFromContextEmbeddedBytes(t *testing.T) {
	data, err := os.ReadFile("/home/han/linglong-installer/assets/logo.png")
	if err != nil {
		t.Skipf("missing sample logo: %v", err)
	}

	ctx := core.NewInstallContext()
	ctx.Set("product.logo.bytes", data)

	key := logoKeyFromContext(ctx)
	if key != "embedded" {
		t.Fatalf("expected embedded key, got %q", key)
	}
}
