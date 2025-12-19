package builtin

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/anthropics/go-pkg-installer/pkg/core"
)

func createTestTarGz(t *testing.T, dir string) string {
	t.Helper()

	archivePath := filepath.Join(dir, "test.tar.gz")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Add a directory
	tw.WriteHeader(&tar.Header{
		Name:     "testdir/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	})

	// Add a file
	content := []byte("test content")
	tw.WriteHeader(&tar.Header{
		Name:     "testdir/file.txt",
		Mode:     0644,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	})
	tw.Write(content)

	return archivePath
}

func createTestZip(t *testing.T, dir string) string {
	t.Helper()

	archivePath := filepath.Join(dir, "test.zip")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	// Add a directory
	zw.Create("testdir/")

	// Add a file
	w, _ := zw.Create("testdir/file.txt")
	w.Write([]byte("test content"))

	return archivePath
}

func TestRegisterUnpackTask(t *testing.T) {
	RegisterUnpackTask()

	if !core.Tasks.Has("unpack") {
		t.Error("expected unpack task to be registered")
	}
}

func TestUnpackTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    *UnpackTask
		wantErr bool
	}{
		{
			name:    "valid",
			task:    &UnpackTask{Source: "/tmp/archive.tar.gz", Destination: "/tmp/extracted"},
			wantErr: false,
		},
		{
			name:    "missing source",
			task:    &UnpackTask{Destination: "/tmp/extracted"},
			wantErr: true,
		},
		{
			name:    "missing destination",
			task:    &UnpackTask{Source: "/tmp/archive.tar.gz"},
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

func TestUnpackTarGz(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := createTestTarGz(t, tmpDir)
	destDir := filepath.Join(tmpDir, "extracted")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &UnpackTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-unpack",
			TaskType: "unpack",
		},
		Source:      archivePath,
		Destination: destDir,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify extracted files
	extractedFile := filepath.Join(destDir, "testdir", "file.txt")
	data, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(data) != "test content" {
		t.Errorf("expected 'test content', got %q", string(data))
	}
}

func TestUnpackZip(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := createTestZip(t, tmpDir)
	destDir := filepath.Join(tmpDir, "extracted")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &UnpackTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-unpack",
			TaskType: "unpack",
		},
		Source:      archivePath,
		Destination: destDir,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify extracted files
	extractedFile := filepath.Join(destDir, "testdir", "file.txt")
	data, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(data) != "test content" {
		t.Errorf("expected 'test content', got %q", string(data))
	}
}

func TestUnpackWithStripPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := createTestTarGz(t, tmpDir)
	destDir := filepath.Join(tmpDir, "extracted")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &UnpackTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-unpack",
			TaskType: "unpack",
		},
		Source:      archivePath,
		Destination: destDir,
		StripPrefix: 1,
	}

	err := task.Execute(ctx, bus)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// File should be directly in destDir, not in testdir
	extractedFile := filepath.Join(destDir, "file.txt")
	data, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(data) != "test content" {
		t.Errorf("expected 'test content', got %q", string(data))
	}
}

func TestUnpackTaskRollback(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := createTestTarGz(t, tmpDir)
	destDir := filepath.Join(tmpDir, "extracted")

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &UnpackTask{
		BaseTask: core.BaseTask{
			TaskID:   "test-unpack",
			TaskType: "unpack",
		},
		Source:      archivePath,
		Destination: destDir,
	}

	// Execute
	if err := task.Execute(ctx, bus); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify can rollback
	if !task.CanRollback() {
		t.Error("expected CanRollback to return true")
	}

	// Rollback
	if err := task.Rollback(ctx, bus); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	// Extracted file should be gone
	extractedFile := filepath.Join(destDir, "testdir", "file.txt")
	if _, err := os.Stat(extractedFile); !os.IsNotExist(err) {
		t.Error("extracted file should have been removed during rollback")
	}
}

func TestUnpackUnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.xyz")
	os.WriteFile(archivePath, []byte("data"), 0644)

	ctx := core.NewInstallContext()
	bus := core.NewEventBus()

	task := &UnpackTask{
		Source:      archivePath,
		Destination: filepath.Join(tmpDir, "extracted"),
	}

	err := task.Execute(ctx, bus)
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}
