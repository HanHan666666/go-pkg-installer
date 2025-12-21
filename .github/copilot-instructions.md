# Copilot 使用说明（go-pkg-installer）

## 开发前必读，否则开发是没有意义的
- 认真阅读本说明文档，理解项目架构和设计理念。
- 认真阅读需求文档，`docs/SPECIFICATION.md`，理解用户需求和预期目标。
- 认真阅读`docs/DEVELOPER.md`,理解开发工作流和项目约定。
- 认真阅读`docs/ARCHITECTURE.md`,理解项目架构和关键模块。
- 认真阅读现有代码，理解代码结构和实现细节。
- **一定要阅读，能看到阅读记录的，一旦发现没阅读，你很快就会被关机，更换下一个AI，一定要小心。**

## 重点（极其重要）
- 在接到用户的任务的时候，先不要着急开始修改代码，要先分析需求，分析代码，列举解决方案，
- 详细的向用户说明你的思路，和你打算如何实现这个需求。
- 要分析整个项目的架构，一切都要从整个项目的角度入手，不能直接看完一个文件就写代码。
- 先问清楚、绝对不允许猜测：遇到需求或现状不确定时，先明确提问，不要主观假设；方案需先得到用户确认再开工。
- 先方案后编码：先梳理背景/现状 → 列备选方案（含改动面、影响范围、取舍理由）→ 让用户确认 → 再动手。**只有在用户确认你的方案后，才开始动手写代码, 不然你很快就会被关机，更换下一个AI，一定要小心。**
- 变更记录：完成功能后，将关键经验和约定同步到本指南，方便后续遵循。



## 代码要求
1. 代码要求结构清晰，不应付事情，长远维护考虑，遵循设计模式最佳实践，遵循项目代码风格。
2. 保证代码逻辑严谨，整洁，结构清晰，容易理解和维护，不要过度设计增加系统复杂性
3. 工程优化，以工程化，能安全正常使用不出错为主，考虑周全，遵循越复杂越容易出错，越简单越容易可控原则，一个健康的系统 越简单越可控
4. 遵循合理的组件化设计原则，要考虑组件复用性的可能。
5. 在你发现架构不合理的时候，要及时的提出来。
6. 编写代码的过程中，必须牢记以下几个原则：
    - 开闭原则（Open Closed Principle，OCP）
    - 单一职责原则（Single Responsibility Principle, SRP）
    - 里氏代换原则（Liskov Substitution Principle，LSP）
    - 依赖倒转原则（Dependency Inversion Principle，DIP）
    - 接口隔离原则（Interface Segregation Principle，ISP）
    - 合成/聚合复用原则（Composite/Aggregate Reuse Principle，CARP）
    - 最少知识原则（Least Knowledge Principle，LKP）或者迪米特法则（Law of  Demeter，LOD）


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
- UI 库使用文档：`docs/tk9_doc`

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
