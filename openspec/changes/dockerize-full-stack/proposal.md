## Why

Amadeus 目前各组件（NapCat QQ 框架、QQ 事件适配器 `cmd/qq`、Agent HTTP Server、Milvus 向量数据库）分散独立运行，部署繁琐、环境不一致。通过统一 Docker Compose 编排，实现一键启动完整运行栈，降低部署门槛并保证环境一致性。

## What Changes

- **新增** `Dockerfile`：基于多阶段构建，同时编译 `cmd/agent-server` 和 `cmd/qq` 两个二进制，runtime 层包含 Node.js（MCP 工具通过 `npx` 启动）
- **新增** `docker-compose.yml`：统一编排 napcat、qq-adapter、amadeus-agent、etcd、minio、milvus、attu 七个服务，共享自定义网络 `amadeus-network`
- **移除** `docker-compose.milvus.yml`：内容合并至统一 compose 文件
- **新增** `napcat-data/` 目录结构：持久化 QQ 登录数据（`ntqq/`）与 NapCat 配置（`config/`）
- **修改** `.env.example`：更新 `NAPCAT_API_URL`、`AGENT_SERVER_URL` 与 `MILVUS_ADDRESS` 为容器内服务名

## Capabilities

### New Capabilities

- `container-orchestration`：定义完整的多容器编排方案，包括 Dockerfile、docker-compose.yml、volume 挂载、网络配置及 NapCat 持久化登录机制

### Modified Capabilities

## Impact

- **新增文件**：`Dockerfile`、`docker-compose.yml`、`napcat-data/.gitkeep`
- **删除文件**：`docker-compose.milvus.yml`（内容迁移）
- **修改文件**：`.env.example`（三处服务地址变更）
- **环境变量变更**：`NAPCAT_API_URL` → `http://napcat:3000`，`AGENT_SERVER_URL` → `http://amadeus-agent:9501`，`MILVUS_ADDRESS` → `milvus:19530`
- **依赖变更**：runtime 镜像需含 Node.js（约增加 ~60MB 镜像体积）
- **现有 `go run` 本地开发流程不受影响**：`.env` 中仍可保留 `localhost` 地址用于本地调试
