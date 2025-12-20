package schema

import (
	"testing"
)

func TestNewValidator(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator failed: %v", err)
	}
	if v == nil {
		t.Fatal("Validator should not be nil")
	}
}

func TestValidateYAMLValid(t *testing.T) {
	v, _ := NewValidator()

	validYAML := `
product:
  name: "Test App"
flows:
  install:
    entry: "welcome"
    steps:
      - id: "welcome"
        title: "Welcome"
        screen:
          type: "richtext"
          content: "Welcome to the installer"
      - id: "finish"
        title: "Finish"
        screen:
          type: "finish"
`
	result := v.ValidateYAML([]byte(validYAML))
	if !result.Valid {
		t.Errorf("Should be valid, errors: %v", result.Errors)
	}
}

func TestValidateYAMLMissingRequired(t *testing.T) {
	v, _ := NewValidator()

	invalidYAML := `
product: {}
flows:
  install:
    entry: "welcome"
    steps:
      - id: "welcome"
        title: "Welcome"
`
	result := v.ValidateYAML([]byte(invalidYAML))
	if result.Valid {
		t.Error("Should be invalid - missing product.name")
	}
	if len(result.Errors) == 0 {
		t.Error("Should have at least one error")
	}
}

func TestValidateYAMLMissingFlows(t *testing.T) {
	v, _ := NewValidator()

	invalidYAML := `
product:
  name: "Test App"
`
	result := v.ValidateYAML([]byte(invalidYAML))
	if result.Valid {
		t.Error("Should be invalid - missing flows")
	}
}

func TestValidateYAMLWithTasks(t *testing.T) {
	v, _ := NewValidator()

	validYAML := `
product:
  name: "Test App"
flows:
  install:
    entry: "install"
    steps:
      - id: "install"
        title: "Installing"
        screen:
          type: "progress"
        tasks:
          - type: "download"
            url: "https://example.com/app.tar.gz"
            sha256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
            to: "/tmp/app.tar.gz"
          - type: "unpack"
            from: "/tmp/app.tar.gz"
            to: "/opt/app"
`
	result := v.ValidateYAML([]byte(validYAML))
	if !result.Valid {
		t.Errorf("Should be valid, errors: %v", result.Errors)
	}
}

func TestValidateYAMLWithGuards(t *testing.T) {
	v, _ := NewValidator()

	validYAML := `
product:
  name: "Test App"
flows:
  install:
    entry: "license"
    steps:
      - id: "license"
        title: "License"
        screen:
          type: "license"
          source: "assets/license.txt"
        guards:
          - type: "mustAccept"
            field: "license.accepted"
          - type: "diskSpace"
            minMB: 500
`
	result := v.ValidateYAML([]byte(validYAML))
	if !result.Valid {
		t.Errorf("Should be valid, errors: %v", result.Errors)
	}
}

func TestValidateYAMLWithBranching(t *testing.T) {
	v, _ := NewValidator()

	validYAML := `
product:
  name: "Test App"
flows:
  install:
    entry: "type"
    steps:
      - id: "type"
        title: "Install Type"
        screen:
          type: "options"
          bind: "install.type"
        branch:
          condition: "install.type"
          branches:
            full: "full_install"
            minimal: "minimal_install"
          default: "full_install"
      - id: "full_install"
        title: "Full Install"
        screen:
          type: "progress"
      - id: "minimal_install"
        title: "Minimal Install"
        screen:
          type: "progress"
`
	result := v.ValidateYAML([]byte(validYAML))
	if !result.Valid {
		t.Errorf("Should be valid, errors: %v", result.Errors)
	}
}

func TestValidateYAMLWithMeta(t *testing.T) {
	v, _ := NewValidator()

	validYAML := `
product:
  name: "Test App"
meta:
  version: "1.0.0"
  payloadUrl: "https://example.com/app.tar.gz"
  payloadSha256: "sha256hash"
flows:
  install:
    entry: "welcome"
    steps:
      - id: "welcome"
        title: "Welcome"
        screen:
          type: "richtext"
          content: "Welcome!"
`
	result := v.ValidateYAML([]byte(validYAML))
	if !result.Valid {
		t.Errorf("Should be valid, errors: %v", result.Errors)
	}
}

func TestValidateYAMLInvalidSyntax(t *testing.T) {
	v, _ := NewValidator()

	invalidYAML := `
product:
  name: "Test
  unclosed: quote
`
	result := v.ValidateYAML([]byte(invalidYAML))
	if result.Valid {
		t.Error("Should be invalid - YAML syntax error")
	}
}

func TestValidateJSON(t *testing.T) {
	v, _ := NewValidator()

	validJSON := `{
		"product": {"name": "Test App"},
		"flows": {
			"install": {
				"entry": "welcome",
				"steps": [
					{"id": "welcome", "title": "Welcome", "screen": {"type": "richtext", "content": "Welcome!"}}
				]
			}
		}
	}`
	result := v.ValidateJSON([]byte(validJSON))
	if !result.Valid {
		t.Errorf("Should be valid, errors: %v", result.Errors)
	}
}

func TestValidateJSONInvalid(t *testing.T) {
	v, _ := NewValidator()

	invalidJSON := `{"product": {}}`
	result := v.ValidateJSON([]byte(invalidJSON))
	if result.Valid {
		t.Error("Should be invalid")
	}
}

func TestValidationErrorFormat(t *testing.T) {
	e := ValidationError{Path: "/product/name", Message: "is required"}
	expected := "/product/name: is required"
	if e.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, e.Error())
	}

	e2 := ValidationError{Message: "global error"}
	if e2.Error() != "global error" {
		t.Errorf("Expected 'global error', got %q", e2.Error())
	}
}

func TestConvertYAMLToJSON(t *testing.T) {
	input := map[any]any{
		"key1": "value1",
		"nested": map[any]any{
			"key2": "value2",
		},
		"list": []any{
			map[any]any{"a": "b"},
		},
	}

	result := convertYAMLToJSON(input)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatal("Should convert to map[string]any")
	}

	if m["key1"] != "value1" {
		t.Error("key1 should be value1")
	}

	nested, ok := m["nested"].(map[string]any)
	if !ok {
		t.Fatal("nested should be map[string]any")
	}
	if nested["key2"] != "value2" {
		t.Error("key2 should be value2")
	}
}

func TestLoadConfig(t *testing.T) {
	yamlContent := `
product:
  name: "Test App"
  logo: "assets/logo.png"
  theme:
    primaryColor: "#ff0000"
meta:
  version: "1.0.0"
flows:
  install:
    entry: "welcome"
    steps:
      - id: "welcome"
        title: "Welcome"
        screen:
          type: "welcome"
          content: "Welcome to the installer"
      - id: "license"
        title: "License"
        screen:
          type: "license"
          source: "assets/license.txt"
          requireScrollToEnd: true
        guards:
          - type: "mustAccept"
            field: "license.accepted"
`
	config, err := LoadConfig([]byte(yamlContent))
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.Product.Name != "Test App" {
		t.Errorf("Expected 'Test App', got %s", config.Product.Name)
	}
	if config.Product.Logo != "assets/logo.png" {
		t.Errorf("Expected logo path, got %s", config.Product.Logo)
	}
	if config.Product.Theme == nil || config.Product.Theme.PrimaryColor != "#ff0000" {
		t.Error("Theme primary color should be #ff0000")
	}

	if config.Meta["version"] != "1.0.0" {
		t.Error("Meta version should be 1.0.0")
	}

	flow, ok := config.Flows["install"]
	if !ok {
		t.Fatal("install flow should exist")
	}
	if flow.Entry != "welcome" {
		t.Errorf("Expected entry 'welcome', got %s", flow.Entry)
	}
	if len(flow.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(flow.Steps))
	}

	licenseStep := flow.Steps[1]
	if licenseStep.Screen.RequireScrollToEnd != true {
		t.Error("requireScrollToEnd should be true")
	}
	if len(licenseStep.Guards) != 1 {
		t.Errorf("Expected 1 guard, got %d", len(licenseStep.Guards))
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	invalidYAML := `
product: {}
`
	_, err := LoadConfig([]byte(invalidYAML))
	if err == nil {
		t.Error("LoadConfig should fail for invalid config")
	}
}

func TestGoExtensionScreenType(t *testing.T) {
	v, _ := NewValidator()

	validYAML := `
product:
  name: "Test App"
flows:
  install:
    entry: "custom"
    steps:
      - id: "custom"
        title: "Custom Screen"
        screen:
          type: "go:customScreen"
`
	result := v.ValidateYAML([]byte(validYAML))
	if !result.Valid {
		t.Errorf("go: extension type should be valid, errors: %v", result.Errors)
	}
}

func TestGoExtensionTaskType(t *testing.T) {
	v, _ := NewValidator()

	validYAML := `
product:
  name: "Test App"
flows:
  install:
    entry: "install"
    steps:
      - id: "install"
        title: "Install"
        screen:
          type: "progress"
        tasks:
          - type: "go:customTask"
            customParam: "value"
`
	result := v.ValidateYAML([]byte(validYAML))
	if !result.Valid {
		t.Errorf("go: extension task should be valid, errors: %v", result.Errors)
	}
}
