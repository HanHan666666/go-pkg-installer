// Package builtin provides the shell task implementation.
package builtin

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/anthropics/go-pkg-installer/pkg/core"
)

// ShellTask executes a shell command.
type ShellTask struct {
	core.BaseTask
	Command string
	Args    []string
	WorkDir string
	Env     map[string]string
	Timeout time.Duration

	// Rollback command (optional)
	RollbackCmd string
}

// RegisterShellTask registers the shell task factory.
func RegisterShellTask() {
	core.Tasks.Register("shell", func(config map[string]any, ctx *core.InstallContext) (core.Task, error) {
		// Parse environment variables
		env := make(map[string]string)
		if envMap, ok := config["env"].(map[string]any); ok {
			for k, v := range envMap {
				if s, ok := v.(string); ok {
					env[k] = ctx.Render(s)
				}
			}
		}

		task := &ShellTask{
			BaseTask: core.BaseTask{
				TaskID:   getConfigString(config, "id"),
				TaskType: "shell",
				Config:   config,
			},
			Command:     ctx.Render(getConfigString(config, "command")),
			Args:        renderStringSlice(ctx, getConfigStringSlice(config, "args")),
			WorkDir:     ctx.Render(getConfigString(config, "workdir")),
			Env:         env,
			Timeout:     time.Duration(getConfigInt(config, "timeout", 300)) * time.Second,
			RollbackCmd: ctx.Render(getConfigString(config, "rollback_command")),
		}

		if task.TaskID == "" {
			// Use first word of command as ID
			parts := strings.Fields(task.Command)
			if len(parts) > 0 {
				task.TaskID = fmt.Sprintf("shell-%s", parts[0])
			} else {
				task.TaskID = "shell-cmd"
			}
		}

		return task, nil
	})
}

// renderStringSlice applies template rendering to a slice of strings.
func renderStringSlice(ctx *core.InstallContext, slice []string) []string {
	result := make([]string, len(slice))
	for i, s := range slice {
		result[i] = ctx.Render(s)
	}
	return result
}

// Validate validates the shell task configuration.
func (t *ShellTask) Validate() error {
	if t.Command == "" {
		return errors.New("shell: command is required")
	}
	return nil
}

// Execute runs the shell command.
func (t *ShellTask) Execute(ctx *core.InstallContext, bus *core.EventBus) error {
	ctx.AddLog(core.LogInfo, fmt.Sprintf("Executing: %s %s", t.Command, strings.Join(t.Args, " ")))

	// Default timeout if not set
	timeout := t.Timeout
	if timeout == 0 {
		timeout = 300 * time.Second
	}

	// Build command
	var cmd *exec.Cmd
	if len(t.Args) > 0 {
		cmd = exec.Command(t.Command, t.Args...)
	} else {
		// Use shell to execute the command
		cmd = exec.Command("sh", "-c", t.Command)
	}

	// Set working directory
	if t.WorkDir != "" {
		cmd.Dir = t.WorkDir
	}

	// Set environment
	if len(t.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range t.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			ctx.AddLog(core.LogError, fmt.Sprintf("Command failed: %v\nstderr: %s", err, stderr.String()))
			return fmt.Errorf("command failed: %w\nstderr: %s", err, stderr.String())
		}
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return fmt.Errorf("command timed out after %v", timeout)
	}

	// Log output
	if stdout.Len() > 0 {
		ctx.AddLog(core.LogInfo, fmt.Sprintf("stdout: %s", stdout.String()))
	}

	return nil
}

// CanRollback returns true if a rollback command is configured.
func (t *ShellTask) CanRollback() bool {
	return t.RollbackCmd != ""
}

// Rollback executes the rollback command.
func (t *ShellTask) Rollback(ctx *core.InstallContext, bus *core.EventBus) error {
	if t.RollbackCmd == "" {
		return nil
	}

	ctx.AddLog(core.LogInfo, fmt.Sprintf("Rolling back: %s", t.RollbackCmd))

	cmd := exec.Command("sh", "-c", t.RollbackCmd)
	if t.WorkDir != "" {
		cmd.Dir = t.WorkDir
	}

	if len(t.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range t.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		ctx.AddLog(core.LogWarn, fmt.Sprintf("Rollback command failed: %v\nstderr: %s", err, stderr.String()))
		return fmt.Errorf("rollback command failed: %w", err)
	}

	return nil
}
