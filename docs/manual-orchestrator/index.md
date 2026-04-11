# Amadeus Agent 模式改造（手动编排）— 文档索引

本文档集说明将 Amadeus 从 CloudWeGo Eino 内置 `ChatModelAgent` 改为「仅使用框架模型能力 + 业务侧主编排」的技术方案，供评审与实施参考。

**改造目标**：不是替换底层模型框架，也不是移除工具体系，而是将推理轮次、工具调度、上下文拼装与终止判断从框架收回到项目代码。

配套执行计划见：[`docs/manual-orchestrator-execution-plan.md`](../manual-orchestrator-execution-plan.md)

## 阅读顺序

| 序号 | 文档 | 内容概要 |
|------|------|----------|
| 1 | [current-state-and-problems.md](current-state-and-problems.md) | 当前实现、链路、问题 |
| 2 | [target-architecture-and-orchestrator.md](target-architecture-and-orchestrator.md) | 改造目标、分层架构、主编排循环 |
| 3 | [protocol-and-tools.md](protocol-and-tools.md) | OpenAI 工具调用协议、工具注册与执行 |
| 4 | [memory-streaming-config.md](memory-streaming-config.md) | 上下文与记忆、流式输出、配置与安全 |
| 5 | [implementation-and-migration.md](implementation-and-migration.md) | 分阶段实施、代码改造点 |
| 6 | [risks-testing-acceptance.md](risks-testing-acceptance.md) | 风险、测试方案、验收标准、结论 |

单文件执行入口已统一为 [`docs/manual-orchestrator-execution-plan.md`](../manual-orchestrator-execution-plan.md)。
