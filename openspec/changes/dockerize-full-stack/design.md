## Context

Amadeus 是一个 Go 语言 AI Agent，当前部署形态为：
- `cmd/qq`：本地运行，监听 8081 端口，接收 NapCat OneBot 11 webhook 事件，路由至 agent-server
- `cmd/agent-server`：本地 `go run` 启动，监听 HTTP 端口 9501
- NapCat：本地或手动部署，通过 `http://127.0.0.1:3000` 提供 QQ HTTP API
- Milvus 栈：由独立的 `docker-compose.milvus.yml` 管理

完整调用链：NapCat → POST webhook → qq-adapter:8081 → amadeus-agent:9501 → 回复 → NapCat API → QQ 用户

四个组件分散部署，无统一编排，部署流程复杂、环境不可复现。

Agent Server 依赖 `npx` 启动 MCP 工具（`tools/toolsConfig.json`），运行时必须具备 Node.js。Skill、工具配置、会话历史均为可变数据，需在容器外持久化。

## Goals / Non-Goals

**Goals:**
- 提供单一 `docker-compose.yml`，一条命令启动完整运行栈（七个服务）
- NapCat 登录数据跨重启持久化（固定 MAC 地址 + volume）
- `skills/`、`tools/`、`checkpoints/` 挂载为 volume，支持热更新无需重新构建镜像
- 保持本地 `go run` 开发流程不受影响

**Non-Goals:**
- 不提供 Kubernetes / Helm 编排
- 不修改 agent-server 业务逻辑
- 不自动化 NapCat QQ 扫码登录流程

## Decisions

### D1: Runtime 镜像选型 — `node:22-alpine` 而非纯 `alpine`

agent-server 通过 `os/exec` 启动 `npx` 子进程来运行 MCP 工具。纯 `alpine` 镜像无 Node.js，启动时 MCP 工具加载会失败。

**选择**: 多阶段构建，build stage 用 `golang:1.25-alpine`，runtime stage 用 `node:22-alpine`（已内置 Node.js + npm/npx），从 builder 复制编译好的两个二进制（`agent-server`、`qq`）。

**替代方案**: 在 `alpine` 上手动安装 `nodejs npm` — 效果相同但维护性差，不选。

### D2: 一个 Dockerfile，两个服务

`cmd/agent-server` 和 `cmd/qq` 同属一个 Go module，在同一 Dockerfile 的 build stage 中一并编译，产出两个二进制。docker-compose 中 `amadeus-agent` 和 `qq-adapter` 使用同一镜像，通过不同的 `command` 分别启动各自二进制。

**替代方案**: 两个独立 Dockerfile — 增加维护负担，镜像层无法复用，不选。

### D3: Volume 挂载策略 — bind mount 而非 named volume

`skills/`、`tools/`、`checkpoints/` 均使用宿主机目录的 bind mount（`./skills:/app/skills`）。

**理由**: 开发时可直接在宿主机编辑 skill 文件，无需 `docker exec` 进容器，调试体验更好。named volume 适合数据库等不需要直接访问的场景。

### D4: NapCat 持久化登录 — 固定 MAC 地址 + ntqq volume

QQ 客户端通过设备 MAC 地址标识登录环境。若 MAC 地址随容器重建变化，QQ 判定为新设备，强制重新扫码。

**选择**: 在 compose 中为 napcat 服务配置固定 `mac_address: "02:42:ac:11:00:02"`，同时挂载 `./napcat-data/ntqq:/app/.config/QQ` 持久化登录 token。两者缺一不可。

### D5: 合并 compose 文件 — 删除 `docker-compose.milvus.yml`

将 milvus 栈（etcd、minio、milvus、attu）直接合并至主 `docker-compose.yml`，避免多文件管理混乱。

**替代方案**: 使用 `docker compose -f a.yml -f b.yml` 叠加 — 命令繁琐，不选。

### D6: 环境变量注入 — `env_file: .env`

amadeus-agent 和 qq-adapter 容器均通过 `env_file: .env` 读取所有配置，与本地开发共享同一 `.env` 文件，只需修改三处服务地址（`NAPCAT_API_URL`、`AGENT_SERVER_URL`、`MILVUS_ADDRESS`）。

## Risks / Trade-offs

- **[Risk] MCP npx 工具首次启动慢** → 容器内无 npm cache，首次 `npx -y` 会下载包。可通过预构建镜像时 `npm install` 或接受首次延迟来缓解。
- **[Risk] NapCat 镜像平台限制** → `mlikiowa/napcat-docker` 仅支持 `linux/amd64` 和 `linux/arm64`，macOS Apple Silicon 需确认 Rosetta 模拟或 ARM 镜像可用。
- **[Trade-off] 镜像体积增大** → node:22-alpine 比纯 alpine 大约 60MB，可接受。
- **[Risk] `.env` 包含密钥被提交** → `.gitignore` 已有 `.env`，`.env.example` 不含真实密钥，风险可控。

## Migration Plan

1. 备份并删除 `docker-compose.milvus.yml`（或保留作参考后删除）
2. 修改 `.env` 中 `NAPCAT_API_URL=http://napcat:3000`，`AGENT_SERVER_URL=http://amadeus-agent:9501`，`MILVUS_ADDRESS=milvus:19530`
3. 创建 `napcat-data/ntqq/` 和 `napcat-data/config/` 目录
4. 首次启动：`NAPCAT_UID=$(id -u) NAPCAT_GID=$(id -g) docker compose up -d`
5. 在 NapCat WebUI（`http://localhost:6099/webui`）完成扫码登录，并配置 HTTP 上报地址为 `http://qq-adapter:8081`
6. 验证 `docker compose ps` 所有服务健康

**回滚**: 重新使用 `docker-compose.milvus.yml` 启动 Milvus，将 `.env` 中地址改回 `localhost`，本地运行 agent-server。

## Open Questions

- NapCat 的 HTTP API（port 3000）需要在 WebUI 中手动开启，首次启动后需进入配置界面启用——文档中应注明此步骤。
