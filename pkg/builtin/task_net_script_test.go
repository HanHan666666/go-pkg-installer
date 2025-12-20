package builtin

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

func TestNetScriptTaskExecute(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "out.txt")
	script := []byte("echo hello > " + outFile + "\n")

	sum := sha256.Sum256(script)
	sha := hex.EncodeToString(sum[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(script)
	}))
	defer server.Close()

	ctx := core.NewInstallContext()
	ctx.Env.IsRoot = true
	bus := core.NewEventBus()

	task := &NetScriptTask{
		BaseTask: core.BaseTask{
			TaskID:   "net-script",
			TaskType: "net_script",
		},
		URL:     server.URL,
		SHA256:  sha,
		WorkDir: tmpDir,
		Timeout: 5 * time.Second,
	}

	if err := task.Execute(ctx, bus); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if _, err := os.Stat(outFile); err != nil {
		t.Fatalf("expected output file, got error: %v", err)
	}
}
