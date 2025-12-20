package builtin

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/HanHan666666/go-pkg-installer/pkg/core"
)

func TestRegisterDownloadTask(t *testing.T) {
	RegisterDownloadTask()

	if !core.Tasks.Has("download") {
		t.Error("expected download task to be registered")
	}
}

func TestDownloadTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    *DownloadTask
		wantErr bool
	}{
		{
			name:    "valid",
			task:    &DownloadTask{URL: "http://example.com/file", Destination: "/tmp/file"},
			wantErr: false,
		},
		{
			name:    "missing url",
			task:    &DownloadTask{Destination: "/tmp/file"},
			wantErr: true,
		},
		{
			name:    "missing destination",
			task:    &DownloadTask{URL: "http://example.com/file"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.task.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestDownloadTaskExecute(t *testing.T) {
	content := "Hello, World!"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "downloaded.txt")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &DownloadTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-download",
			TaskType: "download",
		},
		URL:         server.URL + "/test.txt",
		Destination: destination,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != content {
		t.Errorf("expected %q, got %q", content, string(data))
	}
}

func TestDownloadTaskWithChecksum(t *testing.T) {
	content := "test content"
	expectedHash := "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "downloaded.txt")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	t.Run("valid checksum", func(t *testing.T) {
		task := &DownloadTask{
			BaseTask: core.BaseTask{
				TaskID:   "test-download",
				TaskType: "download",
			},
			URL:         server.URL + "/test.txt",
			Destination: destination,
			SHA256:      expectedHash,
		}

		err := task.Execute(ctx, bus)
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("invalid checksum", func(t *testing.T) {
		destination2 := filepath.Join(tmpDir, "downloaded2.txt")
		task := &DownloadTask{
			BaseTask: core.BaseTask{
				TaskID:   "test-download",
				TaskType: "download",
			},
			URL:         server.URL + "/test.txt",
			Destination: destination2,
			SHA256:      "invalid-hash",
		}

		err := task.Execute(ctx, bus)
		if err == nil {
			t.Error("expected error for invalid checksum")
		}

		if _, err := os.Stat(destination2); !os.IsNotExist(err) {
			t.Error("file should have been removed after checksum failure")
		}
	})
}

func TestDownloadTaskRollback(t *testing.T) {
	content := "test"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "downloaded.txt")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &DownloadTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-download",
			TaskType: "download",
		},
		URL:         server.URL + "/test.txt",
		Destination: destination,
	}

	if err := task.Execute(ctx, bus); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !task.CanRollback() {
		t.Error("expected CanRollback to return true")
	}

	if err := task.Rollback(ctx, bus); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Error("file should have been removed during rollback")
	}
}

func TestDownloadTaskFactory(t *testing.T) {
	RegisterDownloadTask()

	ctx := core.NewInstallContext()
	ctx.Set("url", "http://example.com/file.zip")
	ctx.Set("dest", "/tmp/file.zip")

	factory, ok := core.Tasks.Get("download")
	if !ok {
		t.Fatal("download factory not registered")
	}

	task, err := factory(map[string]any{
		"id":          "dl-1",
		"url":         "{{.url}}",
		"destination": "{{.dest}}",
		"sha256":      "abc123",
	}, ctx)

	if err != nil {
		t.Fatalf("factory error = %v", err)
	}

	dlTask, ok := task.(*DownloadTask)
	if !ok {
		t.Fatal("expected *DownloadTask")
	}

	if dlTask.URL != "http://example.com/file.zip" {
		t.Errorf("expected URL to be rendered, got %q", dlTask.URL)
	}
	if dlTask.Destination != "/tmp/file.zip" {
		t.Errorf("expected destination to be rendered, got %q", dlTask.Destination)
	}
}
