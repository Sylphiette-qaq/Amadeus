# Skill 技术方案

## 目标

本轮只实现 Claude Code 风格 skill 的最小闭环，范围限定为：

1. 新增 `load_skill` 工具，用于按 skill 名称完整加载对应的 `SKILL.md` 内容。
2. 新增一份 `agent.md` 作为 skill 注册表，存放所有 skill 的 `name` 和 `desc`。
3. 程序启动时将 `agent.md` 内容拼入 system prompt，使模型知道当前有哪些 skill 可用。
4. 通过 `.env` 配置 `agent.md` 和 skill 根目录的相对路径、绝对路径。

本阶段不做：

- skill 多文件引用展开
- skill 自动发现并回写 `agent.md`
- skill 热更新
- skill 参数化模板
- skill 安装、启用/禁用、权限控制

## 当前代码现状

当前主链路如下：

1. [cmd/amadeus/main.go](/Users/likanggui/GolandProjects/agent-learn/Amadeus/cmd/amadeus/main.go:14) 启动时加载 `.env`。
2. [internal/tool/registry.go](/Users/likanggui/GolandProjects/agent-learn/Amadeus/internal/tool/registry.go:12) 负责注册所有可调用工具。
3. [internal/model/chat_model.go](/Users/likanggui/GolandProjects/agent-learn/Amadeus/internal/model/chat_model.go:16) 当前 system prompt 是硬编码常量。
4. [internal/orchestrator/loop.go](/Users/likanggui/GolandProjects/agent-learn/Amadeus/internal/orchestrator/loop.go:14) 每轮请求都会把 system prompt 放在消息最前面。

这意味着 skill 能力需要接入两个点：

- 启动阶段：构造最终 system prompt，把 `agent.md` 注入进去。
- 工具阶段：提供 `load_skill`，让模型按需读取完整 `SKILL.md`。

## 总体设计

### 设计原则

1. `agent.md` 只负责“目录级描述”，不承载完整 skill 指令。
2. `SKILL.md` 只在模型明确需要时，通过 `load_skill` 按需读取。
3. 路径配置允许“相对路径优先，绝对路径兜底”，避免部署路径变化导致失效。
4. 所有路径在启动时统一解析并校验，运行期工具只消费解析后的结果，不自行猜路径。

### 模块划分

建议新增以下模块：

- `internal/skill/config.go`
  - 负责从环境变量读取 `agent.md` 和 skill 目录配置。
  - 负责相对/绝对路径解析、存在性校验。
- `internal/skill/registry.go`
  - 负责读取 `agent.md` 原文。
  - 对外提供“注册表文本”。
- `internal/skill/loader.go`
  - 负责根据 skill 名称定位并加载对应的 `SKILL.md`。
- `internal/tool/basetools/load_skill.go`
  - 定义 `load_skill` 工具及其参数、返回值。

## 文件与配置设计

### 1. `agent.md`

职责：

- 作为唯一的 skill 注册入口。
- 启动时直接注入上下文。
- 内容仅包含 skill 列表，不包含完整 skill 指令正文。

推荐格式：

```md
# Skills

- name: openai-docs
  desc: 查询和使用 OpenAI 官方文档，适合 API、模型选型、参数说明等问题。

- name: plugin-creator
  desc: 创建和初始化插件目录结构。
```

也可以允许更自由的 Markdown，但第一期建议约束成稳定格式，原因是：

1. 便于未来做解析校验。
2. 便于在启动时报错提示缺失字段。
3. 便于后续扩展 `version`、`tags`、`examples` 等元数据。

### 2. skill 目录布局

推荐目录结构：

```text
skills/
  openai-docs/
    SKILL.md
  plugin-creator/
    SKILL.md
```

约束：

1. 子目录名必须等于 skill name。
2. 每个 skill 目录下必须存在 `SKILL.md`。
3. `load_skill(name)` 直接映射到 `${skill_root}/${name}/SKILL.md`。

这样可以避免再做一层 name -> path 的注册映射，先把系统做薄。

### 3. `.env` 配置

按你的要求保留相对路径和绝对路径两组配置：

```env
SKILL_AGENT_MD_REL=./skills/agent.md
SKILL_AGENT_MD_ABS=

SKILL_ROOT_REL=./skills
SKILL_ROOT_ABS=
```

解析规则建议固定为：

1. 优先使用绝对路径变量：如果 `*_ABS` 非空，则直接使用。
2. 否则使用相对路径变量：相对于进程工作目录 `cwd` 解析为绝对路径。
3. 若两者都为空，则启动失败。
4. 启动时统一转成绝对路径，后续模块只使用绝对路径。

这样做的原因：

1. 满足本地开发和部署场景。
2. 消除运行期工具对当前工作目录的依赖。
3. 把路径错误前置到启动阶段，而不是对话过程中才暴露。

## `load_skill` 工具设计

### 工具职责

`load_skill` 只做一件事：按 skill 名称返回完整 `SKILL.md` 文本。

不负责：

- 解析 `SKILL.md` 中引用的其他文件
- 扫描所有 skill
- 修改 skill 内容
- 猜测最相近 skill

### 工具接口

工具名：

```text
load_skill
```

参数：

```json
{
  "name": "openai-docs"
}
```

参数约束：

1. `name` 必填。
2. 仅允许 `[a-zA-Z0-9_-]`，防止路径穿越。
3. 禁止出现 `/`、`\`、`..`。

返回结果建议统一成结构化文本或 JSON 字符串，推荐 JSON：

```json
{
  "skill_name": "openai-docs",
  "path": "/abs/path/skills/openai-docs/SKILL.md",
  "content": "# ...完整 SKILL.md 内容..."
}
```

失败场景：

```json
{
  "skill_name": "openai-docs",
  "path": "/abs/path/skills/openai-docs/SKILL.md",
  "error": "skill not found"
}
```

### 工具描述文案

工具描述建议明确模型使用时机：

1. 当用户点名某个 skill。
2. 当系统上下文中的 `agent.md` 显示某个 skill 可能适用，但需要完整指令时。
3. 不要一次性加载多个 skill，按需单个加载。

这样能降低模型在首轮就把所有 skill 全部拉进上下文的概率。

## 启动时上下文注入设计

### 注入内容

当前 system prompt 是固定字符串，建议改成两段拼接：

1. 基础系统提示词
2. skill 注册表内容

示意：

```text
你是一个人工智能助手，名称是 Amadeus。你需要用语气平淡，内容简洁且专业的语气回答问题。

以下是当前可用的 skills 列表。这里只提供名称和简介；当你确认某个 skill 适用时，再调用 load_skill 加载该 skill 的完整说明。

<agent.md 原文>
```

### 为什么不直接把所有 `SKILL.md` 注入 system prompt

原因很直接：

1. token 成本高。
2. 多 skill 时会迅速膨胀上下文。
3. 大部分 skill 在单次对话里根本不会用到。
4. Claude Code 也是“目录摘要常驻 + 具体 skill 按需加载”的模式。

### 注入时机

建议在启动时完成一次：

1. 读取 `.env`
2. 解析路径
3. 读取 `agent.md`
4. 生成最终 system prompt
5. 初始化 orchestrator

这样运行期间不需要重复读取 `agent.md`。

## 代码接入方案

### 1. 新增 skill 配置模块

新增 `internal/skill/config.go`：

- `type Config struct`
- `func LoadConfig() (Config, error)`
- 字段至少包含：
  - `AgentMDPath string`
  - `SkillRootPath string`

职责：

1. 读取四个环境变量。
2. 解析最终绝对路径。
3. 校验 `agent.md` 是否存在。
4. 校验 skill 根目录是否存在且为目录。

### 2. 新增 skill 注册表读取模块

新增 `internal/skill/registry.go`：

- `func LoadAgentMarkdown(cfg Config) (string, error)`

职责：

1. 直接读取 `agent.md` 原文。
2. 可选增加一个轻量校验：
   - 文件非空
   - 至少出现一个 `name:` 和一个 `desc:`

第一期不建议做复杂 Markdown 解析，先把“文本注入 + 人工维护”跑通。

### 3. 新增 skill 加载模块

新增 `internal/skill/loader.go`：

- `func LoadSkillContent(cfg Config, name string) (SkillDocument, error)`

`SkillDocument` 建议包含：

- `Name string`
- `Path string`
- `Content string`

实现逻辑：

1. 校验 `name`。
2. 拼出 `${SkillRootPath}/${name}/SKILL.md`。
3. 校验目标路径必须仍位于 `SkillRootPath` 下。
4. 读取文件并返回。

第 3 条是必要的，避免未来有人绕过正则校验做路径逃逸。

### 4. 新增 `load_skill` 工具

新增 `internal/tool/basetools/load_skill.go`：

实现方式参考现有 [internal/tool/basetools/bash.go](/Users/likanggui/GolandProjects/agent-learn/Amadeus/internal/tool/basetools/bash.go:21)。

建议签名：

- `func GetLoadSkillTool(cfg skill.Config) einotool.InvokableTool`

这样 `basetools.Load()` 需要改成接收配置：

- `func Load(cfg skill.Config) []einotool.InvokableTool`

随后 [internal/tool/registry.go](/Users/likanggui/GolandProjects/agent-learn/Amadeus/internal/tool/registry.go:12) 也要把 skill 配置传进去。

### 5. system prompt 组装

当前 [internal/model/chat_model.go](/Users/likanggui/GolandProjects/agent-learn/Amadeus/internal/model/chat_model.go:16) 中的 `SystemMessage` 是常量，建议改为：

- 保留 `BaseSystemMessage` 常量
- 新增 `func BuildSystemMessage(agentMarkdown string) string`

这样 system prompt 不再是硬编码单值，而是启动阶段动态生成。

### 6. 启动链路调整

建议改造 [cmd/amadeus/main.go](/Users/likanggui/GolandProjects/agent-learn/Amadeus/cmd/amadeus/main.go:14)：

1. 加载 `.env`
2. `skill.LoadConfig()`
3. `skill.LoadAgentMarkdown(cfg)`
4. `model.BuildSystemMessage(agentMarkdown)`
5. `internaltool.LoadInvokableTools(ctx, "./tools/toolsConfig.json", cfg)`
6. 初始化 orchestrator

## 建议的数据流

### 启动阶段

```text
.env
  -> skill.LoadConfig
  -> 解析 agent.md / skill_root 绝对路径
  -> 读取 agent.md 原文
  -> model.BuildSystemMessage(agent.md)
  -> 注册 load_skill 工具
  -> 启动对话
```

### 运行阶段

```text
用户问题
  -> 模型先看到 system prompt 中的 agent.md 摘要
  -> 判断某个 skill 适用
  -> 调用 load_skill(name)
  -> 获取完整 SKILL.md
  -> 基于 skill 指令继续回答或继续执行其他工具
```

## 错误处理策略

### 启动失败类

这些问题建议直接阻止启动：

1. `agent.md` 路径未配置
2. `agent.md` 文件不存在
3. skill 根目录不存在
4. 路径无法转绝对路径

原因是这类问题会导致 skill 机制整体不可用，没必要拖到运行期。

### 工具失败类

`load_skill` 执行失败时返回清晰错误：

1. skill 不存在
2. `SKILL.md` 不存在
3. 名称非法
4. 文件不可读

这里不建议 panic，也不建议吞错，因为模型需要知道失败原因，才能降级处理。

## 安全与边界

### 1. 路径穿越防护

必须做两层：

1. `name` 白名单校验
2. 最终路径 `filepath.Clean` 后，校验仍在 `SkillRootPath` 前缀内

### 2. 输出大小

第一期可以不截断 `SKILL.md`，因为需求就是“完整加载”。

但需要明确两个风险：

1. 超大 skill 会推高 token 消耗。
2. 模型可能多次重复调用同一 skill。

后续可增加：

- 单文件大小限制
- 同轮缓存
- 已加载 skill 提示

### 3. 编码

统一按 UTF-8 文本读取；非法编码直接报错。

## 测试方案

### 单元测试

建议覆盖：

1. 配置解析
   - 仅相对路径
   - 仅绝对路径
   - 绝对路径优先
   - 配置缺失
2. `load_skill`
   - 正常加载
   - skill 不存在
   - 名称非法
   - 路径穿越拦截
3. system prompt 构造
   - 包含基础提示词
   - 包含 `agent.md`

### 集成测试

最小验证流程：

1. 启动程序
2. 提问“有哪些 skill 可用”
3. 确认模型能依据 `agent.md` 回答 skill 列表
4. 提问“请使用 xxx skill”
5. 确认模型会调用 `load_skill`

## 分阶段实施建议

### Phase 1

只做你当前要求的最小版本：

1. `.env` 配置生效
2. 启动注入 `agent.md`
3. `load_skill` 能完整读取单个 `SKILL.md`

### Phase 2

在 Phase 1 稳定后再考虑：

1. `agent.md` 格式校验增强
2. skill 内容缓存
3. skill 目录自动扫描与注册表生成
4. `SKILL.md` 关联资源加载

## 需要你审查确认的决策点

1. `agent.md` 是否接受“固定结构的 Markdown 列表”作为第一期格式约束。
2. `load_skill` 是否只接受 `name`，并强制映射到 `${skill_root}/${name}/SKILL.md`。
3. `.env` 的优先级是否按“绝对路径优先，相对路径兜底”执行。
4. 启动时如果 `agent.md` 或 skill 根目录缺失，是否直接 fail fast。
5. `load_skill` 返回值是否采用 JSON 字段：`skill_name`、`path`、`content`。

## 结论

这个方案的核心是把 skill 分成两层：

1. `agent.md` 负责常驻上下文中的“发现”。
2. `load_skill` 负责运行期按需加载“完整指令”。

这样实现成本低，和你现在的代码结构兼容，也基本贴近 Claude Code 的工作方式。第一期只改启动链路、工具注册和一个文件读取工具，改动面可控，适合先上线验证。
