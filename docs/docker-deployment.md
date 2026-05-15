# Docker 部署指南

## 服务架构

```
                    ┌─────────────────────────────────────┐
                    │           amadeus-network            │
                    │                                      │
  QQ用户  ──────→  napcat:3000/3001     napcat-data/ntqq  │
                    │  (QQ框架 WebUI:6099)  (登录持久化)   │
                    │       │ webhook                      │
                    │       ↓                              │
                    │  qq-adapter:8081                     │
                    │  (OneBot11 事件适配器)                │
                    │       │ HTTP POST /chat              │
                    │       ↓                              │
                    │  amadeus-agent:9501  ←── .env        │
                    │  (AI Agent Server)   ←── skills/     │
                    │       │                  checkpoints/│
                    │       ↓                              │
                    │  milvus:19530  ←── etcd + minio      │
                    │  (向量数据库)      volumes/           │
                    │                                      │
                    │  attu:8080 (Milvus Web 管理)         │
                    └─────────────────────────────────────┘
```

## 服务一览

| 服务 | 镜像 | 宿主机端口 | 说明 |
|------|------|-----------|------|
| `napcat` | `mlikiowa/napcat-docker:latest` | 3000/3001/6099 | QQ 框架 |
| `qq-adapter` | `amadeus`（本地构建） | 8081 | QQ 事件适配器 |
| `amadeus-agent` | `amadeus`（本地构建） | 9501 | AI Agent HTTP Server |
| `etcd` | `quay.io/coreos/etcd:v3.5.25` | — | Milvus 元数据存储 |
| `minio` | `minio/minio:...` | 19000/9001 | Milvus 对象存储（宿主机 9000 被占用，映射到 19000） |
| `milvus` | `milvusdb/milvus:v2.5.10` | 19530/9091 | 向量数据库 |
| `attu` | `zilliz/attu:v2.5` | 8080 | Milvus Web 管理界面 |

---

## 首次启动

### 1. 准备 `.env`

复制 `.env.example` 并填写必填项，**Docker 环境必须使用容器内部域名**：

```dotenv
DEEPSEEK_API_KEY=<your-key>

# Docker 环境专用地址（不能用 localhost）
NAPCAT_API_URL=http://napcat:3000
AGENT_SERVER_URL=http://amadeus-agent:9501
MILVUS_ADDRESS=milvus:19530

QQ_BOT_ID=<bot-qq-number>
```

### 2. 构建并启动

```bash
NAPCAT_UID=$(id -u) NAPCAT_GID=$(id -g) docker compose up -d
```

### 3. 扫码登录 QQ（仅首次）

```bash
# 打开 WebUI（默认 token: napcat）
open http://localhost:6099/webui
```

扫码完成后登录数据写入 `napcat-data/ntqq/`，**后续重启自动登录**，无需再扫。

### 4. 验证服务

```bash
docker compose ps          # 检查所有服务 running
docker logs qq-adapter -f  # 观察事件接收
docker logs amadeus-agent -f  # 观察 AI 响应

# 测试 agent HTTP 接口
curl -X POST http://localhost:9501/chat \
  -H "Content-Type: application/json" \
  -d '{"conversation_id":"test","message":"你好"}'
```

---

## NapCat 网络配置

NapCat 网络通过 `napcat-data/config/onebot11_<QQ号>.json` 配置，首次登录后自动生成：

```json
{
  "network": {
    "httpServers": [
      {
        "name": "httpApi",
        "enable": true,
        "port": 3000,
        "host": "0.0.0.0",
        "token": ""
      }
    ],
    "httpClients": [
      {
        "name": "webhook",
        "enable": true,
        "url": "http://qq-adapter:8081",
        "messagePostFormat": "array",
        "token": ""
      }
    ]
  }
}
```

- `httpServers[0]`：开放 HTTP API（amadeus-agent 调此接口发消息）
- `httpClients[0].url`：NapCat 将事件推送到 qq-adapter（容器内域名）

修改配置文件后需重建容器：`docker compose up -d --force-recreate napcat`

---

## 持久化说明

### Volume 挂载

| 宿主机路径 | 容器路径 | 说明 |
|-----------|---------|------|
| `./napcat-data/ntqq/` | `/app/.config/QQ` | QQ 登录态（必须持久化，否则每次扫码）|
| `./napcat-data/config/` | `/app/napcat/config` | NapCat 网络配置 |
| `./skills/` | `/app/skills` | Skill 定义，修改后重启容器生效 |
| `./tools/` | `/app/tools` | MCP 工具配置 |
| `./checkpoints/` | `/app/checkpoints` | 会话历史 JSONL |
| `./volumes/etcd/` | `/etcd` | Milvus 元数据 |
| `./volumes/minio/` | `/minio_data` | Milvus 向量数据文件 |
| `./volumes/milvus/` | `/var/lib/milvus` | Milvus 状态 |

### NapCat 持久化登录原理

需要两个条件同时满足：
1. `mac_address: "02:42:ac:11:00:02"` 固定（QQ 用 MAC 标识设备，变化则要求重新验证）
2. `napcat-data/ntqq/` volume 挂载（保存登录 token）

二者缺一都会触发重新扫码。

---

## 日常运维

```bash
# 启动所有服务
docker compose up -d

# 停止（数据保留）
docker compose down

# 停止并删除所有 volume（数据清空，慎用）
docker compose down -v

# 重启单个服务（不重新加载 .env）
docker compose restart <service>

# 重建单个服务（重新读取 .env 和镜像）
docker compose up -d --force-recreate <service>

# 重新构建 Go 镜像（代码变更后）
docker compose build
docker compose up -d --force-recreate amadeus-agent qq-adapter

# 查看实时日志
docker compose logs -f napcat qq-adapter amadeus-agent
```

---

## Dockerfile 说明

采用多阶段构建，一个镜像提供两个可执行文件：

```dockerfile
# 阶段1：Go 编译
FROM golang:1.25-alpine AS builder
# 编译 cmd/agent-server → /app/agent-server
# 编译 cmd/qq          → /app/qq

# 阶段2：运行时
FROM node:22-alpine
# 安装 node/npx（MCP 工具依赖）
COPY --from=builder /app/agent-server /app/agent-server
COPY --from=builder /app/qq /app/qq
ENTRYPOINT []   # 覆盖 node 镜像的 docker-entrypoint.sh，防止 Go 二进制被当作 JS 脚本执行
```

`docker-compose.yml` 中两个服务使用同一镜像，仅 `command` 不同：
- `amadeus-agent`: `command: /app/agent-server`
- `qq-adapter`: `command: /app/qq`

---

## 常见问题

### 端口 9000 冲突
MinIO 宿主机端口映射为 `19000:9000`（内部端口仍为 9000，Milvus 连接 `minio:9000` 不受影响）。

### 修改 .env 后不生效
`docker compose restart` 不重新加载环境变量，需用 `--force-recreate`：
```bash
docker compose up -d --force-recreate amadeus-agent qq-adapter
```

### amadeus-agent 启动慢（20-30 秒）
MCP 工具初始化时 `npx -y` 首次会下载 npm 包，属正常现象。

### Milvus 连接超时
`cmd/agent-server/main.go` 中 `NewIndexer` 设有 15 秒超时，超时后自动降级为 noop 模式，agent 继续正常工作，仅 `search_memory` 工具不可用。
