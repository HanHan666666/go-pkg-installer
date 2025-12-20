# Linux 安装流程框架（Go Library First）设计需求文档（初版）

> 目标：做一个类似 macOS `.pkg` 安装器的 **Linux 安装流程框架**，并以 **Go Library** 为主形态。它不是“某一个固定安装向导”，而是一个 **可配置、可扩展、可插拔流程** 的安装基础设施：能灵活增删步骤、定制每一步界面与行为，适配不同应用的安装场景，并支持卸载与后续扩展（如 upgrade/repair）。

> 定位升级：从“一个安装器”升级为“Linux 安装流程基础设施（Install + Uninstall）”，提供可组合的核心能力与参考应用。

---

## 0. 形态与使用层级（Library First）

### 0.1 主形态

* **go-pkg-installer 是 Go Library**，提供流程引擎、任务运行时、屏幕渲染接口与注册机制。
* CLI/GUI 只是基于该库的 **Reference App**，用于演示与直接落地。
* 内置功能保持最小可用，业务能力主要通过扩展实现。

### 0.2 模块划分

* **Core**：流程引擎、上下文、任务运行时、事件总线、Schema 校验
* **Builtin**：内置 Screen/Task/Guard 实现（可直接用）
* **Adapter**：平台适配层（提权、systemd、包管理器、桌面环境）
* **Reference App**：示例 GUI/CLI，展示如何组合与扩展

### 0.3 两种使用层级

1. **纯配置**：仅写 `installer.yaml`，使用内置 Screen/Task/Guard
2. **Go 扩展**：引入库，注册自定义 Screen/Task/Guard

---

## 1. 参考界面分析（你给的截图）

该安装器界面由 4 个稳定区块组成：

1. **顶部标题栏**

* 窗口标题：如“安装 XXX”
* 常见：最小化/关闭/帮助/锁（有些安装器会显示权限状态）

2. **左侧步骤栏（Step Sidebar / Wizard Steps）**

* 纵向列表，显示当前步骤与未完成步骤（常用圆点/勾选/灰显）
* 示例步骤：介绍 → 许可阅读 → 目标位置 → 安装类型 → 安装 → 摘要
* 左下角品牌 Logo

3. **右侧主内容区（Content Area）**

* 标题/说明文案
* 可滚动内容（截图里右侧有滚动条）
* 可包含：富文本、图片、表单控件、提示、链接（如“用户许可协议”）

4. **底部导航（Footer Navigation）**

* 左侧可留空或放“取消”
* 右侧关键按钮：返回 / 继续（继续在末尾变为“安装/完成”）
* 按钮状态受校验控制（不可继续时置灰）

> 结论：这个 UI 本质是 **“向导 + 步骤导航 + 可滚动内容 + 底部导航”** 的经典模型。我们要做的是把“步骤、界面、行为、安装逻辑”全部模块化与配置化。

---

## 2. 产品目标与边界

### 2.1 产品目标

* 提供一个 **Linux 安装流程框架（Go Library）**，支持：

  * 自定义流程：任意增删步骤、调整顺序、条件分支
  * 自定义每一步界面：富文本/图片/表单/进度/日志/协议阅读等
  * 自定义每一步动作：校验、下载、解压、拷贝、写配置、创建服务、权限处理等
  * 可复用：同一套框架服务多个应用（通过配置/插件适配）
  * 生命周期：安装 + 卸载，且共享同一引擎

### 2.2 非目标（初版不做或不承诺）

* 不做“应用商店/多应用统一安装中心”（除非你明确要）
* 不强依赖某一种打包格式（deb/rpm/AppImage/flatpak/snap 都可适配，但框架先抽象）

---

## 3. 核心使用场景（User Stories）

1. 作为软件厂商，我希望只改一个配置（或少量代码），就能生成一个带品牌与流程的安装器。
2. 作为安装器设计者，我希望能把“协议页”替换成我自己的 UI（比如多语言、PDF、滚动到底才能继续）。
3. 作为运维/企业客户，我希望安装过程可无人值守（CLI 模式）并输出结构化日志，但也要有 GUI 模式给普通用户。
4. 作为开发者，我希望安装步骤可以按条件显示（例如检测到已安装旧版本就出现“升级方式”步骤）。
5. 作为平台团队，我希望将其作为库嵌入到自有安装器中，并可通过 Go 扩展私有任务与界面。
6. 作为运维，我希望卸载流程与安装流程一样可配置，且能选择是否保留用户数据。

---

## 4. 功能需求（FR）

### FR-1：流程编排（Workflow）

* 支持定义流程（Wizard Flow）：

  * steps：步骤列表（id、标题、类型、路由、是否可跳转）
  * next/prev：顺序控制
  * guard：前进条件（校验/依赖状态）
  * branch：条件分支（如 detected_old_version ? upgrade_flow : clean_install_flow）
* 支持运行时动态变更步骤：

  * 进入安装器后，根据检测结果插入/移除/禁用步骤
  * Step Sidebar 自动刷新
* 支持 **Action + 多 Flow**：

  * `action` 作为流程选择入口（install/uninstall/upgrade/repair）
  * `flows` 下可声明多个 Flow
  * CLI/GUI 通过 `--action` 或 UI 入口选择 Flow

### FR-2：页面类型（Screen Types）

内置可用页面类型（可扩展）：

* Welcome/介绍页：富文本 + 图片 + 外链
* License/协议页：加载文本/HTML/Markdown/PDF（至少支持文本/HTML），支持“滚动到底才能继续”
* Destination/安装路径：选择目录、空间检查、权限提示
* InstallType/安装类型：典型选项（仅安装主程序/包含服务/桌面图标/开机启动等）
* Summary/摘要：展示将要执行的操作清单
* Progress/安装中：进度条、分阶段任务、实时日志
* Finish/完成：成功/失败、打开应用、查看日志、重试

> 要求：页面 UI 与安装行为解耦（页面负责收集输入与展示状态，行为由任务系统执行）。

### FR-3：任务系统（Install Tasks）

* 安装动作以“任务（Task）”形式声明与执行：

  * DownloadTask、UnpackTask、CopyTask、SymlinkTask、WriteConfigTask、CreateDesktopEntryTask、SystemdServiceTask、DbusServiceTask、PermissionTask、RollbackTask 等
* 每个任务：

  * 输入参数（来自配置 + 用户表单 + 运行时检测）
  * 进度与日志输出（可订阅）
  * 可取消（尽量支持）
  * 失败策略：重试/跳过/中止/回滚
* 支持“事务化/回滚”（至少做到：失败时执行已声明的 rollback tasks）
* 支持受控扩展任务（Go 方法）：

  * 引入 `go:` 命名空间（Task/Guard/Screen）
  * YAML 不能直接写函数名，必须通过 Registry 注册
  * 生命周期：`Validate()` / `Execute()` / `Rollback()`
* 支持脚本任务：

  * `shell` Task：内置脚本执行（支持 embed 脚本、timeout、env、workDir）
  * `net_script` Task：网络脚本执行（url、timeout、env、workDir，可选 sha256）

### FR-4：校验与依赖检测（Preflight）

* 启动后执行环境检测：

  * OS/发行版识别（Ubuntu/Debian/Fedora/Arch…）
  * 架构识别（x86_64/arm64…）
  * 权限状态（是否 root / 是否能提权）
  * 磁盘空间、依赖包、端口占用（可选）
  * 已安装版本检测（包管理器/安装目录/服务存在）
* 检测结果写入全局状态，可用于分支流程与摘要展示

### FR-5：权限与提权

* 支持需要 root 的任务（写 /usr、安装 systemd service、安装依赖包等）
* 提权策略（可插拔）：

  * pkexec / polkit
  * sudo
  * 不提权（仅用户目录安装）
* UI 要能明确提示“哪些操作需要管理员权限”

### FR-6：国际化与品牌定制

* i18n：多语言资源包（至少 zh/en）
* 主题与品牌：logo、产品名、主色、字体、欢迎语、版权信息
* 文案可配置，且支持富文本

### FR-7：日志与诊断

* GUI 中可查看安装日志（Progress 页）
* 导出日志文件（默认保存到用户目录）
* 日志分级：INFO/WARN/ERROR
* 失败时展示“可读的错误原因 + 建议动作”（比如缺依赖、权限不足）

### FR-8：CLI 无人值守模式（建议纳入初版）

* `--mode=cli` 支持静默安装：

  * `--accept-license`
  * `--install-dir=...`
  * `--install-type=...`
* CLI 与 GUI 共用同一套 Workflow + Task 系统（只换“呈现层”）

### FR-9：卸载流程（Uninstall First-class）

* Uninstall 是 **一等 Flow**，不是附属脚本
* 安装与卸载共享同一引擎与数据模型
* 卸载支持：
  * 卸载选项（是否保留用户数据）
  * 停止服务、删除入口、删除目录
  * 卸载删除目标必须在配置中显式声明，并区分：
    * systemPaths：始终删除
    * userDataPaths：仅在不保留用户数据时删除
* 可将卸载结果写入摘要与日志

### FR-10：注册机制（Registry Extension Model）

* 内置 **Registry** 用于扩展 Task/Screen/Guard
* Registry 分三类：

  * Task Registry（任务工厂）
  * Screen Registry（页面渲染器）
  * Guard Registry（前置条件与校验）
* 允许模块或业务代码注册自定义实现（Go 扩展）

---

## 5. 非功能需求（NFR）

* 可靠性：任务中断（断网/关窗）后尽量可恢复或明确提示残留
* 安全性：不执行未签名/不可信脚本；下载支持校验（sha256/signature）
* 可维护性：步骤与任务可独立开发与测试
* 可扩展性：新增一个自定义页面/任务不改核心框架（或最小改动）
* 性能：安装过程进度可流畅更新；大文件解压不阻塞 UI（后台线程/进程）
* 兼容性：至少覆盖主流桌面环境（GNOME/KDE）显示正常

---

## 6. 架构设计（核心：可插拔流程 + 可插拔页面 + 可插拔任务）

### 6.1 分层结构

1. **Presentation（GUI/CLI）**

* Wizard Shell（步骤栏 + 主内容区 + 底部按钮）
* Screen Renderer（渲染某个 step 对应的 screen）

2. **Workflow Engine（流程引擎）**

* 管理步骤状态：currentStep、visited、completed、blocked
* 处理 next/prev/branch/guard
* 维护全局安装上下文 `InstallContext`

3. **Task Runtime（任务运行时）**

* 执行任务图（可线性、可分阶段）
* 输出事件：progress/log/state-change
* 支持取消、重试、回滚

4. **Platform Layer（Linux 适配层）**

* 提权实现、systemd/dbus、desktop entry、包管理器适配等

### 6.2 推荐设计模式（满足“灵活增删流程/自定义界面”）

* **State Machine / 状态机**：管理 step 状态与转移（当前、可跳转、完成、阻塞）
* **Strategy / 策略模式**：提权策略（pkexec/sudo/none）、发行版适配（apt/dnf/pacman）
* **Factory + Registry / 工厂+注册表**：通过 `screen.type` 创建对应 Screen；通过 `task.type` 创建对应 Task
* **Command / 命令模式**：每个 Task 是可执行命令，支持 undo（回滚）
* **Observer(EventBus) / 观察者**：任务进度/日志事件推送到 UI
* **MVVM / Presenter**：Screen 与 InstallContext 双向绑定（表单输入驱动上下文）

### 6.3 模块拆分（Library First）

* Core：流程引擎、上下文、任务运行时、Schema 校验
* Builtin：内置 Screen/Task/Guard
* Adapter：平台与权限适配
* Reference App：示例 GUI/CLI

---

## 7. 配置化方案（关键需求）

### 7.1 配置文件（建议 YAML/JSON）

* `installer.yaml`：品牌信息、流程 steps、每步 UI schema、绑定字段、任务清单、分支规则
* 可支持“内置模板 + 覆盖”：企业版只覆盖文案与品牌即可出新安装器
* `installer.yaml` 必须满足 **强规格 Schema**：

  * 提供完整 JSON Schema
  * 强类型校验（Fail Early）
  * 作为安全审计输入
* 对 Go 扩展类型的校验方式：

  * Registry 注册时必须提供对应的 Schema 或 Validator
  * 运行时装配为合并后的 Schema（内置 + 扩展）
  * 未注册的 `go:` 类型视为校验失败（Fail Early）
* 卸载目标由 uninstall flow 内任务显式声明：

  * `removePath.userData: true` 表示用户数据路径（保留用户数据时跳过）
  * 其他路径在卸载时始终删除

### 7.2 Step 配置示例（示意）

```yaml
product:
  name: "贝锐向日葵"
  logo: "assets/logo.png"
  theme:
    primaryColor: "#E53935"


flows:
  install:
    entry: "welcome"
    steps:
      - id: welcome
        title: "介绍"
        screen:
          type: richtext
          content: "assets/welcome.html"

      - id: license
        title: "许可阅读"
        screen:
          type: license
          source: "assets/eula_zh.txt"
          requireScrollToEnd: true
        guards:
          - type: mustAccept
            field: "license.accepted"

      - id: destination
        title: "安装位置"
        screen:
          type: pathPicker
          bind: "install.dir"
        guards:
          - type: diskSpace
            minMB: 500

      - id: install
        title: "安装"
        screen:
          type: progress
        tasks:
          - type: download
            url: "${meta.payloadUrl}"
            sha256: "${meta.payloadSha256}"
          - type: unpack
            to: "${install.dir}"
          - type: systemdService
            name: "sunlogin"
            action: "enable_and_start"
            requirePrivilege: true

      - id: finish
        title: "摘要"
        screen:
          type: finish
  uninstall:
    entry: "confirm"
    steps:
      - id: confirm
        title: "卸载确认"
        screen:
          type: options
          bind: "uninstall.keepUserData"
      - id: remove
        title: "卸载"
        screen:
          type: progress
        tasks:
          - type: systemdService
            name: "sunlogin"
            action: "stop_and_disable"
            requirePrivilege: true
            userData: false
          - type: removePath
            path: "${install.dir}"
            userData: false
          - type: removePath
            path: /userDataPaths
            userData: true
          - type: removeDesktopEntry
            name: "sunlogin"
            userData: flase
      - id: finish
        title: "完成"
        screen:
          type: finish
```
### 7.3 Registry 与命名空间约定

* 内置类型：直接使用 `type: download` / `type: license` 等
* Go 扩展类型：使用 `go:` 命名空间，例如：

  * `task.type: "go:customTask"`
  * `screen.type: "go:customScreen"`
  * `guard.type: "go:customGuard"`
* YAML 中 **不可直接写函数名**，必须通过 Registry 注册后才能引用


---

## 8. 交互与 UI 规范（对齐截图体验）

* 左侧步骤栏：

  * 当前步骤高亮
  * 已完成步骤显示 ✓
  * 被禁用步骤灰显且不可点击（除非 allowJump）
* 主内容区：

  * 支持滚动
* 底部按钮：

  * `返回`：回到上一步（如当前是第一步则隐藏/置灰）
  * `继续`：根据 step 类型可能变为 `安装` / `完成`
  * guard 不满足时继续按钮置灰，并在页面内显示原因（比如“请先勾选同意协议”）
* 失败处理：

  * 安装任务失败后进入失败态 Finish 页：显示错误、查看日志、重试、回滚（可选）

---

## 9. 数据模型（InstallContext）

建议定义统一上下文，贯穿 GUI/CLI：

* `env`: distro、arch、desktop、isRoot、hasPolkit、diskFreeMB…
* `userInput`: license.accepted、install.dir、install.type、options…
* `plan`: 解析后的任务计划（摘要用）
* `runtime`: currentStep、progress、logs、errors、startTime、action…

---

## 10. 测试要求（初版就要有）

* 单元测试：

  * workflow 分支与 guard
  * task 参数渲染（变量替换）
  * rollback 链
* 集成测试：

  * 模拟无权限/无网络/磁盘不足
  * 发行版适配（apt/dnf）
* UI 测试（可选）：

  * step 切换、按钮状态、长文本滚动

---

## 11. 交付物清单（你这版文档对应的输出）

1. 安装器框架代码（GUI + CLI 共用核心）
2. `installer.yaml` 配置规范与示例
3. JSON Schema（强规格）
4. 内置 Screen 与 Task 插件集合（最小可用）



---

## 12. 初版 MVP 建议范围（最小可用但足够灵活）

* Flow Engine：线性步骤 + guards + 动态禁用 + action 多 Flow
* Screens：welcome、license、pathPicker、progress、finish
* Tasks：download、unpack、copy、writeConfig、desktopEntry、systemdService（可选）、rollback、shell（内置脚本）
* 提权：sudo 或 pkexec 二选一先跑通
* i18n：至少 zh/en 的资源加载机制
