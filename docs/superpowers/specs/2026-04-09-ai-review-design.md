# AI Review Plugin Design

## 背景

当前仓库是一个 Drone plugin 模板，目标是在此基础上实现 AI 评审能力，供 CI 流水线在不同场景下自动执行代码审查。

首版聚焦两类用途：

1. `PR` 增量评审：针对 Pull Request 的变更生成结构化 `JSON`
2. `Full` 全量评审：针对默认分支已检出的仓库工作区生成 `Markdown`

AI 能力通过本地已安装的 `codex` 或 `claude` CLI 调用，不接入 HTTP API。

## 目标

- 提供统一的 Drone plugin 镜像与入口
- 通过显式参数切换 `pr` / `full` 评审模式
- `pr` 模式输出严格符合约定 schema 的 JSON 文件，并标注问题所在文件与行号
- `full` 模式输出可读的 Markdown 审查报告
- 保持实现可扩展，便于后续增加 provider、prompt 策略、过滤规则

## 非目标

- 首版不直接对接 GitHub/GitLab/Gitea API 回写评论
- 首版不做流式输出解析
- 首版不支持 HTTP API provider
- 首版不做复杂的多轮追问式 agent 工作流，仅执行一次 CLI 调用并收集结果

## 使用方式

插件运行在 Drone pipeline step 中，基于已 clone 下来的工作区执行。

核心参数：

- `PLUGIN_MODE=pr|full`
- `PLUGIN_PROVIDER=codex|claude`
- `PLUGIN_OUTPUT=<path>`
- `PLUGIN_WORKDIR=<path>`，可选，默认当前目录
- `PLUGIN_PROMPT_FILE=<path>`，可选，自定义 prompt
- `PLUGIN_CLI_BIN=<path>`，可选，自定义 CLI 可执行文件
- `PLUGIN_MODEL=<name>`，可选，透传给 provider
- `PLUGIN_EXTRA_ARGS=<args>`，可选，透传额外参数
- `PLUGIN_INCLUDE=<glob,...>`，可选，限定评审范围
- `PLUGIN_EXCLUDE=<glob,...>`，可选，排除评审范围

## 运行模式

### PR 模式

触发条件由用户显式设置：

```bash
PLUGIN_MODE=pr
```

插件读取 Drone 官方环境变量：

- `DRONE_PULL_REQUEST`
- `DRONE_SOURCE_BRANCH`
- `DRONE_TARGET_BRANCH`

执行流程：

1. 校验当前目录是 Git 仓库
2. 拉取目标分支引用，确保可生成 diff
3. 生成 `target...HEAD` 的 diff
4. 根据 diff 构造 prompt
5. 调用 `codex` 或 `claude` CLI
6. 解析 CLI 输出为 JSON
7. 校验 JSON schema
8. 写入输出文件

JSON schema 固定为：

```json
{
  "verdict": "pass" | "needs_attention",
  "summary": "short review summary",
  "findings": [
    {
      "path": "relative/file/path",
      "line_start": 1,
      "line_end": 1,
      "severity": "low" | "medium" | "high",
      "title": "short title",
      "body": "actionable explanation"
    }
  ]
}
```

约束：

- `path` 必须是仓库相对路径
- `line_start` / `line_end` 必须是正整数
- `line_end` 不得小于 `line_start`
- `findings` 可为空数组
- 任何非法 JSON 或字段缺失都视为失败

### Full 模式

触发条件由用户显式设置：

```bash
PLUGIN_MODE=full
```

执行流程：

1. 校验工作目录存在且是 Git 仓库
2. 基于当前 clone 下来的仓库工作区执行评审
3. 按 include/exclude 过滤可评审文件
4. 构造仓库级 prompt
5. 调用 `codex` 或 `claude` CLI
6. 将输出作为 Markdown 写入目标文件

说明：

- `full` 模式不额外“收集全量仓库内容”后上传
- 默认由 AI CLI 直接在当前工作区执行审查
- 插件负责准备执行上下文、参数与最终结果落盘

## Prompt 设计

### PR Prompt

PR prompt 应包含：

- 评审目标：关注 bug、风险、可维护性、正确性
- 输出格式：只能输出 JSON，不允许额外说明文字
- 严格字段定义与枚举值
- 行号必须与 diff 中变更行对应
- 仅报告高价值问题，避免噪声

### Full Prompt

Full prompt 应包含：

- 评审范围：当前仓库工作区
- 输出格式：Markdown
- 报告结构建议：概览、主要风险、改进建议、优先级
- 鼓励聚焦高风险问题，避免罗列琐碎风格问题

### 自定义 Prompt

如果设置 `PLUGIN_PROMPT_FILE`，则：

- 使用用户提供内容覆盖默认 prompt 模板
- 仍由插件补充必要的系统约束，如输出格式与路径信息

## Provider 抽象

对外支持两种 provider：

- `codex`
- `claude`

内部定义统一接口，例如：

- 输入：模式、工作目录、prompt、模型名、额外参数
- 输出：CLI 原始 stdout/stderr 与退出状态

provider 负责：

- 定位 CLI 二进制
- 组装命令行参数
- 在指定工作目录运行命令
- 返回执行结果

插件主流程负责：

- prompt 生成
- 模式控制
- 结果解析
- 输出文件写入

## 模块划分

建议新增或调整以下文件：

- `plugin/config.go`：解析 `PLUGIN_*` 配置并做校验
- `plugin/mode.go`：根据 `mode` 调度 `pr` / `full`
- `plugin/git.go`：仓库根目录、diff 生成、分支相关 Git 操作
- `plugin/provider.go`：provider 接口定义
- `plugin/provider_codex.go`：Codex CLI 适配
- `plugin/provider_claude.go`：Claude CLI 适配
- `plugin/prompt.go`：默认 prompt 与自定义 prompt 组合
- `plugin/render_json.go`：PR JSON 解析、校验与输出
- `plugin/render_md.go`：Full Markdown 输出
- `plugin/plugin.go`：保留为总入口，调用配置解析与模式执行

## 错误处理

以下场景直接返回错误并终止：

- `PLUGIN_MODE` 缺失或值非法
- `PLUGIN_PROVIDER` 缺失或值非法
- `PLUGIN_OUTPUT` 缺失
- `pr` 模式缺少必须的 Drone PR 环境变量
- 当前目录不是 Git 仓库
- Git diff 生成失败
- CLI 不存在或执行失败
- `pr` 模式 JSON 解析失败或 schema 校验失败
- `full` 模式输出为空

错误信息要求明确指出失败阶段，便于在 CI 日志中定位。

## 测试策略

首版采用 Go 单元测试 + 小型集成测试。

建议覆盖：

- 配置解析与默认值
- `mode` 路由
- PR 环境变量校验
- provider 命令拼装
- Git diff 生成
- PR JSON schema 校验
- Full Markdown 输出

测试重点：

- 验证 JSON 校验严格执行
- 验证 CLI stdout/stderr 处理正确
- 验证 Git 相关错误能暴露出明确上下文

## 风险与权衡

1. `codex` 与 `claude` CLI 的命令行参数可能不同  
   处理方式：通过 provider 适配层隔离差异

2. PR 行号可能与纯文本 diff 解析存在偏差  
   处理方式：首版要求模型基于统一 diff 上下文返回行号，并在后续如有需要再增加 diff 解析辅助

3. 全量仓库评审可能受上下文窗口限制  
   处理方式：首版优先依赖 CLI 在工作区内执行的能力；若后续发现不足，再增加分块与摘要策略

4. 不同仓库对排除目录要求不同  
   处理方式：提供 include/exclude 参数，不做过度内建规则

## 交付结果

实现完成后，应具备以下最小能力：

- 在 Drone pipeline 中通过 `PLUGIN_MODE=pr` 运行，输出符合 schema 的 `review.json`
- 在 Drone pipeline 中通过 `PLUGIN_MODE=full` 运行，输出 `review.md`
- 可切换 `codex` / `claude` provider
- 关键路径有自动化测试覆盖
