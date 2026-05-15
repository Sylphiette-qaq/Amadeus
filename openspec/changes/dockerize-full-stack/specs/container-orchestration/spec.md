## ADDED Requirements

### Requirement: Dockerfile 多阶段构建 agent-server 和 qq-adapter
系统 SHALL 提供 `Dockerfile`，使用多阶段构建：build stage 基于 `golang:1.25-alpine` 同时编译 `cmd/agent-server` 和 `cmd/qq` 两个二进制；runtime stage 基于 `node:22-alpine`，包含两个编译产物及 Node.js 运行时（供 MCP 工具 npx 调用）。工作目录为 `/app`。

#### Scenario: 镜像构建成功包含两个二进制
- **WHEN** 执行 `docker build -t amadeus .`
- **THEN** 构建成功，产出镜像包含 `/app/agent-server`、`/app/qq` 两个二进制和 `node`/`npx` 可执行文件

#### Scenario: 容器启动监听正确端口
- **WHEN** 容器以正确环境变量启动
- **THEN** agent-server 监听 `AGENT_SERVER_ADDR` 指定端口（默认 `:9501`），`/chat` 端点可访问

### Requirement: 统一 docker-compose.yml 编排七个服务
系统 SHALL 提供 `docker-compose.yml`，编排以下服务：napcat、qq-adapter、amadeus-agent、etcd、minio、milvus、attu，所有服务加入同一自定义网络 `amadeus-network`。

#### Scenario: 一键启动完整栈
- **WHEN** 执行 `NAPCAT_UID=$(id -u) NAPCAT_GID=$(id -g) docker compose up -d`
- **THEN** 七个服务全部启动，`docker compose ps` 显示所有服务状态为 running 或 healthy

#### Scenario: amadeus-agent 等待依赖就绪
- **WHEN** docker compose 启动
- **THEN** amadeus-agent 在 napcat 和 milvus 服务启动后才启动（通过 `depends_on` 配置）

### Requirement: NapCat 持久化登录
系统 SHALL 为 napcat 服务配置固定 MAC 地址及持久化 volume，确保 QQ 登录状态跨容器重建保留。

#### Scenario: 重建容器后免扫码登录
- **WHEN** 已完成首次 QQ 扫码登录后执行 `docker compose down && docker compose up -d`
- **THEN** NapCat 自动恢复登录状态，无需再次扫码

#### Scenario: QQ 登录数据持久化
- **WHEN** NapCat 容器运行并完成登录
- **THEN** 登录数据写入宿主机 `./napcat-data/ntqq/` 目录，容器删除后数据不丢失

### Requirement: amadeus-agent volume 挂载
amadeus-agent 容器 SHALL 通过 bind mount 挂载以下宿主机目录：`./skills → /app/skills`、`./tools → /app/tools`、`./checkpoints → /app/checkpoints`，支持不重建镜像直接修改内容。

#### Scenario: 修改 skill 立即生效
- **WHEN** 在宿主机编辑 `skills/` 下的文件后重启 amadeus-agent 容器
- **THEN** agent-server 加载到更新后的 skill 内容

#### Scenario: 会话历史跨重启保留
- **WHEN** amadeus-agent 容器重启
- **THEN** `checkpoints/sessions/` 下的历史会话文件仍然存在，可被加载恢复

### Requirement: qq-adapter 接收并转发 NapCat 事件
qq-adapter 服务（`cmd/qq`）SHALL 在容器内监听 8081 端口，接收 NapCat OneBot 11 webhook 事件，并调用 amadeus-agent `/chat` 接口获取回复后通过 NapCat HTTP API 发送消息。

#### Scenario: 私聊消息端到端处理
- **WHEN** QQ 用户发送私聊消息，NapCat 向 `http://qq-adapter:8081` POST webhook 事件
- **THEN** qq-adapter 解析事件，调用 amadeus-agent，并通过 NapCat API 回复用户

#### Scenario: 群消息 @bot 触发回复
- **WHEN** QQ 群消息包含 @bot 的 CQ 码，NapCat 向 qq-adapter 上报事件
- **THEN** qq-adapter 过滤出 @bot 消息，调用 agent，向群组发送回复

### Requirement: 容器间通过服务名通信
同一 `amadeus-network` 内的服务 SHALL 通过 Docker 服务名互相访问，不依赖宿主机 IP。

#### Scenario: qq-adapter 访问 NapCat HTTP API
- **WHEN** `NAPCAT_API_URL=http://napcat:3000` 且两服务在同一网络
- **THEN** qq-adapter 可成功调用 NapCat HTTP API 发送消息

#### Scenario: qq-adapter 访问 amadeus-agent
- **WHEN** `AGENT_SERVER_URL=http://amadeus-agent:9501` 且两服务在同一网络
- **THEN** qq-adapter 可成功调用 amadeus-agent `/chat` 端点

#### Scenario: amadeus-agent 访问 Milvus
- **WHEN** `MILVUS_ADDRESS=milvus:19530` 且两服务在同一网络
- **THEN** amadeus-agent 可成功连接 Milvus 向量数据库

### Requirement: 移除独立 docker-compose.milvus.yml
`docker-compose.milvus.yml` SHALL 被删除，其内容完整迁移至新的统一 `docker-compose.yml`。

#### Scenario: 删除旧 compose 文件
- **WHEN** 实现完成后
- **THEN** 项目根目录不存在 `docker-compose.milvus.yml`，所有 Milvus 相关服务由 `docker-compose.yml` 管理
