# Go Package Installer - Architecture

## Overview

The Go Package Installer is a YAML-driven framework for creating Linux application installers. It provides a declarative configuration system, pluggable task execution, and an optional TK9-based GUI.

## Core Components

### 1. Schema Layer (`pkg/schema`)

Handles YAML configuration loading and validation.

```
┌─────────────────────────────────────────────────┐
│                 YAML Config                       │
│  (installer.yaml)                                │
└─────────────────┬───────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────┐
│            Schema Validator                       │
│  - JSON Schema validation                        │
│  - YAML parsing                                  │
│  - Type conversion                               │
└─────────────────┬───────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────┐
│            Config Struct                          │
│  - Product info                                  │
│  - Flows & Steps                                 │
│  - Variables                                     │
└─────────────────────────────────────────────────┘
```

### 2. Core Engine (`pkg/core`)

The heart of the installer framework.

```
┌────────────────────────────────────────────────────────────────────┐
│                          Core Engine                                 │
├────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐        │
│  │  Workflow   │    │  TaskRunner │    │   EventBus      │        │
│  │  Manager    │◄──►│             │◄──►│                 │        │
│  └─────────────┘    └─────────────┘    └─────────────────┘        │
│         │                  │                    │                    │
│         │                  │                    │                    │
│         ▼                  ▼                    ▼                    │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐        │
│  │   Guards    │    │   Tasks     │    │  Subscribers    │        │
│  │  Registry   │    │  Registry   │    │  (UI, Logging)  │        │
│  └─────────────┘    └─────────────┘    └─────────────────┘        │
│                                                                      │
└────────────────────────────────────────────────────────────────────┘
```

#### Key Components:

- **InstallContext**: Thread-safe key-value store for installation state
- **Workflow**: Manages flow selection, step navigation, and guards
- **TaskRunner**: Executes tasks with progress tracking and rollback support
- **EventBus**: Pub/sub event system for loose coupling
- **Task Registry**: Plugin registry for task types
- **Guard Registry**: Plugin registry for navigation guards

### 3. Builtin Tasks (`pkg/builtin`)

Pre-built task implementations:

| Task Type | Description |
|-----------|-------------|
| `download` | HTTP/HTTPS file downloads with progress |
| `unpack` | Archive extraction (tar.gz, zip) |
| `copy` | File/directory copying with globs |
| `symlink` | Symbolic link creation |
| `shell` | Shell command execution |
| `writeConfig` | Configuration file generation |
| `desktopEntry` | .desktop file creation |
| `removePath` | File/directory removal |

### 4. UI Layer (`pkg/ui`)

TK9-based graphical interface.

```
┌────────────────────────────────────────────────────────────────────┐
│                            Window                                    │
├────────────────────────────────────────────────────────────────────┤
│  ┌────────────────────────────────────────────────────────────┐   │
│  │                     Step Indicator                           │   │
│  │  [Welcome] ─ [License] ─ [Options] ─ [Install] ─ [Finish]   │   │
│  └────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │                     Content Area                             │   │
│  │                  (Current Screen)                            │   │
│  │                                                              │   │
│  │  Screen Types:                                               │   │
│  │  - Welcome Screen                                            │   │
│  │  - License Screen                                            │   │
│  │  - Form Screen                                               │   │
│  │  - Progress Screen                                           │   │
│  │  - Summary Screen                                            │   │
│  │  - Finish Screen                                             │   │
│  │                                                              │   │
│  └────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────┐   │
│  │              Navigation Buttons                              │   │
│  │                                                              │   │
│  │     [◄ Back]              [Cancel]              [Next ►]     │   │
│  └────────────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────────────┘
```

## Data Flow

### Installation Flow

```
User Input
    │
    ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    UI       │────►│  Workflow   │────►│   Guard     │
│  Events     │     │  Manager    │     │  Evaluation │
└─────────────┘     └─────────────┘     └─────────────┘
                           │                    │
                           │                    │
                    ┌──────┴──────┐      (pass/fail)
                    │             │             │
                    ▼             ▼             ▼
              ┌──────────┐ ┌──────────┐ ┌──────────┐
              │  Step A  │ │  Step B  │ │  Step C  │
              │  (form)  │ │(progress)│ │ (finish) │
              └──────────┘ └──────────┘ └──────────┘
                               │
                               ▼
                        ┌─────────────┐
                        │ TaskRunner  │
                        │             │
                        │  Task 1     │
                        │  Task 2     │
                        │  Task 3     │
                        └─────────────┘
                               │
                               ▼
                        ┌─────────────┐
                        │  EventBus   │
                        │  (progress, │
                        │   logs)     │
                        └─────────────┘
                               │
                               ▼
                        ┌─────────────┐
                        │    UI       │
                        │  Update     │
                        └─────────────┘
```

## Extension Points

### Custom Tasks

```go
// Register a custom task type
core.Tasks.Register("myTask", func(config map[string]any, ctx *core.InstallContext) (core.Task, error) {
    return &MyTask{
        BaseTask: core.BaseTask{
            TaskID:   "my-task",
            TaskType: "myTask",
            Config:   config,
        },
    }, nil
})
```

### Custom Guards

```go
// Register a custom guard type
core.Guards.Register("myGuard", func(config map[string]any) (core.Guard, error) {
    return &MyGuard{}, nil
})
```

### Custom Screens

```go
// Register a custom screen type
ui.RegisterScreen("myScreen", func() ui.Screen {
    return &MyScreen{}
})
```

## Thread Safety

- **InstallContext**: All operations are protected by RWMutex
- **TaskRunner**: Uses mutex for state management
- **EventBus**: Thread-safe publish/subscribe
- **Workflow**: Atomic flow/step transitions

## Error Handling

### Task Failure Policies

1. **Abort**: Stop immediately on failure (default)
2. **Skip**: Log warning and continue to next task
3. **Rollback**: Execute rollback on all completed tasks
4. **Retry**: Retry failed task up to N times

### Rollback Support

Tasks can implement `CanRollback()` and `Rollback()` methods:

```go
func (t *MyTask) CanRollback() bool {
    return true
}

func (t *MyTask) Rollback(ctx *core.InstallContext, bus *core.EventBus) error {
    // Undo the changes made by Execute()
    return nil
}
```

## File Structure

```
go-pkg-installer/
├── cmd/
│   └── installer/          # CLI entry point
├── pkg/
│   ├── core/               # Core engine
│   │   ├── context.go      # InstallContext
│   │   ├── workflow.go     # Workflow management
│   │   ├── task.go         # Task interface & runner
│   │   ├── guard.go        # Guards
│   │   ├── eventbus.go     # Event system
│   │   └── registry.go     # Plugin registries
│   ├── schema/             # Configuration
│   │   ├── config.go       # Config structures
│   │   └── validator.go    # JSON Schema validation
│   ├── builtin/            # Built-in tasks
│   │   ├── task_download.go
│   │   ├── task_unpack.go
│   │   ├── task_copy.go
│   │   ├── task_shell.go
│   │   └── ...
│   └── ui/                 # TK9 UI
│       ├── window.go       # Main window
│       ├── screens.go      # Screen management
│       └── screen_*.go     # Screen implementations
├── examples/               # Example installers
├── tests/
│   └── e2e/                # End-to-end tests
└── docs/                   # Documentation
```
