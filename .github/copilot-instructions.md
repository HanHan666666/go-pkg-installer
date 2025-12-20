# Copilot 使用说明（go-pkg-installer）

## 总览
- 这是一个 **Go library-first** 的 Linux 安装框架：YAML 配置 → `pkg/schema` 校验 → `pkg/core` 工作流/任务引擎 → 可选的 TK9 GUI（`pkg/ui`）。
- 参考应用入口在 `cmd/installer/main.go`（命令行参数、配置加载、环境探测、提权、GUI/Headless）。

## 关键模块（从哪里开始读）
- 工作流状态机：`pkg/core/engine.go`（`Workflow`/`Flow`/`Step`，guards + 分支）。
- 任务运行时：`pkg/core/task.go`（`TaskRunner`，失败策略、回滚支持）。
- 共享状态 + 模板渲染：`pkg/core/context.go`（`InstallContext` 点路径 key + `${...}` 渲染）。
- 可扩展机制：`pkg/core/registry.go`（全局注册表 `Tasks`/`Guards`/`Screens`；支持 `go:` 类型前缀）。
- 内置能力注册：`pkg/builtin/builtin.go` 与 `pkg/builtin/task_*.go`。
- GUI 适配层：`pkg/ui/window.go` + `pkg/ui/screen_*.go`（TK9；goroutine 更新 UI 通过 `PostEvent`）。
- 配置示例：`examples/demo-installer.yaml`。

## 开发者工作流（直接用这些命令）
- 构建：`make build`（生成 `build/installer`）。
- 单元测试（race）：`make test`（实际运行 `go test -v -race ./...`）。
- Lint（本地可选）：`make lint`（需要 `golangci-lint`）。
- 运行参考安装器：`go run ./cmd/installer --config examples/demo-installer.yaml`。
- 仅校验配置：`go run ./cmd/installer --config … --validate`。
- 无界面模式：`go run ./cmd/installer --config … --headless`。

## 项目约定与模式
- **Context key 使用点路径**（例如 `license.accepted`、`install.dir`、`product.name`）。优先用 `ctx.Set("a.b", v)` 写入，并用 `ctx.Get*` 系列读取。
- **字符串插值使用 `${path}`**，由 `InstallContext.Render` 解析（见 `cmd/installer/main.go` 与 `pkg/core/context.go`）。新增支持模板的配置字段时，应在执行时渲染。
- **任务/守卫通过注册表创建**：
  - Task 类型来自 YAML `tasks: - type: ...`，由 `TaskRunner.QueueConfig` 通过 registry 创建。
  - Guard 类型来自 YAML `guards: - type: ...`，由 `Workflow.CanGoNext` 评估。
  - 支持 `go:` 前缀扩展（例如 `type: go:myTask`），在查找 registry 前会去掉 `go:`。
- **新增内置 Task 的方式**：
  - 实现 `core.Task`（通常嵌入 `core.BaseTask`），并在 `core.Tasks` 注册 factory。
  - 参考 `pkg/builtin/task_*.go` 的写法，并在相邻位置补充/更新测试（`pkg/builtin/*_test.go`）。
- **UI Screen**：
  - Screen 渲染器基于 TK9 组件；从 goroutine 更新 UI 必须通过 `PostEvent`（见 `pkg/ui/screen_progress.go`）。
  - Screen 类型名与 YAML `screen.type` 匹配（renderer 的自动注册见 `pkg/ui/window.go`）。
- **Schema/校验**：
  - 配置通过 `pkg/schema/validator.go` 使用内嵌 `installer.schema.json` 做 schema 校验。

## 需要注意的集成点
- 提权 + 预检：参考应用会调用 `core.DetectEnv(ctx)`，并可能基于配置/上下文提权（见 `cmd/installer/main.go`）。新增需要 root 的任务时，确保其在 `privilege.strategy` 下行为合理。
- 事件驱动的进度/日志：任务应通过 `EventBus` 发布 progress/log，使 Progress 屏能展示（见 `pkg/ui/screen_progress.go`，事件类型在 `pkg/core`）。
