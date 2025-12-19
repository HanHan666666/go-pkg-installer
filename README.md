# go-pkg-installer

A Linux installation framework written in Go, inspired by macOS `.pkg` installers.

## Overview

`go-pkg-installer` is a **Go library** that provides a configurable, extensible installation framework for Linux applications. It offers:

- **Declarative Configuration**: Define installation flows in YAML
- **Wizard UI**: Built-in tk9-based GUI with step navigation
- **Task System**: Modular tasks for file operations, configuration, and services
- **Extensibility**: Register custom screens, tasks, and guards

## Quick Start

### Installation

```bash
go get github.com/anthropics/go-pkg-installer
```

### Basic Usage

1. Create an `installer.yaml` configuration:

```yaml
product:
  name: "My Application"

flows:
  install:
    entry: welcome
    steps:
      - id: welcome
        title: "Welcome"
        screen:
          type: welcome
          content: "Welcome to My Application installer!"
      
      - id: license
        title: "License"
        screen:
          type: license
          source: "path/to/LICENSE.txt"
        guards:
          - type: mustAccept
            field: license_accepted
      
      - id: destination
        title: "Install Location"
        screen:
          type: pathPicker
          bind: install_dir
      
      - id: progress
        title: "Installing"
        screen:
          type: progress
        tasks:
          - type: shell
            command: "mkdir"
            args: ["-p", "${install_dir}"]
          - type: writeConfig
            destination: "${install_dir}/config.json"
            format: json
            content:
              installed: true
      
      - id: finish
        title: "Complete"
        screen:
          type: finish
```

2. Run the installer:

```bash
go run ./cmd/installer --config installer.yaml
```

### Command Line Options

```
Usage: installer [options]

Options:
  -config string    Path to installer configuration YAML file
  -action string    Action to perform: install, uninstall (default "install")
  -validate         Only validate the configuration file
  -headless         Run in headless/CLI mode (no GUI)
  -verbose          Enable verbose logging
  -version          Show version information
```

## Architecture

### Modules

- **Core**: Workflow engine, context management, task runner, event bus
- **Builtin**: Built-in screen types, tasks, and guards
- **Schema**: YAML/JSON schema validation
- **UI**: tk9-based graphical user interface

### Screen Types

| Type | Description |
|------|-------------|
| `welcome` | Welcome/introduction screen with rich text |
| `license` | License agreement with acceptance checkbox |
| `pathPicker` | Directory selection for install location |
| `options` | Multiple choice options |
| `summary` | Pre-install summary of selections |
| `progress` | Progress bar with task execution |
| `finish` | Completion screen |

### Built-in Tasks

| Task | Description |
|------|-------------|
| `shell` | Execute shell scripts |
| `copy` | Copy files/directories |
| `symlink` | Create symbolic links |
| `writeConfig` | Write configuration files |
| `removePath` | Remove files/directories |
| `desktop_entry` | Create .desktop files |
| `download` | Download files from URLs |
| `unpack` | Extract archives (tar, zip) |

### Guards

| Guard | Description |
|-------|-------------|
| `mustAccept` | Require checkbox acceptance |
| `diskSpace` | Check available disk space |
| `fieldNotEmpty` | Ensure field has value |
| `expression` | Custom expression evaluation |

## Library Usage

```go
package main

import (
    "github.com/anthropics/go-pkg-installer/pkg/core"
    "github.com/anthropics/go-pkg-installer/pkg/builtin"
    "github.com/anthropics/go-pkg-installer/pkg/ui"
)

func main() {
    // Register built-in tasks and guards
    builtin.RegisterAll()
    
    // Create context and workflow
    ctx := core.NewInstallContext()
    eventBus := core.NewEventBus()
    workflow := core.NewWorkflow(ctx, eventBus)
    
    // Add custom flow programmatically
    flow := &core.Flow{
        ID:    "install",
        Entry: "welcome",
        Steps: []*core.Step{
            {ID: "welcome", Title: "Welcome"},
            {ID: "finish", Title: "Complete"},
        },
    }
    workflow.AddFlow(flow)
    workflow.SelectFlow("install")
    
    // Create and run UI
    win := ui.NewInstallerWindow(ctx, workflow, eventBus)
    win.OnComplete(func() {
        // Handle completion
    })
    win.Run()
}
```

## Extending

### Custom Task

```go
func init() {
    core.Tasks.MustRegister("myTask", func(params map[string]any, ctx *core.InstallContext) (core.Task, error) {
        return &MyTask{BaseTask: core.NewBaseTask("myTask", "my-task", params)}, nil
    })
}

type MyTask struct {
    core.BaseTask
}

func (t *MyTask) Execute(ctx *core.InstallContext) error {
    // Implementation
    return nil
}

func (t *MyTask) Validate(ctx *core.InstallContext) error {
    return nil
}
```

### Custom Guard

```go
func init() {
    core.Guards.MustRegister("myGuard", func(params map[string]any, ctx *core.InstallContext) (core.Guard, error) {
        return &MyGuard{params: params}, nil
    })
}
```

## Development

### Build

```bash
make build
```

### Test

```bash
make test
```

### Lint

```bash
make lint
```

## License

MIT License - see [LICENSE](LICENSE) for details.
