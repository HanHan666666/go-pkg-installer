package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndValidateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-installer.yaml")

	configContent := `
product:
  name: "Test App"

flows:
  install:
    entry: welcome
    steps:
      - id: welcome
        title: "Welcome"
        screen:
          type: welcome
          content: "Hello World"
      - id: finish
        title: "Done"
        screen:
          type: finish
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := loadAndValidateConfig(configPath)
	if err != nil {
		t.Fatalf("loadAndValidateConfig failed: %v", err)
	}

	if cfg.Product == nil {
		t.Fatal("Expected product to be set")
	}
	if cfg.Product.Name != "Test App" {
		t.Errorf("Expected product name 'Test App', got '%s'", cfg.Product.Name)
	}
	if len(cfg.Flows) != 1 {
		t.Errorf("Expected 1 flow, got %d", len(cfg.Flows))
	}
}

func TestLoadAndValidateConfigNotFound(t *testing.T) {
	_, err := loadAndValidateConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLoadAndValidateConfigWithTasks(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "tasks.yaml")

	configContent := `
product:
  name: "Test App"

flows:
  install:
    entry: progress
    steps:
      - id: progress
        title: "Installing"
        screen:
          type: progress
        tasks:
          - type: shell
            script: "echo hello"
          - type: writeConfig
            path: "/tmp/test.txt"
            content: "test content"
      - id: finish
        title: "Done"
        screen:
          type: finish
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := loadAndValidateConfig(configPath)
	if err != nil {
		t.Fatalf("loadAndValidateConfig failed: %v", err)
	}

	installFlow := cfg.Flows["install"]
	if installFlow == nil {
		t.Fatal("Expected install flow")
	}

	if len(installFlow.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(installFlow.Steps))
	}
}
