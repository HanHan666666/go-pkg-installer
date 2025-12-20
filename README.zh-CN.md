# go-pkg-installer

面向 Linux 的 Go 安装框架，灵感来自 macOS `.pkg` 安装器。

- English version: `README.md`

## 项目概述

`go-pkg-installer` 是一个 **Go 库**，提供可配置、可扩展的 Linux 应用安装框架，适用于构建图形化安装器与卸载器。

### 核心特点

- **可插拔 UI**：内置 tk9 向导界面，也支持自定义 screen
- **可插拔任务**：内置常用任务类型，并支持注册自定义 task
- **可配置流程**：YAML 驱动的流程，支持 guards/分支/跳转
- **卸载支持**：同一引擎支持 install/uninstall/repair/upgrade 动作

## 截图

![Screenshot](docs/assets/screenshot.png)

> 请将截图放到 `docs/assets/screenshot.png`。

## 快速开始

### 安装依赖

```bash
go get github.com/HanHan666666/go-pkg-installer
```

### 运行内置 GUI 安装器

```bash
go run ./cmd/installer --config installer.yaml
```

### 命令行参数

```
Usage: installer [options]

Options:
  -config string    安装器配置 YAML 路径
  -action string    动作: install, uninstall (默认 "install")
  -validate         仅校验配置文件
  -headless         纯命令行模式（无 GUI）
  -verbose          输出详细日志
  -version          显示版本信息
```

## 配置说明

### 最小示例

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

      - id: summary
        title: "Ready to Install"
        screen:
          type: summary

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

### 支持的动作

- `install`: 安装流程
- `uninstall`: 卸载流程
- `upgrade`: 升级流程（可选）
- `repair`: 修复流程（可选）

## 架构说明

### 核心模块

- **core**：流程引擎、上下文、任务执行、事件总线
- **builtin**：内置任务/guards/screen 默认实现
- **schema**：YAML/JSON 结构校验
- **ui**：tk9 图形界面渲染

### 内置 Screen 类型

| 类型 | 说明 |
|------|------|
| `welcome` | 欢迎页（富文本） |
| `license` | 许可协议（滚动到末尾/勾选） |
| `pathPicker` | 安装目录选择 |
| `options` | 选项配置 |
| `summary` | 安装前摘要 |
| `progress` | 进度与日志 |
| `finish` | 完成页 |

### 内置 Task 类型

| Task | 说明 |
|------|------|
| `shell` | 执行 shell 命令 |
| `copy` | 复制文件/目录 |
| `symlink` | 创建软链接 |
| `writeConfig` | 写配置文件 |
| `removePath` | 删除文件/目录 |
| `desktop_entry` | 创建 .desktop 文件 |
| `download` | 下载文件 |
| `unpack` | 解压归档（tar/zip） |

### 内置 Guards

| Guard | 说明 |
|-------|------|
| `mustAccept` | 必须勾选同意 |
| `diskSpace` | 磁盘空间检查 |
| `fieldNotEmpty` | 字段非空校验 |
| `expression` | 自定义表达式 |

## 扩展开发

### 自定义 Task

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
    // 实现逻辑
    return nil
}

func (t *MyTask) Validate(ctx *core.InstallContext) error {
    return nil
}
```

### 自定义 Guard

```go
func init() {
    core.Guards.MustRegister("myGuard", func(params map[string]any, ctx *core.InstallContext) (core.Guard, error) {
        return &MyGuard{params: params}, nil
    })
}
```

### 自定义 Screen

```go
func init() {
    ui.RegisterScreenRenderer("myScreen", NewMyScreen)
}
```

## 开发

### 构建

```bash
make build
```

### 测试

```bash
make test
```

### Lint

```bash
make lint
```

## License

MIT License - 详见 [LICENSE](LICENSE)。
