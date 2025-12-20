# Developer Guide

This guide explains how to use the Go Package Installer framework to build custom installers.

## Getting Started

### Installation

```bash
go get github.com/HanHan666666/go-pkg-installer
```

### Quick Start

1. Create an installer configuration (`installer.yaml`)
2. Build the installer binary
3. Package your application with the installer

## Using the Library

### Basic Usage

```go
package main

import (
    "github.com/HanHan666666/go-pkg-installer/pkg/builtin"
    "github.com/HanHan666666/go-pkg-installer/pkg/core"
    "github.com/HanHan666666/go-pkg-installer/pkg/schema"
    "github.com/HanHan666666/go-pkg-installer/pkg/ui"
)

func main() {
    // Register builtin tasks and guards
    builtin.RegisterAll()
    core.RegisterBuiltinGuards()

    // Load configuration
    config, err := schema.LoadConfigFile("installer.yaml")
    if err != nil {
        panic(err)
    }

    // Create core components
    ctx := core.NewInstallContext()
    eventBus := core.NewEventBus()
    workflow := core.NewWorkflow(ctx, eventBus)

    // Initialize from config
    for name, vars := range config.Variables {
        ctx.Set(name, vars)
    }

    // Add flows from config
    for flowName, flowCfg := range config.Flows {
        flow := &core.Flow{
            ID:    flowName,
            Entry: flowCfg.Entry,
        }
        for _, stepCfg := range flowCfg.Steps {
            step := &core.Step{
                ID:     stepCfg.ID,
                Title:  stepCfg.Title,
                Config: stepCfg,
            }
            flow.Steps = append(flow.Steps, step)
        }
        workflow.AddFlow(flow)
    }

    // Create and run UI
    window := ui.NewWindow(config.Product.Name, workflow, ctx, eventBus)
    window.Run()
}
```

### Headless Installation

For automated/silent installations:

```go
func runHeadless(config *schema.Config, vars map[string]string) error {
    builtin.RegisterAll()
    core.RegisterBuiltinGuards()

    ctx := core.NewInstallContext()
    eventBus := core.NewEventBus()
    
    // Apply command-line variables
    for k, v := range vars {
        ctx.Set(k, v)
    }

    workflow := core.NewWorkflow(ctx, eventBus)
    runner := core.NewTaskRunner(ctx, eventBus)

    // Select install flow
    workflow.SelectFlow("install")

    // Execute all steps
    for !workflow.IsComplete() {
        step := workflow.CurrentStep()
        
        // Execute step tasks
        if step.Config != nil && len(step.Config.Tasks) > 0 {
            for _, taskCfg := range step.Config.Tasks {
                cfg := core.TaskConfig{
                    Type:   taskCfg.Type,
                    Params: taskCfg.Params,
                }
                if err := runner.QueueConfig(cfg); err != nil {
                    return err
                }
            }
            if err := runner.Run(); err != nil {
                return err
            }
        }

        // Move to next step
        if !workflow.IsLastStep() {
            workflow.Next()
        } else {
            workflow.Complete()
        }
    }

    return nil
}
```

## Creating Custom Tasks

### Task Interface

```go
type Task interface {
    ID() string
    Type() string
    Validate() error
    Execute(ctx *InstallContext, bus *EventBus) error
    CanRollback() bool
    Rollback(ctx *InstallContext, bus *EventBus) error
}
```

### Example: Custom Task

```go
package custom

import (
    "github.com/HanHan666666/go-pkg-installer/pkg/core"
)

// DatabaseTask creates a database
type DatabaseTask struct {
    core.BaseTask
    Host     string
    Port     int
    Database string
}

func init() {
    RegisterDatabaseTask()
}

func RegisterDatabaseTask() {
    core.Tasks.Register("createDatabase", func(config map[string]any, ctx *core.InstallContext) (core.Task, error) {
        task := &DatabaseTask{
            BaseTask: core.BaseTask{
                TaskID:   "create-database",
                TaskType: "createDatabase",
                Config:   config,
            },
        }

        // Get parameters with variable substitution
        if host, ok := config["host"].(string); ok {
            task.Host = ctx.Render(host)
        }
        if port, ok := config["port"].(float64); ok {
            task.Port = int(port)
        }
        if db, ok := config["database"].(string); ok {
            task.Database = ctx.Render(db)
        }

        return task, nil
    })
}

func (t *DatabaseTask) Validate() error {
    if t.Host == "" {
        return errors.New("host is required")
    }
    if t.Database == "" {
        return errors.New("database name is required")
    }
    return nil
}

func (t *DatabaseTask) Execute(ctx *core.InstallContext, bus *core.EventBus) error {
    bus.PublishLog(core.LogInfo, "Creating database: "+t.Database)
    
    // Create database logic here...
    
    bus.PublishLog(core.LogInfo, "Database created successfully")
    return nil
}

func (t *DatabaseTask) CanRollback() bool {
    return true
}

func (t *DatabaseTask) Rollback(ctx *core.InstallContext, bus *core.EventBus) error {
    bus.PublishLog(core.LogInfo, "Dropping database: "+t.Database)
    // Drop database logic here...
    return nil
}
```

Usage in configuration:

```yaml
tasks:
  - type: createDatabase
    host: localhost
    port: 5432
    database: "${app_name}_db"
```

## Creating Custom Guards

### Guard Interface

```go
type Guard interface {
    Check(ctx *InstallContext) error
}
```

### Example: Custom Guard

```go
package custom

import (
    "errors"
    "os/exec"
    
    "github.com/HanHan666666/go-pkg-installer/pkg/core"
)

func init() {
    RegisterCommandExistsGuard()
}

func RegisterCommandExistsGuard() {
    core.Guards.Register("commandExists", func(config map[string]any) (core.Guard, error) {
        command, _ := config["command"].(string)
        message, _ := config["message"].(string)
        
        if command == "" {
            return nil, errors.New("command is required")
        }
        if message == "" {
            message = "Required command not found: " + command
        }
        
        return &CommandExistsGuard{
            Command: command,
            Message: message,
        }, nil
    })
}

type CommandExistsGuard struct {
    Command string
    Message string
}

func (g *CommandExistsGuard) Check(ctx *core.InstallContext) error {
    _, err := exec.LookPath(g.Command)
    if err != nil {
        return errors.New(g.Message)
    }
    return nil
}
```

Usage in configuration:

```yaml
guards:
  - type: commandExists
    command: docker
    message: "Docker is required to install this application"
```

## Creating Custom Screens

### Screen Interface

```go
type Screen interface {
    ID() string
    Init(step *core.Step, ctx *core.InstallContext, bus *core.EventBus)
    Render(parent tk.Widget) tk.Widget
    OnEnter()
    OnLeave()
}
```

### Registering Custom Screens

```go
package custom

import (
    "github.com/HanHan666666/go-pkg-installer/pkg/core"
    "github.com/HanHan666666/go-pkg-installer/pkg/ui"
    tk "modernc.org/tk9.0"
)

func init() {
    ui.RegisterScreen("customScreen", func() ui.Screen {
        return &CustomScreen{}
    })
}

type CustomScreen struct {
    ui.BaseScreen
    content tk.Widget
}

func (s *CustomScreen) Render(parent tk.Widget) tk.Widget {
    frame := tk.TFrame(parent)
    
    // Build your custom UI here
    tk.TLabel(frame, "-text", "Custom Content")
    
    s.content = frame
    return frame
}

func (s *CustomScreen) OnEnter() {
    // Called when screen becomes active
}

func (s *CustomScreen) OnLeave() {
    // Called when leaving screen
}
```

## Event System

### Subscribing to Events

```go
eventBus.Subscribe(core.EventProgress, func(e core.Event) {
    if p := e.ProgressPayload(); p != nil {
        fmt.Printf("Progress: %.0f%% - %s\n", p.Progress*100, p.Message)
    }
})

eventBus.Subscribe(core.EventLog, func(e core.Event) {
    if p := e.LogPayload(); p != nil {
        fmt.Printf("[%s] %s\n", p.Level, p.Message)
    }
})

eventBus.Subscribe(core.EventTaskComplete, func(e core.Event) {
    if p := e.TaskPayload(); p != nil {
        fmt.Printf("Task completed: %s\n", p.TaskID)
    }
})
```

### Event Types

| Event Type | Description |
|------------|-------------|
| `EventProgress` | Task progress update |
| `EventLog` | Log message |
| `EventTaskStart` | Task execution started |
| `EventTaskComplete` | Task execution completed |
| `EventTaskError` | Task execution failed |
| `EventStepChange` | Workflow step changed |
| `EventFlowComplete` | Flow finished |

## Error Handling

### Failure Policies

```go
runner := core.NewTaskRunner(ctx, eventBus)

// Stop on first failure (default)
runner.SetFailurePolicy(core.FailureAbort)

// Skip failed tasks and continue
runner.SetFailurePolicy(core.FailureSkip)

// Rollback completed tasks on failure
runner.SetFailurePolicy(core.FailureRollback)

// Retry failed tasks
runner.SetFailurePolicy(core.FailureRetry)
runner.SetMaxRetries(3)
```

## Testing

### Unit Testing Tasks

```go
func TestMyTask(t *testing.T) {
    ctx := core.NewInstallContext()
    ctx.Set("install_dir", t.TempDir())
    
    eventBus := core.NewEventBus()
    
    task := &MyTask{
        // ... configuration
    }
    
    if err := task.Validate(); err != nil {
        t.Fatalf("Validation failed: %v", err)
    }
    
    if err := task.Execute(ctx, eventBus); err != nil {
        t.Fatalf("Execution failed: %v", err)
    }
    
    // Verify results
}
```

### Integration Testing

```go
func TestInstallFlow(t *testing.T) {
    builtin.RegisterAll()
    core.RegisterBuiltinGuards()
    
    ctx := core.NewInstallContext()
    ctx.Set("install_dir", t.TempDir())
    
    eventBus := core.NewEventBus()
    runner := core.NewTaskRunner(ctx, eventBus)
    
    // Queue tasks
    runner.QueueConfig(core.TaskConfig{
        Type: "shell",
        Params: map[string]any{
            "command": "mkdir",
            "args":    []any{"-p", "${install_dir}/bin"},
        },
    })
    
    // Execute
    if err := runner.Run(); err != nil {
        t.Fatalf("Failed: %v", err)
    }
    
    // Verify
    if _, err := os.Stat(filepath.Join(ctx.GetString("install_dir"), "bin")); err != nil {
        t.Error("Directory not created")
    }
}
```

## Building the Installer

### Standard Build

```bash
go build -o installer ./cmd/installer
```

### Static Build (for distribution)

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o installer ./cmd/installer
```

### Embedding Resources

Use Go's embed package to include files in the binary:

```go
//go:embed installer.yaml
var configData []byte

//go:embed files/*
var filesFS embed.FS
```

## Best Practices

1. **Use meaningful task IDs** - Makes logs and debugging easier
2. **Validate early** - Check prerequisites in guards before installation
3. **Implement rollback** - For critical tasks that modify system state
4. **Use variable substitution** - Keep configuration DRY
5. **Test in isolation** - Unit test each task independently
6. **Log appropriately** - Use proper log levels (Info, Warn, Error)
7. **Handle cancellation** - Respect user cancellation requests
8. **Provide feedback** - Update progress during long operations
