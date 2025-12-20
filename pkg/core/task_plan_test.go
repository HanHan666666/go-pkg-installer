package core

import "testing"

func TestBuildTaskPlan(t *testing.T) {
	flow := &FlowConfig{
		Entry: "start",
		Steps: []*StepConfig{
			{
				ID:    "start",
				Title: "Start",
				Tasks: []TaskConfig{
					{
						Type: "copy",
						ID:   "copy-files",
						Params: map[string]any{
							"requirePrivilege": true,
						},
					},
				},
			},
		},
	}

	plan := BuildTaskPlan(flow)
	if plan == nil || len(plan.Tasks) != 1 {
		t.Fatalf("expected plan with one task")
	}
	if plan.Tasks[0].Description != "copy-files" {
		t.Fatalf("expected description to use task ID, got %q", plan.Tasks[0].Description)
	}
	if !plan.Tasks[0].RequiresRoot {
		t.Fatalf("expected RequiresRoot to be true")
	}
}
