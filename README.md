# drone-ai-review

基于 Drone CI 的 AI 代码评审插件，支持：

- `PR` 增量评审，输出结构化 `JSON`
- 默认分支全量仓库评审，输出 `Markdown`
- provider 可切换 `codex` 或 `claude` CLI

## 环境要求

- 流水线工作目录已由 Drone clone 为 Git 仓库
- 容器内已完成认证：
  - `codex`
  - 或 `claude`

当前仓库内置安装：

- `@openai/codex`
- Claude Code 原生 CLI

## 插件参数

### 必填参数

- `PLUGIN_MODE=pr|full`
- `PLUGIN_PROVIDER=codex|claude`
- `PLUGIN_OUTPUT=<output file>`

### 可选参数

- `PLUGIN_WORKDIR=<path>`：默认当前目录
- `PLUGIN_PROMPT_FILE=<path>`：自定义 prompt 文件
- `PLUGIN_CLI_BIN=<path>`：自定义 CLI 可执行文件
- `PLUGIN_MODEL=<name>`：指定模型
- `PLUGIN_EXTRA_ARGS=<arg1,arg2>`：附加 CLI 参数，使用逗号分隔
- `PLUGIN_INCLUDE=<glob1,glob2>`：提示 AI 优先关注的路径
- `PLUGIN_EXCLUDE=<glob1,glob2>`：提示 AI 跳过的路径
- `PLUGIN_DEBUG=true`：开启调试日志

## Drone 环境变量

`pr` 模式依赖 Drone 注入的以下环境变量：

- `DRONE_PULL_REQUEST`
- `DRONE_SOURCE_BRANCH`
- `DRONE_TARGET_BRANCH`

插件会在当前仓库内自动执行 `git diff` 生成评审上下文。

## PR 模式

`PLUGIN_MODE=pr` 时：

- 校验 Drone PR 环境变量
- 获取目标分支 diff
- 调用 `codex` 或 `claude` CLI
- 校验输出 JSON schema
- 将结果写入 `PLUGIN_OUTPUT`

输出结构：

```json
{
  "verdict": "pass",
  "summary": "short review summary",
  "findings": [
    {
      "path": "relative/file/path",
      "line_start": 12,
      "line_end": 12,
      "severity": "medium",
      "title": "short title",
      "body": "actionable explanation"
    }
  ]
}
```

### PR 示例

```yaml
kind: pipeline
type: docker
name: default

steps:
  - name: ai-pr-review
    image: your-registry/drone-ai-review
    environment:
      PLUGIN_MODE: pr
      PLUGIN_PROVIDER: codex
      PLUGIN_OUTPUT: review.json
```

## Full 模式

`PLUGIN_MODE=full` 时：

- 直接在当前 clone 工作区执行仓库级审查
- 调用 `codex` 或 `claude` CLI
- 将 Markdown 报告写入 `PLUGIN_OUTPUT`

### Full 示例

```yaml
kind: pipeline
type: docker
name: default

steps:
  - name: ai-full-review
    image: your-registry/drone-ai-review
    environment:
      PLUGIN_MODE: full
      PLUGIN_PROVIDER: claude
      PLUGIN_OUTPUT: review.md
```

## Provider 说明

### Codex

`codex` provider 使用非交互 `exec` 模式，并在 `pr` 模式下附带 JSON schema 约束。

### Claude

`claude` provider 使用 `-p` 非交互模式，并在 `pr` 模式下通过 `--json-schema` 约束输出结构。

## 本地调试

```bash
export PLUGIN_MODE=full
export PLUGIN_PROVIDER=codex
export PLUGIN_OUTPUT=review.md

go run .
```

## 镜像内 CLI 安装来源

- Codex CLI：官方文档当前推荐 `npm i -g @openai/codex`
- Claude Code：官方文档当前推荐原生安装方式；本镜像使用官方 `install.sh`

示例流水线文件见：`.drone.yml.example`
