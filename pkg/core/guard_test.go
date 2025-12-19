package core

import (
	"testing"
)

func TestMustAcceptGuard(t *testing.T) {
	// Create guard
	guard, err := NewMustAcceptGuard(map[string]any{
		"field": "license.accepted",
	})
	if err != nil {
		t.Fatalf("Failed to create guard: %v", err)
	}

	if guard.Type() != "mustAccept" {
		t.Errorf("Expected type 'mustAccept', got %s", guard.Type())
	}

	ctx := NewInstallContext()

	// Should fail when not accepted
	err = guard.Check(ctx)
	if err == nil {
		t.Error("Guard should fail when field is not set")
	}

	// Should fail when explicitly false
	ctx.Set("license.accepted", false)
	err = guard.Check(ctx)
	if err == nil {
		t.Error("Guard should fail when field is false")
	}

	// Should pass when accepted
	ctx.Set("license.accepted", true)
	err = guard.Check(ctx)
	if err != nil {
		t.Errorf("Guard should pass when field is true: %v", err)
	}
}

func TestMustAcceptGuardCustomMessage(t *testing.T) {
	guard, _ := NewMustAcceptGuard(map[string]any{
		"field":   "terms.accepted",
		"message": "You must accept the terms",
	})

	if guard.Message() != "You must accept the terms" {
		t.Errorf("Expected custom message, got %s", guard.Message())
	}

	ctx := NewInstallContext()
	err := guard.Check(ctx)
	if err.Error() != "You must accept the terms" {
		t.Errorf("Error message should match: %s", err.Error())
	}
}

func TestMustAcceptGuardMissingField(t *testing.T) {
	_, err := NewMustAcceptGuard(map[string]any{})
	if err == nil {
		t.Error("Should error when field is missing")
	}

	_, err = NewMustAcceptGuard(map[string]any{"field": ""})
	if err == nil {
		t.Error("Should error when field is empty")
	}
}

func TestDiskSpaceGuard(t *testing.T) {
	guard, err := NewDiskSpaceGuard(map[string]any{
		"minMB": 500,
	})
	if err != nil {
		t.Fatalf("Failed to create guard: %v", err)
	}

	if guard.Type() != "diskSpace" {
		t.Errorf("Expected type 'diskSpace', got %s", guard.Type())
	}

	ctx := NewInstallContext()

	// Should fail when disk space is low
	ctx.Env.DiskFreeMB = 100
	err = guard.Check(ctx)
	if err == nil {
		t.Error("Guard should fail when disk space is insufficient")
	}

	// Should pass when enough space
	ctx.Env.DiskFreeMB = 1000
	err = guard.Check(ctx)
	if err != nil {
		t.Errorf("Guard should pass when disk space is sufficient: %v", err)
	}

	// Exactly at threshold
	ctx.Env.DiskFreeMB = 500
	err = guard.Check(ctx)
	if err != nil {
		t.Errorf("Guard should pass when disk space equals minimum: %v", err)
	}
}

func TestDiskSpaceGuardNumericTypes(t *testing.T) {
	// Test int
	g1, err := NewDiskSpaceGuard(map[string]any{"minMB": 100})
	if err != nil {
		t.Errorf("Should accept int: %v", err)
	}
	if g1.(*DiskSpaceGuard).MinMB != 100 {
		t.Error("MinMB should be 100")
	}

	// Test int64
	g2, err := NewDiskSpaceGuard(map[string]any{"minMB": int64(200)})
	if err != nil {
		t.Errorf("Should accept int64: %v", err)
	}
	if g2.(*DiskSpaceGuard).MinMB != 200 {
		t.Error("MinMB should be 200")
	}

	// Test float64 (common from JSON/YAML)
	g3, err := NewDiskSpaceGuard(map[string]any{"minMB": float64(300)})
	if err != nil {
		t.Errorf("Should accept float64: %v", err)
	}
	if g3.(*DiskSpaceGuard).MinMB != 300 {
		t.Error("MinMB should be 300")
	}
}

func TestDiskSpaceGuardInvalid(t *testing.T) {
	// Missing minMB
	_, err := NewDiskSpaceGuard(map[string]any{})
	if err == nil {
		t.Error("Should error when minMB is missing")
	}

	// Zero minMB
	_, err = NewDiskSpaceGuard(map[string]any{"minMB": 0})
	if err == nil {
		t.Error("Should error when minMB is zero")
	}

	// Negative minMB
	_, err = NewDiskSpaceGuard(map[string]any{"minMB": -100})
	if err == nil {
		t.Error("Should error when minMB is negative")
	}

	// Wrong type
	_, err = NewDiskSpaceGuard(map[string]any{"minMB": "500"})
	if err == nil {
		t.Error("Should error when minMB is string")
	}
}

func TestFieldNotEmptyGuard(t *testing.T) {
	guard, err := NewFieldNotEmptyGuard(map[string]any{
		"field": "install.dir",
	})
	if err != nil {
		t.Fatalf("Failed to create guard: %v", err)
	}

	if guard.Type() != "fieldNotEmpty" {
		t.Errorf("Expected type 'fieldNotEmpty', got %s", guard.Type())
	}

	ctx := NewInstallContext()

	// Should fail when empty
	err = guard.Check(ctx)
	if err == nil {
		t.Error("Guard should fail when field is empty")
	}

	// Should pass when set
	ctx.Set("install.dir", "/opt/myapp")
	err = guard.Check(ctx)
	if err != nil {
		t.Errorf("Guard should pass when field is set: %v", err)
	}
}

func TestFieldNotEmptyGuardMissingField(t *testing.T) {
	_, err := NewFieldNotEmptyGuard(map[string]any{})
	if err == nil {
		t.Error("Should error when field is missing")
	}
}

func TestExpressionGuard(t *testing.T) {
	guard, err := NewExpressionGuard(map[string]any{
		"expression": "env.isRoot",
		"expected":   true,
		"message":    "Root access required",
	})
	if err != nil {
		t.Fatalf("Failed to create guard: %v", err)
	}

	if guard.Type() != "expression" {
		t.Errorf("Expected type 'expression', got %s", guard.Type())
	}

	ctx := NewInstallContext()

	// Should fail when not root
	ctx.Env.IsRoot = false
	err = guard.Check(ctx)
	if err == nil {
		t.Error("Guard should fail when condition not met")
	}

	// Should pass when root
	ctx.Env.IsRoot = true
	err = guard.Check(ctx)
	if err != nil {
		t.Errorf("Guard should pass when condition met: %v", err)
	}
}

func TestExpressionGuardStringComparison(t *testing.T) {
	guard, _ := NewExpressionGuard(map[string]any{
		"expression": "env.distro",
		"expected":   "ubuntu",
	})

	ctx := NewInstallContext()
	ctx.Env.Distro = "fedora"

	err := guard.Check(ctx)
	if err == nil {
		t.Error("Guard should fail when distro doesn't match")
	}

	ctx.Env.Distro = "ubuntu"
	err = guard.Check(ctx)
	if err != nil {
		t.Errorf("Guard should pass when distro matches: %v", err)
	}
}

func TestExpressionGuardDefaultExpected(t *testing.T) {
	// No expected value - defaults to true
	guard, _ := NewExpressionGuard(map[string]any{
		"expression": "feature.enabled",
	})

	ctx := NewInstallContext()
	ctx.Set("feature.enabled", true)

	err := guard.Check(ctx)
	if err != nil {
		t.Errorf("Guard should pass when value is true: %v", err)
	}

	ctx.Set("feature.enabled", false)
	err = guard.Check(ctx)
	if err == nil {
		t.Error("Guard should fail when value is false")
	}
}

func TestExpressionGuardFieldNotFound(t *testing.T) {
	guard, _ := NewExpressionGuard(map[string]any{
		"expression": "nonexistent.field",
	})

	ctx := NewInstallContext()
	err := guard.Check(ctx)
	if err == nil {
		t.Error("Guard should fail when field not found")
	}
}

func TestExpressionGuardMissingExpression(t *testing.T) {
	_, err := NewExpressionGuard(map[string]any{})
	if err == nil {
		t.Error("Should error when expression is missing")
	}
}

func TestRegisterBuiltinGuards(t *testing.T) {
	Guards.Clear()
	RegisterBuiltinGuards()

	expectedGuards := []string{"mustAccept", "diskSpace", "fieldNotEmpty", "expression"}
	for _, name := range expectedGuards {
		if !Guards.Has(name) {
			t.Errorf("Built-in guard %s should be registered", name)
		}
	}
}
