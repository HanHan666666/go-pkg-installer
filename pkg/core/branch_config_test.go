package core

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBranchConfigUnmarshalNewFields(t *testing.T) {
	input := `
condition: env.distro
branches:
  ubuntu: step_ubuntu
default: step_default
`
	var cfg BranchConfig
	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Condition != "env.distro" {
		t.Fatalf("expected condition to be set, got %q", cfg.Condition)
	}
	if cfg.Branches["ubuntu"] != "step_ubuntu" {
		t.Fatalf("expected branch mapping, got %#v", cfg.Branches)
	}
	if cfg.Default != "step_default" {
		t.Fatalf("expected default, got %q", cfg.Default)
	}
}

func TestBranchConfigUnmarshalLegacyFields(t *testing.T) {
	input := `
when: env.isRoot
then: step_root
else: step_user
`
	var cfg BranchConfig
	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Condition != "env.isRoot" {
		t.Fatalf("expected condition to be set, got %q", cfg.Condition)
	}
	if cfg.Branches["true"] != "step_root" {
		t.Fatalf("expected then mapping, got %#v", cfg.Branches)
	}
	if cfg.Default != "step_user" {
		t.Fatalf("expected else mapping, got %q", cfg.Default)
	}
}
