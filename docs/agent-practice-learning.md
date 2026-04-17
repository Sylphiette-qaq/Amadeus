# 通过项目实践理解 AI Agent 的本质

## 一、总结

这个项目的目的，不是单纯实现一个运行在 CLI 中的 AI 助手，而是通过一个可运行、可观察、可调试的系统，系统理解 AI agent 的关键机制，并将这些理解转化为实际工作方法。

围绕这个目标，项目重点验证了几件事：

- agent 的本质不是单轮回答，而是围绕目标持续推进任务的闭环系统
- `LLM` 的行为不仅取决于模型能力，也取决于上下文的组织方式
- `skill` 的核心作用是注入规则，而不是堆积知识
- `CLI/tool` 的核心作用是提供稳定的执行接口，而不是零散能力展示

通过这次实践，形成了三类直接收获：

- 对 agent 的系统结构有了更清晰的理解
- 对 `skill`、`CLI`、上下文装配之间的关系有了更具体的认识
- 对业务落地路径有了更务实的方法，即“接口 CLI 化，经验 skill 化”

---

## 二、展开

### 1. 做这个项目的目的

这个项目主要服务于三个目标：

- 验证 agent 的最小闭环  
  即“用户输入 -> 模型判断 -> 工具调用 -> 结果回填 -> 继续推进”
- 理解 agent 的关键组成  
  包括 system prompt、history、user input、tool call、tool result、skill 等部分如何协同
- 探索实际落地方式  
  即如何通过 `skill + CLI` 的方式，把业务操作转成 agent 可执行流程

因此，这个项目更适合作为 agent 机制的实践样本，而不只是一个功能演示。

### 2. 对 Agent 本质的理解

通过项目实践，可以将 agent 理解为一个任务推进系统，而不是一次性回答系统。

它的基本闭环是：

```text
理解目标
  -> 判断下一步
  -> 调用外部能力
  -> 获取反馈
  -> 基于反馈继续推进
```

这意味着：

- 普通问答模型主要解决“输出回答”
- agent 主要解决“持续推进任务直到完成”

因此，agent 的关键不只是会不会调用工具，而是能不能把工具调用纳入任务闭环。

### 3. 对 LLM 上下文结构的理解

这个项目最重要的收获之一，是把 `LLM` 上下文从“一个 prompt”拆成了更清晰的结构。

可以概括为五部分：

- `System Prompt`  
  定义角色、语气、全局规则，以及可用 skill 摘要
- `Conversation History`  
  保存已经发生的任务过程
- `Current User Input`  
  当前这一轮的任务目标
- `Tool Calls`  
  模型决定调用哪些工具
- `Tool Results`  
  外部执行后的反馈结果

对应到当前项目的实现，消息装配遵循以下顺序：

```text
system
  -> history
  -> current user input
  -> assistant tool calls
  -> tool results
```

这里有三点尤其重要：

- `system prompt` 是全局约束层，不只是开场白
- 用户输入只是当前轮目标，不等于全部上下文
- `tool result` 必须回到消息流中，模型才能形成真正闭环

这也说明，agent 的表现很大程度上取决于上下文结构，而不只是模型本身。

### 4. 对 Skill 作用机制的理解

项目实践说明，`skill` 的核心作用不是补充背景知识，而是定义任务执行规则。

它主要负责：

- 说明什么场景下使用某项能力
- 规定调用工具前要确认什么
- 规定步骤顺序
- 约束错误路径和禁止行为

因此，`skill` 更接近：

- SOP
- 操作手册
- 决策规则
- 工具使用规范

这也是为什么 `skill` 对 agent 稳定性非常重要。模型知道一件事，不等于模型会按正确流程处理这件事。

### 5. 对 Skill 渐进式披露的理解

在外部资料中，Anthropic 将 skill 的核心原则概括为 `progressive disclosure`，即“渐进式披露”。

它的核心逻辑是：

- 先把 skill 的名称和简介放入 system prompt
- 只有在任务相关时，再加载完整 `SKILL.md`
- 如果还有更细资料，再继续按需展开

这个原则解决的是上下文预算问题：

- 避免一次性加载过多无关信息
- 降低 token 成本
- 提高当前任务的上下文密度

当前项目已经实现了这个原理的最小闭环：

- `agent.md` 提供 skill 摘要
- `load_skill` 按需读取完整 `SKILL.md`
- `load_skill` 成功后，将完整 skill 持久化到当前 session 的上下文资产中
- 后续轮重建上下文时，继续把已加载 skill 作为 `system` 级规则注入

因此，当前系统已经具备两级披露能力：

- 第一级：skill 摘要常驻
- 第二级：完整 skill 按需加载并持久激活

可以把当前实现理解成下面这个顺序：

```text
用户请求
  -> 模型先看到 system prompt 中的 `agent.md` 摘要
  -> 判断某个 skill 适用
  -> 调用 `load_skill(name)`
  -> 读取对应 `SKILL.md`
  -> 将完整 skill 写入 session 的 `loaded_skills.jsonl`
  -> 当前轮立刻把该 skill 作为一条 `system` 消息加入上下文
  -> 后续轮恢复时继续注入这个 skill
```

一个更直观的时序图如下：

```text
User           Model            load_skill tool        Session Store        Next Turn Builder
 |               |                    |                     |                     |
 | ask task      |                    |                     |                     |
 |-------------->| sees agent.md      |                     |                     |
 |               | decide skill needed|                     |                     |
 |               |------------------->| read SKILL.md       |                     |
 |               |                    |-------------------->| append loaded skill |
 |               |<-------------------| return full skill   |                     |
 |               | inject as system msg in current turn     |                     |
 |               | continue reasoning |                     |                     |
 | next user msg |                    |                     |                     |
 |-------------->|                    |                     |                     |
 |               |                                            load loaded skills  |
 |               |<--------------------------------------------------------------|
 |               | rebuild context: base system + loaded skills + history + user |
```

这意味着现在的 `load_skill` 不只是“读取一个文件”，而是“读取并激活为当前 session 的持久规则上下文”。

### 6. 对 Skill 和 CLI 分工的理解

这个项目也帮助明确了 `skill` 和 `CLI/tool` 的职责边界。

可以概括为：

- `skill` 解决“怎么做”
- `CLI/tool` 解决“拿什么做”

具体来说：

- `skill` 是行为层  
  负责规则、顺序、约束
- `CLI/tool` 是执行层  
  负责把能力暴露成模型可调用接口

两者缺一不可：

- 只有 skill，没有 tool，知道流程但无法执行
- 只有 tool，没有 skill，有执行能力但容易误用

因此，更合理的理解方式是：

- `skill` 是规则接口
- `tool` 是动作接口

### 7. 对实际工作的帮助

这些理解对实际工作有比较直接的帮助，主要体现在三个方面。

#### 7.1 更容易识别适合 agent 化的工作

现在判断一个任务是否适合 agent 化，会优先看：

- 是否高频重复
- 是否有明确步骤
- 是否有前置条件
- 是否能封装为 CLI
- 是否能写成 skill 规则

这比从“模型能不能做”出发更务实。

#### 7.2 更容易把经验沉淀为可复用能力

很多工作效率低，不是因为能力不够，而是经验只存在于人脑中。

通过 `skill + CLI` 这种方式，可以把经验拆成两部分：

- 业务规则沉淀为 skill
- 业务动作沉淀为 CLI/tool

这样更容易复用，也更容易标准化。

#### 7.3 对 prompt 的理解从“写句子”转向“做上下文工程”

项目实践之后，关注点不再只是提示词怎么写，而是：

- 哪些信息应该常驻
- 哪些信息应该按需加载
- 哪些执行结果必须回灌
- 不同消息在上下文里的优先级如何安排

这会让 agent 的设计更稳定，也更接近工程问题而不是文案问题。

### 8. 实际落地案例：元数据平台 CLI 与 Skill

元数据平台是一个适合作为独立案例的业务场景。

在这个案例里：

- CLI 负责提供动作能力  
  例如列出 API、查看帮助、发起调用、返回 JSON 结果
- Skill 负责提供执行规则  
  例如先确认 `base_url` 和 `token`，先查 API，再看帮助，最后调用

这个案例体现得很清楚：

- 工具提供能力
- skill 提供方法
- orchestrator 负责闭环推进

这也是一条比较现实的业务落地路径：

- 先把业务接口 CLI 化
- 再把业务经验 skill 化
- 最后接入 agent 系统

---

## 三、总结

这个项目带来的价值，不只是实现了一个可运行的 agent 原型，更重要的是形成了一套更清晰的理解框架：

- agent 是任务推进闭环，不是单轮回答
- `LLM` 的行为依赖上下文结构，而不是只依赖模型能力
- `skill` 的价值在于规则注入
- `CLI/tool` 的价值在于提供执行抓手
- `progressive disclosure` 是 skill 设计中非常关键的上下文管理原则

对实际工作的帮助主要体现在：

- 更容易判断哪些业务适合 agent 化
- 更容易把经验沉淀为自动化能力
- 更容易从上下文工程角度设计稳定的 agent 系统

如果继续沿着这条路径推进，比较现实的方向是：

- 将更多高频业务接口 CLI 化
- 将更多业务经验 skill 化
- 逐步把重复性工作转化为可复用的 agent 流程

---

## 参考资料

- Anthropic, "Equipping agents for the real world with Agent Skills", 2025-10-16  
  https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills
- Anthropic Docs, "Context windows"  
  https://platform.claude.com/docs/en/build-with-claude/context-windows
- OpenAI Docs, "Prompt engineering"  
  https://platform.openai.com/docs/guides/prompt-engineering
- OpenAI Docs, "Text generation / message roles"  
  https://platform.openai.com/docs/guides/text-generation/chat-completions-api
- OpenAI Docs, "Function calling"  
  https://platform.openai.com/docs/guides/function-calling
- OpenAI Docs, "Function calling lifecycle"  
  https://platform.openai.com/docs/guides/function-calling/lifecycle
